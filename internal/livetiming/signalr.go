package livetiming

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	liveTimingHub = "https://livetiming.formula1.com/signalr"
	wsHub         = "wss://livetiming.formula1.com/signalr"

	// Topics we subscribe to.
	topicTimingData          = "TimingData"
	topicTimingAppData       = "TimingAppData"
	topicDriverList          = "DriverList"
	topicCarData             = "CarData.z"
	topicPosition            = "Position.z"
	topicRaceControlMessages = "RaceControlMessages"
	topicWeatherData         = "WeatherData"
	topicSessionStatus       = "SessionStatus"
	topicLapCount            = "LapCount"
	topicSessionInfo         = "SessionInfo"
)

var subscribedTopics = []string{
	topicTimingData,
	topicTimingAppData,
	topicDriverList,
	topicCarData,
	topicPosition,
	topicRaceControlMessages,
	topicWeatherData,
	topicSessionStatus,
	topicLapCount,
	topicSessionInfo,
}

// negotiateResp is the partial response from /negotiate.
type negotiateResp struct {
	ConnectionToken string `json:"ConnectionToken"`
	ConnectionID    string `json:"ConnectionId"`
}

// signalrMsg is a raw SignalR server→client message.
type signalrMsg struct {
	// Hub invocation (method call from server)
	H string          `json:"H"` // hub name
	M string          `json:"M"` // method name
	A []json.RawMessage `json:"A"` // arguments

	// Feed / keepalive
	C string            `json:"C"` // cursor
	M2 []feedItem       `json:"M,omitempty"` // re-used field name — handled below
}

// feedItem is one item in the "M" array of a data frame.
type feedItem struct {
	H string          `json:"H"`
	M string          `json:"M"`
	A []json.RawMessage `json:"A"`
}

// Client connects to the F1 live-timing SignalR endpoint and feeds
// updates into a *State.
type Client struct {
	state  *State
	conn   *websocket.Conn
	done   chan struct{}
	logger *log.Logger
}

// NewClient creates a Client backed by the given state.
func NewClient(state *State, logger *log.Logger) *Client {
	return &Client{
		state:  state,
		done:   make(chan struct{}),
		logger: logger,
	}
}

// Run connects, negotiates, subscribes and reads messages until ctx is
// cancelled or the connection drops.  It reconnects automatically with
// exponential back-off.
func (c *Client) Run(ctx context.Context) {
	bo := time.Second
	for {
		if err := c.runOnce(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Printf("[signalr] disconnected: %v — reconnecting in %s", err, bo)
			select {
			case <-time.After(bo):
			case <-ctx.Done():
				return
			}
			if bo < 60*time.Second {
				bo *= 2
			}
		} else {
			bo = time.Second
		}
	}
}

func (c *Client) runOnce(ctx context.Context) error {
	token, err := negotiate(ctx)
	if err != nil {
		return fmt.Errorf("negotiate: %w", err)
	}

	wsURL := buildWSURL(token)
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL, http.Header{
		"User-Agent": []string{"termf1/1.0"},
	})
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	defer conn.Close()

	// Send the SignalR start/subscribe message.
	if err := c.subscribe(conn); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	c.logger.Printf("[signalr] connected to %s — subscribed to %d topics", wsHub, len(subscribedTopics))

	// Read loop.
	readErr := make(chan error, 1)
	go func() {
		readErr <- c.readLoop(conn)
	}()

	select {
	case err := <-readErr:
		return err
	case <-ctx.Done():
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return nil
	}
}

// negotiate performs the SignalR HTTP negotiate handshake and returns the
// connection token.
func negotiate(ctx context.Context) (string, error) {
	u, _ := url.Parse(liveTimingHub + "/negotiate")
	q := u.Query()
	q.Set("connectionData", `[{"name":"streaming"}]`)
	q.Set("clientProtocol", "1.5")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	req.Header.Set("User-Agent", "termf1/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var nr negotiateResp
	if err := json.NewDecoder(resp.Body).Decode(&nr); err != nil {
		return "", err
	}
	if nr.ConnectionToken == "" {
		return "", fmt.Errorf("empty connection token")
	}
	return nr.ConnectionToken, nil
}

// buildWSURL constructs the WebSocket URL from the connection token.
func buildWSURL(token string) string {
	u, _ := url.Parse(wsHub + "/connect")
	q := u.Query()
	q.Set("connectionData", `[{"name":"streaming"}]`)
	q.Set("clientProtocol", "1.5")
	q.Set("transport", "webSockets")
	q.Set("connectionToken", token)
	u.RawQuery = q.Encode()
	return u.String()
}

// subscribe sends the SignalR hub subscription message.
func (c *Client) subscribe(conn *websocket.Conn) error {
	// Build the hub subscribe invocation.
	type subscribeMsg struct {
		H string        `json:"H"`
		M string        `json:"M"`
		A []interface{} `json:"A"`
		I int           `json:"I"`
	}
	msg := subscribeMsg{
		H: "streaming",
		M: "Subscribe",
		A: []interface{}{subscribedTopics},
		I: 1,
	}
	b, _ := json.Marshal(msg)
	return conn.WriteMessage(websocket.TextMessage, b)
}

// readLoop processes incoming WebSocket frames.
func (c *Client) readLoop(conn *websocket.Conn) error {
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		// SignalR keepalive is `{}`
		if bytes.Equal(bytes.TrimSpace(raw), []byte("{}")) {
			continue
		}
		c.dispatch(raw)
	}
}

// dispatch routes a raw SignalR JSON frame to the appropriate state update.
func (c *Client) dispatch(raw []byte) {
	// The frame can have an "M" field that is either an array of feed items
	// OR be a method invocation.  We unmarshal into a generic map first.
	var frame map[string]json.RawMessage
	if err := json.Unmarshal(raw, &frame); err != nil {
		return
	}

	mRaw, hasMFeed := frame["M"]
	if !hasMFeed {
		return
	}

	// Try as array of feed items first.
	var items []feedItem
	if err := json.Unmarshal(mRaw, &items); err != nil || len(items) == 0 {
		return
	}

	for _, item := range items {
		if !strings.EqualFold(item.H, "streaming") || item.M != "feed" {
			continue
		}
		if len(item.A) < 2 {
			continue
		}
		// A[0] = topic name (string), A[1] = payload (raw JSON or base64+zlib)
		var topic string
		if err := json.Unmarshal(item.A[0], &topic); err != nil {
			continue
		}
		payload := item.A[1]

		// Compressed topics end in ".z" — payload is a base64+zlib string.
		if strings.HasSuffix(topic, ".z") {
			decompressed, err := decompressPayload(payload)
			if err != nil {
				c.logger.Printf("livetiming: decompress %s: %v", topic, err)
				continue
			}
			payload = decompressed
			// Strip the ".z" for routing.
			topic = strings.TrimSuffix(topic, ".z")
		}

		c.route(topic, payload)
	}
}

// route dispatches a decoded payload to the appropriate state method.
func (c *Client) route(topic string, payload json.RawMessage) {
	c.logger.Printf("[signalr] %-28s  %5d B", topic, len(payload))
	switch topic {
	case topicTimingData:
		c.state.applyTimingData(payload)
	case topicTimingAppData:
		c.state.applyTimingAppData(payload)
	case topicDriverList:
		c.state.applyDriverList(payload)
	case "CarData":
		c.state.applyCarData(payload)
	case "Position":
		c.state.applyPosition(payload)
	case topicRaceControlMessages:
		c.state.applyRaceControlMessages(payload)
	case topicWeatherData:
		c.state.applyWeatherData(payload)
	case topicSessionStatus:
		c.state.applySessionStatus(payload)
	case topicLapCount:
		c.state.applyLapCount(payload)
	case topicSessionInfo:
		c.state.applySessionInfo(payload)
	default:
		c.logger.Printf("[signalr] unhandled topic: %s", topic)
	}
}

// decompressPayload decodes the base64+zlib payload used by CarData.z and
// Position.z topics.
func decompressPayload(raw json.RawMessage) (json.RawMessage, error) {
	// raw is a JSON string, e.g. `"eJy..."`
	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil, err
	}
	// The F1 feed pads with "=" that may be missing; add padding.
	padded := encoded
	switch len(padded) % 4 {
	case 2:
		padded += "=="
	case 3:
		padded += "="
	}
	b, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		// Try RawStdEncoding as fallback.
		b, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("base64: %w", err)
		}
	}
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("zlib: %w", err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("zlib read: %w", err)
	}
	return json.RawMessage(out), nil
}

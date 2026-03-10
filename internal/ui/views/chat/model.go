package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devkeshwani/termf1/internal/api/groq"
	"github.com/devkeshwani/termf1/internal/ui/styles"
	"github.com/devkeshwani/termf1/internal/ui/views/common"
)

// ── Types ─────────────────────────────────────────────────────────────────────

type chatMessage struct {
	role    string // "user" | "assistant" | "error"
	content string
}

// ── Messages ──────────────────────────────────────────────────────────────────

type chatResponseMsg struct{ content string }
type chatErrMsg struct{ err error }

// ── Model ─────────────────────────────────────────────────────────────────────

type Chat struct {
	client              *groq.Client
	width               int
	height              int
	history             []chatMessage  // conversation display history
	messages            []groq.Message // API conversation history
	viewport            viewport.Model
	input               textarea.Model
	spin                spinner.Model
	waiting             bool // waiting for Groq response
	needsScrollToBottom bool // scroll to bottom on next View()
}

func NewChat(client *groq.Client) *Chat {
	ta := textarea.New()
	ta.Placeholder = "Ask anything about F1… (Enter to send, Esc to clear)"
	ta.CharLimit = 1000
	ta.MaxHeight = 4
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.Focus()

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorF1Red)

	return &Chat{
		client:   client,
		input:    ta,
		viewport: vp,
		spin:     s,
		messages: []groq.Message{
			{
				Role:    "system",
				Content: "You are an expert F1 analyst and assistant embedded in a terminal dashboard called termf1. Answer questions about Formula 1 concisely and accurately. You can discuss race strategy, regulations, driver stats, team performance, history, and current news. Respond in plain text without markdown formatting.",
			},
		},
	}
}

func (c *Chat) SetSize(w, h int) {
	c.width = w
	c.height = h

	// View() renders: title(1)+sep(1)+blank(1)+vpBorder(2)+waitStr(1)+inputBox(5)+hint(1) = 12 lines overhead.
	vpH := h - 12
	if vpH < 3 {
		vpH = 3
	}

	c.viewport.Width = w - 4
	c.viewport.Height = vpH
	c.input.SetWidth(w - 4)
}

func (c *Chat) InputFocused() bool {
	return c.input.Focused()
}

func (c *Chat) Init() tea.Cmd {
	return textarea.Blink
}

func (c *Chat) UpdateChat(msg tea.Msg) (*Chat, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			c.input.Blur()
			return c, nil

		case tea.KeyEnter:
			// Shift+Enter → newline; plain Enter → send
			if msg.Alt {
				break
			}
			text := strings.TrimSpace(c.input.Value())
			if text == "" || c.waiting {
				return c, nil
			}
			c.input.Reset()
			c.input.Focus()
			c.addUserMessage(text)
			c.waiting = true
			cmds = append(cmds, c.spin.Tick, c.sendToGroq(text))
			return c, tea.Batch(cmds...)

		case tea.KeyCtrlL:
			// Clear chat
			c.history = nil
			c.messages = c.messages[:1] // keep system prompt
			c.viewport.SetContent("")
			return c, nil
		}

	case chatResponseMsg:
		c.waiting = false
		c.addAssistantMessage(msg.content)

	case chatErrMsg:
		c.waiting = false
		c.addErrorMessage(msg.err.Error())

	case spinner.TickMsg:
		if c.waiting {
			var cmd tea.Cmd
			c.spin, cmd = c.spin.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Route key events to avoid conflicts between viewport and textarea.
	// When the input is focused, all keys go to the textarea (so the user
	// can type freely). When blurred, scroll keys go to the viewport only.
	if kMsg, isKey := msg.(tea.KeyMsg); isKey {
		if c.input.Focused() {
			// Textarea has focus — forward to textarea only.
			var taCmd tea.Cmd
			c.input, taCmd = c.input.Update(kMsg)
			cmds = append(cmds, taCmd)
		} else {
			// Input blurred — forward to viewport only for scrolling.
			var vpCmd tea.Cmd
			c.viewport, vpCmd = c.viewport.Update(kMsg)
			cmds = append(cmds, vpCmd)
		}
	} else {
		// Non-key msgs (window resize, ticks, etc.) → update both.
		var vpCmd tea.Cmd
		c.viewport, vpCmd = c.viewport.Update(msg)
		cmds = append(cmds, vpCmd)

		var taCmd tea.Cmd
		c.input, taCmd = c.input.Update(msg)
		cmds = append(cmds, taCmd)
	}

	return c, tea.Batch(cmds...)
}

func (c *Chat) View() string {
	title := styles.Title.Render(" 🤖 Ask AI  ") +
		styles.DimStyle.Render("powered by Groq  •  ctrl+l to clear history")
	sep := styles.Divider.Render(strings.Repeat("─", c.width))

	// Render conversation — only snap to bottom when a new message arrives,
	// so the user can freely scroll up through history.
	content := c.renderHistory()
	c.viewport.SetContent(content)
	if c.needsScrollToBottom {
		c.viewport.GotoBottom()
		c.needsScrollToBottom = false
	}

	vpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Padding(0, 1).
		Width(c.width - 2).
		Render(c.viewport.View())

	// Input area
	waitStr := ""
	if c.waiting {
		waitStr = "  " + c.spin.View() + " Thinking…"
	}
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorF1Red).
		Padding(0, 1).
		Width(c.width - 2).
		Render(c.input.View())

	hint := styles.DimStyle.Render("  Enter: send  │  Esc: blur input  │  Ctrl+L: clear  │  Ctrl+C / q: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		title, sep, "",
		vpBox,
		waitStr,
		inputBox,
		hint,
	)
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (c *Chat) addUserMessage(text string) {
	c.history = append(c.history, chatMessage{role: "user", content: text})
	c.messages = append(c.messages, groq.Message{Role: "user", Content: text})
	c.needsScrollToBottom = true
}

func (c *Chat) addAssistantMessage(text string) {
	c.history = append(c.history, chatMessage{role: "assistant", content: text})
	c.messages = append(c.messages, groq.Message{Role: "assistant", Content: text})
	c.needsScrollToBottom = true
}

func (c *Chat) addErrorMessage(text string) {
	c.history = append(c.history, chatMessage{role: "error", content: text})
}

func (c *Chat) renderHistory() string {
	if len(c.history) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.ColorSubtle).
			Italic(true).
			Render("  Ask anything about F1 — live races, regulations, history, strategy…\n\n  Examples:\n  • What are the current championship standings?\n  • Explain DRS and when drivers can use it\n  • What happened in the last race?\n  • Compare Verstappen and Hamilton's careers")
	}

	var sb strings.Builder
	for _, msg := range c.history {
		switch msg.role {
		case "user":
			label := lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.ColorBlue).
				Render("You")
			text := lipgloss.NewStyle().
				Foreground(styles.ColorText).
				Render(msg.content)
			sb.WriteString(fmt.Sprintf("%s  %s\n\n", label, text))

		case "assistant":
			label := lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.ColorF1Red).
				Render("AI ")
			// Wrap long lines
			wrapped := wordWrap(msg.content, c.viewport.Width-6)
			text := lipgloss.NewStyle().
				Foreground(styles.ColorText).
				Render(wrapped)
			sb.WriteString(fmt.Sprintf("%s  %s\n\n", label, text))

		case "error":
			label := lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.ColorOrange).
				Render("ERR")
			text := styles.ErrorStyle.Render(msg.content)
			sb.WriteString(fmt.Sprintf("%s  %s\n\n", label, text))
		}
	}
	return sb.String()
}

func (c *Chat) sendToGroq(userMsg string) tea.Cmd {
	// Keep system prompt + last 10 messages (5 exchanges) to avoid hitting
	// the Groq request-size limit (413 Request Entity Too Large).
	const maxHistory = 10
	src := c.messages
	var msgs []groq.Message
	if len(src) > 1+maxHistory {
		msgs = append([]groq.Message{src[0]}, src[len(src)-maxHistory:]...)
	} else {
		msgs = make([]groq.Message, len(src))
		copy(msgs, src)
	}

	return func() tea.Msg {
		resp, err := c.client.Chat(common.ContextBG(), msgs)
		if err != nil {
			return chatErrMsg{err}
		}
		return chatResponseMsg{resp}
	}
}

// wordWrap wraps text at word boundaries within maxWidth columns.
func wordWrap(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}
	var sb strings.Builder
	for _, line := range strings.Split(text, "\n") {
		words := strings.Fields(line)
		col := 0
		for i, word := range words {
			wl := len(word)
			if i > 0 {
				if col+1+wl > maxWidth {
					sb.WriteByte('\n')
					col = 0
				} else {
					sb.WriteByte(' ')
					col++
				}
			}
			sb.WriteString(word)
			col += wl
		}
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

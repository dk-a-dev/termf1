package dashboard

// StaticDrivers holds a hardcoded mapping for the 2024/2025 F1 grid.
// This ensures driver TLAs and names are always populated even if the
// DriverList topic is missing or delayed in a live session.
var StaticDrivers = map[string]struct{ Tla, Name string }{
	"1":  {"VER", "Max Verstappen"},
	"11": {"PER", "Sergio Perez"},
	"44": {"HAM", "Lewis Hamilton"},
	"63": {"RUS", "George Russell"},
	"4":  {"NOR", "Lando Norris"},
	"81": {"PIA", "Oscar Piastri"},
	"16": {"LEC", "Charles Leclerc"},
	"55": {"SAI", "Carlos Sainz"},
	"14": {"ALO", "Fernando Alonso"},
	"18": {"STR", "Lance Stroll"},
	"10": {"GAS", "Pierre Gasly"},
	"31": {"OCO", "Esteban Ocon"},
	"23": {"ALB", "Alexander Albon"},
	"2":  {"SAR", "Logan Sargeant"},
	"43": {"COL", "Franco Colapinto"},
	"22": {"TSU", "Yuki Tsunoda"},
	"3":  {"RIC", "Daniel Ricciardo"},
	"30": {"LAW", "Liam Lawson"},
	"77": {"BOT", "Valtteri Bottas"},
	"24": {"ZHO", "Zhou Guanyu"},
	"27": {"HUL", "Nico Hulkenberg"},
	"20": {"MAG", "Kevin Magnussen"},
	"38": {"BEA", "Oliver Bearman"},
	"87": {"BEA", "Oliver Bearman"}, // used 87 in some sessions
	"12": {"DOO", "Jack Doohan"},    // used 12 in some sessions
	"7":  {"ANT", "Kimi Antonelli"}, // used 7 in some sessions
}

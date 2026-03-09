package views

import "context"

// contextBG returns a background context shared among view fetch commands.
func contextBG() context.Context {
	return context.Background()
}

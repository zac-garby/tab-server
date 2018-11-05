package main

import "github.com/Zac-Garby/tab-server/src"

func main() {
	// Make a new Server instance, addr = "" and
	// port = 8000. Everything else is zero values
	// of the respective types.
	s := &src.Server{
		Address: "",
		Port:    8000,
	}

	// Start listening on port 8000.
	s.Listen()
}

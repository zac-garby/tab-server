package main

import (
	"fmt"
	"os"

	"github.com/Zac-Garby/tab-server/src"
	"github.com/go-redis/redis"
)

func main() {
	// Open a connection to the Redis server so
	// the data can be fetched.
	db := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Load the settings from the database, potentially
	// handling an error. An error will cause the
	// program to exit early without starting a web
	// server.
	settings, err := src.LoadSettings(db)
	if err != nil {
		fmt.Println("Could not load settings. Reason:", err)
		os.Exit(1)
	}

	// Make a new Server instance, addr = "" and
	// port = 8000. Everything else is zero values
	// of the respective types.
	s := &src.Server{
		Address:  "",
		Port:     8000,
		Settings: settings,
		Database: db,
	}

	// Start listening on port 8000.
	s.Listen()
}

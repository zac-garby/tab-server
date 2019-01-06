package src

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

// A Server is used to handle HTTP requests.
type Server struct {
	// Address is the address to listen on, usually an empty string
	// which indicates that any address will be listened to.
	Address string

	// Port is the port to listen on.
	Port int

	// HTTPS says whether or not HTTPS should be used to communicate.
	HTTPS bool

	// Certificate is the filename of the HTTPS certificate.
	Certificate string

	// Key is the filename of the HTTPS key.
	Key string

	// Settings stores the settings of this server.
	Settings *Settings

	// Database allows access to the database from server methods.
	Database *redis.Client
}

// Listen starts the HTTP server running on the given address and port.
func (s *Server) Listen() {
	// Create a new router, which will be used to listen to HTTP requests and
	// decide what to do to respond back.
	r := mux.NewRouter()

	r.HandleFunc("/", s.handleIndex)
	r.HandleFunc("/settings", s.handleSettings)

	r.HandleFunc("/api/tabs", s.handleTabsAPI)

	// Handle static files
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./www/")),
		),
	)

	// Starts the HTTP server listening using the router defined previously.
	fmt.Printf("Server is running at %s:%d...\n", s.Address, s.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", s.Address, s.Port), r)
}

// handleIndex is called to respond to a HTTP request to /.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	//  Disable caching for this route.
	w.Header().Set("Cache-Control", "max-age=0")

	// Shorthand for:
	//  - opening www/html/index.html
	//  - reading its contents
	//  - serving that text, along with relavent metadata
	http.ServeFile(w, r, "www/html/index.html")
}

// handleSettings is called to respond to a HTTP request to /settings.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	//  Disable caching for this route.
	w.Header().Set("Cache-Control", "max-age=0")

	// Shorthand for:
	//  - opening www/html/settings.html
	//  - reading its contents
	//  - serving that text, along with relavent metadata
	http.ServeFile(w, r, "www/html/settings.html")
}

// handleTabsAPI is called to respond to a HTTP request to /api/tabs.
func (s *Server) handleTabsAPI(w http.ResponseWriter, r *http.Request) {
	// Disable caching for this request - caching will be managed
	// manually by this program.
	w.Header().Set("Cache-Control", "max-age=0")

	// Set the content type of the response to JSON so browsers
	// don't attempt to display it as HTML.
	w.Header().Set("Content-Type", "text/json")

	// Get a list of tabs.
	tabs, err := s.getTabs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the tabs into JSON so they can be transmitted over HTTP.
	jsonData, err := json.Marshal(tabs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

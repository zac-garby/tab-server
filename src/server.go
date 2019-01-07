package src

import (
	"crypto/sha512"
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
	r.HandleFunc("/api/reset-cache", s.handleResetCacheAPI)
	r.HandleFunc("/api/change-password", s.handleChangePassword)

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
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	tabs, err := s.getTabs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the tabs into JSON so they can be transmitted over HTTP.
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	jsonData, err := json.Marshal(tabs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// handleResetCacheAPI is called to respond to a HTTP request to
// /api/reset-cache.
func (s *Server) handleResetCacheAPI(w http.ResponseWriter, r *http.Request) {
	// Remove all keys in the database with the prefix tab:*.
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	if err := s.Database.Eval(
		`return redis.call('del', unpack(redis.call('keys', ARGV[1])))`,
		nil, "tab:*",
	).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Empty the tab ID list and the filename-ID map.
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	if err := s.Database.Del("tabs", "filenames").Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reset the tab counter to 0, so the next tab will be
	// assigned the ID of (0 + 1) = 1.
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	if err := s.Database.Set("tab-counter", 0, 0).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleChangePassword is called to respond to a HTTP request to
// /api/change-password. It will only accept POST requests.
func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	// If the request method isn't POST, send an error back to the client
	// telling them that only POST will work, with a Method Not Allowed status.
	if r.Method != "POST" {
		http.Error(w, "only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	// Get the 'old' and 'new' form values from the request. Also, fetch the
	// actual password's hash from the database (stored under the key
	// password-hash).
	var (
		oldPassword     = r.PostFormValue("old")
		newPassword     = r.PostFormValue("new")
		actualHash, err = s.Database.Get("password-hash").Result()
	)

	// If there is an error while fetching the password's hash from the
	// database, send the error to the client with an Internal Server
	// Error status.
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Hash the password which the client believes to be the existing
	// password using a SHA512 hash. This is done using the Sum512
	// function to compute the SHA512 digest of the specified password.
	//
	// The %x format option converts the byte array to a string representing
	// the hash in hexadecimal format.
	requestHash := fmt.Sprintf("%x", sha512.Sum512([]byte(oldPassword)))

	// If the requested hash is not equal to the actual hash of the
	// password, send an error telling the client exactly that, with
	// a Bad Request status code.
	if requestHash != actualHash {
		http.Error(w, "wrong password", http.StatusBadRequest)
		return
	}

	// At this point, we know that the user has entered the correct
	// password, implying that they are in fact the admin. So now
	// the new password they want will be stored in the database,
	// reporting any errors to the user.
	//
	// To do this, the new password must first be hashed. Then,
	// the SET redis command is used to set the new password.
	newHash := fmt.Sprintf("%x", sha512.Sum512([]byte(newPassword)))
	if err := s.Database.Set("password-hash", newHash, 0).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

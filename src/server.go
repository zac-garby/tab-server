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
	r.HandleFunc("/api/delete-tab", s.handleDeleteTab)
	r.HandleFunc("/api/settings", s.handleSettingsAPI)
	r.HandleFunc("/api/change-settings", s.handleChangeSettingsAPI)

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
	w.Header().Set("Content-Type", "application/json")

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
	// Validate the user's entered password, in the form field 'password', and
	// if it is wrong send them a message and exit the function.
	if status, err := s.validatePassword(r, "old"); err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	// At this point, we know that the user has entered the correct
	// password, implying that they are in fact the admin. So now
	// the new password they want will be stored in the database,
	// reporting any errors to the user.
	//
	// To do this, the new password must first be hashed. Then,
	// the SET redis command is used to set the new password.
	newHash := fmt.Sprintf("%x", sha512.Sum512([]byte(r.PostFormValue("new"))))
	if err := s.Database.Set("password-hash", newHash, 0).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleDeleteTab is called to respond to a HTTP request to
// /api/delete-tab. It will only accept POST requests because the
// password is sent in the POST form data.
func (s *Server) handleDeleteTab(w http.ResponseWriter, r *http.Request) {
	// Validate the user's entered password, in the form field 'password', and
	// if it is wrong send them a message and exit the function.
	if status, err := s.validatePassword(r, "password"); err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	// Now we know that the user has entered the correct password, the
	// tab can be deleted. This is done through the 'deleteTab' function
	// inside the api.go file.
	if err := s.deleteTab(r.PostFormValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleSettingsAPI is called to a HTTP request to /api/settings. It will
// respond with the current settings encoded in JSON. It will be able to
// accept any request method type because the password is not transmitted.
func (s *Server) handleSettingsAPI(w http.ResponseWriter, r *http.Request) {
	// Disable caching for this request - caching will be managed
	// manually by this program.
	w.Header().Set("Cache-Control", "max-age=0")

	// Set the content type of the response to JSON so browsers
	// don't attempt to display it as HTML.
	w.Header().Set("Content-Type", "application/json")

	passwordHash := s.Settings.PasswordHash
	s.Settings.PasswordHash = ""

	// Convert the settings into JSON so they can be transmitted over HTTP.
	// If there is an error, it will be returned as a HTTP error with the
	// status code 500, or Internal Server Error.
	jsonData, err := json.Marshal(s.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Settings.PasswordHash = passwordHash

	w.Write(jsonData)
}

// handleChangeSettingsAPI is called to respond to a HTTP request to
// /api/change-settings. It will update the settings in the running program's
// memory and also in the database. It requires an admin password in the
// 'password' form value, so only POST requests are accepted.
func (s *Server) handleChangeSettingsAPI(w http.ResponseWriter, r *http.Request) {
	// Validate the user's entered password, in the form field 'password', and
	// if it is wrong send them a message and exit the function.
	if status, err := s.validatePassword(r, "password"); err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	// Now we know that the user has entered the correct password, the
	// settings can be updated using the 'changeSettings' server method.
	if err := s.changeSettings(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

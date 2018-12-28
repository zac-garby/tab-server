package src

import (
	"github.com/go-redis/redis"
)

// Settings is used to store the user settings of the application
// and holds all the relavent fields from the database.
// This struct can be converted into a JSON form, which is
// necessary for sending to clients via HTTP.
type Settings struct {
	// PasswordHash stores the SHA-256 hash of the admin
	// password.
	PasswordHash string `json:"password-hash"`

	// TabDirectory is the absolute path to the directory
	// in which to look for tabs.
	TabDirectory string `json:"tab-directory"`

	// FilenamePattern is the pattern to parse tabs with.
	FilenamePattern string `json:"filename-pattern"`

	// NonCapitalWords is the set of words which should
	// not be capitalised when capitalising metadata.
	NonCapitalWords []string `json:"non-capital-words"`

	// CharactersToRemove is the set of characters to
	// get rid of from metadata.
	CharactersToRemove string `json:"characters-to-remove"`
}

// LoadSettings creates a new instance of Settings by fetching
// the settings from the given database connection. If there
// is an error while fetching the data, an error will be
// returned.
func LoadSettings(db *redis.Client) (*Settings, error) {
	// Get the password's SHA256 hash from the database, and
	// check for any errors. If there is an error, this is
	// returned from the function.
	pw, err := db.Get("password-hash").Result()
	if err != nil {
		return nil, err
	}

	// The same thing is done for each other field which must
	// be fetched.
	dir, err := db.Get("tab-directory").Result()
	if err != nil {
		return nil, err
	}

	pattern, err := db.Get("filename-pattern").Result()
	if err != nil {
		return nil, err
	}

	// This field is slightly different in that it is a set of
	// strings instead of just a single string, which means
	// that a different function must be used to fetch it.
	nonCap, err := db.SMembers("non-capital-words").Result()
	if err != nil {
		return nil, err
	}

	charsToRemove, err := db.Get("characters-to-remove").Result()
	if err != nil {
		return nil, err
	}

	// Create a new Settings instance populated with the fetched
	// fields and return it.
	return &Settings{
		PasswordHash:       pw,
		TabDirectory:       dir,
		FilenamePattern:    pattern,
		NonCapitalWords:    nonCap,
		CharactersToRemove: charsToRemove,
	}, nil
}

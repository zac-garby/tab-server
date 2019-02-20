package src

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// tabFilenames returns a list of the filenames in the tab directory.
func (s *Server) tabFilenames() ([]string, error) {
	// Get a list of information about each file in the tab directory. If an
	// error occurs - i.e. if the directory doesn't exist - that error is
	// returned and the function exits early.
	files, err := ioutil.ReadDir(s.Settings.TabDirectory)
	if err != nil {
		return nil, err
	}

	// Make a new list of strings, allocating enough memory to store a string
	// for each file in 'files'.
	filenames := make([]string, 0, len(files))

	// Iterate through the file informations which are stored in 'files',
	// keeping track the current iteration index and the file information
	// on each iteration.
	for _, file := range files {
		// If the filename begins with a '.' character, ignore it. A '.'
		// before a filename implies that it is hidden (in macOS, anyway),
		// and thus shouldn't be processed by the program.
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// Append the filename to the filenames list.
		filenames = append(filenames, file.Name())
	}

	// Return the list of filenames, and a nil error since the function was
	// successful.
	return filenames, nil
}

// filterFilenames returns two lists, one containing all filenames which need to
// be processed further and one containing all filenames which have already been
// cached and thus don't need any more processing (except from fetching the data
// from the database).
func (s *Server) filterFilenames(filenames []string) (toProcess, cached []string, err error) {
	// Fetch the set containing all cached tab IDs from the database, which is
	// stored inside the key 'tabs'. If there is an error, return it along with
	// nil values for the two lists.
	tabIDs, err := s.Database.SMembers("tabs").Result()
	if err != nil {
		return nil, nil, err
	}

	// Initialise the cached filename list. Since we already have a list containing
	// all of the tab IDs, we know that there will be len(tabIDs) tabs to put in
	// the list, allowing the list to be initialised to the correct capacity
	// beforehand, which will very slightly increase the performance because
	// memory allocation takes time.
	cached = make([]string, len(tabIDs))

	// Iterate over each ID in the list of tab IDs, also keeping track of the current
	// index of the iteration.
	for index, id := range tabIDs {
		// Find the filename corresponding to the current iteration's ID. If an
		// error occurs, return from the function, returning the error. The tab's
		// key is calculated as the concatenation of "tab:" and the ID. HGET is a
		// Redis command which gets a particular value from a hashmap, in this
		// case the value with the key "filename".
		filename, err := s.Database.HGet("tab:"+id, "filename").Result()
		if err != nil {
			return nil, nil, err
		}

		cached[index] = filename
	}

	// Create the list in which the filenames of files which need to be processed
	// will be put.
	toProcess = make([]string, 0)

	// Iterate through the given filenames to check which of them have been cached
	// and which haven't yet. Also, label this loop as 'outerLoop' so it can be
	// referenced by 'continue' statements.
outerLoop:
	for _, filename := range filenames {
		// Go through the list of filenames which have been cached, checking for
		// each one if it is equal to the current iteration's filename. If it is,
		// then this file has already been cached the the next iteration of the
		// outer loop can be skipped to. This loop implements a linear search.
		for _, existing := range cached {
			if existing == filename {
				continue outerLoop
			}
		}

		// If the loop finished without and filename matching, this file needs
		// to be processed further, and as such it is appended to the toProcess
		// list.
		toProcess = append(toProcess, filename)
	}

	// A return with no "arguments" here means that the two lists are returned
	// implicitly, because they are named return values.
	return
}

// getTabs returns a list of all of the tabs in the system, getting cached ones
// from the database and parsing new ones if necessary from the filesystem.
func (s *Server) getTabs() (tabs []*Tab, err error) {
	// Initialise the tabs list, which was declared in the return parameters.
	// It is defined as initially having a length of 0, because at this point
	// we don't know how long it should be.
	tabs = make([]*Tab, 0)

	// Get the list of filenames in the tab directory, which will be used to
	// find the tabs which haven't been cached. Any error will be propogated
	// to the error of the getTabs function, which will make an early return
	filenames, err := s.tabFilenames()
	if err != nil {
		return nil, err
	}

	// Split the list of filenames into filenames which should be parsed from
	// scratch and ones which have already been cached, again propogating any
	// errors to the error return value of this function.
	toProcess, cached, err := s.filterFilenames(filenames)
	if err != nil {
		return nil, err
	}

	// Iterate through the list of cached filenames, fetching the relavent data
	// from the database and appending a new tab to 'tabs' for each cached item
	for _, filename := range cached {
		// Using the filenames hashmap in the database, find the ID corresponding
		// to the current cached filename, which can be used to find the tab data
		id, err := s.Database.HGet("filenames", filename).Result()
		if err != nil {
			return nil, err
		}

		// Fetch the tab with the ID from the database. If there is an error, return
		// that error from the getTabs function, if the tab doesn't exist, something
		// weird has happened so give the server a message saying that it should not
		// happen and should be debugged. Otherwise, append the tab to the tab list.
		tab, ok, err := s.fetchTab(id)
		if err != nil {
			return nil, err
		} else if !ok {
			fmt.Println("this point shouldn't be reached (Server.getTabs)")
		} else {
			tabs = append(tabs, tab)
		}
	}

	// Convert the filename pattern from the settings into a list of tokens which
	// will be used to parse and extract the metadata from each of the filenames.
	tokens := tokenizePattern(s.Settings.FilenamePattern)

	// Iterate through the list of filenames which need to be parsed from the disk,
	// for each one reading the file and extracting the metadata from the filename.
	for _, filename := range toProcess {
		// Extract the title, artist name, and list of tags from the filename, using
		// the tokens lexed from the filename pattern earlier. If there is no parse,
		// log a message to the server and skip to the next filename in the list.
		title, artist, tags, ok := parseFilename(
			strings.TrimSuffix(filename, filepath.Ext(filename)),
			tokens,
		)
		if !ok {
			fmt.Printf("The filename %s could not be parsed.\n", filename)
			continue
		}

		// Read the content of the file. If the file does not exist, and error will
		// be returned and the function will exit early. The content is returned from
		// this function as a list of bytes representing the characters instead of a
		// string so it is converted to a string when the tab is created. Another thing
		// to note is that the filepath of the file is calculated by joining the tab
		// directory and the current filename, where the join function inserts a /
		// or a \ between the two arguments based on the system on which it's running.
		content, err := ioutil.ReadFile(filepath.Join(s.Settings.TabDirectory, filename))
		if err != nil {
			return nil, err
		}

		// Construct the tab instance, excluding the ID as this will be added when
		// cacheNewTab is called.
		tab := &Tab{
			Title:    title,
			Artist:   artist,
			Tags:     tags,
			Filename: filename,
			Content:  string(content),
		}

		// Write the tab to the database and if there is an error, skip to the
		// next filename to process, not adding this tab to the list of tabs.
		// Also, write the error to the console.
		if err := s.cacheNewTab(tab); err != nil {
			fmt.Printf(
				"The tab with filename %s could not be added to the database: %s\n",
				filename,
				err,
			)

			continue
		}

		// Append the tab to the list of tabs.
		tabs = append(tabs, tab)
	}

	for _, tab := range tabs {
		tab.applyTransformations(s.Settings.CharactersToRemove, s.Settings.NonCapitalWords)
	}

	return
}

func (s *Server) deleteTab(id string) error {
	// Fetch the filename of the tab with the specified ID, so the filename-ID
	// mapping can later be removed from the filename-ID hashmap.
	filename, err := s.Database.HGet(fmt.Sprintf("tab:%s", id), "filename").Result()
	if err != nil {
		return err
	}

	// Delete the tab's data hashmap and its tags set, returning any errors which
	// are encountered.
	if err := s.Database.Del(
		fmt.Sprintf("tab:%s", id),
		fmt.Sprintf("tab:%s:tags", id)).Err(); err != nil {
		return err
	}

	// Remove the tab's ID from the ID set, meaning that it will no longer be
	// included when looking up the list of all tabs.
	if err := s.Database.SRem("tabs", id).Err(); err != nil {
		return err
	}

	// Delete the filename from the hashmap in the database which maps the filenames
	// to their tab IDs.
	if err := s.Database.HDel("filenames", filename).Err(); err != nil {
		return err
	}

	// Remove the file from the filesystem, calculating it's filepath relative to
	// the working directory as <tab-directory>/<filename>.
	if err := os.Remove(filepath.Join(s.Settings.TabDirectory, filename)); err != nil {
		return err
	}

	// At this point, the tab has been completely removed from the database, as if
	// it were never there. So, the function has completed successfully and can
	// return a nil error meaning that there was no problem.

	return nil
}

// validatePassword gets the password from the given form field (specified in the
// passwordField parameter) and checks it against the password hash from the database.
// If it is incorrect, an error and error status will be returned.
func (s *Server) validatePassword(r *http.Request, passwordField string) (int, error) {
	// If the request method isn't POST, send an error back to the client
	// telling them that only POST will work, with a Method Nod Allowed status.
	if r.Method != "POST" {
		return http.StatusMethodNotAllowed, errors.New("only POST is supported")
	}

	// Get the entered password from the request form and fetch the hash of the
	// password from the database.
	var (
		enteredPassword = r.PostFormValue(passwordField)
		actualHash, err = s.Database.Get("password-hash").Result()
	)

	// If there is an error while fetching the password's hash from the
	// database, send the error to the client with an Internal Server
	// Error status.
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Hash the password which the client believes to be the existing
	// password using a SHA512 hash. This is done using the Sum512
	// function to compute the SHA512 digest of the specified password.
	//
	// The %x format option converts the byte array to a string representing
	// the hash in hexadecimal format.
	requestHash := fmt.Sprintf("%x", sha512.Sum512([]byte(enteredPassword)))

	// If the requested hash is not equal to the actual hash of the
	// password, send an error telling the client exactly that, with
	// a Bad Request status code.
	if requestHash != actualHash {
		return http.StatusBadRequest, errors.New("wrong password")
	}

	// No problems have come up so just return no error, along with an OK
	// status.
	return http.StatusOK, nil
}

// changeSettings updates the server's settings, both in the database and also in
// the Settings instance in s.Settings. An error will be returned if there is a
// problem communicating with the database.
func (s *Server) changeSettings(r *http.Request) error {
	// Get all of the new settings values from the request form, except from
	// non-capital-words. The set of non capital words is initialised as an empty
	// list of strings.
	var (
		tabDirectory       = r.PostFormValue("tab-directory")
		filenamePattern    = r.PostFormValue("filename-pattern")
		nonCapitalWords    = make([]string, 0)
		charactersToRemove = r.PostFormValue("characters-to-remove")
	)

	// Parse the JSON-encoded non-capital-words into the nonCapitalWords list,
	// returning an error if the JSON data is malformed.
	if err := json.Unmarshal(
		[]byte(r.PostFormValue("non-capital-words")), &nonCapitalWords,
	); err != nil {
		return err
	}

	// Use the MSET command (sets multiple scalar values) to set the new settings
	// data into the database.
	if err := s.Database.MSet(
		"tab-directory", tabDirectory,
		"filename-pattern", filenamePattern,
		"characters-to-remove", charactersToRemove,
	).Err(); err != nil {
		return err
	}

	// Remove the database's set of non capital words in preparation for when
	// the new non-capital-words will be added.
	if err := s.Database.Del("non-capital-words").Err(); err != nil {
		return err
	}

	// Create a list of type []interface{} containing the same data as
	// nonCapitalWords but in the correct type to pass to the SADD command below.
	nonCapitalWordsI := make([]interface{}, len(nonCapitalWords))

	for i, s := range nonCapitalWords {
		nonCapitalWordsI[i] = interface{}(s)
	}

	// Use the SADD command to add each of the non capital words to the database's
	// non-capital-words set.
	if err := s.Database.SAdd("non-capital-words", nonCapitalWordsI...).Err(); err != nil {
		return err
	}

	// Now the database has been fully updated, also update the in-memory settings
	// values to the new values.
	s.Settings = &Settings{
		CharactersToRemove: charactersToRemove,
		FilenamePattern:    filenamePattern,
		NonCapitalWords:    nonCapitalWords,
		PasswordHash:       s.Settings.PasswordHash,
		TabDirectory:       tabDirectory,
	}

	return nil
}

// resetCache removes all tabs from the database, meaning they will have to be
// reloaded when the first request is made.
func (s *Server) resetCache() error {
	// Remove all keys in the database with the prefix tab:*.
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	if err := s.Database.Eval(
		`return redis.call('del', unpack(redis.call('keys', ARGV[1])))`,
		nil, "tab:*",
	).Err(); err != nil {
		return err
	}

	// Empty the tab ID list and the filename-ID map.
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	if err := s.Database.Del("tabs", "filenames").Err(); err != nil {
		return err
	}

	// Reset the tab counter to 0, so the next tab will be
	// assigned the ID of (0 + 1) = 1.
	// If there is an error, it will be returned as a HTTP error
	// with the status code 500, or Internal Server Error.
	if err := s.Database.Set("tab-counter", 0, 0).Err(); err != nil {
		return err
	}

	return nil
}

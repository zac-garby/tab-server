package src

import (
	"fmt"
	"strings"
)

// A Tab represents a tab from the database.
type Tab struct {
	Title    string   `json:"title"`
	Artist   string   `json:"artist"`
	Content  string   `json:"content"`
	ID       string   `json:"ID"`
	Filename string   `json:"filename"`
	Tags     []string `json:"tags"`
}

// tokenizePattern takes a string representing a filename pattern
// and returns a list of its tokens, which can be given to the
// parser to be parsed into the set of metadata of that particular
// file.
func tokenizePattern(pattern string) []string {
	var (
		// Initialise the list of tokens as a zero-length slice of
		// strings. It makes sense to initialise no memory to start
		// because the amount of tokens is not known at this point.
		tokens = make([]string, 0)

		// Buffer is used to build up the tokens as the pattern is
		// traversed, and will be appended to tokens frequently
		// during the execution of the algorithm.
		buffer = ""
	)

	// Iterate through each character in the pattern, not keeping
	// track of the index however since that is useless.
	for _, character := range pattern {
		// If the character denotes the beginning or end of a
		// variable, act accordingly. Otherwise, just append
		// the character to the buffer.
		switch character {
		case '[':
			tokens = append(tokens, buffer)
			buffer = "["

		case ']':
			buffer += string(character)
			tokens = append(tokens, buffer)
			buffer = ""

		default:
			buffer += string(character)
		}
	}

	// If the buffer is non-empty, append it to tokens. This
	// means that the piece of text at the end of the filename
	// will still be tokenized.
	if len(buffer) > 0 {
		tokens = append(tokens, buffer)
	}

	return tokens
}

// parseFilename parses a filename using the given tokens (which
// will probably have been returned from tokenizePattern). This
// will extract all of the metadata of the filename, returning
// each piece of data as a separate return value. If the pattern
// does not agree with the filename, the final return value: ok,
// will be equal to false.
func parseFilename(
	filename string,
	tokens []string,
) (
	title,
	artist string,
	tags []string,
	ok bool,
) {
	// Initialise the metadata values to placeholder values which
	// will be overwritten
	title = "Untitled"
	artist = "Unnamed"
	tags = make([]string, 0)

	// Set ok to true, since it will only be false in one case
	// and that case can set ok to false when it needs to.
	ok = true

	for index, token := range tokens {
		// If the current token denotes a variable:
		if isVariable(token) {
			var (
				// stop keeps track of the character which this
				// token should stop parsing at.
				stop byte

				// buffer keeps track of the value of the
				// variable being parsed.
				buffer string
			)

			// If the next token is not a variable, set stop to
			// the first character of that token.
			if index+1 < len(tokens) && !isVariable(tokens[index+1]) {
				stop = tokens[index+1][0]
			}

			// While the filename is non-empty and its first
			// character is not equal to 'stop'.
			for len(filename) > 0 && filename[0] != stop {
				// Append the first character of filename to
				// the buffer.
				buffer += string(filename[0])

				// Remove the first character from filename
				if len(filename) == 1 {
					filename = ""
				} else {
					filename = filename[1:]
				}
			}

			// If the current token is a valid variable name, assign
			// the buffer's value to the appropriate variable.
			switch token {
			case "[title]":
				title = buffer
			case "[artist]":
				artist = buffer
			case "[tag]":
				tags = append(tags, buffer)
			}
		} else {
			if strings.HasPrefix(filename, token) {
				// In this case, filename begins with the correct
				// characters such that it matches the pattern. The
				// token is removed from the start of the filename
				// and the function carries on iterating.
				filename = filename[len(token):]
			} else {
				// In this case, the filename doesn't match the
				// pattern so the function is returned from with
				// ok = false.
				ok = false
				return
			}
		}
	}

	return
}

// isVariable checks whether a string, str, begins and ends
// with square brackets.
func isVariable(str string) bool {
	return len(str) > 1 && str[0] == '[' && str[len(str)-1] == ']'
}

// removeCharacters removes the specified characters from the
// tab's title and artist name.
func (t *Tab) removeCharacters(chars string) {
	for _, character := range chars {
		// Replace all instances of the current character with a
		// space. The -1 signifies that infinitely many replacements
		// can take place (as opposed to, if I gave a number like 5,
		// a maximum of 5 replacements could happen.)
		t.Title = strings.Replace(t.Title, string(character), " ", -1)
		t.Artist = strings.Replace(t.Artist, string(character), " ", -1)
	}
}

// capitaliseString capitalises the first letter of each word except
// words which exist in the set of words to not capitalise.
func capitaliseString(str string, blacklist []string) string {
	var (
		// Initialise the output string to an empty string.
		output = ""

		// Define the list of words in str using the Fields function,
		// which splits a string by whitespace.
		words = strings.Fields(str)
	)

	// Iterate over each word in the list of words, keeping track of
	// the index of each iteration so the first word can be capitalised
	// regardless of if it's blacklisted.
	for index, word := range words {
		output += " "

		// Check whether the current word is in the list of words which
		// should not be capitalised.
		blacklisted := false
		for _, b := range blacklist {
			if strings.ToLower(str) == strings.ToLower(b) {
				blacklisted = true
				break
			}
		}

		// If this is the first word, or the word is not blacklisted,
		// append the 'titlecase' of the word to the output. The title-
		// case of a string is a copy of the string where the first
		// letter of each string is capitalised.
		// If it isn't the first word and the word is blacklisted, just
		// append the word to the output.
		if index == 0 || !blacklisted {
			output += strings.Title(word)
		} else {
			output += word
		}
	}

	// Return all but the first character of the output string, which will
	// exclude the leading space generated due to the method of joining
	// the words by spaces.
	return output[1:]
}

// applyTransformations applies both metadata transformations to the tab.
// characterCutset is the string containing the characters to be removed
// from the metadata, and then capitalisationBlacklist contains the words
// which should not be capitalised.
func (t *Tab) applyTransformations(characterCutset string, capitalisationBlacklist []string) {
	t.removeCharacters(characterCutset)
	t.Title = capitaliseString(t.Title, capitalisationBlacklist)
	t.Artist = capitaliseString(t.Artist, capitalisationBlacklist)
}

// fetchTab finds the tab corresponding to the given ID in the database
// and constructs a *Tab value to hold the information about that tab.
// If the tab does not exist, the second return parameter will be false,
// otherwise it will be true. Transformations will not be applied
func (s *Server) fetchTab(id string) (*Tab, bool, error) {
	// Compute the value of the tab's database key.
	key := "tab:" + id

	// Check whether the tab actually exists, by checking if it is a
	// member of the 'tabs' set (recall that 'tabs' is a set containing
	// the IDs of all tabs).
	exists, err := s.Database.SIsMember("tabs", id).Result()
	if err != nil {
		return nil, false, err
	} else if !exists {
		return nil, false, nil
	}

	// Use the HGETALL Redis command to get all key-value pairs from
	// the tab's hashmap. This will get all of the relavent data, except
	// for the tags, which are stored in a separate key in the database
	// and will have to be fetched separately.
	data, err := s.Database.HGetAll(key).Result()
	if err != nil {
		return nil, false, err
	}

	// Use the SMEMBERS Redis command to get a list of tags of the tab.
	tags, err := s.Database.SMembers(key + ":tags").Result()
	if err != nil {
		return nil, false, err
	}

	// Create the tab to return.
	tab := &Tab{
		ID:       data["id"],
		Artist:   data["artist"],
		Content:  data["content"],
		Title:    data["title"],
		Filename: data["filename"],
		Tags:     tags,
	}

	return tab, true, nil
}

// cacheNewTab stores a tab into the database, setting its ID to the next
// available ID. It will return an error if there is a problem with
// communicating with the database.
func (s *Server) cacheNewTab(tab *Tab) error {
	// Increment the tab-counter in the database, using the new value
	// as the ID.
	id, err := s.Database.Incr("tab-counter").Result()
	if err != nil {
		return err
	}

	// Set the tab's ID to the ID from the database, converted to a
	// string first.
	tab.ID = fmt.Sprintf("%v", id)

	// Append the ID to the tabs set.
	if err := s.Database.SAdd("tabs", id).Err(); err != nil {
		return err
	}

	// Add the filename-ID mapping to the filenames hashmap.
	if err := s.Database.HSet("filenames", tab.Filename, id).Err(); err != nil {
		return err
	}

	// Create the tab's data hashmap, in the tab:ID key.
	if err := s.Database.HMSet(fmt.Sprintf("tab:%v", id), map[string]interface{}{
		"title":    tab.Title,
		"artist":   tab.Artist,
		"content":  tab.Content,
		"id":       id,
		"filename": tab.Filename,
	}).Err(); err != nil {
		return err
	}

	// Construct a list of type []interface{} to hold the list of tags,
	// since the SADD command requires the data to be in that format.
	// This is done simply by casting each element into the interface{}
	// type and then putting them in the correct positions in the new list.
	tags := make([]interface{}, len(tab.Tags))

	for i, tag := range tab.Tags {
		tags[i] = interface{}(tag)
	}

	// Create the tab's tag set, in the tab:ID:tags key.
	if err := s.Database.SAdd(fmt.Sprintf("tab:%v:tags", id), tags...).Err(); err != nil {
		return err
	}

	return nil
}

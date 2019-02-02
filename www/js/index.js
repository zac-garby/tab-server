var tabs = []
var selectedID
var chords
var chordSymbols

// This function will be called after the DOM has been completely
// loaded, meaning that the DOM elements can be referenced from
// inside this function.
function onLoad() {
    updateTabList()
    loadChords()
}

function updateTabList() {
    // Create a new HTTP request object, which will be used
    // to fetch the list of tabs from the server.
    var req = new XMLHttpRequest()

    // The onreadystatechange method of the HTTP request is called
    // when the state of the request changes. In this case, only
    // ready state 4 (which means that the response has been received)
    // is relevant.
    req.onreadystatechange = function() {
        if (this.readyState == 4) {
            // If the status of the response is 200, the request was
            // successful. 200 = OK.
            if (this.status == 200) {
                // Parse the JSON response into a list of objects.
                tabs = JSON.parse(this.responseText)

                // Show the tabs which are now stored in the global
                // variable, taking into account the sorting and
                // filtering options.
                showTabs()

                // If there is at least one tab in the list of tabs,
                // select it initially so there isn't a huge blank
                // area covering most of the page.
                if (tabs.length > 0) {
                    selectTab(tabs[0].ID)
                }
            } else {
                // If the execution gets here, an error has occured. Thus,
                // send an error message to the user via an alert.
                alert(this.status + ": " + this.responseText)
            }
        }
    }

    // Send the HTTP GET request to /api/tabs. location.origin is
    // the URL without the current path appended, so if I'm running
    // the server locally it would be http://localhost:8000. true
    // as the third parameter indicates that the request is
    // asynchronous, meaning that the user can still interact with
    // the page while the request is loading.
    req.open("GET", location.origin + "/api/tabs", true)
    req.send()
}

function showTabs() {
    // Get the user's selected sort type from the sorting combo box.
    // This will be one of:
    //  - title-asc
    //  - title-desc
    //  - artist-asc
    //  - artist-desc
    var sortType = document.getElementById("sorting").value
    var searchTerm = document.getElementById("search-bar").value

    // Get an array of the tabs which are to be displayed by sorting
    // and filtering the entire list of tabs.
    var tabsToDisplay = tabs.sort(sortFunction(sortType))
                     .filter(filterFunction(searchTerm))

    // Find the tab-list element by its ID and remove all
    // of its children by setting its inner HTML source to
    // an empty string.
    var ul = document.getElementById("tab-list")
    ul.innerHTML = ""

    // For each tab in the list of tabs, append a new <li>
    // element with its title and artist's name to the tab
    // list element.
    for (var tab of tabsToDisplay) {
        const id = tab.ID
        var li = document.createElement("li")
        li.innerHTML = "<strong>" + tab.title + "</strong> &mdash; " + tab.artist
        li.onclick = () => selectTab(id)
        ul.appendChild(li)
    }
}

// sortFunction returns a function which compares two tab
// objects, based on the given sortOption.
// sortOption can be one of:
//  - title-asc
//  - title-desc
//  - artist-asc
//  - artist-desc
function sortFunction(sortOption) {
    switch (sortOption) {
    case "title-asc":
        return (a, b) => a.title > b.title
    case "title-desc":
        return (a, b) => a.title < b.title
    case "artist-asc":
        return (a, b) => a.artist > b.artist
    case "artist-desc":
        return (a, b) => a.artist < b.artist
    default:
        return () => 0
    }
}

// filterFunction returns a function which returns true
// if an object should be in the filtered array, and false
// otherwise, based on the search term.
function filterFunction(searchTerm) {
    // If the search term is empty, return a function which
    // always will just return true.
    if (searchTerm == "") {
        return () => true
    }

    // Extract a list of words from the search term by first
    // removing whitespace from either end of the string and
    // then using the split function with the regex /\s+/ to
    // split the string by whitespace.
    var words = searchTerm.trim().split(/\s+/)

    // Return a function which will only return true if all
    // of the words are present in the argument's title or
    // artist. This is a closure so this scope is still
    // available to the returned function, such that 'words'
    // will be in scope.
    return (s) => words.every(word =>
        s.title.toLowerCase().match(word.toLowerCase()) ||
        s.artist.toLowerCase().match(word.toLowerCase()))
}

// selectTab displays the information about the tab with the
// given ID in the text areas on the right of the page.
function selectTab(id) {
    // selected will store the tab with the given ID once the
    // linear search below has been completed, or will be
    // undefined if the tab doesn't exist.
    var selected

    // Linearly search through the list of tabs, looking for
    // the one with the requested ID.
    for (var tab of tabs) {
        if (tab.ID == id) {
            selected = tab
            break
        }
    }

    // If no matching tab was found, return from the function
    // without doing anything.
    if (selected == undefined) return

    // Update the global selected ID to the ID of the tab which
    // has just been selected.
    selectedID = id

    // Set the inner HTML fields of each of the elements which need
    // to be updated to their new values, as found in the selected
    // tab object.
    document.getElementById("title").innerHTML = selected.title
    document.getElementById("info").innerHTML = selected.artist + " (" + selected.tags + ")"
    document.getElementById("content").innerHTML = selected.content

    showChords()
}

// deleteSelected sends a HTTP request to /api/delete-tab to
// delete the currently selected tab. A prompt dialog box is
// opened to ask the user to enter their password.
function deleteSelected() {
    // Ask the user to enter their password by opening up a
    // prompt dialog, displaying the message "Enter your password:".
    // No validation needs to be done here, as the ID will be
    // validated server-side.
    var password = prompt("Enter your password:")

    // Create a new HTTP request object, which will be used
    // to request the tab be deleted.
    var req = new XMLHttpRequest()

    // The onreadystatechange method of the HTTP request is called
    // when the state of the request changes. In this case, only
    // ready state 4 (which means that the response has been received)
    // is relevant.
    req.onreadystatechange = function() {
        if (this.readyState == 4) {
            // If the status of the response is 200, the request was
            // successful. 200 = OK.
            if (this.status == 200) {
                // In this case, the tab has been successfully deleted from
                // the server. So now, the tab list should be reloaded to reflect those changes.
                updateTabList()
            } else {
                // If the execution gets here, an error has occured. Thus,
                // send an error message to the user via an alert.
                alert(this.status + ": " + this.responseText)
            }
        }
    }

    // Create a URLSearchParams object to store the form values which will
    // be sent in the POST request to the server.
    var params = new URLSearchParams()
    params.set("password", password)
    params.set("id", selectedID)

    // Send the HTTP GET request to /api/delete-tab. location.origin is
    // the URL without the current path appended, so if I'm running
    // the server locally it would be http://localhost:8000. true
    // as the third parameter indicates that the request is
    // asynchronous, meaning that the user can still interact with
    // the page while the request is loading.
    req.open("POST", location.origin + "/api/delete-tab", true)
    req.send(params)
}

function loadChords() {
    var req = new XMLHttpRequest()

    req.onreadystatechange = function() {
        if (this.readyState == 4) {
            if (this.status == 200) {
                chords = JSON.parse(this.responseText)
                chordSymbols = Object.keys(chords).sort((a, b) => a.length < b.length)
                showChords()
            } else {
                console.log("error loading chords:", this.status)
            }
        }
    }

    req.open("GET", location.origin + "/static/lib/chord-collection/chords.json")
    req.send()
}

// showChords will highlight each of the chords in the currently selected
// tab and will add event handlers such that when the chord symbol is
// clicked it will show how to play that chord.
function showChords() {
    // This function cannot work if either no tab is selected or the
    // chords list hasn't been loaded yet, so these conditions must be
    // tested and verified before going any further. This isn't an error
    // situation though, the the function is silently returned from.
    if (selectedID === undefined || chords === undefined) {
        return
    }

    // selected will store the tab with the selected ID once the
    // linear search below has been completed, or will be
    // undefined if the tab doesn't exist.
    var selected

    // Linearly search through the list of tabs, looking for
    // the one with the selected ID.
    for (var tab of tabs) {
        if (tab.ID == selectedID) {
            selected = tab
            break
        }
    }

    // If no tab is selected, return early from the function as nothing
    // can be done.
    if (selected === undefined) {
        return
    }

    // This list will store the indexes and lengths of each chord in the content.
    var indexes = []

    for (var symbol of chordSymbols) {
        // Create a RegExp object using the current chord symbol. Plus
        // signs are escaped because they are special characters in
        // regular expressions and as such they would have a specific meaning
        // if they were not escaped.
        let sym = symbol.replace("+", "\\+")
        const pattern = new RegExp(`\\s(${sym})\\s`, "g")

        // For each RegExp match in the currently selected tab content:
        while ((match = pattern.exec(selected.content)) != null) {
            let index = match.index + 1
            // If the index of the chord which was just found lies inside any
            // previously found chord, skip to the next match.
            if (indexes.find(c => index >= c.index && index <= c.index + c.length) != undefined) {
                continue
            }

            indexes.push({
                index: index,
                length: symbol.length,
            })
        }
    }

    // Sort the indexes such that they are in order for splitContent. splitContent
    // only works on sorted inputs.
    indexes.sort((a, b) => a.index > b.index)

    // Format the indexes list in the format which the splitContent function likes
    // so it can be processed further.
    formattedIndexes = indexes.map(c => [c.index, c.index + c.length]).flat()

    // Split the content into an array of consecutive substrings to separate out
    // the chord symbols from the rest of the tab.
    var split = splitContent(selected.content, formattedIndexes)

    var pre = document.getElementById("content")
    document.getElementById("content").innerHTML = ""

    // Initialise a flag variable to keep track of whether the current content part
    // is a chord or not. Each iteration of the loop below, this variable is inverted,
    // because the parts are arranged in an alternating fashion.
    var isChord = false

    // Iterate over each of the parts of the split content.
    for (var part of split) {
        // Add a <span> element to the <pre> to represent the current content part. If
        // it is a chord, set its class name to chord-symbol.
        var span = document.createElement("span")
        span.innerHTML = part
        if (isChord) span.className = "chord-symbol"
        pre.appendChild(span)

        isChord = !isChord
    }
}

function splitContent(content, indexes) {
    // The base case. If there are no indexes, the content is placed
    // inside a list and returned straight away.
    if (indexes.length == 0) {
        return [content]
    }

    // Otherwise, the main chunk of the algorithm is carried out. First, the
    // first index is separated from the rest of them and the rest are mapped
    // such that they are all lowered by the value of the first index.
    var first = indexes[0]
    var rest = indexes.slice(1).map(n => n - first)

    // The recursive step. Where n = first, return a list where the first element
    // is the first n characters of the content and the rest of the list is another
    // call to splitContent with the rest of the content and the rest of the indexes.
    return [content.slice(0, first)].concat(splitContent(content.slice(first), rest))
}
var tabs = []

// This function will be called after the DOM has been completely
// loaded, meaning that the DOM elements can be referenced from
// inside this function.
function onLoad() {
    updateTabList()
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
        var li = document.createElement("li")
        li.innerHTML = "<strong>" + tab.title + "</strong> &mdash; " + tab.artist
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
function onLoad() {
    // Create a new HTTP request object, which will be used to fetch
    // the current settings from the server.
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
                settings = JSON.parse(this.responseText)

                // Set the initial values of the various inputs to their
                // corresponding settings values.
                document.getElementById("tab-directory").value = settings["tab-directory"]
                document.getElementById("filename-pattern").value = settings["filename-pattern"]
                document.getElementById("non-capital-words").value = settings["non-capital-words"]
                document.getElementById("characters-to-remove").value = settings["characters-to-remove"]
            } else {
                // If the execution gets here, an error has occured. Thus,
                // send an error message to the user via an alert.
                alert(this.status + ": " + this.responseText)
            }
        }
    }

    // Send the HTTP GET request to /api/settings. location.origin is
    // the URL without the current path appended, so if I'm running
    // the server locally it would be http://localhost:8000. true
    // as the third parameter indicates that the request is
    // asynchronous, meaning that the user can still interact with
    // the page while the request is loading.
    req.open("GET", location.origin + "/api/settings", true)
    req.send()
}

// changeSettings sends a request to /api/change-settings, sending the four
// parameters as POST values. It will also prompt the user to enter their
// password in a dialog box.
function changeSettings(tabDirectory, filenamePattern, nonCapitalWords, charactersToRemove) {
    // Ask the user to enter their password by opening up a
    // prompt dialog, displaying the message "Enter your password:".
    // No validation needs to be done here, as the ID will be
    // validated server-side.
    var password = prompt("Enter your password:")

    // Create a new HTTP request object, which will be used to change the
    // settings by sending a request.
    var req = new XMLHttpRequest()

    // The onreadystatechange method of the HTTP request is called
    // when the state of the request changes. In this case, only
    // ready state 4 (which means that the response has been received)
    // is relevant.
    req.onreadystatechange = function() {
        if (this.readyState == 4) {
            // If the status is 200, the request was OK and the settings
            // change was successful.
            if (this.status == 200) {
                alert("The settings have been updated! You may want to reload the\
tabs from their files, otherwise the changes won't show up until you do.")
            } else {
                // At this point, some error has occured, so send an error
                // message to the user.
                alert(this.status + ": " + this.responseText)
            }
        }
    }

    // Create a URLSearchParams object to store the form values which will
    // be sent in the POST request to the server.
    var params = new URLSearchParams()
    params.set("password", password)
    params.set("tab-directory", tabDirectory)
    params.set("filename-pattern", filenamePattern)
    params.set("non-capital-words", nonCapitalWords)
    params.set("characters-to-remove", charactersToRemove)

    // Send the HTTP GET request to /api/change-settings. location.origin is
    // the URL without the current path appended, so if I'm running
    // the server locally it would be http://localhost:8000. true
    // as the third parameter indicates that the request is
    // asynchronous, meaning that the user can still interact with
    // the page while the request is loading.
    req.open("POST", location.origin + "/api/change-settings", true)
    req.send(params)
}

// apply gets the values from the input fields in the settings form and calls
// changeSettings with those values to update the server's settings.
function apply() {
    // Get the value of each input field, storing them in variables.
    var tabDirectory = document.getElementById("tab-directory").value
    var filenamePattern = document.getElementById("filename-pattern").value
    var charactersToRemove = document.getElementById("characters-to-remove").value

    // Perform input validation. The only constraints are that both the
    // tab directory and filename pattern at at least one character long.
    if (tabDirectory.length == 0) {
        alert("You must enter a value for the tab directory")
        return
    } else if (filenamePattern.length == 0) {
        alert("You must enter a value for the filename pattern")
        return
    }

    // nonCapitalWords is expected to be  JSON-encoded array of strings,
    // where each string is a word to be ignored when capitalising metadata.
    // This piece of code:
    //  - finds the entered value --> "a, b, c"
    //  - splits it by commas     --> ["a", " b", " c"]
    //  - trims whitespace        --> ["a", "b", "c"]
    //  - encodes it with JSON    --> "[\"a\", \"b\", \"c\"]"
    var nonCapitalWords = JSON.stringify(
        document
        .getElementById("non-capital-words")
        .value
        .split(",")
        .map(s => s.trim()))
    
    changeSettings(tabDirectory, filenamePattern, nonCapitalWords, charactersToRemove)
}

// reloadTabs removes all of the cached tabs from the database by sending
// a HTTP request to /api/reset-cache. It will send an alert to the user
// to say if that was successful or not.
function reloadTabs() {
    // Make a new HTTP request object which will be used to send the HTTP
    // request.
    var req = new XMLHttpRequest()

    // The readystatechange will run the following function when the state
    // of the request is changed, which includes receiving a response.
    req.onreadystatechange = function() {
        if (this.readyState == 4) {
            // Check the response status and send an alert message based
            // on that status, saying either that the request was successful
            // or not.
            if (this.status == 200) {
                alert("The tabs have been removed from the database. They can be reloaded by navigating back to the home page.")
            } else {
                alert("Something went wrong and the tabs couldn't be reloaded!")
            }
        }
    }

    // Send the HTTP request using the GET method to /api/reset-cache.
    req.open("GET", location.origin + "/api/reset-cache", true)
    req.send()
}
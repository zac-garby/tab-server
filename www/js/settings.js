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
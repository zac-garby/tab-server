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
                var tabs = JSON.parse(this.responseText)

                // Find the tab-list element by its ID and remove all
                // of its children by setting its inner HTML source to
                // an empty string.
                var ul = document.getElementById("tab-list")
                ul.innerHTML = ""

                // For each tab in the list of tabs, append a new <li>
                // element with its name to the tab list element.
                for (var tab of tabs) {
                    var li = document.createElement("li")
                    li.innerHTML = tab.title
                    ul.appendChild(li)
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
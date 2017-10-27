var data = null; // the mountain and camera info recieved from the api.
var urlBase = ""; // the base url on which to build api requests.

// Sets up urlBase, the mountain/camera 'dropdowns' + associated data,
// and attatches an onchange listener to the mountain dropdown.
window.onload = function (e) {
    urlBase = window.location.protocol + "//" + window.location.hostname;
    if (window.location.port != '' && window.location.port != 0) {
        urlBase = urlBase + ":" + window.location.port;
    }
    var mtSelect = document.getElementById("mountain");
    var camSelect = document.getElementById("cam");

    // attach functionality to the mountain selection dropdown
    mtSelect.onchange = function (ev) {
        data.forEach(function (mt) { // inefficient 'search' for data
            if (mt["id"] == mtSelect.value) {
                // remove all old children
                rmAllChildren(camSelect);
                // make and add new children
                mt["cams"].forEach(function (cam) {
                    //create new element
                    var e = createCamSelectOption(cam);
                    camSelect.appendChild(e);
                }, this);
            }
        }, this);
    };

    setupMtCamSelection();
}

// Requests scraperecords from the api between the dates specified in the ui,
// then populates rows of the table.
function getScrapes() {
    var start = document.getElementById("start").value;
    var end = document.getElementById("end").value;

    // if (start == '' || start == null) {
    //     alert("'From' must be specified.");
    // }
    // if (end == '' || end == null) {
    //     alert("'To' must be specified");
    // }

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (request.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
            if (request.status == 200) {
                var scrapes = JSON.parse(request.responseText);
                var sTable = document.getElementById("scrapes");

                // empty table
                rmAllChildren(sTable);

                // add header
                sTable.appendChild(createScrapeHeader());

                // add each item to the table
                scrapes.forEach(function (scrape) {
                    var r = createScrapeRow(scrape);
                    sTable.appendChild(r);
                }, this);
            }
            else if (request.status == 400) {
                alert('There was an error 400');
            }
            else {
                alert('something else other than 200 was returned');
            }
        }
    };

    // prepare and make api request
    var mt = document.getElementById("mountain").value;
    var cam = document.getElementById("cam").value;
    var asLocal = document.getElementById("as-local").checked;
    var url = urlBase + "/api/mountains/" + mt +
        "/cams/" + cam +
        "/scrapes?start=" + start +
        "&end=" + end +
        "&as_local_time=" + asLocal;
    request.open("GET", url, true);
    request.send();
}

// Gets mountain and cam data from the api, then populates the
// respective 'dropdown' items in the ui.
function setupMtCamSelection() {

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (request.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
            if (request.status == 200) {
                data = JSON.parse(request.responseText);
                var mtSelect = document.getElementById("mountain");

                // do mountain select
                data.forEach(function (mt) {
                    mtSelect.appendChild(createMtSelectOption(mt));
                }, this);
                mtSelect.onchange();
            }
            else if (request.status == 400) {
                alert('There was an error 400');
            }
            else {
                alert('something else other than 200 was returned');
            }
        }
    };

    // prepare and make api request
    var url = urlBase + "/api/data";
    request.open("GET", url, true);
    request.send();
}

// remove all child elements of element.
function rmAllChildren(element) {
    while (element.firstChild) {
        element.removeChild(element.firstChild);
    }
}

// create a single `<option>` element for the mountain 'dropdown'
function createMtSelectOption(mt) {
    var e = document.createElement("option");
    e.value = mt["id"];
    e.innerText = mt["name"] + " (" + mt["state"] + ")";
    return e;
}

// create a single `<option>` element for the camera 'dropdown'
function createCamSelectOption(cam) {
    var e = document.createElement("option");
    e.value = cam["id"];
    e.innerText = cam["name"] + " (" + cam["elevation_ft"] + "ft)";
    return e;
}

// create header `tr` for scrape display table
function createScrapeHeader() {
    var tr = document.createElement("tr");
    var c1 = document.createElement("th");
    var c2 = document.createElement("th");
    var c3 = document.createElement("th");
    c1.innerText = "Time (Pacific Time)";
    c2.innerText = "Result";
    c3.innerText = "Detail";
    tr.appendChild(c1);
    tr.appendChild(c2);
    tr.appendChild(c3);
    return tr;
}

// create a single `tr` for the given scrape record
function createScrapeRow(scrape) {
    var tr = document.createElement("tr");
    var c1 = document.createElement("td");
    var c2 = document.createElement("td");
    var c3 = document.createElement("td");

    c1.innerText = scrape["time"];
    c1.classList.add("time");

    if (scrape["result"] == "success") {
        c2.innerHTML = "<a href=" + scrape["file"] + " target=\"_blank\">" + scrape["result"] + "</a>";
    } else {
        c2.innerText = scrape["result"];
    }
    c2.classList.add("result");

    c3.innerText = scrape["detail"];
    c3.classList.add("detail");

    tr.appendChild(c1);
    tr.appendChild(c2);
    tr.appendChild(c3);
    return tr;
}
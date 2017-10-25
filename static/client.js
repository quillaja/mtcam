var data = null;
var urlBase = "";

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
                scrapes.forEach(function (cam) {
                    var r = createScrapeRow(cam);
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
    var mt = document.getElementById("mountain").value;
    var cam = document.getElementById("cam").value;
    var url = urlBase + "/api/mountains/" + mt +
        "/cams/" + cam +
        "/scrapes?start=" + start +
        "&end=" + end;
    request.open("GET", url, true);
    request.send();
}


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

function createMtSelectOption(mt) {
    var e = document.createElement("option");
    e.value = mt["id"];
    e.innerText = mt["name"] + " (" + mt["state"] + ")";
    return e;
}

function createCamSelectOption(cam) {
    var e = document.createElement("option");
    e.value = cam["id"];
    e.innerText = cam["name"] + " (" + cam["elevation_ft"] + "ft)";
    return e;
}

function createScrapeHeader() {
    var tr = document.createElement("tr");
    var c1 = document.createElement("th");
    var c2 = document.createElement("th");
    var c3 = document.createElement("th");
    var c4 = document.createElement("th");
    c1.innerText = "Time";
    c2.innerText = "Result";
    c3.innerText = "Image";
    c4.innerText = "Detail";
    tr.appendChild(c1);
    tr.appendChild(c2);
    tr.appendChild(c3);
    tr.appendChild(c4);
    return tr;
}

function createScrapeRow(cam) {
    var tr = document.createElement("tr");
    var c1 = document.createElement("td");
    var c2 = document.createElement("td");
    var c3 = document.createElement("td");
    var c4 = document.createElement("td");
    c1.innerText = cam["time"];
    c2.innerText = cam["result"];
    if (cam["result"] == "success") {
        c3.innerHTML = "<a href=" + cam["file"] + " target=\"_blank\">image</a>";
    } else {
        c3.innerHTML = "";
    }
    c4.innerText = cam["detail"];
    tr.appendChild(c1);
    tr.appendChild(c2);
    tr.appendChild(c3);
    tr.appendChild(c4);
    return tr;
}
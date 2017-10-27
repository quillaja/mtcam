var data = null; // the mountain and camera info recieved from the api.
var scrapes = null; // the scraperecords from the api (upon user action)
var urlBase = ""; // the base url on which to build api requests.

// tabs data structure
var tabData = {
    "Log": "scrapes",
    "Timelapse": "timelapse",
    "Location Info": "info"
};

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

    // setup tab bar
    setupTabBar();

    // get data for 'dropdowns' from api
    setupMtCamSelection();

    // attach functionality to "load data" button
    document.getElementById("submit").onclick = loadAndDisplayData;
}

// Requests scraperecords from the api between the dates specified in the ui,
// then populates rows of the table.
function loadAndDisplayData() {

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (request.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
            if (request.status == 200) {
                // parse scrapes
                scrapes = JSON.parse(request.responseText);

                // display various data
                populateScrapeTable();
                showInfo();
                // makeTimelapse();
                document.getElementById("tab-area").classList.remove("hidden");
            }
            else if (request.status == 400) {
                alert('There was an error 400');
            }
            else {
                alert('(getScrapes) something else other than 200 was returned');
            }
        }
    };

    // prepare and make api request
    var start = document.getElementById("start").value;
    var end = document.getElementById("end").value;
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

function populateScrapeTable() {
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
                alert('(setupMtCam) something else other than 200 was returned');
            }
        }
    };

    // prepare and make api request
    var url = urlBase + "/api/data";
    request.open("GET", url, true);
    request.send();
}

// creates the tabs from the information in `tabData`
function setupTabBar() {
    var tabBar = document.getElementById("tab-area");

    for (var k in tabData) {
        t = document.createElement("span");
        t.classList.add("tab");
        t.innerText = k;
        t.onclick = function () { tabClicked(this); }
        tabBar.appendChild(t);
    }

    tabClicked(tabBar.children[0]);
}

// toggles the tab clicked.
// TODO: make tab click change display area visibility
// TODO: Ignore above TODO. Decided "load data" will initiate everything
function tabClicked(tab) {

    var selected = document.querySelector(".tab.selected");
    if (selected != null) {
        var oldContentId = tabData[selected.innerText];
        selected.classList.remove("selected");
        document.getElementById(oldContentId).classList.add("hidden");
    }

    tab.classList.add("selected");
    var newContentId = tabData[tab.innerText];
    document.getElementById(newContentId).classList.remove("hidden");
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

// TODO: eliminate repetition. Each type of object (mt, cam, scrapes)
// will get customized UI presentation, so can't totally refactor this =(
function showInfo() {

    // most of this section sucks
    var mtId = document.getElementById("mountain").value;
    var camId = document.getElementById("cam").value;
    var mt = null;
    var cam = null;

    data.forEach(function (m) {
        if (m["id"] == mtId) {
            mt = m;
            mt["cams"].forEach(function (c) {
                if (c["id"] == camId) {
                    cam = c;
                    return;
                }
            }, this);
            return;
        }
    }, this);

    var locBox = document.getElementById("info");
    rmAllChildren(locBox);

    if (mt != null) {
        var mInfoBox = document.createElement("span");
        mInfoBox.classList.add("info-box");

        // Name, state
        addPropValueToElement(mInfoBox, mt["name"] + "  (" + mt["state"] + ")", "");
        // elevation
        addPropValueToElement(mInfoBox, "Elevation (ft)", mt["elevation_ft"]);
        // lat, lon (link)
        addPropValueToElement(mInfoBox, "Location", mapLink(mt["latitude"], mt["longitude"]));
        // timeZoneName, rawOffset, dstOffset
        var tz = mt["tz"]["timeZoneName"] + "<br>(UTC+" +
            secToHr(mt["tz"]["rawOffset"]) + "hr, +" + secToHr(mt["tz"]["dstOffset"]) + "hr DST)";
        addPropValueToElement(mInfoBox, "Timezone", tz);

        locBox.appendChild(mInfoBox);
    }

    if (cam != null) {
        var cInfoBox = document.createElement("span");
        cInfoBox.classList.add("info-box");

        // Name, active
        addPropValueToElement(cInfoBox, cam["name"], "active = " + cam["is_active"]);
        // elevation
        addPropValueToElement(cInfoBox, "Elevation (ft)", cam["elevation_ft"]);
        // lat,lon (link)
        addPropValueToElement(cInfoBox, "Location", mapLink(cam["latitude"], cam["longitude"]));
        // interval
        addPropValueToElement(cInfoBox, "Interval (min)", cam["interval"]);
        // comment
        addPropValueToElement(cInfoBox, "Comment", cam["comment"]);

        locBox.appendChild(cInfoBox);
    }

    if (scrapes != null) {
        var sInfoBox = document.createElement("span");
        sInfoBox.classList.add("info-box");
        stats = {
            "Scrape Statistics": "",
            "total": 0,
            "success": 0,
            "failure": 0,
            "idle": 0,
            "success rate": 0.0
        };
        scrapes.forEach(function (scrape) {
            stats["total"] += 1;
            stats[scrape["result"]] += 1;
        }, this);
        stats["success rate"] = (100 * stats["success"] / (stats["success"] + stats["failure"])).toFixed(1) + "%";

        for (var k in stats) {
            addPropValueToElement(sInfoBox, k, stats[k]);
        }
        locBox.appendChild(sInfoBox);
    }
}

// constructs and adds a row of "<property name, value>" 
// to elem ( which is class .info-box)
function addPropValueToElement(elem, prop, val) {
    var p = document.createElement("span");
    p.classList.add("property");
    p.innerHTML = prop;
    var v = document.createElement("span");
    v.classList.add("value");
    v.innerHTML = val;
    elem.appendChild(p);
    elem.appendChild(v);
}

function secToHr(sec) {
    return Math.trunc(sec / 3600.0);
}

function mapLink(lat, lon) {
    return '<a href="https://www.google.com/maps/place/' + lat + ',' + lon +
        '" target="_blank">' + lat + ', ' + lon + '</a>';
}
var data = null; // the mountain and camera info recieved from the api.
var scrapes = null; // the scraperecords from the api (upon user action)
var urlBase = ""; // the base url on which to build api requests.
var tldisp = null; // element holding the timelapse images
var tlProg = null; // elemebt holding timelapse 'progress' text
var tlFrame = 0; // frame the timelapse is displaying
var tlFrameTime = 1.0; // time in sec between each timelapse frame, from speed dropdown
var tlPaused = true; // to pause/play the timelapse

// tabs data structure
var tabData = {
    "Help": "help",
    "Info": "info",
    "Timelapse": "timelapse",
    "Log": "scrapes",
    "Weather": "weather",
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
        mt = data[mtSelect.value];
        // remove all old children
        rmAllChildren(camSelect);
        // make and add new children
        for (var k in mt["cams"]) {
            var cam = mt["cams"][k];
            //create new element
            camSelect.appendChild(createCamSelectOption(cam));
        }
        showInfo();

    };

    // attach functionality to camera selection dropdown
    camSelect.onchange = function (ev) {
        // selecting a camera will show it's info and hide
        // whatever is showing in timelapse and scrapes(log) tabs
        showInfo();
        tabClicked("info-tab");

        tabVisible("timelapse-tab", false);
        tabVisible("scrapes-tab", false);
    }

    // setup tab bar
    setupTabBar();

    // get data for 'dropdowns' from api
    setupMtCamSelection();

    // attach functionality to "load photos" button
    document.getElementById("submit-photos").onclick = function () {
        let requestMade = loadAndDisplayPhotos();

        if (requestMade) {

            // show previously hidden tabs
            tabVisible("info-tab", true);
            tabVisible("timelapse-tab", true);
            tabVisible("scrapes-tab", true);

            // flash 2 'background' tabs
            tabFlash("timelapse-tab");
            tabFlash("scrapes-tab");

            // for space savings on narrower devices (phones), hide 'help'
            // tab. I decided to show the 'info' tab by default instead.
            tabClicked("info-tab");
            tabVisible("help-tab", false);
        } else {
            alert("Sorry, but you've asked for a time span longer than 7 days. To maintain performance of both the server and your browser, please select a time span 7 days or less. Thanks!\n-Ben");
        }
    };

    // attach functionality to "load weather" button
    // document.getElementById("submit-weather").onclick = function () {
    //     loadAndDisplayWeather();

    //     tabVisible("weather-tab", true);
    // };

    // set the default 'start' to be Today at 00:00 (12am)
    var dt = new Date();
    dt.setHours(0);
    dt.setMinutes(0);
    dt.setSeconds(0);
    dt.setMilliseconds(0);
    document.getElementById("start").value = strDatetimeLocal(dt);
}

// Requests scraperecords from the api between the dates specified in the ui,
// then populates rows of the table.
function loadAndDisplayPhotos() {

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (request.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
            if (request.status == 200) {
                // parse scrapes
                scrapes = JSON.parse(request.responseText);

                // display various data
                showInfo();
                populateScrapeTable();
                makeTimelapse();

                // show tabs for the 3 displays created above
                // weather tab is unhidden here, but populated with content
                // in another independent function because it's a separate
                // API call.
                // TODO: streamline this? if separate, then should be really
                // separate.
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

    let dt = new Date(end) - new Date(start);
    let ok = isNaN(dt); // NaN means one or both of start/end are unset (which is ok)
    if (!ok) {
        const msperday = 86400000; // ms in 1 day
        let lenDays = Math.ceil(dt / msperday);
        ok = lenDays <= 7; // if the timespan is 1 week or less, ok.
    }

    // do request if ok
    if (ok) {
        var url = urlBase + "/api/mountains/" + mt +
            "/cams/" + cam +
            "/scrapes?start=" + start +
            "&end=" + end;

        request.open("GET", url, true);
        request.send();
    }

    return ok;
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

// Request bokeh html elements from the API for the weather data of the
// mountain and dates desired. Show the result in the client page.
function loadAndDisplayWeather() {
    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (request.readyState == XMLHttpRequest.DONE) {
            if (request.status == 200) {
                // take recieved html and javascript, insert into document,
                // and unhide weather tab.
                var data = JSON.parse(request.responseText);

                var weather = document.getElementById("weather");
                rmAllChildren(weather);
                weather.innerHTML = data['div'];

                // the cleanest way I found to 'inject' a script into the DOM
                // had to remove <script> tags from bokeh output on back end
                var script = document.createElement("script");
                var innerScript = document.createTextNode(data["script"]);
                script.appendChild(innerScript);
                weather.appendChild(script);
            }
            else {
                alert('(loadWeather) error ' + request.status + ' was returned');
            }
        }
    };

    var start = document.getElementById("start").value;
    var end = document.getElementById("end").value;
    var mt = document.getElementById("mountain").value;
    var asLocal = document.getElementById("as-local").checked;

    var url = urlBase + "/api/mountains/" + mt +
        "/weather?format=bokeh" +
        "&start=" + start +
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

                // create mountain select options
                for (var k in data) {
                    mtSelect.appendChild(createMtSelectOption(data[k]));
                }
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
        t.id = tabData[k] + "-tab";
        t.classList.add("tab");
        t.classList.add("hidden");
        t.innerText = k;
        t.onclick = function () { tabClicked(this.id); }
        tabBar.appendChild(t);
    }

    // remove 'hidden' from tab, then simulate click to unhide the
    // associated content and apply correct styles.
    tabVisible("help-tab", true);
    tabClicked("help-tab");

    tabVisible("info-tab", true);
}

// set tab visbility.
function tabVisible(tabId, visible) {
    let tab = document.getElementById(tabId)
    if (visible) {
        tab.classList.remove("hidden");
    } else {
        tab.classList.add("hidden");
    }
}

function tabFlash(tabId) {
    let tab = document.getElementById(tabId);
    tab.classList.add("flash");
    window.setTimeout(() => { tab.classList.remove("flash"); }, 200);
}

// toggles the tab clicked.
function tabClicked(tabId) {

    var selected = document.querySelector(".tab.selected");
    if (selected != null) {
        var oldContentId = tabData[selected.innerText];
        selected.classList.remove("selected");
        tabVisible(oldContentId, false);
    }

    var tab = document.getElementById(tabId);
    tab.classList.add("selected");
    var newContentId = tabData[tab.innerText];
    tabVisible(newContentId, true);
}

// remove all child elements of element.
function rmAllChildren(element) {
    while (element.firstChild) {
        element.removeChild(element.firstChild);
    }
}

// converts a Date() object into a string for use in the datetime-local input.
// use lame conversion function because JS+HTML is too fucking stupid
// return format YYYY-MM-DDTHH:MM which is necessary for api request.
function strDatetimeLocal(date) {
    var
        ten = function (i) {
            return (i < 10 ? '0' : '') + i;
        },
        YYYY = date.getFullYear(),
        MM = ten(date.getMonth() + 1),
        DD = ten(date.getDate()),
        HH = ten(date.getHours()),
        II = ten(date.getMinutes());
    //SS = ten(date.getSeconds());

    return YYYY + '-' + MM + '-' + DD;// + 'T' +
    // HH + ':' + II;
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
    c1.innerText = "Time";
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

// TODO: eliminate repetition? Each type of object (mt, cam, scrapes)
// will get customized UI presentation, so can't totally refactor this =(
function showInfo() {

    // get the keys (id) to access the correct data
    var mtId = document.getElementById("mountain").value;
    var camId = document.getElementById("cam").value;
    var mt = data[mtId];
    var cam = mt["cams"][camId];

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
        var tz = mt["tz"];
        // var tz = mt["tz"]["timeZoneName"] + "<br>(UTC+" +
        //     secToHr(mt["tz"]["rawOffset"]) + "hr, +" + secToHr(mt["tz"]["dstOffset"]) + "hr DST)";
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


function makeTimelapse() {

    // initialize/reset state
    tlFrame = 0;
    tlPaused = true;
    document.getElementById("play").innerHTML = "Play"; // meh, hackish
    tldisp = document.getElementById("timelapse-display");
    tlProg = document.getElementById("progress");

    // attach button actions
    document.getElementById("speed").onchange = function () {
        setTimelapseSpeed(this);
    };
    setTimelapseSpeed();

    // create img children
    hasImgs = createTimelapseImages();

    // do not make buttons work if there are no images to play
    if (hasImgs) {
        document.getElementById("previous").onclick = prevTimelapseImg;
        document.getElementById("next").onclick = nextTimelapseImg;
        document.getElementById("play").onclick = toggleTimelapsePause;
        updateTimelapseProgress();
    }

}

function createTimelapseImages() {
    rmAllChildren(tldisp); // clear it out

    // fill it up
    scrapes.forEach(function (scrape) {
        if (scrape["result"] == "success") {
            var img = document.createElement("img");
            img.src = scrape["file"];
            img.classList.add("hidden");
            tldisp.appendChild(img);
        }
    }, this);

    if (tldisp.children.length > 0) {
        tldisp.children[0].classList.remove("hidden"); // unhide first
        return true;
    } else {
        tldisp.innerText = "No images to display.";
        return false;
    }
}

function updateTimelapseProgress() {
    tlProg.innerText = `${1 + tlFrame}/${tldisp.children.length}`;
}

function setTimelapseSpeed(speedDropdown = null) {
    if (speedDropdown == null) {
        speedDropdown = document.getElementById("speed");
    }

    tlFrameTime = 1.0 / speedDropdown.value;
}

function nextTimelapseImg() {
    var oldFrame = tlFrame;
    tlFrame++;
    if (tlFrame >= tldisp.children.length) {
        tlFrame = 0;
    }
    tldisp.children[tlFrame].classList.remove("hidden");
    tldisp.children[oldFrame].classList.add("hidden");
    updateTimelapseProgress();
}

function prevTimelapseImg() {
    var oldFrame = tlFrame;
    tlFrame--;
    if (tlFrame < 0) {
        tlFrame = tldisp.children.length - 1;
    }
    tldisp.children[tlFrame].classList.remove("hidden");
    tldisp.children[oldFrame].classList.add("hidden");
    updateTimelapseProgress();
}

function playTimelapse() {
    nextTimelapseImg();
    if (!tlPaused) {
        window.setTimeout(playTimelapse, tlFrameTime * 1000);
    }
}

function toggleTimelapsePause() {
    var playBtn = document.getElementById("play");
    if (tlPaused) {
        playBtn.innerHTML = "Pause";
        tlPaused = false;
        playTimelapse();
    } else {
        playBtn.innerHTML = "Play";
        tlPaused = true;
    }
}

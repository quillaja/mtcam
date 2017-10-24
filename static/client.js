var data = null;

window.onload = function (e) {
    // alert("load");
    var mtSelect = document.getElementById("mountain");
    var camSelect = document.getElementById("cam");
    mtSelect.onchange = function (ev) {
        data.forEach(function (mt) { // inefficient 'search' for data
            if (mt["id"] == mtSelect.value) {
                // remove all old children
                while (camSelect.firstChild) {
                    camSelect.removeChild(camSelect.firstChild);
                }
                // make and add new children
                mt["cams"].forEach(function (cam) {
                    //create new element
                    var e = document.createElement("option");
                    e.value = cam["id"];
                    e.innerText = cam["name"];

                    // add each new element
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

    if (start == '' || start == null) {
        alert("'From' must be specified.");
    }
    if (end == '' || end == null) {
        alert("'To' must be specified");
    }

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (request.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
            if (request.status == 200) {
                var sTable = document.getElementById("scrapes");
                var scrapes = JSON.parse(request.responseText);
                scrapes.forEach(function (element) {
                    var r = document.createElement("tr");
                    r.textContent = JSON.stringify(element)
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
    var url = "http://127.0.0.1:5000/api/mountains/" + mt +
        "/cams/" + cam +
        "/scrapes?start=" + start +
        "&end=" + end;
    alert(url);
    request.open("GET", url, true);
    request.send();
}


function setupMtCamSelection() {

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (request.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
            if (request.status == 200) {
                // alert(request.responseText)
                var mtSelect = document.getElementById("mountain");
                data = JSON.parse(request.responseText);

                // do mountain select
                data.forEach(function (mt) {
                    var e = document.createElement("option");
                    e.value = mt["id"];
                    e.innerText = mt["name"] + " (" + mt["state"] + ")"
                    mtSelect.appendChild(e);
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

    var url = "http://127.0.0.1:5000/api/data";
    request.open("GET", url, true);
    request.send();
}
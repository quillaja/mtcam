package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/quillaja/mtcam/db"
	"github.com/quillaja/mtcam/log"
	"github.com/quillaja/mtcam/model"
)

// content type header
const contenttype = "Content-Type"

// json mime
const jsonMime = "application/json"

// HTTP strict transport security
// see https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Strict_Transport_Security_Cheat_Sheet.html
const hstsHeader = "Strict-Transport-Security"
const hstsValue = "max-age=86400; includeSubDomains" // 24 hr lifetime

// time.Format() formats
const (
	datefmt     = "2006-01-02"
	datetimefmt = "2006-01-02 15:04"
	datetzfmt   = "2006-01-02 -0700"
)

// addHSTS is a 'middleware' that writes HSTS headers to all requests.
func addHSTS(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add(hstsHeader, hstsValue)
			h.ServeHTTP(w, req)
		})
}

// CreateHandler creates a ServeMux for the given serverd config.
func CreateHandler(cfg *ServerdConfig) http.Handler {
	mux := http.NewServeMux()

	// add handlers for API endpoints
	mux.HandleFunc(cfg.Routes.Api+"data/", ApiData())
	mux.HandleFunc(cfg.Routes.Api+"mountains/", ApiScrapes(cfg.Routes.Api, cfg.Routes.Image))

	// add handlers for image folder
	mux.Handle(cfg.Routes.Image, http.StripPrefix(
		cfg.Routes.Image, http.FileServer(http.Dir(cfg.ImageRoot))))

	// add handler for root (static files)
	// use "StaticRoot" if set, fallback to embedded client
	if cfg.StaticRoot != "" {
		mux.Handle("/", addHSTS(http.FileServer(http.Dir(cfg.StaticRoot))))
	} else {
		mux.Handle("/", addHSTS(http.FileServer(client)))
	}

	return mux
}

// ApiData returns a HandlerFunc that responds to requests for the publicly
// accessible lump sum of mountains and cameras.
func ApiData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqstart := time.Now()
		status := http.StatusOK
		defer func() {
			took := time.Since(reqstart)
			log.Printf(log.Info, "%s %d %s %s (%s)",
				r.RemoteAddr, status, http.StatusText(status), r.RequestURI,
				took)
		}()

		// fetch all mountains from db
		mts, err := db.Mountains()
		// cams, err := db.Cameras()
		if err != nil {
			log.Printf(log.Error, "ApiData db error getting mts or cams: %s", err)
			status = http.StatusInternalServerError
			http.Error(w, "", status)
			return
		}

		// fetch cameras for each mountain from db
		// assigning them to the "Camera" field on model.Mountain.
		// this field is really only used for json encoding
		// groups := db.GroupCamerasByMountain(cams)
		for id, mt := range mts {
			// TODO: prefetch all cameras instead of going to
			// the db each time
			mt.Cameras, err = db.CamerasOnMountain(id)
			if err != nil {
				log.Printf(log.Error, "getting cameras for mtID(%d): %s", id, err)
			}
			mts[id] = mt
		}

		// encode mountains map to json and return
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		err = enc.Encode(mts)
		if err != nil {
			status = http.StatusInternalServerError
			http.Error(w, "", status)
			return
		}
		w.Header().Set(contenttype, jsonMime)
	}
}

// ApiScrapes returns a HandlerFunc to respond to requests for scrapes.
func ApiScrapes(apiRoute, imgRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqstart := time.Now()
		msg := ""
		status := http.StatusOK
		defer func() {
			took := time.Since(reqstart)
			log.Printf(log.Info, "%s %d %s %s (%s) %s",
				r.RemoteAddr, status, http.StatusText(status), r.RequestURI,
				took, msg)
		}()

		// process url path to extract params for this request
		// first group is mountainID, second is cameraID.
		//  return 404 for non-matching urls
		expression := apiRoute + `mountains/(\d+)/cams/(\d+)/scrapes`
		re, err := regexp.Compile(expression)
		if err != nil {
			log.Print(log.Error, "ApiScrapes regexp: %s", err)
			status = http.StatusInternalServerError
			http.Error(w, "", status)
			return
		}

		// get mt and cam ids from url
		matches := re.FindStringSubmatch(r.URL.Path)
		if matches == nil || len(matches) != 3 {
			status = http.StatusNotFound
			http.Error(w, "", status)
			return
		}

		// fetch the requested mt and cam from db
		mtID, camID := processIDs(matches)
		mt, err := db.Mountain(mtID)
		cam, err := db.Camera(camID)
		if err != nil {
			log.Printf(log.Error, "ApiScrapes db error getting mt or cam: %s", err)
			status = http.StatusInternalServerError
			http.Error(w, "", status)
			return
		}
		if mt.ID == 0 || cam.ID == 0 {
			// if a mt or cam isn't found in db, return 404
			status = http.StatusNotFound
			http.Error(w, "", status)
			return
		}

		// get and process url query times (start and/or end)
		start, end := processQuery(r.URL.Query(), mt.TzLocation)

		// fetch scrapes from db
		scrapes, err := db.Scrapes(camID, start.UTC(), end.UTC()) // UTC() required
		if err != nil {
			log.Printf(log.Error, "ApiScrapes db error getting scrapes: %s", err)
			status = http.StatusInternalServerError
			http.Error(w, "", status)
			return
		}

		// 1) process scrapes by replaces their filename with the complete
		// path to the image file
		// 2) change 'created' time to be in mountain's timezone
		tz, _ := time.LoadLocation(mt.TzLocation)
		for i := range scrapes {
			if scrapes[i].Result == model.Success {
				scrapes[i].Filename = path.Join(imgRoute, mt.Pathname, cam.Pathname, scrapes[i].Filename)
			} else {
				scrapes[i].Filename = ""
			}

			scrapes[i].Created = scrapes[i].Created.In(tz)
		}

		msg = fmt.Sprintf("%d scrapes for %s(%d) %s(%d) in (%s) to (%s)",
			len(scrapes), mt.Name, mt.ID, cam.Name, cam.ID,
			start.Format(datetzfmt), end.Format(datetzfmt))

		// encode scrapes array into json and return
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		err = enc.Encode(scrapes)
		if err != nil {
			log.Printf(log.Error, "ApiScrapes couldn't encode scrapes for mtID(%d), camID(%d): %s", mtID, camID, err)
			status = http.StatusInternalServerError
			http.Error(w, "", status)
			return
		}
		w.Header().Set(contenttype, jsonMime)
	}
}

// converts the matched url params to mt and cam ids.
func processIDs(matches []string) (mtID, camID int) {
	mtID, _ = strconv.Atoi(matches[1])
	camID, _ = strconv.Atoi(matches[2])
	return
}

// converts the start/end query params to useable time.Times.
func processQuery(query url.Values, tzname string) (start, end time.Time) {

	tz, _ := time.LoadLocation(tzname)
	start, _ = time.ParseInLocation(datefmt, query.Get("start"), tz)
	end, _ = time.ParseInLocation(datefmt, query.Get("end"), tz)
	// if err != nil {
	// 	log.Print(log.Error, err)
	// }

	const negDay = -24 * time.Hour
	const posDay = 24 * time.Hour

	switch {
	case start.IsZero() && end.IsZero():
		// user entered nothing.
		// so end at "now" and start from 24 hours earlier.
		end = time.Now().In(tz)
		start = end.Add(negDay)

	case start.IsZero() && !end.IsZero():
		// user entered only end.
		// so the "previous 24 hours" from
		// the last moment of the "end" day would start at the beginning
		// of "end", and end at end+24hrs.
		start = end
		end = end.Add(posDay)

	case !start.IsZero() && end.IsZero():
		// user entered only start.
		// start at the beginning of that day and end 24 hours later.
		end = start.Add(posDay)

	default:
		// user entered both start and end.
		// ensure end is the last moment of the day, so when start's
		// date is equal to end's date, the user gets the scrapes
		// from that day. AKA, make 'end' inclusive [start, end] instead
		// of [start, end)
		end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	return
}

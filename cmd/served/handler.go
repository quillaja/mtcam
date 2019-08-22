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

	"github.com/quillaja/mtcam/model"

	"github.com/quillaja/mtcam/db"

	"github.com/quillaja/mtcam/log"
)

// content type header
const contenttype = "Content-Type"

// json mime
const jsonMime = "application/json"

// time.Format() formats
const (
	datefmt     = "2006-01-02"
	datetimefmt = "2006-01-02 15:04"
	datetzfmt   = "2006-01-02 -0700"
)

// CreateHandler creates a ServeMux for the given serverd config.
func CreateHandler(cfg *ServerdConfig) http.Handler {
	mux := http.NewServeMux()

	// add handlers for API endpoints
	mux.HandleFunc(cfg.ApiRoute+"data/", ApiData())
	mux.HandleFunc(cfg.ApiRoute+"mountains/", ApiScrapes(cfg.ApiRoute, cfg.ImageRoute))

	// add handlers for image folder
	mux.Handle(cfg.ImageRoute, http.StripPrefix(
		cfg.ImageRoute, http.FileServer(http.Dir(cfg.ImageRoot))))

	// add handler for root (static files)
	mux.Handle("/", http.FileServer(http.Dir(cfg.StaticRoot)))

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
			log.Printf(log.Debug, "%s %d %s %s (%s)",
				r.RemoteAddr, status, http.StatusText(status), r.RequestURI,
				took)
		}()

		// fetch all mountains from db
		mts, err := db.Mountains()
		// cams, err := db.Cameras()
		if err != nil {
			log.Printf(log.Error, "ApiData db error getting mts or cams: %s", err)
			status = http.StatusInternalServerError
			w.WriteHeader(status)
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
			w.WriteHeader(status)
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
			log.Printf(log.Debug, "%s %d %s %s (%s) %s",
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
			w.WriteHeader(status)
			return
		}

		// get mt and cam ids from url
		matches := re.FindStringSubmatch(r.URL.Path)
		if matches == nil || len(matches) != 3 {
			status = http.StatusNotFound
			w.WriteHeader(status)
			return
		}

		// fetch the requested mt and cam from db
		mtID, camID := processIDs(matches)
		mt, err := db.Mountain(mtID)
		cam, err := db.Camera(camID)
		if err != nil {
			log.Printf(log.Error, "ApiScrapes db error getting mt or cam: %s", err)
			status = http.StatusInternalServerError
			w.WriteHeader(status)
			return
		}
		if mt.ID == 0 || cam.ID == 0 {
			// if a mt or cam isn't found in db, return 404
			status = http.StatusNotFound
			w.WriteHeader(status)
			return
		}

		// get and process url query times (start and/or end)
		start, end := processQuery(r.URL.Query(), mt.TzLocation)

		// fetch scrapes from db
		scrapes, err := db.Scrapes(camID, start.UTC(), end.UTC()) // UTC() required
		if err != nil {
			log.Printf(log.Error, "ApiScrapes db error getting scrapes: %s", err)
			status = http.StatusInternalServerError
			w.WriteHeader(status)
			return
		}

		// process scrapes by replaces their filename with the complete
		// path to the image file
		for i := range scrapes {
			if scrapes[i].Result == model.Success {
				scrapes[i].Filename = path.Join(imgRoute, mt.Pathname, cam.Pathname, scrapes[i].Filename)
			} else {
				scrapes[i].Filename = ""
			}
		}
		// log.Printf(log.Debug, "%s requested %d scrapes for %s(%d) %s(%d) in period (%s) to (%s)",
		// 	r.RemoteAddr,
		// 	len(scrapes), mt.Name, mt.ID, cam.Name, cam.ID,
		// 	start.Format(time.RFC3339), end.Format(time.RFC3339))
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
			w.WriteHeader(status)
			return
		}
		w.Header().Set(contenttype, jsonMime)
	}
}

func processIDs(matches []string) (mtID, camID int) {
	mtID, _ = strconv.Atoi(matches[1])
	camID, _ = strconv.Atoi(matches[2])
	return
}

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
		end = time.Now().In(tz)
		start = end.Add(negDay)

	case start.IsZero() && !end.IsZero():
		start = end.Add(negDay)

	case !start.IsZero() && end.IsZero():
		end = start.Add(posDay)
	}

	return
}

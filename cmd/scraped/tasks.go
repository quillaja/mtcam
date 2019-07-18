package main

import (
	"bytes"
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lucasb-eyer/go-colorful"

	"github.com/pkg/errors"

	"github.com/quillaja/mtcam/model"

	"github.com/disintegration/imaging"

	"github.com/quillaja/mtcam/astro"
	"github.com/quillaja/mtcam/db"
	"github.com/quillaja/mtcam/log"
)

// user agent header
const useragent = "User-Agent"

// Scrape returns a function to be used in a Task. The returned function
// will attempt to scrape the cam with camID using the time passed to it.
//
// This process will take perhaps 10-30 seconds depending on the network and
// camera configuration. It involves reading from the database multiple times,
// downloading an image, resizing and comparing this image, saving the image
// to disk, and ultimately adding a scrape record to the database.
//
// In the event of errors, generally the task is abandoned but a detailed
// error is logged and, if it makes sense, a "failure" scrape is recorded
// in the database with a note about the failure. This note also appears in the
// error log.
func Scrape(mtID, camID int, cfg *ScrapedConfig) func(time.Time) {
	// TODO: this is kinda a shitshow (is it?) and could use refactoring

	return func(now time.Time) {

		// create new scrape record
		scrape := model.Scrape{
			CameraID: camID,
			Created:  now,
			Result:   model.Failure,
		}
		// defer to end to always make an attempt at writing a scrape
		// record even when failing during scrape
		defer func() {
			err := db.InsertScrape(&scrape)
			if err != nil {
				err = errors.Wrapf(err, "(mtID=%d camID=%d) failed to insert scrape into db", mtID, camID)
				log.Print(log.Critical, err)
				// can't do anything now... total failure
			}
		}()

		// func to simplify life
		var err error // used throughout Scrape()
		setDetailAndLog := func(detail string) {
			scrape.Detail = detail
			msg := fmt.Sprintf("(mtID=%d camID=%d) %s", mtID, camID, detail)
			if err != nil {
				err = errors.Wrapf(err, msg)
			} else {
				err = errors.New(msg)
			}
			log.Print(log.Error, err)
		}

		// read mt and cam
		mt, err := db.Mountain(mtID)
		cam, err := db.Camera(camID)
		if err != nil {
			setDetailAndLog("could't read db")
			return
		}

		// wait cam delay
		time.Sleep(time.Duration(cam.Delay) * time.Second)

		// process the url template
		tz, err := time.LoadLocation(mt.TzLocation)
		data := UrlData{
			Camera:   cam,
			Mountain: mt,
			Now:      now.In(tz)} // send the url template the local time
		url, err := cam.ExecuteUrl(data)
		if err != nil {
			setDetailAndLog("couldn't execute url template")
			return
		}

		// perform the scrape
		// setting a custom timeout and useragent
		client := http.Client{Timeout: time.Duration(cfg.RequestTimeoutSec) * time.Second}
		request, err := http.NewRequest(http.MethodGet, url, nil)
		request.Header.Set(useragent, cfg.UserAgent)
		resp, err := client.Do(request)
		if resp.StatusCode != http.StatusOK {
			err = errors.Wrapf(err, "status code %d %s", resp.StatusCode, resp.Status)
		}
		if err != nil {
			setDetailAndLog("trouble downloading image")
			return
		}
		defer resp.Body.Close()

		// extract the image
		img, err := imaging.Decode(resp.Body)
		if err != nil {
			setDetailAndLog("couldn't decode downloaded image")
			return
		}
		// resize the image
		img = imaging.Resize(img, cfg.ImageWidth, 0, imaging.Lanczos)

		// build (and create if necessary) the directory where the scraped
		// images will live
		camImgDir := filepath.Join(cfg.ImageRoot, mt.AsPathname(), cam.AsPathname())
		// check if directory exists and create if not
		err = os.MkdirAll(camImgDir, 0755)
		if err != nil {
			// NOTE: this is once case where it would not be appropriate to
			// save a (failed) scrape to the database
			err = errors.Wrapf(err, "couldn't make path %s", camImgDir)
			log.Print(log.Critical, err)
			return
		}

		if cfg.ImageEqualityTesting {
			// a function to "encapsulate" getting the previously scraped image
			getPreviousImage := func() image.Image {
				// fetch previously (successfully) scraped image
				prevScrape, err := db.MostRecentScrape(camID, model.Success)
				if err != nil {
					err = errors.Wrapf(err, "(mtID=%d camID=%d) couldn't get previous scrape from db", mtID, camID)
					log.Print(log.Error, err)
					return nil
				}
				prevImgPath := filepath.Join(camImgDir, prevScrape.Filename)
				prevImg, err := imaging.Open(prevImgPath)
				if err != nil {
					err = errors.Wrapf(err, "(mtID=%d camID=%d) couldn't open previous image", mtID, camID)
					log.Print(log.Error, err)
					return nil
				}

				return prevImg
			}

			// a function to "encapsulate" processing of the downloaded image
			// so it can be tested against previously scraped image
			getTestImage := func() image.Image {
				// NOTE: !important! the JPEG quality setting alters the scraped images
				// beyond resizing, which prevents simple equality testing from working.
				//
				// Fix: encode to a memory buffer and decode back to image.Image. This will
				// perform the same processing on the freshly downloaded image as was
				// previously preformed on the prior images.
				//
				// Question: given the same input, will jpeg compression produce
				// identical output?? Minor testing shows same-in-same-out.
				buf := new(bytes.Buffer)
				err = imaging.Encode(buf, img, imaging.JPEG, imaging.JPEGQuality(cfg.ImageQuality))
				testimg, err := imaging.Decode(buf)
				if err != nil {
					err = errors.Wrapf(err, "(mtID=%d camID=%d) mem encode/decode of downloaded img", mtID, camID)
					log.Print(log.Error, err)
					return nil
				}

				return testimg
			}

			prev, cur := getPreviousImage(), getTestImage()
			// actually test for equality
			if equal(prev, cur, cfg.ImageEqualityTolerance) {
				setDetailAndLog("image identical to previously scraped image")
				return
			}
		}

		// save image to disk
		// filename is sec since unix epoc in UTC
		scrape.Filename = strings.ToLower(fmt.Sprintf("%d.%s", now.UTC().Unix(), cam.FileExtension))
		imgPath := filepath.Join(camImgDir, scrape.Filename)
		err = imaging.Save(img, imgPath, imaging.JPEGQuality(cfg.ImageQuality))
		if err != nil {
			setDetailAndLog("couldn't save image " + scrape.Filename + " to disk")
			return
		}
		log.Printf(log.Info, "wrote %s", imgPath)

		// if we make it this far, everything was ok
		scrape.Result = model.Success
		scrape.Detail = ""

	}
}

func ScheduleScrapes(mID int, app *Application) func(time.Time) {

	return func(now time.Time) {

		// read mt and cams
		mt, err := db.Mountain(mID)
		cams, err := db.CamerasOnMountain(mID)
		if err != nil {
			log.Print(log.Error, err)
			return // can't continue if can't read DB
		}

		// get tz info for mt
		tz, err := time.LoadLocation(mt.TzLocation)
		if err != nil {
			log.Print(log.Error, err)
			return // can't continue if can't get tz
		}
		now = now.In(tz) // convert time to correct tz
		log.Printf(log.Debug, "processing mountain %s at %s", mt.Name, now.Format(time.RFC3339))

		// get astro data for mt
		const maxTries = 5
		var tries int
		var sun astro.Data
		for ; tries < maxTries; tries++ {
			sun, err = astro.Get(mt.Latitude, mt.Longitude, now)
			if err == nil {
				log.Printf(log.Info, "got astro data after %d tries", tries+1)
				break
			}
			// log.Print(log.Error, err)
			time.Sleep(1 * time.Second)
		}
		if tries >= maxTries {
			log.Print(log.Error, "too many tries")
			return
		}

		// for each cam
		// byMt := db.GroupCamerasByMountain(cams)
		for _, cam := range cams { //byMt[mt.ID] {
			log.Printf(log.Debug, " processing camera %s", cam.Name)
			// skip inactive cams
			if !cam.IsActive {
				log.Printf(log.Info, " skipping inactive cam %s", cam.Name)
				continue
			}
			// round current time to nearest cam interval
			// for each time+interval until end-of-day
			interval := time.Duration(cam.Interval) * time.Minute
			stop := startOfNextDay(now)
			for t := roundup(now, interval); t.Before(stop); t = t.Add(interval) {
				// determine if the cam should be scraped
				if sun.SunTransit[astro.StartCivilTwilight].Before(t) && t.Before(sun.SunTransit[astro.EndCivilTwilight]) {
					// schedule a scrape
					log.Printf(log.Info, "   scrape scheduled for %s at %s", cam.Name, t.Format(time.RFC3339))
				}
			}
		}

		// schedule ScheduleScrapes() for next day
		next := startOfNextDay(now)
		log.Printf(log.Info, "ScheduleScrapes(%s) scheduled at %s", mt.Name, next.Format(time.RFC3339))
	}
}

// roundup rounds t up to the nearest d. Works best for d <=60m and in
// divisors of 60 (60/2=30m, 60/3=20m, ...)
func roundup(t time.Time, d time.Duration) time.Time {
	rounded := t.Round(d)
	if rounded.Before(t) {
		rounded = rounded.Add(d)
	}
	return rounded
}

// startOfNextDay returns the day after t at 0:00:00.
func startOfNextDay(t time.Time) time.Time {
	next := t.Add(24 * time.Hour)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
}

// equal determines if 2 images are (about) the same
func equal(a, b image.Image, tolerance float64) bool {
	if a == nil || b == nil {
		log.Print(log.Debug, "a or b nil")
		return false
	}
	if a.Bounds() != b.Bounds() {
		log.Print(log.Debug, "a and b have different sizes")
		return false
	}

	for y := 0; y < a.Bounds().Dy(); y++ {
		for x := 0; x < a.Bounds().Dx(); x++ {
			ca, cb := a.At(x, y), b.At(x, y)
			la, _ := colorful.MakeColor(ca) // images don't have alpha, so no
			lb, _ := colorful.MakeColor(cb) // worry about 0 in alpha channel
			dE := la.DistanceLab(lb)
			if dE > tolerance {
				log.Printf(log.Debug, "(%d, %d) %v != %v dE=%f", x, y, ca, cb, dE)
				return false
			}
		}
	}

	return true
}

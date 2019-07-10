# mtcam
Enhanced version of "palmer" which is database backed and can scrape multiple webcams.
This new version is in Go.

# Notes

## Ideas

#### scraping
- scraping done via a scheduler 'daemon' (instead of using cron)
    - upon startup, tasks are enqueued for the scheduler
        - tasks: scrape, update ancillary data, enqueue next set of tasks, etc.
        - need function that figures out what tasks to enqueue for each mt/cam
- scraping is performed as scheduled, and using a set of 'rules' which are
    'scripted' via a go text template evaluating to True or False.
- url to scrape is generated via go text template.
    - allows more 'dynamic' urls, such as ones containing a date/time
    - many urls will still just be static
- get sunrise/set data from us navy api
    - https://aa.usno.navy.mil/data/docs/api.php
    - also has moon phase data

#### config
- program config in a file (ie config.json)
    - config system 'watches' file for updates and reloads settings live

#### post-scrape
- image processing:
    - resize image (save disk space for large images)
    - (?) check if scraped image is identical to previously scraped (resized)
        image, and discard if so (to avoid duplicates/"frozen" cams)

#### binaries
- scrape daemon
- frontend server
- cmd line tool to manipulate (CRUD) mountains and cams

#### front end
- frontend changes
    - remove time from selection; dates only, midnight-midnight
    - show mt/cam info before fetching photos
        - add 'statistics' after fetch
        - changes when mt/cam selection changes

#### other
- remove
    - weather? kinda sucks

### internal packages
- astro - gets sun/moon data from navy api
    - `func Get(lat, lon double, date time.Time, apikey string) (data SOMESTRUCT)`
    - various constants for phemonenon
- db
- model
- scheduler
- googletz - get tz location id (eg "America/Los_Angeles") for lat/lon
- log - provides simple logging to systemd via stdout

## Dependencies
1. github.com/mattn/go-sqlite3 - for sqlite
1. github.com/disintegration/imaging - for image resizing
1. github.com/gorilla/mux - easier handling of api routes
1. ~~http://github.com/sirupsen/logrus - might have to make my own formatter for systemd~~
1. github.com/shibukawa/configdir - don't really need if i assume linux (can just use os.GetEnv())

# API
Generally not changed from python version

    https://<whatever address>/ -> /
        homepage and static root

    /api
        root of api. returns nothing.
    /api/data
        GET: returns json dict<id,obj> of mountains containing dict<id,obj> of cams
    /api/mountains/<mt_id>/cams/<cam_id>/scrapes[?start=<datetime>&end=<datetime>]
        GET: returns json list of scrape records
    /api/mountains/<mt_id>/weather[?start=<datetime>&end=<datetime>&format=<'json'|'bokeh'>]
        GET: returns weather data for the time period, 
        either as json list or a bunch of html for a bokeh graph.


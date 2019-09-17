# Notes

## TODO

Sooner
- [ ] fallback to some 'default' astro or start/end time if Sun/Moon served is inaccessible
- [ ] robots.txt ?
- [ ] import older "palmer" data (note logs are in Y-D-M format, not Y-M-D)

Later
- [ ] documentation on everything
- [ ] rewrite client.js
- [ ] tasks to update mountain timezones
- [ ] config file watch for changes
- [ ] cache data served from `/api/data` to improve performance/reduce queries. goroutine to refresh data periodically.
- [ ] web app manifest https://developers.google.com/web/fundamentals/web-app-manifest/

Done
- [x] resize image on height, not width (can do either or both. 0='auto' to maintain aspect ratio)
- [X] resize image only if bigger than height/width
- [x] simulate 'click' on info tab when hiding other tabs after user selects a camera
- [x] fix 1-off error in timelapse counter (thought i did before?)
- [x] limit timespan of request to 2 weeks or do something else sensible to prevent huge 58000 scrape requests
- [x] css styles for "flashing" tabs when they're first unhidden
- [x] server to use StaticRoot if available, fallback to embedded
- [x] HTTP strict transport? see: https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Strict_Transport_Security_Cheat_Sheet.html
- [x] block (404) any HTTP requests not for specfic domains (hosts). (and also anything that isn't for root ("/"))
- [x] "end" scrape param still needs to include that day for all cases.
- [x] shutdown scraped on signal (requires change to scheduler to allow tasks to complete)

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
- image processing:
    - resize image (save disk space for large images)
    - check if scraped image is identical (or nearly so) to previously scraped (resized)
        image, and discard if so (to avoid duplicates/"frozen" cams)
- get sunrise/set data from us navy api
    - https://aa.usno.navy.mil/data/docs/api.php
    - also has moon phase data

#### config
- program config in a file (ie config.json)
    - config system 'watches' file for updates and reloads settings live

#### binaries
- scrape daemon
- frontend server
- ~~cmd line tool to manipulate (CRUD) mountains and cams~~

#### front end
- frontend changes
    - remove time from selection; dates only, midnight-midnight
    - show mt/cam info before fetching photos
        - add 'statistics' after fetch
        - changes when mt/cam selection changes
    - show 'log' time in mountain's local time

#### other
- remove
    - weather? kinda sucks

### internal packages
- astro - gets sun/moon data from navy api
    - various constants for phemonenon
- db - database connection and queries
- model - data structs
- scheduler - executes tasks at pre-scheduled times
- googletz - get tz location id (eg "America/Los_Angeles") for lat/lon
- log - provides simple logging to systemd via stdout
- config - suite wide config structure and helper functions for config file watching

## Dependencies
1. github.com/mattn/go-sqlite3 - for sqlite
1. github.com/disintegration/imaging - for image resizing
1. ~~github.com/gorilla/mux - easier handling of api routes~~
1. ~~http://github.com/sirupsen/logrus - might have to make my own formatter for systemd~~
1. ~~github.com/shibukawa/configdir - don't really need if i assume linux (can just use os.GetEnv())~~

## Directories
- binaries and cfg in /opt/mtcam
- images and db in ~/mtcam

or 

- all files (bin, cfg, db) in /opt/mtcam
- images in /opt/mtcam/img

or

- binaries in /opt/mtcam
- config files in /etc/mtcam
- database in /var/opt/mtcam
- images in /var/opt/mtcam/img

# Migration
1. stop old mtcam scraper (on pi)
    1. remove 'idle' scrapes from old db
2. `$ sqlite3 new.db` create new database file
3. `.read new_table.sql` create the new tables
4. `.read migration.sql`  will pull in old.db, set new/updated fields on old data
5. `go run cmd/convert_tz/main.go new.db`  converts all times in db from PST to UTC
6. move images from pi to nuc (~18GB)

# API
Generally not changed from python version

    https://<whatever address>/ -> /
        homepage and static root

    /img/
        root of image folders

    /api/
        root of api. returns nothing.
    /api/data/
        GET: returns json dict<id,obj> of mountains containing dict<id,obj> of cams
    /api/mountains/<mt_id>/cams/<cam_id>/scrapes[?start=<datetime>&end=<datetime>]
        GET: returns json list of scrape records

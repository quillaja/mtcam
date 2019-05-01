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


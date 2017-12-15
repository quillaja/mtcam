# mtcam
Enhanced version of "palmer" which is database backed and can scrape multiple webcams.

# Notes
lat, lon must be a str for ephem
elevation must be in meters for ephem (util.ft_m())
ephem times are always UTC

Can store google timezone info (in Mountain.tz) as string using JSON
encoding/decoding from/to dict.

To check if scrape should happen use this (simplified):
`rise_time <= ephem.now() <= set_time`
 - rise_time and set_time are ephem.Date() type
 - use util.floor() on each part of comparison to simulate floor()/ceiling() for dates

For the sake of simplicity, update timezone info at 3AM Pacific Time.
This is because most (all) of the webcams will be in the PT zone, and
updating at this time should prevent most DST oddities, compared to using UTC.
Mounains/cams in other timezones will just have potential time errors.

cron jobs:
1. scrape every 5 mins (in scrape.py)
    1. Done - set delay for cams?? --10 sec seems good in early testing
2. update Mountain.tz daily at 3AM PT (in tz_update.py)
3. get weather data every hour (TODO)

should program configuration also be saved in database?

path to save each scraped image is `/archiveroot/mountain/cam/timestamp.jpg`

at each scrape:
1. load model data from database (see http://docs.peewee-orm.com/en/latest/peewee/querying.html#using-aggregate-rows)
2. get utc timestamp (use same timestamp for each scrape)
3. for each mountain:
    1. create observer (lat, lon, elev, local-noon as utc, horizon)
    2. get sunrise and sunset times
    3. if between sunrise and sunset, then for each cam
        1. download cam photo from web
        2. write new ScrapeRecord with appropriate info to database

# Todo
1. DONE - write API
    1. queries.py
    2. views.py (using flask)
2. DONE - Change JSON returned from API to be dict<id,obj> instead of just a list.
2. clean up util.py
2. DONE - decide (in db or file) and create program settings
3. DONE - decide on timestamp or datetime for 'modified' ModelBase field
4. DONE - Write script to update Mountain.tz_json field
5. write script to convert/import original palmer data to mtcam data
6. DONE - Create "client" (webpage/javascript/css)
    1. html (index.html)
    2. javascript (client.js)
        1. timelapse.js or gallery.js for the "moving picture" display?
    3. style (style.css)
7. DONE - Weather
    1. DONE - write scraper to get weather data for each (US) mountain each hour
        1. data: min temp, max temp, temp, wind spd, wind gust, wind direction, amount rain, amount snow
        2. DONE (yes, get all) - also get this data? probabilty of precip, sky cover (%), ...
        3. DONE - convert values to float, datetime; convert to correct units (knots->mph)
        4. do NOT invert wind direction--this will be done only when a bokeh plot is requested.
    2. DONE - update model.py
        1. DONE - class(es) for weather data, FK to Mountain
        2. DONE - add flag to Mountain indicating if its weather should/n't be queried?
            1. Boolean flag for older xml API, but for new API just use the existence of a request url to signal that data should be gotten.
    3. DONE - update api.py
    4. DONE - update client.js and index.html
    5. DONE - Oh hey, NOAA has new interface that returns JSON... good i guess. Now I have to re-write the whole thing.
        1. The data I can get hasn't changed. (good)
        2. easier to parse--just use python's great json support. (good.. don't need beautifulsoup and lxml)
        3. all data is in SI (metric) units.(?) (bad)
            1. have to convert
            2. is there a way to request imperial/english units? -- no documentation of this
        4. Documentation is pretty crappy. (bad)
        5. Basic pattern: `<data-dict>['properties']['<parameter>']['values'][0]['value']`

# API
Was previously in `api.md`.

    mtcam.quillaja.net/ -> /

    /api
        root of api. returns nothing.
    /api/data
        GET: returns json dict<id,obj> of mountains containing dict<id,obj> of cams
    /api/mountains/<mt_id>/cams/<cam_id>/scrapes[?start=<datetime>&end=<datetime>]
        GET: returns json list of scrape records
    /api/mountains/<mt_id>/weather[?start=<datetime>&end=<datetime>&format=<'json'|'bokeh'>]
        GET: returns weather data for the time period, 
        either as json list or a bunch of html for a bokeh graph.


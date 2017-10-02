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
1. scrape every 5 mins
    1. set delay for cams??
2. update Mountain.tz daily at 3AM PT

should program configuration can also be saved in database?

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
1. write API
    1. queries.py
    2. views.py (using flask)
2. clean up util.py
2. decide (in db or file) and create program settings
3. decide on timestamp or datetime for 'modified' ModelBase field
4. Write script to update Mountain.tz_json field
5. write script to convert/import original palmer data to mtcam data
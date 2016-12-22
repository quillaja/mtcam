# mtcam
Enhanced version of "palmer" which is database backed and can scrape multiple webcams.

# Notes
lat, lon must be a str for ephem
elevation must be in meters for ephem
ephem times are always UTC

Can store google timezone info as string using str() and
restore it to a dict using eval()

To check if scrape should happen use this (simplified):
rise_time <= ephem.Date(datetime.utcnow()) <= set_time
 use util.strip_to_datehour() on each part of comparison to
 simulate floor()/ceiling() for dates

For the sake of simplicity, update timezone info at 3AM Pacific Time.
This is because most (all) of the webcams will be in the PT zone, and
updating at this time should prevent most DST oddities, compared to using UTC.
Cams in other timezones will just have potential time errors.
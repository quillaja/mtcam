import json
import time
from datetime import datetime, timedelta, tzinfo

import requests
import ephem

url = 'https://maps.googleapis.com/maps/api/timezone/json?location={},{}&timestamp={}&key=AIzaSyC7B3hbUJ0m-2F-pOp0W6IIirO4nOUwWrU'
fmt = '%Y-%m-%dT%H:%M:%S.%f%z'

pdx = (45.31, -121.84)
london = (51.5, -0.05)
fuji = (35.360730, 138.727359)
everest = (27.988056, 86.925278)

##local->UTM: local_time - (offset + dstoffset)
##UTM->local: UTM_time + (offset + dstoffset)

## datetime.now().replace(tzinfo=pst).astimezone(timezone.utc)  'pst' from get_tzinfo()


def strip_to_datehour(dt):
    if isinstance(dt, ephem.Date):
        tt = dt.tuple()
        return datetime(tt[0], tt[1], tt[2], tt[3])
    elif isinstance(dt, datetime):
        return datetime.replace(minute=0, second=0, microsecond=0)
    else:
        raise ValueError('argument must be datetime.datetime or ephem.Date')


def get_tz(lat, lon, t=None):
    """
    latitude and longitude be on land. the "ocean" apparently has no timezeone.
    t is seconds since epoch in UTM (defaults to time.time() )
    """
    if not t:
        t = int(time.time())
    #use url to get json
    result = requests.get(url.format(lat, lon, t))
    #return converted json
    return json.loads(result.content.decode('utf-8'))


def utmnoon_at_loc(lat, lon):
    """
    Returns the UTM time it'll be noon (12:00) at the given coordinates.
    """
    tz = get_tz(lat, lon)
    if (tz['status'] != 'OK'):
        return None

    total_offset_s = tz['rawOffset'] + tz['dstOffset']
    noon = datetime.utcnow().replace(
        hour=12, minute=0, second=0, microsecond=0) - timedelta(
            seconds=total_offset_s)

    return noon


def get_tzinfo(lat, lon, t=None):
    """
    Returns a datetime.tzinfo object for the given coordinates.
    Relies on Python 3's datetime.timezone() function.
    """
    tz = get_tz(lat, lon, t)
    return timezone(
        timedelta(seconds=(tz['rawOffset'] + tz['dstOffset'])),
        tz['timeZoneName'])

# REDUNDANT: use datetime.utcnow()
# def utm_now():
#     """
#     Returns a datetime object for the present time in UTC.
#     """
#     return datetime.now() + timedelta(seconds=time.timezone)


def now_at_loc(lat, lon):
    """
    Returns the present time at the given coordinates.
    """
    tz = get_tz(lat, lon)
    return datetime.utcnow() + timedelta(seconds=(tz['rawOffset'] + tz['dstOffset']))


class Gtz(tzinfo):
    """
    Creates a tzinfo object for the given coordinates 
    for use with python's datetime utilities
    """
    def __init__(self, latitude, longitude):
        self.latitude = latitude
        self.longitude = longitude
        self.tz = get_tz(latitude, longitude)
        self.offset = timedelta(seconds=(self.tz['rawOffset'] + self.tz['dstOffset']))
        self.dst_offset = timedelta(seconds=self.tz['dstOffset'])

    def utcoffset(self, dt):
        return self.offset

    def dst(self, dt):
        return self.dst_offset

    def tzname(self, dt):
        return self.tz['timeZoneName']
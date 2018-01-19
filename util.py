# """
# Requires Python 3 for certain timezone objects/functions.
# """

import json
import time
from datetime import datetime, timedelta, timezone, tzinfo

import ephem
import requests
from bokeh.embed import components
from bokeh.plotting import figure
from bokeh.layouts import column
from bokeh.models import Range1d, LinearAxis

import settings

url = 'https://maps.googleapis.com/maps/api/timezone/json?location={},{}&timestamp={}&key={}'
fmt = '%Y-%m-%dT%H:%M:%S.%f%z'

##local->UTM: local_time - (offset + dstoffset)
##UTM->local: UTM_time + (offset + dstoffset)

## Converion functions --------------------------------------------------


def identity(value):
    '''returns the value unchanged'''
    return value


def c_to_f(temp_c):
    '''converts deg C to deg F'''
    return temp_c * 1.8 + 32


def ms_to_mph(speed_ms):
    '''converts meters per second to miles per hour'''
    return speed_ms * 2.23694


def mm_to_in(mm):
    '''converts mm to inch'''
    return mm * 0.0393701


def m_to_ft(meters):
    '''converts meters to feet'''
    return meters * 3.28084


def ft_to_m(feet):
    '''Converts feet to meters.'''
    return feet * 0.3048


def floor(dt):
    '''Replaces minute, second, and microsecond parts of datetime or ephem.Date
    with zeros, essentially 'flooring' it to the hour.'''
    if isinstance(dt, ephem.Date):
        tt = dt.tuple()
        return datetime(tt[0], tt[1], tt[2], tt[3])
    elif isinstance(dt, datetime):
        return datetime.replace(minute=0, second=0, microsecond=0)
    else:
        raise ValueError('argument must be datetime.datetime or ephem.Date')


## Google TZ Api call -----------------------------------------------------


def get_tz(lat, lon, t=None):
    """
    latitude and longitude must be on land. the "ocean" 
    apparently has no timezeone.
    t is seconds since epoch in UTM (defaults to time.time() )
    """
    if not t:
        t = int(time.time())
    #use url to get json
    result = requests.get(url.format(lat, lon, t, settings.GOOGLE_TZAPI_KEY))
    #return converted json
    return result.json()


## Used in api.py for various things -------------------------------------


def to_sys_time(t: datetime, mttz: dict) -> datetime:
    '''Uses mttz and system tz info to recalculate `t` into system time.'''

    if 'rawOffset' not in mttz or 'dstOffset' not in mttz:
        return t

    return (t - timedelta(seconds=(mttz['rawOffset'] + mttz['dstOffset']))) - (
        timedelta(seconds=time.timezone) -
        timedelta(hours=time.localtime().tm_isdst))


def convert_input(time_input: str,
                  as_local_time: bool = False,
                  mttz: dict = None) -> datetime:
    '''
    Parses a time string. If `as_local_time` is True, it will use the
    timezone info provided in `mttz` and the system's timezone info to
    convert `time_input` from the local time (at the mountain) to the
    system's time.
    '''
    try:
        time_input = datetime.strptime(time_input, '%Y-%m-%dT%H:%M')
        if as_local_time and (mttz is not None):
            time_input = to_sys_time(time_input, mttz)
    except ValueError:
        time_input = None

    return time_input


def weather_to_json(forecasts):
    '''
    Make list of WeatherForecast into JSON list of dicts.
    '''

    data = list()

    for f in forecasts:
        f_dict = {
            'date': f.created.strftime('%Y-%m-%d %H:%M'),
            'temp': f.temp,
            'temp_min': f.temp_min,
            'temp_max': f.temp_max,
            'wind_spd': f.wind_spd,
            'wind_gust': f.wind_gust,
            'wind_dir': f.wind_dir,
            'prob_precip': f.prob_precip,
            'cloud': f.cloud,
            'snow_level': f.snow_level,
            'rain': f.rain,
            'snow': f.snow
        }

        data.append(f_dict)

    return json.dumps(data, indent=2, sort_keys=True)


def weather_to_bokeh(forecasts, mt_name):
    '''makes nice plot from the forecast data'''

    if len(forecasts) == 0:
        return json.dumps({'script': '', 'div': '<div>No data.</div>'})

    xend = forecasts[-1].created + timedelta(hours=1)
    xstart = forecasts[-1].created - timedelta(days=1)

    # used for all x axes
    created = [w.created for w in forecasts]

    top = figure(
        title=mt_name + ' Summit - Forecast Weather Each Hour',
        x_range=(xstart, xend),
        x_axis_label='time',
        x_axis_type='datetime',
        y_axis_label='unit',
        tools='xpan,wheel_zoom,box_zoom,reset,save',
        width=640,
        height=400)

    # temperature related lines
    top.line(
        created, [w.temp for w in forecasts],
        line_width=2,
        color='black',
        legend='Temp (F)')
    top.line(
        created, [w.temp_max for w in forecasts],
        legend='Max Temp',
        line_dash='4 4',
        color='red',
        line_width=0.5)
    top.line(
        created, [w.temp_min for w in forecasts],
        legend='Min Temp',
        line_dash='4 4',
        color='blue',
        line_width=0.5)

    # wind related lines
    # for some stupid reason bokeh draws angles counter clockwise, so
    # i have to negate all the wind_dir entries to make it display right.
    #
    wind_dir_inverted = [
        -(w.wind_dir if w.wind_dir is not None else 0) for w in forecasts
    ]
    wind_spd = [w.wind_spd for w in forecasts]
    top.line(
        created,
        wind_spd,
        line_width=2,
        color='lightgreen',
        legend='Wind (mph)')
    top.vbar(  # looks better than circle for wind gusts
        x=created,
        bottom=wind_spd,
        color='lightgreen',
        width=0.5,
        top=[w.wind_gust
             for w in forecasts])  # not pretty for missing values (drawn as 0)
    top.inverted_triangle(  # use inverted_triangle to make display 'map view'
        created,
        wind_spd,
        angle=wind_dir_inverted,
        angle_units='deg',
        color='green',
        size=8)
    top.rect(
        created,
        wind_spd,
        angle=wind_dir_inverted,
        angle_units='deg',
        color='green',
        width=0.5,
        height=16,
        height_units='screen')

    # legend setting has to be here to actually take effect
    top.legend.location = 'top_left'
    top.legend.orientation = 'horizontal'
    top.legend.background_fill_alpha = 0.5
    top.legend.click_policy = 'hide'

    # grid settings
    top.xaxis.minor_tick_line_color = 'black'
    top.xaxis.minor_tick_line_width = 1
    top.xgrid.minor_grid_line_color = 'gray'
    top.xgrid.minor_grid_line_alpha = 0.2
    top.ygrid.minor_grid_line_color = 'gray'
    top.ygrid.minor_grid_line_alpha = 0.2

    bottom = figure(
        title=mt_name + ' Summit - Forecast Weather Each Hour',
        x_axis_label='time',
        x_axis_type='datetime',
        x_range=top.x_range,
        y_axis_label='unit',
        y_range=(0, 24),
        tools='xpan,wheel_zoom,box_zoom,reset,save',
        width=640,
        height=400)

    # precipitation related graphs
    shifted_time = created[1:] + [created[-1] + timedelta(hours=1)
                                  ] if len(forecasts) > 0 else []
    bottom.quad(
        legend='Rain (in)',
        left=created,
        right=shifted_time,
        top=[w.rain for w in forecasts],
        bottom=0,
        color='blue',
        alpha=0.2)
    bottom.quad(
        legend='Snow (in)',
        left=created,
        right=shifted_time,
        top=[w.snow for w in forecasts],
        bottom=0,
        color='red',
        alpha=0.2)

    # probability of precipitation and cloud cover, on 2nd axis
    bottom.extra_y_ranges = {'prob': Range1d(start=0, end=105)}
    bottom.line(
        created, [w.prob_precip for w in forecasts],
        legend='Prob.Precip (%)',
        color='blue',
        line_width=1,
        y_range_name='prob')
    bottom.line(
        created, [w.cloud for w in forecasts],
        legend='Cloud Cover (%)',
        color='black',
        line_width=1,
        y_range_name='prob')

    # legend setting has to be here to actually take effect
    bottom.legend.location = 'top_left'
    bottom.legend.orientation = 'horizontal'
    bottom.legend.background_fill_alpha = 0.5
    bottom.legend.click_policy = 'hide'

    # grid settings
    bottom.add_layout(
        LinearAxis(y_range_name='prob', axis_label='percent'), 'left')

    bottom.xaxis.minor_tick_line_color = 'black'
    bottom.xaxis.minor_tick_line_width = 1
    bottom.xgrid.minor_grid_line_color = 'gray'
    bottom.xgrid.minor_grid_line_alpha = 0.2
    bottom.ygrid.minor_grid_line_color = 'gray'
    bottom.ygrid.minor_grid_line_alpha = 0.2

    script, div = components(column(top, bottom))

    # hackish, but I had to remove the <script> tags here in order to 'inject'
    # the javascript into the DOM on the client side. Could also consider
    # using regex or something to do this, but simply slicing was easy
    # and convenient, and probably faster.
    script = script[32:-9]

    return json.dumps({'script': script, 'div': div})


# Unused? ----------------------------------------------------------------


def utmnoon_at_loc(lat, lon, tz=None):
    """
    Returns the UTM time it'll be noon (12:00) at the given coordinates.
    """
    if not tz:
        tz = get_tz(lat, lon)
    if (tz['status'] != 'OK'):
        return None

    total_offset_s = tz['rawOffset'] + tz['dstOffset']
    noon = datetime.utcnow().replace(
        hour=12, minute=0, second=0,
        microsecond=0) - timedelta(seconds=total_offset_s)

    return noon


def get_tzinfo(lat, lon, tz=None, t=None):
    """
    Returns a datetime.tzinfo object for the given coordinates.
    Relies on Python 3's datetime.timezone() function.
    """
    if not tz:
        tz = get_tz(lat, lon, t)

    return timezone(
        timedelta(seconds=(tz['rawOffset'] + tz['dstOffset'])),
        tz['timeZoneName'])


def now_at_loc(lat, lon, tz=None):
    """
    Returns the present local time at the given coordinates.
    """
    if not tz:
        tz = get_tz(lat, lon)

    return datetime.utcnow() + timedelta(
        seconds=(tz['rawOffset'] + tz['dstOffset']))


class Gtz(tzinfo):
    """
    Creates a tzinfo object for the given coordinates 
    for use with python's datetime utilities
    """

    def __init__(self, latitude, longitude):
        self.latitude = latitude
        self.longitude = longitude
        self.tz = get_tz(latitude, longitude)
        self.offset = timedelta(
            seconds=(self.tz['rawOffset'] + self.tz['dstOffset']))
        self.dst_offset = timedelta(seconds=self.tz['dstOffset'])

    def utcoffset(self, dt):
        return self.offset

    def dst(self, dt):
        return self.dst_offset

    def tzname(self, dt):
        return self.tz['timeZoneName']

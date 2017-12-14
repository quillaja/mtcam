import datetime as dt
import json
import time
from distutils.util import strtobool

from bokeh.plotting import figure, output_file, show
from bokeh.embed import components, file_html
from flask import Flask, abort, request
from peewee import DoesNotExist

import queries
import settings
from model import Cam, Mountain, ScrapeRecord

app = Flask(__name__)


def to_sys_time(t: dt.datetime, mttz: dict) -> dt.datetime:
    '''Uses mttz and system tz info to recalculate `t` into system time.'''

    if 'rawOffset' not in mttz or 'dstOffset' not in mttz:
        return t

    return (t - dt.timedelta(seconds=(mttz['rawOffset'] + mttz['dstOffset']))
            ) - (dt.timedelta(seconds=time.timezone) -
                 dt.timedelta(hours=time.localtime().tm_isdst))


def convert_input(time_input: str,
                  as_local_time: bool = False,
                  mttz: dict = None) -> dt.datetime:
    '''
    Parses a time string. If `as_local_time` is True, it will use the
    timezone info provided in `mttz` and the system's timezone info to
    convert `time_input` from the local time (at the mountain) to the
    system's time.
    '''
    try:
        time_input = dt.datetime.strptime(time_input, '%Y-%m-%dT%H:%M')
        if as_local_time and (mttz is not None):
            time_input = to_sys_time(time_input, mttz)
    except ValueError:
        time_input = None

    return time_input


@app.route('/api/data')
def do_data():
    '''Gets all the mountain and camera info, reformulates it into an
    easily JSONified data structure, and then returns it as JSON.'''

    r = queries.prefetch_all_mts_cams()
    data = dict()
    for m in r:
        md = {
            'id': m.id,
            'name': m.name,
            'state': m.state,
            'elevation_ft': m.elevation_ft,
            'latitude': m.latitude,
            'longitude': m.longitude,
            'tz': json.loads(m.tz_json),
            'pathname': m.as_pathname(),
            'cams': dict()
        }

        for c in m.cams_prefetch:
            cd = {
                'id': c.id,
                'name': c.name,
                'elevation_ft': c.elevation_ft,
                'latitude': c.latitude,
                'longitude': c.longitude,
                'is_active': c.is_active,
                'interval': c.interval,
                'comment': c.comment,
                'pathname': c.as_pathname()
            }
            md['cams'][c.id] = cd

        data[m.id] = md

    return json.dumps(data, indent=2, sort_keys=True)


@app.route('/api/mountains/<int:mt_id>/cams/<int:cam_id>/scrapes')
def do_scrapes(mt_id, cam_id):
    '''Gets and JSONifies the scrape records from the given mountain/cam 
    and dates, if provided.'''

    try:
        mt = Mountain.get(Mountain.id == mt_id)
        cam = Cam.get(Cam.id == cam_id)
    except DoesNotExist:
        return abort(404)

    end = request.args.get('end', '')
    start = request.args.get('start', '')
    as_local_time = strtobool(request.args.get('as_local_time', 'false'))

    try:
        mttz = json.loads(mt.tz_json)
    except ValueError:
        # mt.tz_json was invalid json (including '')
        mttz = None

    # convert GET date params into datetime objects
    # also will adjust the datetimes to the provided timezone
    # if as_local_time is True and mttz is a valid data structure
    end = convert_input(end, as_local_time, mttz)
    start = convert_input(start, as_local_time, mttz)

    # get records from db
    r = queries.scraperecords_for_cam(cam_id, start, end)
    data = list()
    mt_path = mt.as_pathname()
    cam_path = cam.as_pathname()

    # translate the db records into a more accessible form for
    # JSON conversion and use on the client-end
    for s in r:
        if s.result == ScrapeRecord.SUCCESS:
            filename = '{}/{}/{}/{}'.format(settings.IMG_ROOT, mt_path,
                                            cam_path, s.filename)
        else:
            filename = ''

        sd = {
            'time': s.created.strftime('%Y-%m-%d %H:%M'),
            'result': s.result,
            'detail': s.detail,
            'file': filename
        }
        data.append(sd)

    return json.dumps(data, indent=2, sort_keys=True)


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
    p = figure(
        title=mt_name + ' Summit - Forecast Weather Each Hour',
        x_axis_label='time',
        x_axis_type='datetime',
        y_axis_label='unit',
        width=1000)

    created = [w.created for w in forecasts]

    # temperature related lines
    p.line(
        created, [w.temp for w in forecasts],
        line_width=2,
        color='black',
        legend='Temp (F)')
    p.line(
        created, [w.temp_max for w in forecasts],
        legend='Max Temp',
        line_dash='4 4',
        color='red',
        line_width=0.5)
    p.line(
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
    p.line(
        created,
        wind_spd,
        line_width=2,
        color='lightgreen',
        legend='Wind (mph)')
    p.vbar(  # looks better than circle for wind gusts
        x=created,
        bottom=wind_spd,
        color='lightgreen',
        width=0.5,
        top=[w.wind_gust
             for w in forecasts])  # not pretty for missing values (drawn as 0)
    p.inverted_triangle(  # use inverted_triangle to make display 'map view'
        created,
        wind_spd,
        angle=wind_dir_inverted,
        angle_units='deg',
        color='green',
        size=8)
    p.rect(
        created,
        wind_spd,
        angle=wind_dir_inverted,
        angle_units='deg',
        color='green',
        width=0.5,
        height=16,
        height_units='screen')

    # precipitation related graphs
    shifted_time = created[1:] + [created[-1] + dt.timedelta(hours=1)
                                  ] if len(forecasts) > 0 else []
    p.quad(
        legend='Rain (in)',
        left=created,
        right=shifted_time,
        top=[w.rain for w in forecasts],
        bottom=0,
        color='blue',
        alpha=0.2)
    p.quad(
        legend='Snow (in)',
        left=created,
        right=shifted_time,
        top=[w.snow for w in forecasts],
        bottom=0,
        color='red',
        alpha=0.2)

    # legend setting has to be here to actually take effect
    p.legend.location = 'top_left'
    p.legend.orientation = 'horizontal'
    p.legend.background_fill_alpha = 0.5
    p.legend.click_policy = 'hide'

    # grid settings
    p.xaxis.minor_tick_line_color = 'black'
    p.xaxis.minor_tick_line_width = 1
    p.xgrid.minor_grid_line_color = 'gray'
    p.xgrid.minor_grid_line_alpha = 0.2
    p.ygrid.minor_grid_line_color = 'gray'
    p.ygrid.minor_grid_line_alpha = 0.2

    from bokeh.resources import CDN
    return file_html(p, CDN, mt_name)


@app.route('/api/mountains/<int:mt_id>/weather')
def do_weather(mt_id):
    '''Returns weather data for the mountain during the dates specified by
    the params 'start' and 'end'. Data can be had in JSON format or 'Bokeh
    format' (html for a plot).'''

    try:
        mt = Mountain.get(Mountain.id == mt_id)
    except DoesNotExist:
        abort(404)

    end = request.args.get('end', '')
    start = request.args.get('start', '')
    as_local_time = strtobool(request.args.get('as_local_time', 'false'))
    frmt = request.args.get('format', 'bokeh')

    try:
        mttz = json.loads(mt.tz_json)
    except ValueError:
        # mt.tz_json was invalid json (including '')
        mttz = None

    # convert GET date params into datetime objects
    # also will adjust the datetimes to the provided timezone
    # if as_local_time is True and mttz is a valid data structure
    end = convert_input(end, as_local_time, mttz)
    start = convert_input(start, as_local_time, mttz)

    forecasts = queries.weatherforecasts_for_mt(mt.id, start, end)

    # bokeh is the default if nothing is specified in the GET format
    # param, but if invalid junk is specified, send json by default
    # because it's less expensive to process.
    if frmt == 'bokeh':
        # script, div = weather_to_bokeh(forecasts, mt.name)
        # return div + script  # seems to work
        return weather_to_bokeh(forecasts, mt.name)
    else:
        return weather_to_json(forecasts)

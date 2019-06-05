import datetime as dt
import json
import time
from distutils.util import strtobool

from flask import Flask, abort, request
from peewee import DoesNotExist

import queries
import settings
from util import convert_input, to_sys_time, weather_to_bokeh, weather_to_json
from model import Cam, Mountain, ScrapeRecord

app = Flask(__name__)


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

        for c in m.cams:
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

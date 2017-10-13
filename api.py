import json
import datetime as dt
from flask import Flask, request
import queries
from model import Mountain, Cam, ScrapeRecord
import settings

app = Flask(__name__)


@app.route('/api/data')
def data():
    r = queries.prefetch_all_mts_cams()
    data = list()
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
            'cams': list()
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
            md['cams'].append(cd)

        data.append(md)

    return json.dumps(data, indent=2, sort_keys=True)


@app.route('/api/mountains/<int:mt_id>/cams/<int:cam_id>/scrapes')
def scrapes(mt_id, cam_id):
    end = request.args.get('end', dt.datetime.now())
    start = request.args.get('start', end - dt.timedelta(hours=24))

    mt_path = Mountain.get(Mountain.id == mt_id).as_pathname()
    cam_path = Cam.get(Cam.id == cam_id).as_pathname()

    r = queries.scraperecords_for_cam(cam_id, start, end)
    data = list()

    for s in r:
        if s.result == ScrapeRecord.SUCCESS:
            filename = '{}/{}/{}/{}'.format(settings.IMG_ROOT, mt_path,
                                            cam_path, s.filename)
        else:
            filename = ''

        sd = {
            'time': s.created,
            'result': s.result,
            'detail': s.detail,
            'file': filename
        }
        data.append(sd)

    return json.dumps(data, indent=2, sort_keys=True)

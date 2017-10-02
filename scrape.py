import datetime as dt
import os
import threading
import json
import requests
import ephem

from model import Mountain, Cam, ScrapeRecord, _db
from queries import prefetch_all_mts_cams

USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36'
IMG_ROOT = 'img'


class ScrapeJob(threading.Thread):
    def __init__(self, cam: Cam, tstamp: dt.datetime):
        super().__init__()

        self.cam = cam

        self.record = ScrapeRecord()
        self.record.cam = self.cam
        self.record.timestamp = tstamp

    def run(self):
        if self.is_between_sunrise_sunset():
            # do the scrape
            try:
                headers = requests.utils.default_headers()
                headers.update({'User-Agent': USER_AGENT})

                url = self.cam.url_fmt
                result = requests.get(url, headers=headers, timeout=10)
                if result.status_code == requests.codes.ok:
                    # request worked
                    self.record.result = ScrapeRecord.SUCCESS
                    self.record.filename = '{}.{}'.format(
                        int(self.record.timestamp.timestamp()),
                        self.cam.file_ext)

                    # write image to file
                    picdir = os.path.join(IMG_ROOT,
                                          self.cam.mountain.as_pathname(),
                                          self.cam.as_pathname())
                    try:
                        os.makedirs(picdir)
                    except OSError:
                        # this exception raised if directory exists, so skip
                        pass
                    filename = os.path.join(picdir, self.record.filename)
                    with open(filename, 'wb') as f:
                        f.write(result.content)

                else:
                    # request failed
                    result.raise_for_status()

            except requests.exceptions.RequestException as err:
                # some sort of failure
                self.record.result = ScrapeRecord.FAILURE
                self.record.detail = str(err)
        else:
            # idle
            self.record.result = ScrapeRecord.IDLE

    def is_between_sunrise_sunset(self):
        mt = self.cam.mountain
        tz = json.loads(mt.tz_json)

        obs = ephem.Observer()
        obs.horizon = '-12'
        obs.elevation = 0.3048 * mt.elevation_ft
        obs.lat = str(mt.latitude)
        obs.lon = str(mt.longitude)

        # change UTC->local, make local = 12 noon, then change local->UTC
        total_offset_s = tz['rawOffset'] + tz['dstOffset']
        noon = dt.datetime.utcnow() + dt.timedelta(seconds=total_offset_s)
        noon = noon.replace(hour=12, minute=0, second=0, microsecond=0)
        noon = noon - dt.timedelta(seconds=total_offset_s)
        obs.date = noon.strftime('%Y/%m/%d %H:%M:%S')

        srise = obs.previous_rising(ephem.Sun(), use_center=True)
        sset = obs.next_setting(ephem.Sun(), use_center=True)
        now = ephem.now()

        print(noon, srise, now, sset)
        return srise <= now <= sset  #srise.datetime().hour <= now.datetime().hour <= sset.datetime().hour


def main():
    data = prefetch_all_mts_cams()
    now = dt.datetime.now()  #maybe make this and other time UTC?
    jobs = list()

    # perform scrapes on every cam which
    # 1) is active
    # 2) it's the appropriate time of the hour
    for mt in data:
        for cam in mt.cams_prefetch:
            if cam.is_active:  # and now.minute % cam.every_mins == 0:
                j = ScrapeJob(cam, now)
                jobs.append(j)
                j.start()

    # wait for all threads to finish, up to 30 sec
    for j in jobs:
        j.join(timeout=30)

    # save all new scrape records to the database in 1 transaction
    with _db.atomic():
        for j in jobs:
            j.record.save()


if __name__ == '__main__':
    main()
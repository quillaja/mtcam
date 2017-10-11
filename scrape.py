import datetime as dt
import os
import threading
import json
import requests
import ephem

import settings
from model import Mountain, Cam, ScrapeRecord, _db
from queries import prefetch_all_mts_cams
from util import floor, ft_m


class ScrapeJob(threading.Thread):
    """
    Does the main work of scraping a webcam, including calculating if the
    mountain seen in the webcam is between sun rise and sun set. As this is
    a subclass of `Thread`, use it as one would a standard thread, by calling
    `join`.
    """

    def __init__(self, cam: Cam, tstamp: str):
        super().__init__()

        self.cam = cam

        self.record = ScrapeRecord()
        self.record.cam = self.cam
        self.timestamp = tstamp

    def run(self):
        '''Does the scrape'''

        try:
            if self.is_between_sunrise_sunset():
                # do the scrape
                headers = requests.utils.default_headers()
                headers.update({'User-Agent': settings.USER_AGENT})
                url = self.cam.url
                result = requests.get(
                    url, headers=headers, timeout=settings.REQUEST_TIMEOUT)

                if result.status_code == requests.codes.ok:
                    # request worked
                    self.record.result = ScrapeRecord.SUCCESS
                    self.record.filename = '{}.{}'.format(
                        self.timestamp, self.cam.file_ext)

                    # write image to file
                    picdir = os.path.join(settings.IMG_ROOT,
                                          self.cam.mountain.as_pathname(),
                                          self.cam.as_pathname())
                    os.makedirs(picdir, exist_ok=True)  #no err if dir exists
                    filename = os.path.join(picdir, self.record.filename)
                    with open(filename, 'wb') as f:
                        f.write(result.content)

                else:
                    # request failed
                    result.raise_for_status()

            else:
                # idle
                self.record.result = ScrapeRecord.IDLE

        except requests.exceptions.RequestException as err:
            # some sort of failure with the request
            self.record.result = ScrapeRecord.FAILURE
            self.record.detail = str(err)
        except OSError as err:
            # some sort of I/O failure
            self.record.result = ScrapeRecord.FAILURE
            self.record.detail = str(err)

    def is_between_sunrise_sunset(self):
        '''Uses pyephem to determine if the cam should be scraped based on if
        the time at the mountain is between sun rise and sun set.'''

        mt = self.cam.mountain

        obs = ephem.Observer()
        obs.horizon = settings.HORIZON
        obs.elevation = ft_m(mt.elevation_ft)
        obs.lat = str(mt.latitude)
        obs.lon = str(mt.longitude)

        # change UTC->local, make local = 12 noon, then change local->UTC
        try:
            tz = json.loads(mt.tz_json)
            total_offset_s = tz['rawOffset'] + tz['dstOffset']
        except Exception:
            total_offset_s = 0

        noon = dt.datetime.utcnow() + dt.timedelta(seconds=total_offset_s)
        noon = noon.replace(hour=12, minute=0, second=0, microsecond=0)
        noon = noon - dt.timedelta(seconds=total_offset_s)
        obs.date = noon.strftime('%Y/%m/%d %H:%M:%S')

        # get sun rise/set times, convert to datetime, and "floor"
        srise = floor(obs.previous_rising(ephem.Sun(), use_center=True))
        sset = floor(obs.next_setting(ephem.Sun(), use_center=True))
        now = floor(ephem.now())

        return srise <= now <= sset


def main():
    try:
        data = prefetch_all_mts_cams()
        minute = dt.datetime.now().minute
        filename_time = str(int(dt.datetime.utcnow().timestamp()))
        jobs = list()

        # perform scrapes on every cam which
        # 1) is active
        # 2) it's the appropriate time of the hour
        for mt in data:
            for cam in mt.cams_prefetch:
                if cam.is_active and minute % cam.interval == 0:
                    j = ScrapeJob(cam, filename_time)
                    jobs.append(j)
                    j.start()

        # wait for all threads to finish, up to 30 sec
        for j in jobs:
            j.join(timeout=settings.SCRAPE_JOB_TIMEOUT)

        # save all new scrape records to the database in 1 transaction
        with _db.atomic():
            for j in jobs:
                j.record.save()

    except Exception as err:
        # catch all other exceptions and print
        print('\n', dt.datetime.now().isoformat())
        print(err.with_traceback())


if __name__ == '__main__':
    main()
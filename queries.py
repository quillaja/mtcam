import datetime as dt
import peewee
from model import Mountain, Cam, ScrapeRecord


def prefetch_all_mts_cams():
    '''
    Gets all the `Mountain` and `Cam` instances from the database.
    '''

    mountains = Mountain.select()
    cams = Cam.select()

    return peewee.prefetch(mountains, cams)


def scraperecords_for_cam(cam_id, start=None, end=None):
    '''
    Gets all `ScrapeRecord`s for the `Cam` with the given id between `start` 
    and `end`. If no start and end dates are specified, 
    it'll return the scrapes of the previous 24 hours.
    '''

    end = end or dt.datetime.now()
    start = start or (end - dt.timedelta(hours=24))

    return ScrapeRecord.select().where(
        ScrapeRecord.cam_id == cam_id, ScrapeRecord.created.between(
            start, end)).order_by(ScrapeRecord.created)

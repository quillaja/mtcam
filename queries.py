import peewee
from model import Mountain, Cam


def prefetch_all_mts_cams():
    mountains = Mountain.select()
    cams = Cam.select()

    return peewee.prefetch(mountains, cams)

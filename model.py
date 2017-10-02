import datetime
import peewee as p

_db = p.SqliteDatabase('mtcam_test.db')


class ModelBase(p.Model):
    created = p.DateTimeField(default=datetime.datetime.now)
    modified = p.TimestampField()

    def save(self, *args, **kwargs):
        '''Overrides default Model.save() to enable the modified field
        to be updated every time the model is written to the database.'''
        self.modified = datetime.datetime.now()
        super().save(*args, **kwargs)
        # super(ModelBase, self).save(*args, **kwargs)

    class Meta(object):
        database = _db


class Mountain(ModelBase):
    name = p.CharField()
    state = p.CharField()
    elevation_ft = p.IntegerField()
    latitude = p.FloatField()
    longitude = p.FloatField()
    tz_json = p.TextField(default='')  #timezone info from google. req Lat,Lon

    def get_location(self):
        '''Return 3-tuple of latitude, logitude, and elevation(ft).'''
        return (self.latitude, self.longitude, self.elevation_ft)

    # def set_location(self, location):
    #     '''Set location from provided 3-tuple of latitude, longitude,
    #     and elevatiion(ft).'''
    #     self.latitude, self.longitude, self.elevation_ft = location
    #     self.save()

    def as_pathname(self):
        return '{}_{}'.format(self.name, self.state).lower().replace(' ', '_')

    def __repr__(self):
        return '{} ({})'.format(self.name, self.state)


class Cam(ModelBase):
    mountain = p.ForeignKeyField(Mountain, related_name='cams')
    name = p.CharField()
    elevation_ft = p.IntegerField()
    latitude = p.FloatField()
    longitude = p.FloatField()
    url_fmt = p.CharField()
    file_ext = p.CharField(default='jpg')
    is_active = p.BooleanField(default=True)
    every_mins = p.IntegerField(default=5)  # scrape every X mins
    comment = p.TextField(default='')

    def get_location(self):
        return (self.latitude, self.longitude, self.elevation_ft)

    # def set_location(self, location):
    #     self.latitude, self.longitude, self.elevation_ft = location
    #     self.save()

    def as_pathname(self):
        return str(self.name).lower().replace(' ', '_')

    def __repr__(self):
        return '{} ({})'.format(self.name, str(self.mountain))


class ScrapeRecord(ModelBase):
    # status codes to use in the 'result' field
    SUCCESS = 'success'
    FAILURE = 'failure'
    IDLE = 'idle'

    cam = p.ForeignKeyField(Cam, related_name='scrapes')
    timestamp = p.DateTimeField(
        default=datetime.datetime.now())  # time the image was downloaded
    result = p.CharField()
    detail = p.TextField(default='')
    filename = p.CharField(default='')  #does not include path

    def __repr__(self):
        return '{}\t{}\t{}'.format(self.timestamp, self.cam.name, self.result)


def create_tables():
    _db.connect()
    _db.create_tables([Mountain, Cam, ScrapeRecord], safe=True)
    _db.close()


def create_test_data():
    # import random
    import json
    import util

    #-- mountains ---

    hood = Mountain.create(
        name='Hood',
        state='OR',
        elevation_ft=11200,
        latitude=45.373439,
        longitude=-121.695962)
    hood.tz_json = json.dumps(util.get_tz(hood.latitude, hood.longitude))
    hood.save()

    sisters = Mountain.create(
        name='Three Sisters',
        state='OR',
        elevation_ft=10358,
        latitude=44.103241,
        longitude=-121.769253)
    sisters.tz_json = json.dumps(
        util.get_tz(sisters.latitude, sisters.longitude))
    sisters.save()

    rainier = Mountain.create(
        name='Rainier',
        state='WA',
        elevation_ft=14411,
        latitude=46.851736,
        longitude=-121.760398)
    rainier.tz_json = json.dumps(
        util.get_tz(rainier.latitude, rainier.longitude))
    rainier.save()

    fuji = Mountain.create(
        name="Fuji",
        state="Japan",
        elevation_ft=12388,
        latitude=35.360388,
        longitude=138.727724)
    fuji.tz_json = json.dumps(util.get_tz(fuji.latitude, fuji.longitude))
    fuji.save()

    #-- Cams ------
    palmer = Cam.create(
        mountain=hood,
        name='Palmer',
        elevation_ft=7000,
        latitude=45.373439,
        longitude=-121.695962,
        url_fmt=
        '''https://www.timberlinelodge.com/snowcameras/palmerbottom.jpg''')
    bachelor = Cam.create(
        mountain=sisters,
        name='Bachelor',
        elevation_ft=6040,
        latitude=44.103241,
        longitude=-121.769253,
        url_fmt='''https://www.mtbachelor.com/webcams/southsisteroutback.jpg'''
    )
    paradise = Cam.create(
        mountain=rainier,
        name='Paradise',
        elevation_ft=5400,
        latitude=46.851736,
        longitude=-121.760398,
        url_fmt='''https://www.nps.gov/webcams-mora/mountain.jpg''')

    subaru = Cam.create(
        mountain=fuji,
        name='Fuji Subaru 5th',
        elevation_ft=5000,
        latitude=35.360388,
        longitude=138.727724,
        url_fmt='''http://www.mfi.or.jp/goraikou223/cam04/0000.jpg''')

    # t = datetime.datetime.now()
    # for td in range(0, 25, 5):
    #     t = t + datetime.timedelta(minutes=td)

    #     status = ScrapeRecord.SUCCESS if random.random(
    #     ) >= 0.1 else ScrapeRecord.FAILURE
    #     ScrapeRecord.create(
    #         cam=palmer,
    #         timestamp=t,
    #         result=status,
    #         filename='{}.jpg'.format(int(t.timestamp())))

    #     status = ScrapeRecord.SUCCESS if random.random(
    #     ) >= 0.1 else ScrapeRecord.FAILURE
    #     ScrapeRecord.create(
    #         cam=cooper,
    #         timestamp=t,
    #         result=status,
    #         filename='{}.jpg'.format(int(t.timestamp())))

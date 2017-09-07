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
        super(ModelBase, self).save(*args, **kwargs)

    class Meta(object):
        database = _db


class Mountain(ModelBase):
    name = p.CharField()
    state = p.CharField()
    elevation_ft = p.IntegerField()
    latitude = p.FloatField()
    longitude = p.FloatField()
    tz_json = p.TextField(default='') #timezone info from google. req Lat,Lon

    def get_location(self):
        '''Return 3-tuple of latitude, logitude, and elevation(ft).'''
        return (self.latitude, self.longitude, self.elevation_ft)

    # def set_location(self, location):
    #     '''Set location from provided 3-tuple of latitude, longitude, 
    #     and elevatiion(ft).'''
    #     self.latitude, self.longitude, self.elevation_ft = location
    #     self.save()

    def as_pathname(self):
        return '{}_{}'.format(self.name,self.state).lower().replace(' ', '_')

    def __repr__(self):
        return '{} ({})'.format(self.name, self.state)


class Cam(ModelBase):
    mountain = p.ForeignKeyField(Mountain, related_name='cams')
    name = p.CharField()
    elevation_ft = p.IntegerField()
    latitude = p.FloatField()
    longitude = p.FloatField()
    url_fmt = p.CharField()
    is_active = p.BooleanField(default=True)
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
    filename = p.CharField()  #does not include path
    file_ext = p.CharField(default='jpg')

    def __repr__(self):
        return '{}\t{}\t{}'.format(self.timestamp, self.cam.name, self.result)


def create_tables():
    _db.connect()
    _db.create_tables([Mountain, Cam, ScrapeRecord], safe=True)
    _db.close()


def create_test_data():
    import random

    hood = Mountain.create(
        name='Hood',
        state='OR',
        elevation_ft=11200,
        latitude=45.5,
        longitude=-120.1)
    sisters = Mountain.create(
        name='Three Sisters',
        state='OR',
        elevation_ft=10000,
        latitude=41.2,
        longitude=-119.0)
    palmer = Cam.create(
        mountain=hood,
        name='Palmer',
        elevation_ft=7000,
        latitude=45.4,
        longitude=-120.12,
        url_fmt='''poop.com/palmer''')
    cooper = Cam.create(
        mountain=hood,
        name='Cooper Spur',
        elevation_ft=6040,
        latitude=45.8,
        longitude=-120.0,
        url_fmt='''poop.com/cooper''')

    t = datetime.datetime.now()
    for td in range(0, 25, 5):
        t = t + datetime.timedelta(minutes=td)

        status = ScrapeRecord.SUCCESS if random.random(
        ) >= 0.1 else ScrapeRecord.FAILURE
        ScrapeRecord.create(
            cam=palmer,
            timestamp=t,
            result=status,
            filename='{}.jpg'.format(int(t.timestamp())))

        status = ScrapeRecord.SUCCESS if random.random(
        ) >= 0.1 else ScrapeRecord.FAILURE
        ScrapeRecord.create(
            cam=cooper,
            timestamp=t,
            result=status,
            filename='{}.jpg'.format(int(t.timestamp())))

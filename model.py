import peewee as p

_db = p.SqliteDatabase('mtcam_test.db')


class ModelBase(p.Model):
    class Meta(object):
        database = _db


class Mountain(ModelBase):
    name = p.CharField()
    state = p.CharField()
    elevation_ft = p.IntegerField()
    latitude = p.FloatField()
    longitude = p.FloatField()

    def get_location(self):
        return (self.latitude, self.longitude, self.elevation_ft)

    def set_location(self, location):
        self.latitude, self.longitude, self.elevation_ft = location
        self.save()


class Cam(ModelBase):
    mountain = p.ForeignKeyField(Mountain, related_name='cams')
    name = p.CharField(unique=True)
    elevation_ft = p.IntegerField()
    latitude = p.FloatField()
    longitude = p.FloatField()
    url_fmt = p.CharField()
    is_active = p.BooleanField(default=True)
    comment = p.TextField()

    def get_location(self):
        return (self.latitude, self.longitude, self.elevation_ft)

    def set_location(self, location):
        self.latitude, self.longitude, self.elevation_ft = location
        self.save()


class ScrapeRecord(ModelBase):
    # status codes to use in the 'result' field
    SUCCESS = 'success'
    FAILURE = 'failure'
    IDLE = 'idle'

    cam = p.ForeignKeyField(Cam, related_name='scrapes')
    timestamp = p.DateTimeField() # time the image was downloaded
    result = p.CharField()
    detail = p.TextField()
    filename = p.CharField()  #does not include path


def create_tables():
    _db.connect()
    _db.create_tables([Mountain, Cam, ScrapeRecord], safe=True)
    _db.close()

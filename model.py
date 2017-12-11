import datetime
import peewee as p
import settings

if settings.DB_TYPE == 'sqlite':
    _db = p.SqliteDatabase(settings.DB_CONNECTION)
elif settings.DB_TYPE == 'mysql':
    _db = p.MySQLDatabase(settings.DB_CONNECTION)
elif settings.DB_TYPE == 'postgresql':
    _db = p.PostgresqlDatabase(settings.DB_CONNECTION)
else:
    raise RuntimeError('Invalid DB settings')


class ModelBase(p.Model):
    created = p.DateTimeField(default=datetime.datetime.now)
    modified = p.DateTimeField()

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
    get_weather = p.BooleanField(default=True) #if should scrape nws weather

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
    url = p.CharField()
    file_ext = p.CharField(default='jpg')
    is_active = p.BooleanField(default=True)
    interval = p.IntegerField(default=5)  # scrape every N mins
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
    result = p.CharField()
    detail = p.TextField(default='')
    filename = p.CharField(default='')  #does not include path

    def __repr__(self):
        return '{}\t{}\t{}'.format(self.created, self.cam.name, self.result)

class WeatherForecast(ModelBase):
    mountain = p.ForeignKeyField(Mountain, related_name='weather')
    #retrieved = p.DateTimeField() # just set created to sec=0,microsec=0
    temp = p.FloatField(null=True, default=None)
    temp_max = p.FloatField(null=True, default=None)
    temp_min = p.FloatField(null=True, default=None)
    wind_spd = p.FloatField(null=True, default=None)
    wind_gust = p.FloatField(null=True, default=None)
    wind_dir = p.FloatField(null=True, default=None)
    prob_precip = p.FloatField(null=True, default=None)
    rain = p.FloatField(null=True, default=None)
    snow = p.FloatField(null=True, default=None)
    cloud = p.FloatField(null=True, default=None)



def create_tables():
    _db.connect()
    _db.create_tables([Mountain, Cam, ScrapeRecord, WeatherForecast], safe=True)
    _db.close()

def _migrate_add_weather():
    import playhouse.migrate as m
    migrator = m.SqliteMigrator(_db)

    m.migrate(
        migrator.add_column('mountain','get_weather', Mountain.get_weather)
    )

    _db.create_table(WeatherForecast,safe=True)


def create_test_data():
    # import random
    import json
    import util

    #-- mountains ---

    hood = Mountain.create(
        name='Mt Hood',
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
        name='Mt Rainier',
        state='WA',
        elevation_ft=14411,
        latitude=46.851736,
        longitude=-121.760398)
    rainier.tz_json = json.dumps(
        util.get_tz(rainier.latitude, rainier.longitude))
    rainier.save()

    fuji = Mountain.create(
        name="Mt Fuji",
        state="Japan",
        elevation_ft=12388,
        latitude=35.360388,
        longitude=138.727724)
    fuji.tz_json = json.dumps(util.get_tz(fuji.latitude, fuji.longitude))
    fuji.save()

    blanc = Mountain.create(
        name='Mt Blanc',
        state='France',
        elevation_ft=15774,
        latitude=45.833611,
        longitude=6.865)
    blanc.tz_json = json.dumps(util.get_tz(blanc.latitude, blanc.longitude))
    blanc.save()

    #-- Cams ------
    palmer = Cam.create(
        mountain=hood,
        name='Palmer',
        elevation_ft=7000,
        latitude=45.373439,
        longitude=-121.695962,
        url='''https://www.timberlinelodge.com/snowcameras/palmerbottom.jpg''')
    vista = Cam.create(
        mountain=hood,
        name='Vista',
        elevation_ft=5000,
        latitude=45.373439,
        longitude=-121.695962,
        url='''https://www.skihood.com/cams/vista''')

    bachelor = Cam.create(
        mountain=sisters,
        name='Bachelor',
        elevation_ft=6040,
        latitude=44.103241,
        longitude=-121.769253,
        url='''https://www.mtbachelor.com/webcams/southsisteroutback.jpg''')

    paradise = Cam.create(
        mountain=rainier,
        name='Paradise',
        elevation_ft=5400,
        latitude=46.851736,
        longitude=-121.760398,
        url='''https://www.nps.gov/webcams-mora/mountain.jpg''')

    fuji_n = Cam.create(
        mountain=fuji,
        name='North',
        elevation_ft=5000,
        latitude=35.360388,
        longitude=138.727724,
        url=
        '''http://www.sizenken.biodic.go.jp/system/camera_image/biodic/117_c/NCS.jpg'''
    )

    blanc_tourism = Cam.create(
        mountain=blanc,
        name='Tourism',
        elevation_ft=5500,
        latitude=45.833611,
        longitude=6.865,
        url=
        '''http://www.chamonix.com/webcam/webcam-argentiere-mont-blanc.jpg''')

import json
import datetime as dt

import requests

import settings
from model import Mountain, WeatherForecast


def get_forecast_url(lat, lon):
    '''get the url to use for weather info.'''
    # example https://api.weather.gov/points/45.3734,-121.696

    url = 'https://api.weather.gov/points/{},{}'.format(lat, lon)
    headers = requests.utils.default_headers()
    headers.update({'User-Agent': settings.USER_AGENT})

    result = requests.get(
        url, headers=headers, timeout=settings.REQUEST_TIMEOUT)

    if result.status_code == requests.codes.ok:
        return result.json()['properties']['forecastGridData']
    else:
        result.raise_for_status()


def identity(value):
    '''returns the value unchanged'''
    return value


def c_to_f(temp_c):
    '''converts deg C to deg F'''
    return temp_c * 1.8 + 32


def ms_to_mph(speed_ms):
    '''converts meters per second to miles per hour'''
    return speed_ms * 2.23694


def mm_to_in(mm):
    '''converts mm to inch'''
    return mm * 0.0393701


def m_to_ft(meters):
    '''converts meters to feet'''
    return meters * 3.28084


# This is used to make it easy to process all the information contained
# in the NOAA forecast data without having to "hard code" each line.
#
SCHEMA = {
    # my name: (noaa key, conversion func) # my unit, their unit
    'temp': ('temperature', c_to_f),  #F, C
    'temp_max': ('maxTemperature', c_to_f),  #F, C
    'temp_min': ('minTemperature', c_to_f),  #F, C
    'temp_apparent': ('apparentTemperature', c_to_f),  #F, C
    'humidity': ('relativeHumidity', identity),  #%, %
    'dewpoint': ('dewpoint', c_to_f),  #F, C
    'wind_chill': ('windChill', c_to_f),  #F, C
    #dont care ('heat_index', c_to_f): 'heatIndex', #F, C
    'wind_spd': ('windSpeed', ms_to_mph),  #mph, m_s-1 (m/s)
    'wind_gust': ('windGust', ms_to_mph),  #mph, m/s
    'wind_dir': ('windDirection', identity),  #deg, deg
    'prob_precip': ('probabilityOfPrecipitation', identity),  #%, %
    'rain': ('quantitativePrecipitation', mm_to_in),  #in, mm -FLOAT
    'snow': ('snowfallAmount', mm_to_in),  #in, mm -FLOAT
    'snow_level': ('snowLevel', m_to_ft),  #ft, m
    'ice_accumulation': ('iceAccumulation', mm_to_in),  #in, mm -FLOAT
    'cloud': ('skyCover', identity),  #%, %
    #is null? 'ceiling': ('ceilingHeight', ), #NULL
    # has other NULL values or non-applicable things
}


def get_raw_noaa(url):
    '''returns a dict of the raw NOAA data, 
    or None if there was an error, etc'''

    try:
        headers = requests.utils.default_headers()
        headers.update({'User-Agent': settings.USER_AGENT})
        result = requests.get(
            url, headers=headers, timeout=settings.REQUEST_TIMEOUT)

        if result.status_code == requests.codes.ok:
            # request worked
            return result.json()

        else:
            # request failed
            result.raise_for_status()

    except:
        # some kind of error happened... whatevers
        return None


def noaa_accessor(raw, key):
    '''easily get the desired data from the noaa forecast by key (in SCHEMA).
    Returns None if any error occurs.'''

    try:
        return raw['properties'][key]['values'][0]['value']
    except:
        return None


def noaa_to_model(raw_noaa):
    '''gets and converts the raw data from NOAA to the internal weather
    forecast representation in model.py. Does not set WeatherForecast.mountain,
    or save() it to the database.'''

    data = dict()
    for mykey in SCHEMA.keys():
        noaakey, convert = SCHEMA[mykey]
        item = noaa_accessor(raw_noaa, noaakey)
        data[mykey] = convert(item) if item is not None else item

    # populate WeatherForecast model
    wf = WeatherForecast()
    wf.temp = int(data['temp']) if data['temp'] is not None else None
    wf.temp_max = int(
        data['temp_max']) if data['temp_max'] is not None else None
    wf.temp_min = int(
        data['temp_min']) if data['temp_min'] is not None else None
    wf.temp_apparent = int(
        data['temp_apparent']) if data['temp_apparent'] is not None else None
    wf.humidity = int(
        data['humidity']) if data['humidity'] is not None else None
    wf.dewpoint = int(
        data['dewpoint']) if data['dewpoint'] is not None else None
    wf.wind_chill = int(
        data['wind_chill']) if data['wind_chill'] is not None else None
    wf.wind_spd = int(
        data['wind_spd']) if data['wind_spd'] is not None else None
    wf.wind_gust = int(
        data['wind_gust']) if data['wind_gust'] is not None else None
    wf.wind_dir = int(
        data['wind_dir']) if data['wind_dir'] is not None else None
    wf.prob_precip = int(
        data['prob_precip']) if data['prob_precip'] is not None else None
    wf.rain = float(data['rain']) if data['rain'] is not None else None
    wf.snow = float(data['snow']) if data['snow'] is not None else None
    wf.snow_level = int(
        data['snow_level']) if data['snow_level'] is not None else None
    wf.ice_accumulation = float(data[
        'ice_accumulation']) if data['ice_accumulation'] is not None else None
    wf.cloud = int(data['cloud']) if data['cloud'] is not None else None

    return wf


def scrape(mountain):
    '''scrape and save the current NOAA forecast for the given mountain.'''
    forecast = noaa_to_model(get_raw_noaa(mountain.weather_url))
    forecast.mountain = mountain
    forecast.created = forecast.created.replace(second=0, microsecond=0)
    forecast.save()


def main():
    '''Do the actual full scrape and save for each mountain with a
    NOAA forecast weather url.'''

    mts = Mountain.select().where(Mountain.weather_url != None)
    for m in mts:
        scrape(m)


if __name__ == '__main__':
    main()
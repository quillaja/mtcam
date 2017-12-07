import sys
import json
from datetime import datetime

import requests
from bs4 import BeautifulSoup
from bokeh.plotting import figure, output_file, show

# URL = "https://graphical.weather.gov/xml/sample_products/browser_interface/ndfdBrowserClientByDay.php?lat={}&lon={}&format=24+hourly&numDays=1&startDate={}"

# params = {'lat': '45.373439', 'lon': '-121.695962', 'date': '2017-12-05'}
# req_url = URL.format(params['lat'], params['lon'], params['date'])
# responseSummarized = requests.get(req_url)

# parsedSummarized = BeautifulSoup(responseSummarized.content, 'xml')

# max temp
# parsed.data.parameters.find_all('temperature')[0].value.text


def scrape():
    '''get xml from NWS'''
    params = {
        'lat': '45.373439',
        'lon': '-121.695962',
        'product': 'time-series',
        # regular (forecast)
        'temp': 'temp',
        'maxt': 'maxt',
        'mint': 'mint',
        'snow': 'snow',
        'qpf': 'qpf',
        'wspd': 'wspd',
        'wgust': 'wgust',
        'wdir': 'wdir',
        # real time mesoscale analysis
        # 'precipa_r': 'precipa_r',
        # 'sky_r': 'sky_r',
        # 'td_r': 'td_r',
        # 'temp_r': 'temp_r',
        # 'wdir_r': 'wdir_r',
        # 'wspd_r': 'wspd_r'
    }

    # TODO: make more robust, timeouts, etc
    URL_NON_SUMMARIZED = "https://graphical.weather.gov/xml/sample_products/browser_interface/ndfdXMLclient.php"
    responseNonsummarized = requests.get(URL_NON_SUMMARIZED, params=params)
    pns = BeautifulSoup(responseNonsummarized.content, 'xml')
    return pns


def clean(soup_obj):
    '''use bs4 to parse xml and extract data. put data into dict'''

    # TODO: make more fault tolerant
    # TODO: check for 'error' results
    temp_max = soup_obj.find_all('temperature')[0].value.text
    temp_min = soup_obj.find_all('temperature')[1].value.text
    temp = soup_obj.find_all('temperature')[2].value.text
    rain = soup_obj.find_all('precipitation')[0].value.text
    snow = soup_obj.find_all('precipitation')[1].value.text
    wind_spd = soup_obj.find_all('wind-speed')[0].value.text
    wind_gust = soup_obj.find_all('wind-speed')[1].value.text
    wind_dir = soup_obj.find_all('direction')[0].value.text
    # weather = soup_obj.find_all('weather')[0].weather_conditions

    retrieved = datetime.now()
    retrieved = retrieved.replace(microsecond=0)

    data = {
        'retrieved': retrieved.isoformat(),
        'temp_max': temp_max,
        'temp_min': temp_min,
        'temp': temp,
        'rain': rain,
        'snow': snow,
        'wind_spd': wind_spd,
        'wind_gust': wind_gust,
        'wind_dir': wind_dir,
        # 'weather': weather
    }

    return data


def write(data):
    '''append jsonified data to file'''
    data_json = json.dumps(data)

    with open('weather.json', 'a') as f:
        f.write(data_json + '\n')


def read():
    '''read data from file and put it in a list of dicts'''
    data_list = list()
    with open('weather.json') as f:
        for line in f.readlines():
            data_list.append(json.loads(line))

    return data_list


def to_series(data_list):
    '''convert list of dicts into dict of lists (grouped by parameter)'''
    series = {
        'retrieved': list(),
        'temp_max': list(),
        'temp_min': list(),
        'temp': list(),
        'rain': list(),
        'snow': list(),
        'wind_spd': list(),
        'wind_gust': list(),
        'wind_dir': list(),
        # 'weather': weather
    }

    for data in data_list:
        for k in series.keys():
            v = data.get(k) # returns the item or None
            if v is None:
                item = v # customize what to do with non-existent data
            elif k == 'retrieved':
                item = datetime.strptime(v, '%Y-%m-%dT%H:%M:%S').replace(
                    second=0)
            else:
                item = float(v)

            series[k].append(item)

    return series


def plot(series):
    '''draw a nice graph of the data'''

    p = figure(
        title='Mt Hood Summit - Forecast Weather Each Hour',
        x_axis_label='time',
        x_axis_type='datetime',
        y_axis_label='unit',
        width=1000)

    # temperature related lines
    p.line(
        series['retrieved'],
        series['temp'],
        line_width=2,
        color='black',
        legend='Temp (F)')
    p.line(
        series['retrieved'], series['temp_max'], line_dash='4 4', color='red')
    p.line(
        series['retrieved'], series['temp_min'], line_dash='4 4', color='blue')

    # wind related lines
    p.line(
        series['retrieved'],
        series['wind_spd'],
        line_width=2,
        color='lightgreen',
        legend='Wind (mph)')
    p.triangle(  # original data apparently in 'map view', not normal wind reporting format
        series['retrieved'],
        series['wind_spd'],
        angle=series['wind_dir'],
        angle_units='deg',
        color='lightgreen',
        size=12)
    p.rect(
        series['retrieved'],
        series['wind_spd'],
        angle=series['wind_dir'],
        angle_units='deg',
        color='lightgreen',
        width=1,
        height=25,
        height_units='screen')
    p.circle(
        series['retrieved'],
        series['wind_spd'],
        radius=series['wind_gust'],
        radius_units='screen',
        color='green',
        fill_color=None,
        legend='wind gust')
    # p.line(series['retrieved'], series['wind_dir'], line_dash='4 4', color='green', legend='wind dir deg')

    output_file('weather_plot.html')
    show(p)


if __name__ == '__main__':
    if len(sys.argv) > 1 and sys.argv[1] == 'plot':
        plot(to_series(read()))
    else:
        soup_obj = scrape()
        data = clean(soup_obj)
        write(data)


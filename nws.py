import sys
import json
from datetime import datetime, timedelta

import requests
from bs4 import BeautifulSoup
from bokeh.plotting import figure, output_file, show


def scrape():
    '''get xml from NWS'''
    params = {
        # Mt Hood, approximate in Devil's Kitchen
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
    url_non_summarized = "https://graphical.weather.gov/xml/sample_products/browser_interface/ndfdXMLclient.php"
    response = requests.get(url_non_summarized, params=params)
    pns = BeautifulSoup(response.content, 'xml')
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

    retrieved = datetime.now()
    retrieved = retrieved.replace(microsecond=0)

    # TODO: convert to numbers instead of leaving as strings
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
    }

    return data


def write(data):
    '''append jsonified data to file'''
    data_json = json.dumps(data, sort_keys=True)

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
    }

    # do conversion by processing each dict "data" with a loop over
    # the keys in "series". This ensures that none of the lists in "series"
    # are missing a data point if a particular parameter was not in
    # the original data.
    for data in data_list:
        for k in series.keys():
            v = data.get(k)  # returns the item or None
            if v is None:
                item = v  # customize what to do with non-existent data
            elif k == 'retrieved':
                item = datetime.strptime(
                    v, '%Y-%m-%dT%H:%M:%S').replace(second=0)
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
        size=10)
    p.rect(
        series['retrieved'],
        series['wind_spd'],
        angle=series['wind_dir'],
        angle_units='deg',
        color='lightgreen',
        width=0.5,
        height=20,
        height_units='screen')
    p.rect(  # looks better than circle for wind gusts
        series['retrieved'],
        series['wind_spd'],
        # angle=series['wind_dir'],
        # angle_units='deg',
        color='green',
        width=1,
        height=series['wind_gust'],
        height_units='screen')

    # precipitation related graphs
    shifted_time = series['retrieved'][1:] + [
        series['retrieved'][-1] + timedelta(hours=1)
    ]
    p.quad(
        legend='Rain (in)',
        left=series['retrieved'],
        right=shifted_time,
        top=series['rain'],
        bottom=0,
        color='blue',
        alpha=0.2)
    p.quad(
        legend='Snow (in)',
        left=series['retrieved'],
        right=shifted_time,
        top=series['snow'],
        bottom=0,
        color='red',
        alpha=0.2)

    # legend setting has to be here to actually take effect
    p.legend.location = 'top_left'

    output_file('weather_plot.html')
    show(p)


def main():
    if len(sys.argv) > 1 and sys.argv[1] == 'plot':
        plot(to_series(read()))
    else:
        soup_obj = scrape()
        data = clean(soup_obj)
        write(data)


if __name__ == '__main__':
    main()

import sys
import json
from datetime import datetime, timedelta

import requests
from bs4 import BeautifulSoup
from bokeh.plotting import figure, output_file, show

#region: scraping

# a sort of schema to help me request and parse data from the NWS service
nws_xml = {
    # my name : (NWS resp tag, NWS resp type, request type)
    'temp': ('temperature', 'hourly', 'temp'),
    'temp_max': ('temperature', 'maximum', 'maxt'),
    'temp_min': ('temperature', 'minimum', 'mint'),
    'wind_spd': ('wind-speed', 'sustained', 'wspd'),
    'wind_gust': ('wind-speed', 'gust', 'wgust'),
    'wind_dir': ('direction', 'wind', 'wdir'),
    'prob_precip': ('probability-of-precipitation', '12 hour', 'pop12'),
    'rain': ('precipitation', 'liquid', 'qpf'),
    'snow': ('precipitation', 'snow', 'snow'),
    'cloud': ('cloud-amount', 'total', 'sky')
}


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
        'pop12': 'pop12',
        'sky': 'sky'
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
    soup_obj = BeautifulSoup(response.content, 'xml')
    return soup_obj


def clean(soup_obj):
    '''use bs4 to parse xml and extract data. put data into dict'''

    # TODO: check for 'error' results

    data = {'retrieved': datetime.now().replace(second=0, microsecond=0)}

    # NOTE: value.string throws AttributeError if no 'value' is found
    # TODO: make this able to accept response from multi-point request
    #
    # iterate the data items i want, extracting each from xml via the
    # beautifulsoup 'find' method. If the data item is not in the xml
    # the 'value.string' will throw AttributeError.
    point_xml = soup_obj.find('parameters', **{'applicable-location': 'point1'})
    for var, v in nws_xml.items():
        tag, ttype, _ = v
        try:
            data[var] = float(point_xml.find(tag, type=ttype).value.string)
        except AttributeError:
            data[var] = None

    return data


# NOTE: will not work with above functions because datetime cannot be
# jsonified. Doesn't matter because data will ultimately in a database.
def write(data):
    '''append jsonified data to file'''
    data_json = json.dumps(data, sort_keys=True)

    with open('weather.json', 'a') as f:
        f.write(data_json + '\n')

#endregion

#region: plotting
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
        series['retrieved'],
        series['temp_max'],
        legend='Max Temp',
        line_dash='4 4',
        color='red',
        line_width=0.5)
    p.line(
        series['retrieved'],
        series['temp_min'],
        legend='Min Temp',
        line_dash='4 4',
        color='blue',
        line_width=0.5)

    # wind related lines
    # for some stupid reason bokeh draws angles counter clockwise, so
    # i have to negate all the wind_dir entries to make it display right.
    #
    # wind speed and gust are in knots. 1 kn = 1.15078 mph
    wind_dir_inverted = [-w for w in series['wind_dir']]
    p.line(
        series['retrieved'],
        series['wind_spd'],
        line_width=2,
        color='lightgreen',
        legend='Wind (knots)')
    p.vbar(  # looks better than circle for wind gusts
        x=series['retrieved'],
        bottom=series['wind_spd'],
        color='lightgreen',
        width=0.5,
        top=series['wind_gust'])  # not pretty for missing values (drawn as 0)
    p.inverted_triangle(  # use inverted_triangle to make display 'map view'
        series['retrieved'],
        series['wind_spd'],
        angle=wind_dir_inverted,
        angle_units='deg',
        color='green',
        size=8)
    p.rect(
        series['retrieved'],
        series['wind_spd'],
        angle=wind_dir_inverted,
        angle_units='deg',
        color='green',
        width=0.5,
        height=16,
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
    p.legend.orientation = 'horizontal'
    p.legend.background_fill_alpha = 0.5
    p.legend.click_policy = 'hide'

    # grid settings
    p.xaxis.minor_tick_line_color = 'black'
    p.xaxis.minor_tick_line_width = 1
    p.xgrid.minor_grid_line_color = 'gray'
    p.xgrid.minor_grid_line_alpha = 0.2
    p.ygrid.minor_grid_line_color = 'gray'
    p.ygrid.minor_grid_line_alpha = 0.2

    output_file('weather_plot.html')
    show(p)

#endregion

def main():
    if len(sys.argv) > 1 and sys.argv[1] == 'plot':
        plot(to_series(read()))
    else:
        soup_obj = scrape()
        data = clean(soup_obj)
        write(data)


if __name__ == '__main__':
    main()

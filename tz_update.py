import json
import util
from model import Mountain


def main():
    """
    Updates the tz_json field of each `Mountain`. This script is intended
    to be called as a cron job.
    """

    for m in Mountain.select():
        tz = json.dumps(util.get_tz(m.latitude, m.longitude))
        m.tz_json = tz
        m.save()


if __name__ == '__main__':
    main()
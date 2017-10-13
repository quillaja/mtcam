mtcam.quillaja.net/ -> /

/api
    root of api. returns nothing.
/api/data
    returns json with list of mountains containing list of cams
/api/mountains/<mt_id>/cams/<cam_id>/scrapes[?start=<datetime>&end=<datetime>]
    returns json list of scrape records
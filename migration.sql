/* 'open' old database */
ATTACH DATABASE 'mtcam_test.db' AS old;

/* 
convert mountain .
will have to manually add TzLocation to the new table.
*/
INSERT INTO mountain (rowid, created, modified, name, state, elevation_ft, latitude, longitude, tz_location)
    SELECT rowid, created, modified, name, state, elevation_ft, latitude, longitude, 'UTC'
    FROM old.mountain;

/* 
convert cam.
will have to manually add rules
and update url.
*/
INSERT INTO camera (rowid, created, modified, name, elevation_ft, latitude, longitude, url, file_ext, is_active, interval, comment, mountain_id, rules)
    SELECT rowid, created, modified, name, elevation_ft, latitude, longitude, url, file_ext, is_active, interval, comment, mountain_id, ''
    FROM old.cam;

/* 
convert scraperecord (renamed 'scrape')
doesn't have any new/modified columns so
no need for manual intervention.
*/
INSERT INTO scrape (rowid, created, result, detail, filename, camera_id)
    SELECT rowid, created, result, detail, filename, cam_id 
    FROM old.scraperecord;

/* Done */
DETACH DATABASE old;
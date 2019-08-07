/* 'open' old database */
ATTACH DATABASE 'mtcam_test.db' AS old;

/* 
convert mountain .
will have to manually add TzLocation to the new table.
note: pathname set below. 
*/
INSERT INTO mountain (rowid, created, modified, name, state, elevation_ft, latitude, longitude, tz_location, pathname)
    SELECT rowid, created, modified, name, state, elevation_ft, latitude, longitude, 'UTC', ''
    FROM old.mountain;

/* 
convert cam.
will have to manually add rules, delay,
and update url.
note: pathname set below.
*/
INSERT INTO camera (rowid, created, modified, name, elevation_ft, latitude, longitude, url, file_ext, is_active, interval, delay, comment, mountain_id, rules, pathname)
    SELECT rowid, created, modified, name, elevation_ft, latitude, longitude, url, file_ext, is_active, interval, 0, comment, mountain_id, '', ''
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

/* set pathname for mountain and camera */
UPDATE mountain SET pathname=lower(replace(name,' ','_'));
UPDATE camera SET pathname=lower(replace(name,' ','_'));
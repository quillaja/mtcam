/*
migrate from v1.0 (python) to v2.0 (go)
*/

/* 'open' old database */
ATTACH DATABASE 'mtcam_test.db' AS old;

/* 
convert mountain .
will have to manually add TzLocation to the new table. (see below)
note: pathname set below. 
*/
INSERT INTO mountain (rowid, created, modified, name, state, elevation_ft, latitude, longitude, tz_location, pathname)
    SELECT rowid, created, modified, name, state, elevation_ft, latitude, longitude, 'UTC', ''
    FROM old.mountain;

/* 
convert cam.
will have to manually add rules, delay, (see below)
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
UPDATE mountain SET pathname=lower(replace(name,' ','_')||'_'||state);
UPDATE camera SET pathname=lower(replace(name,' ','_'));

/* update mountain timezones to new 'style' */
UPDATE mountain SET tz_location='America/Los_Angeles' WHERE rowid=1; /*hood*/
UPDATE mountain SET tz_location='America/Los_Angeles' WHERE rowid=2; /*sisters*/
UPDATE mountain SET tz_location='America/Los_Angeles' WHERE rowid=3; /*rainier*/
UPDATE mountain SET tz_location='Asia/Tokyo' WHERE rowid=4; /*fuji*/
UPDATE mountain SET tz_location='Europe/Paris' WHERE rowid=5; /*blanc*/
UPDATE mountain SET tz_location='America/Los_Angeles' WHERE rowid=6; /*shasta*/
UPDATE mountain SET tz_location='America/Los_Angeles' WHERE rowid=7; /*whitney*/
-- UPDATE mountain SET tz_location='' WHERE rowid=8; /*eiger*/

/* update camera rules, delay (url can stay same on current dataset) */
UPDATE camera SET delay=30, rules='{{ or (betweenRiseSet .Now .Astro 2) (brightMoon .Astro) }}' WHERE rowid=1; /*palmer(hood)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 1 }}' WHERE rowid=2; /*bachelor(sisters)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 1 }}' WHERE rowid=3; /*paradise(rainier)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 1 }}' WHERE rowid=4; /*north(fuji)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 0 }}' WHERE rowid=5; /*tourism(blanc)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 1 }}' WHERE rowid=6; /*vista(hood)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 1 }}' WHERE rowid=7; /*snowcrest(shasta)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 1 }}' WHERE rowid=8; /*schurman(rainier)*/
UPDATE camera SET delay=30, rules='{{ betweenRiseSet .Now .Astro 0 }}' WHERE rowid=9; /*lone pine(whitney)*/
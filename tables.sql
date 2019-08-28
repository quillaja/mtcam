CREATE TABLE IF NOT EXISTS "mountain" (
    -- rowid auto PK
    
    -- time mountain was added
    "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    -- time mountain was last modified
    "modified" DATETIME NOT NULL, 
    -- display name
    "name" TEXT NOT NULL, 
    -- US state or other political region (country)
    "state" TEXT NOT NULL, 
    -- elevation of peak in feet, coordinates
    "elevation_ft" INTEGER NOT NULL, 
    "latitude" REAL NOT NULL, 
    "longitude" REAL NOT NULL, 
    -- timezone database name eg "America/Los_Angeles"
    "tz_location" TEXT NOT NULL,
    -- string to use as pathname representation. eg 'mt_hood_or'
    -- must be unique, but not adding that constraint since it's kinda a pain
    "pathname" TEXT NOT NULL DEFAULT '');

CREATE TABLE IF NOT EXISTS "camera" (
    -- rowid auto PK

    -- time camera was created and last modified
    "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    "modified" DATETIME NOT NULL, 
    -- camera display name
    "name" TEXT NOT NULL, 
    -- elevation of camera in feet, coordinates
    "elevation_ft" INTEGER NOT NULL, 
    "latitude" REAL NOT NULL, 
    "longitude" REAL NOT NULL, 
    -- go text template evaluating to an URL for the camera image
    "url" TEXT NOT NULL,
    -- file extention (eg 'jpg') of image, NO period
    "file_ext" TEXT NOT NULL, 
    -- main camera on/off switch
    "is_active" BOOLEAN NOT NULL,
    -- time in minutes between scrapes 
    "interval" INTEGER NOT NULL, 
    -- number of sec to wait before actually scraping
    "delay" INTEGER NOT NULL DEFAULT 0,
    -- go text template evaluating to True/False which determines if the
    -- camera should be scraped at a particular time
    "rules" TEXT NOT NULL,
    -- notes, etc
    "comment" TEXT NOT NULL DEFAULT '',
    -- string to use as pathname representation. eg 'palmer'
    -- must be unique, but not adding that constraint since it's kinda a pain
    "pathname" TEXT NOT NULL DEFAULT '',
    -- FK to mountain
    "mountain_id" INTEGER NOT NULL, 
    FOREIGN KEY ("mountain_id") REFERENCES "mountain" ("rowid"));

CREATE INDEX "camera_mountain_id" ON "camera" ("mountain_id");

CREATE TABLE IF NOT EXISTS "scrape" (
    -- rowid auto PK

    -- time this scrape was performed
    "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    -- result such as 'success', 'failure', etc.
    "result" TEXT NOT NULL, 
    -- details relating to the result
    "detail" TEXT NOT NULL DEFAULT '', 
    -- filename of image on disk with extension.
    -- filename is UTC timestamp (eg 1565257200.jpg)
    "filename" TEXT NOT NULL, 
    -- FK to camera
    "camera_id" INTEGER NOT NULL, 
    FOREIGN KEY ("camera_id") REFERENCES "camera" ("rowid"));
    
CREATE INDEX "scrape_camera_id" ON "scrape" ("camera_id");

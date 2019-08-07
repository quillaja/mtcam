CREATE TABLE IF NOT EXISTS "mountain" (
    -- rowid auto PK
    "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    "modified" DATETIME NOT NULL, 
    "name" TEXT NOT NULL, 
    "state" TEXT NOT NULL, 
    "elevation_ft" INTEGER NOT NULL, 
    "latitude" REAL NOT NULL, 
    "longitude" REAL NOT NULL, 
    "tz_location" TEXT NOT NULL, -- will only contain tz name eg "Los_Angeles/America"
    "pathname" TEXT NOT NULL DEFAULT ''); -- must be unique, but not adding that constraint since it's kinda a pain

CREATE TABLE IF NOT EXISTS "camera" (
    -- rowid auto PK
    "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    "modified" DATETIME NOT NULL, 
    "name" TEXT NOT NULL, 
    "elevation_ft" INTEGER NOT NULL, 
    "latitude" REAL NOT NULL, 
    "longitude" REAL NOT NULL, 
    "url" TEXT NOT NULL, -- will be text template 
    "file_ext" TEXT NOT NULL, 
    "is_active" BOOLEAN NOT NULL, 
    "interval" INTEGER NOT NULL, 
    "delay" INTEGER NOT NULL DEFAULT 0, -- number of sec to wait before actually scraping
    "rules" TEXT NOT NULL, -- text template evaluating to True/False
    "comment" TEXT NOT NULL DEFAULT '',
    "pathname" TEXT NOT NULL DEFAULT '', -- must be unique
    "mountain_id" INTEGER NOT NULL, 
    FOREIGN KEY ("mountain_id") REFERENCES "mountain" ("rowid"));

CREATE INDEX "camera_mountain_id" ON "camera" ("mountain_id");

CREATE TABLE IF NOT EXISTS "scrape" (
    -- rowid auto PK
    "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    "result" TEXT NOT NULL, 
    "detail" TEXT NOT NULL DEFAULT '', 
    "filename" TEXT NOT NULL, 
    "camera_id" INTEGER NOT NULL, 
    FOREIGN KEY ("camera_id") REFERENCES "camera" ("rowid"));
    
CREATE INDEX "scrape_camera_id" ON "scrape" ("camera_id");

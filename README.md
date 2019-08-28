# Mountain Cameras
Enhanced version of "palmer" which is database backed and can scrape multiple webcams.
This new version (v2.0) is in Go, with a few added enhancements over the previous Python version (v1.0).

## Installation

By default installs 2 binaries, `scraped` and `served`, to `/opt/mtcam`. No database or 
config files are created.

    $ git clone https://github.com/quillaja/mtcam.git
    $ cd mtcam
    $ make build
    $ make install
    $ make service-install

A database can be created with `$ sqlite3 YOUR_DATABASE '.read tables.sql'`.

## Usage

`scraped` and `served` both take 1 required flag: `-cfg PATH_TO_CONFIG`. If `-cfg default` is used,
a blank config file for the binary is written to disk alonside the binary.
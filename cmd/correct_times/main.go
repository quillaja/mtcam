package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var pacifictz *time.Location

func init() {
	pacifictz, _ = time.LoadLocation("America/Los_Angeles")
}

func main() {
	dbstring := os.Args[1]
	fmt.Println("converting ", dbstring)

	db, err := sql.Open("sqlite3", dbstring)
	kill(err)

	fmt.Println("converting scape table (takes about 30 mins)")
	convertScrape(db)

	fmt.Println("finished converting")
	db.Close()
}

// took over 45 mins at which point I killed it.
// probably because the final UPDATE is O(n²).
/*func convertScrapeUltraSlow(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE utcscrape (
		"rowid" INTEGER,
		"created" DATETIME)`)
	kill(err)
	defer db.Exec("DROP TABLE utcscrape")

	rows, err := db.Query("SELECT rowid, created FROM scrape")
	kill(err)

	var rowid int
	var created time.Time
	var insert strings.Builder
	insert.WriteString("INSERT INTO utcscrape (rowid, created) VALUES ")
	pacifictz, _ := time.LoadLocation("America/Los_Angeles")

	for rows.Next() {
		err = rows.Scan(&rowid, &created)
		kill(err)

		created = time.Date(created.Year(), created.Month(), created.Day(),
			created.Hour(), created.Minute(), created.Second(), 0,
			pacifictz)

		createdstring := created.In(time.UTC).Format(sqlite3.SQLiteTimestampFormats[0])
		insert.WriteString(fmt.Sprintf("(%d, '%s'), ", rowid, createdstring))
	}
	rows.Close()

	fmt.Println("\tbuild INSERT")
	_, err = db.Exec(insert.String()[:insert.Len()-2])
	kill(err)

	fmt.Println("\tstarting UPDATE")
	_, err = db.Exec(
		`UPDATE scrape
	SET created = (SELECT created
					FROM utcscrape
					WHERE utcscrape.rowid=scrape.rowid)`)
	kill(err)
}*/

// takes about 30 mins to run over 500,000 rows. Is about O(n).
func convertScrape(db *sql.DB) {
	const printmod = 100

	rows, err := db.Query("SELECT rowid, created FROM scrape")
	kill(err)

	// get all rows
	type scrape struct {
		rowid   int
		created time.Time
	}
	var s scrape
	allScrapes := []scrape{}
	for rows.Next() {
		err = rows.Scan(&s.rowid, &s.created)
		// &s.ID,
		// &s.Created,
		// &s.Result,
		// &s.Detail,
		// &s.Filename,
		// &s.CameraID)
		kill(err)

		if s.rowid%printmod == 0 {
			fmt.Println("read rowid", s.rowid)
		}

		// convert the created time, first by 'labeling' it Pacific Time,
		// then converting to UTC
		s.created = changeToUTC(s.created, pacifictz)

		allScrapes = append(allScrapes, s)
	}
	rows.Close()
	fmt.Println("read", len(allScrapes), "rows")

	// update the rows
	const update = `UPDATE scrape
		SET
			rowid = ?,
			created = ?
		WHERE
			rowid = ?`
	fmt.Println("updating")

	for _, s := range allScrapes {
		if s.rowid%printmod == 0 {
			fmt.Println("updating rowid", s.rowid)
		}

		_, err = db.Exec(update, s.rowid, s.created, s.rowid)
		kill(err)
	}
}

func convertCamera(db *sql.DB) {
}

func convertMountain(db *sql.DB) {
}

func kill(err error) {
	if err != nil {
		panic(err)
	}
}

func changeToUTC(t time.Time, current *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), 0,
		current).In(time.UTC)
}

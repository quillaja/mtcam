// special program crated to convert the times in the mtcam database
// from "unlabeled" pacific timezone to labeled UTC (or other) timezone.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	pacifictz, _ := time.LoadLocation("America/Los_Angeles")

	dbstring := os.Args[1]
	fmt.Printf("converting %s\n\n", dbstring)

	db, err := sql.Open("sqlite3", dbstring)
	kill(err)

	fmt.Println("converting mountain table")
	convert(db, pacifictz, time.UTC, "mountain", "created", "modified")

	fmt.Println("converting camera tables")
	convert(db, pacifictz, time.UTC, "camera", "created", "modified")

	fmt.Println("converting scape table (takes about 30 mins)")
	convert(db, pacifictz, time.UTC, "scrape", "created")

	fmt.Println("finished converting")
	db.Close()
}

// converts table containing time columns from fromTz to toTz.
// takes about 30 mins to run over 500,000 rows. Is about O(n).
func convert(db *sql.DB, fromTz, toTz *time.Location, table string, columns ...string) {
	const idmod = 100

	selectQuery := fmt.Sprintf("SELECT rowid, %s FROM %s", strings.Join(columns, ", "), table)
	fmt.Println("running:", selectQuery)
	rows, err := db.Query(selectQuery)
	kill(err)

	// get all rows
	// set up a couple of maps and arrays to hold the data
	// and also in which to Scan() the data
	allData := map[int][]time.Time{}
	var id int
	timecols := make([]time.Time, len(columns))
	pointers := make([]interface{}, 1+len(columns))
	pointers[0] = &id
	for i := 1; i < len(pointers); i++ {
		pointers[i] = &timecols[i-1]
	}

	// do the scanning
	for rows.Next() {

		err = rows.Scan(pointers...)
		kill(err)

		// status display
		if id%idmod == 0 || id == 1 {
			fmt.Printf("\r read rowid %-10d\r", id)
		}

		// convert the created time, first by 'labeling' it fromTz,
		// then converting to toTz
		for i := range timecols {
			timecols[i] = changeTz(timecols[i], fromTz, toTz)
		}

		// store the time columns associated with id
		allData[id] = append([]time.Time{}, timecols...)
	}
	rows.Close()
	fmt.Printf(" read %d rows           \n", len(allData))

	// update the rows
	updateQuery := fmt.Sprintf(
		"UPDATE %s SET rowid = ?, %s = ? WHERE rowid = ?",
		table,
		strings.Join(columns, " = ?, "))
	fmt.Println("running:", updateQuery)

	values := make([]interface{}, 2+len(columns))
	var updates int
	var total, done, prevdone float64
	total = float64(len(allData))

	for id, times := range allData {
		// superfluous "status" display
		done = 100 * float64(1+updates) / total
		if done-prevdone >= 0.1 {
			fmt.Printf("\r %.1f%% complete\r", done)
			prevdone = done
		}

		// fill array with the values for the row
		values[0] = id
		values[len(values)-1] = id
		for i := 1; i < len(values)-1; i++ {
			values[i] = times[i-1]
		}

		_, err = db.Exec(updateQuery, values...)
		// time.Sleep(1 * time.Millisecond) // simulate update in testing
		kill(err)

		updates++
	}

	fmt.Printf("\nfinished converting %s. %d rows updated.\n\n", table, updates)
}

func kill(err error) {
	if err != nil {
		panic(err)
	}
}

func changeTz(t time.Time, from, to *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), 0, from).In(to)
}

package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var count int64

func main() {
	customColsPtr := flag.String("cols", "", "Custom column definition: i.e. \"col1,col2,col3\"")
	commitRatePtr := flag.Int64("crate", 10000, "How many transactions between SQL commits.\nNote: Higher values may be faster but use more RAM")
	dbPath := flag.String("out", "./sqlite.db", "sqliteDB output path.")
	table := flag.String("table", "data", "table name for CSV data")
	delimiterPrt := flag.String("delimiter", ",", "Custom delimiter")
	flag.Parse()
	filePath := flag.Arg(0)
	if len(filePath) == 0 {
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("A CSV Input file is required.")
		return
	}

	file, err := os.Open(filePath)
	ckErrFatal(err, "failed to open file @ "+filePath)

	scanner := bufio.NewScanner(file)
	// get categories
	var cols []string
	if *customColsPtr != "" {
		cols = strings.Split(*customColsPtr, *delimiterPrt)
	} else { // read cols from first line
		scanner.Scan()
		cols = strings.Split(scanner.Text(), *delimiterPrt)
	}

	if len(cols) <= 1 {
		log.Fatalln("No columns found.")
	}

	fmt.Println("cols found:", len(cols), cols)
	sqlTx, database := initDB(*dbPath, *table, cols...)
	defer database.Close()
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), *delimiterPrt)
		stmt, err := sqlTx.Prepare(genInsertStr(*table, cols...))
		ckErrFatal(err, "failed to insert prepared statement")
		dataArgs := make([]interface{}, len(data))
		for i, v := range data {
			dataArgs[i] = v
		}
		for i := len(dataArgs) + 1; i < len(cols); i++ {
			dataArgs[i] = ""
		}
		_, err = stmt.Exec(dataArgs...)
		ckErrFatal(err, "failed to execute prepared statement")
		count++
		fmt.Printf("\r%d", count)

		// For memory management purposes we split our commits
		// this likely has a performance penalty but will use far less RAM
		if count%*commitRatePtr == 0 {
			ckErrFatal(sqlTx.Commit(), "failed to commit sql (not final commit)")
			ckErrFatal(database.Close(), "failed to close database (mid-commit)")
			database, err = sql.Open("sqlite3", *dbPath)
			ckErrFatal(err, "failed to reopen DB after mid-process commit")
			_, err = database.Exec(`PRAGMA shrink_memory;`)
			ckErrFatal(err, "failed to shrink memory")
			sqlTx, err = database.Begin()
			ckErrFatal(err, "could not Begin database")
		}
	}

	ckErrFatal(sqlTx.Commit(), "failed to commit sql")
}

func genInsertStr(table string, columns ...string) string {
	if len(columns) < 1 {
		log.Fatalln("genInsertStr() requires at least one column")
	}
	outStr := "INSERT into \"" + table + "\" ("
	for _, value := range columns {
		outStr += "\"" + value + "\", "
	}
	outStr = outStr[0:len(outStr)-2] + ") VALUES ("
	for i := 0; i < len(columns); i++ {
		outStr += "?, "
	}
	//fmt.Println(outStr[0:len(outStr)-2] + ");")
	return outStr[0:len(outStr)-2] + ");"
}

func initDB(path, table string, columns ...string) (*sql.Tx, *sql.DB) {
	if len(columns) == 0 {
		log.Fatalln("at least one colum required to initDB")
	}

	db, err := sql.Open("sqlite3", path)
	ckErrFatal(err, "Could not open sqlite3 db @ "+path)

	ddl := `
	       PRAGMA automatic_index = ON;
	       PRAGMA cache_size = 32768;
	       PRAGMA cache_spill = OFF;
	       PRAGMA foreign_keys = ON;
	       PRAGMA journal_size_limit = 67110000;
	       PRAGMA locking_mode = NORMAL;
	       PRAGMA page_size = 4096;
	       PRAGMA recursive_triggers = ON;
	       PRAGMA secure_delete = OFF;
	       PRAGMA synchronous = OFF;
	       PRAGMA temp_store = MEMORY;
	       PRAGMA journal_mode = OFF;
	       PRAGMA wal_autocheckpoint = 16384;
	       CREATE TABLE IF NOT EXISTS `
	ddl += "\"" + table + "\" ("
	for _, col := range columns {
		ddl += "\"" + col + "\" TEXT NOT NULL, "
	}
	ddl = strings.TrimSuffix(ddl, ", ") + ");"

	_, err = db.Exec(ddl)
	ckErrFatal(err, "failed to exec ddl template"+ddl)

	tx, err := db.Begin()
	ckErrFatal(err, "failed to begin db")

	return tx, db
}

func ckErrFatal(err error, reason string) {
	if err != nil {
		log.Fatalln(reason+":", err)
	}
}

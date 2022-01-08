package main

import (
	"flag"
	"fmt"
	"os"

	"golang.binggl.net/monorepo/internal/bookmarks/store"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// migrate the existing data of the mysql bookmarks db into a new sqlite db

func main() {
	var mysqlConStr, sqliteConStr string
	mysqlConStr = os.Getenv("MYSQL_DSN")
	sqliteConStr = os.Getenv("SQLITE_DSN")

	if mysqlConStr == "" || sqliteConStr == "" {
		flag.Usage()
		panic("Missing required parameters!\n")
	}

	mysqlCon, err := gorm.Open(mysql.Open(mysqlConStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection to mysql: %v", err))
	}
	fmt.Printf("established connection to MYSQL\n")

	sqliteCon, err := gorm.Open(sqlite.Open(sqliteConStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection to sqlite: %v", err))
	}
	fmt.Printf("established connection to SQLITE\n")

	defer func() {
		db, _ := mysqlCon.DB()
		db.Close()

		db, _ = sqliteCon.DB()
		db.Close()
	}()

	// Ensure the schema creation for sqlite
	sqliteCon.AutoMigrate(&store.Bookmark{})

	var items []store.Bookmark
	h := mysqlCon.Raw("SELECT * FROM BOOKMARKS").Scan(&items)
	if h.Error != nil {
		panic(fmt.Sprintf("could not get data from mysql DB: %v\n", h.Error))
	}
	fmt.Printf("got %d entries from MYSQL!", len(items))
	fmt.Printf("will store the entries in the sqlite-DB!\n")
	for _, item := range items {
		result := sqliteCon.Create(&item)
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "could not save bookmark entry to SQLITE; %v\n", result.Error)
		}
	}
	fmt.Printf("migrated data to the sqlite-DB!\n")
}

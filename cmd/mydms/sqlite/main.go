package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// define a document entity for GORM

type UserSiteEntity struct {
	Name      string    `gorm:"primaryKey;TYPE:varchar(128);COLUMN:name;NOT NULL;INDEX:IX_USERSITE_PK"`
	User      string    `gorm:"primaryKey;TYPE:varchar(128);COLUMN:user;NOT NULL;INDEX:IX_USERSITE_PK"`
	URL       string    `gorm:"TYPE:varchar(256);COLUMN:url;NOT NULL"`
	PermList  string    `gorm:"TYPE:varchar(256);COLUMN:permission_list;NOT NULL"`
	CreatedAt time.Time `gorm:"COLUMN:created;NOT NULL"`
}

type documentEntity struct {
	ID            string         `gorm:"primaryKey;TYPE:varchar(36);COLUMN:id;NOT NULL;INDEX:IX_DOCUMENTS_PK"`
	Title         string         `gorm:"TYPE:varchar(255);COLUMN:title;NOT NULL"`
	FileName      string         `gorm:"TYPE:varchar(255);COLUMN:filename;NOT NULL"`
	AltID         string         `gorm:"TYPE:varchar(128);COLUMN:alternativeid;"`
	PreviewLink   sql.NullString `gorm:"TYPE:varchar(128);COLUMN:previewlink;"`
	Amount        float32        `gorm:"TYPE:decimal(10,0);COLUMN:amount;"`
	Created       time.Time      `gorm:"TYPE:date;COLUMN:created;NOT NULL;autoCreateTime"`
	Modified      sql.NullTime   `gorm:"TYPE:date;COLUMN:modified;"`
	TagList       string         `gorm:"TYPE:text;COLUMN:taglist;"`
	SenderList    string         `gorm:"TYPE:text;COLUMN:senderlist;"`
	InvoiceNumber sql.NullString `gorm:"TYPE:varchar(128);COLUMN:invoicenumber;"`
}

func (d documentEntity) String() string {
	return fmt.Sprintf("documentEntity: '%s,%s'", d.ID, d.Title)
}

// TableName specifies the name of the Table used
func (documentEntity) TableName() string {
	return "DOCUMENTS"
}

// migrate the existing data of the mysql mydms-documents db into a new sqlite db

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
	sqliteCon.AutoMigrate(&documentEntity{})

	var items []documentEntity
	h := mysqlCon.Raw("SELECT * FROM DOCUMENTS").Scan(&items)
	if h.Error != nil {
		panic(fmt.Sprintf("could not get data from mysql DB: %v\n", h.Error))
	}
	fmt.Printf("got %d entries from MYSQL!\n", len(items))
	fmt.Printf("will store the entries in the sqlite-DB!\n")
	for _, item := range items {
		result := sqliteCon.Create(&item)
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "could not save bookmark entry to SQLITE; %v\n", result.Error)
		}
	}
	fmt.Printf("migrated data to the sqlite-DB!\n")
}

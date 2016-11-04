package database

import (
	"database/sql"
	// Database driver
	_ "github.com/mattn/go-sqlite3"
	"log"

	"github.com/Skarlso/google-oauth-go-sample/structs"
)

// Save saves the passed in Character.
func Save(c *structs.Character) error {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		return err
	}
	defer db.Close()
	return nil
}

// InitDb creates an initial database file if it's not present yet.
func InitDb() error {
	db, err := sql.Open("sqlite3", "foo.db")
	if err != nil {
		log.Println("Error occured in the database layer.")
		return err
	}
	defer db.Close()
	return nil
}

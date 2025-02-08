package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// New initializes a new SQLite database connection using GORM.
func New(dbPath string) (*gorm.DB, error) {
	// Open a SQLite database connection.
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

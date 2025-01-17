package sqllite

import (
	"database/sql"
	"fmt"
)

type Database struct {
	db *sql.DB
}

func New(dbPath string) (*Database, error) {
	const mark = "database.sqllite.New"

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", mark, err)
	}

	return &Database{db: db}, nil
}

package sqllite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/FacelessWayfarer/urlshortner/internal/database"
	"modernc.org/sqlite"
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

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s:%w", mark, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s:%w", mark, err)
	}

	return &Database{db: db}, nil
}

func (d *Database) SaveURL(longURL, alias string) (int64, error) {
	const mark = "database.sqllite.SaveURL"

	stmt, err := d.db.Prepare("INSERT INTO url(url,alias) VALUES(?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s:%w", mark, err)
	}

	rst, err := stmt.Exec(longURL, alias)
	if err != nil {
		sqliteErr := err.(*sqlite.Error)
		if sqliteErr.Code() == 2067 {
			return 0, fmt.Errorf("%s: %w", mark, database.ErrURLAlreadyExists)
		}
		return 0, fmt.Errorf("%s:%w", mark, err)
	}
	id, err := rst.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s:%w", mark, err)
	}

	return id, nil

}

func (d *Database) GetURL(alias string) (url string, err error) {
	const mark = "database.sqllite.GetURL"

	stmt, err := d.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s:%w", mark, err)
	}

	err = stmt.QueryRow(alias).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", database.ErrURLNotFound
		}
		return "", fmt.Errorf("%s:%w", mark, err)
	}

	return url, nil
}

func (d *Database) DeleteURL(alias string) error {
	const mark = "database.sqllite.DeleteURL"

	stmt, err := d.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s:%w", mark, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s:%w", mark, err)
	}
	return nil
}

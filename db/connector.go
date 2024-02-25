package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"

	log "github.com/sirupsen/logrus"
)

type DBConfig struct {
	UID   string
	_path string
}

type DB struct {
	DB     *sql.DB
	DBPath string
}

func NewDBConfig(UID string) (*DBConfig, error) {
	return &DBConfig{
		UID:   UID,
		_path: fmt.Sprintf("%s/%s.db", os.Getenv("FPBASE"), UID),
	}, nil
}

func (c *DBConfig) Connect() (DB, error) {
	if _, err := os.Stat(c._path); err == nil {
		//DB Exists. Directly Load from DB file
		db, err := sql.Open("sqlite3", c._path)
		if err != nil {
			log.Fatal("Failed to connect to SQLite DB: ", err)
			return DB{}, err
		}
		return DB{DB: db, DBPath: c._path}, nil
	} else {
		return DB{}, fmt.Errorf("DB Doesn't exist yet")
	}
}

func (c *DBConfig) Create() (DB, error) {
	//create by opening connection, sqlite does it by default
	db, err := sql.Open("sqlite3", c._path)
	if err != nil {
		log.Fatal("Failed to Create SQLite DB: ", err)
		return DB{}, err
	}
	return DB{DB: db, DBPath: c._path}, nil
}

func (db *DB) CreateNewBasket(bid string) (bool, error) {
	query := `
		CREATE TABLE %s (
		date integer,
		position text,
		trigger integer,
		etime integer,
		entry real,
		xtime integer,
		exit real,
		result real
	);
	`
	query = fmt.Sprintf(query, bid)
	_, err := db.DB.Exec(query)
	if err != nil {
		log.Error("Error in creating basket table")
		return false, err
	}
	return true, nil
}

func (db *DB) LoadOutput(bid string) (bool, error) {
	// path := fmt.Sprintf("%s/%s.csv", db.DBPath, bid)
	path := "/home/hakaitech/data/ZKxPLCKEXSQDm6eGaQtM4o.csv"
	cmd := exec.Command("sqlite3", db.DBPath, ".mode csv", fmt.Sprintf(".import %s %s", path, bid))
	var (
		stderr, out bytes.Buffer
	)
	cmd.Stderr = &stderr
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Error(err, stderr.String(), out.String())
	}
	return true, nil
}

package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"sync"

	_ "github.com/lib/pq"
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

type Singleton struct {
	Db *sql.DB
	/*
		Read *sql.DB
		Write *sql.DB
	*/
	orgsCnx map[string]*sql.DB
}

type Config struct {
	// Hostname
	Host string
	// Port
	Port string
	// User
	User string
	// Password
	Pwd string
	// Database Name
	Name string
}

type Cnx struct {
	Cnx   *sql.DB
	Error error
}

var instance *Singleton
var once sync.Once

func NewDBConfig(UID string) (*DBConfig, error) {
	return &DBConfig{
		UID:   UID,
		_path: fmt.Sprintf("%s/%s.db", os.Getenv("UDB"), UID),
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
		db, err := c.Create()
		if err != nil {
			log.Error("DB DNE, Creation Failed")
			return DB{}, err
		}
		return db, nil
	}
}

func (c *DBConfig) Create() (DB, error) {
	//create by opening connection, sqlite does it by default
	db, err := sql.Open("sqlite3", c._path)
	if err != nil {
		log.Fatal("Failed to Create SQLite DB: ", err)
		return DB{}, err
	}
	// out := DB{DB: db, DBPath: c._path}
	// _, err = out.InitUserDB()
	// if err != nil {
	// 	log.Error("Can't init udb: ", out)
	// 	return DB{}, nil
	// }
	return DB{DB: db, DBPath: c._path}, nil
}

func (db *DB) CreateNewBasket(bid string) (bool, error) {
	initquery := "INSERT INTO baskets (bid,name,opfp,status) VALUES('%s','%s','%s','%s')"
	name := bid
	initquery = fmt.Sprintf(initquery, bid, name, "NA", "OP")
	_, err := db.DB.Exec(initquery)
	if err != nil {
		log.Error("Error in inserting into userdb: ", initquery)
		return false, err
	}
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
	query = fmt.Sprintf(query, name)
	// log.Debug(query)
	_, err = db.DB.Exec(query)
	if err != nil {
		log.Error("Error in creating basket table: ", query, err)
		return false, err
	}
	return true, nil
}

func (db *DB) InitUserDB() (bool, error) {
	query := `
		CREATE TABLE %s (
		bid varchar,
		name varchar,
		opfp varchar,
		status varchar
	);
	`
	query = fmt.Sprintf(query, "baskets")
	_, err := db.DB.Exec(query)
	if err != nil {
		log.Error("Error in creating basket table")
		return false, err
	}
	return true, nil
}

func (db *DB) LoadOutput(bid string) (bool, error) {
	path := fmt.Sprintf("%s/%s.csv", os.Getenv("FPBASE"), bid)
	// path := "/home/hakaitech/data/ZKxPLCKEXSQDm6eGaQtM4o.csv"
	name := bid
	cmd := exec.Command("sqlite3", db.DBPath, ".mode csv", fmt.Sprintf(".import --skip 1 %s %s", path, name))
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

func ConnectToDB(config *Config) chan Cnx {
	ch := make(chan Cnx)
	go func() {
		defer close(ch)
		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Pwd, config.Name)
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Error("Error in DB Connection.")
			ch <- Cnx{Error: err}
			return
		}

		// db.SetConnMaxLifetime(-1)
		_, err = db.Exec("SET datestyle = \"SQL, DMY\"")
		if err != nil {
			log.Error(err)
		}
		ch <- Cnx{Cnx: db}
	}()

	return ch
}

// Returns a Singleton instance that keeps the connections to the Database
func GetInstance() *Singleton {
	once.Do(func() {
		// Get the  Database configuration
		config := &Config{
			Host: os.Getenv("PSQL_HOST"),
			Port: "5432",
			User: os.Getenv("PSQL_USER"),
			Pwd:  os.Getenv("PSQL_PWD"),
			Name: os.Getenv("PSQL_NAME"),
		}
		cnxResult := <-ConnectToDB(config)

		//build connections cache
		cnxCache := make(map[string]*sql.DB)
		instance = &Singleton{
			Db:      cnxResult.Cnx,
			orgsCnx: cnxCache,
		}
	})
	return instance
}

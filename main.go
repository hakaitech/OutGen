package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/stock-jarvis/OutGen/cmd"
	"github.com/stock-jarvis/OutGen/db"
)

func init() {
	cmd := exec.Command("apt", "update")
	var (
		stderr, out bytes.Buffer
	)
	cmd.Stderr = &stderr
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Error(err, stderr.String(), out.String())
	}
	cmd = exec.Command("apt", "upgrade", "-y")
	cmd.Stderr = &stderr
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Error(err, stderr.String(), out.String())
	}
	cmd = exec.Command("apt", "install", "-y", "sqlite3", "libsqlite3-dev")
	cmd.Stderr = &stderr
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Error(err, stderr.String(), out.String())
	}
	// fmt.Println("GOMAXPROCS Set to: ", runtime.GOMAXPROCS(int(15)))
	err = godotenv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load env ")
		os.Exit(0)
	}
	fmt.Println("env: \nFile Path Source DIR: ", os.Getenv("FPBASE"), "\n UserDB Source DIR: ", os.Getenv("UDB"), "\n Reports DIR: ", os.Getenv("OPBASE"))
	fmt.Println("Initializing Logs....")
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
		PadLevelText:  true,
	})
	log.SetLevel(log.DebugLevel)
	log.Debug("Debug [x]")
	log.Info("Info [x]")
	log.Error("Error [x]")
}

func main() {
	config := &db.Config{
		Host: os.Getenv("PSQL_HOST"),
		Port: "5432",
		User: os.Getenv("PSQL_USER"),
		Pwd:  os.Getenv("PSQL_PWD"),
		Name: os.Getenv("PSQL_NAME"),
	}
	cnx := <-db.ConnectToDB(config)
	if cnx.Error == nil {
		log.Info("DB Conn Successful")
	} else {
		log.Fatal(cnx.Error)
	}

	go func() {
		for {
			err := cnx.Cnx.Ping()
			if err == nil {
				log.Info("DB Health Check Success")
			} else {
				log.Warn("DB Health Check Failed")
			}
			time.Sleep(time.Hour * 3)
		}
	}()

	log.Println("Starting up http server and registering routes: ")
	r := gin.Default()
	webhooks := r.Group("/webhooks")
	webhooks.GET("/stsup", cmd.CheckStatus)
	webhooks.GET("/report", cmd.Report)
	webhooks.GET("/initUser", cmd.InitUDB)
	webhooks.GET("/cnb", cmd.CreateNewBasket)
	webhooks.GET("/loadop", cmd.LoadOP)
	r.Run(os.Getenv("WEBHOOK_PORT"))

}

package cmd

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stock-jarvis/OutGen/db"
	"github.com/stock-jarvis/OutGen/metadata"
)

func Report(ctx *gin.Context) {
	uid, ok := ctx.GetQuery("cid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	bid, ok := ctx.GetQuery("bid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// log.Debug("Starting Report Generation for: ", bid)
	out := metadata.GenerateReport(uid, bid)
	ctx.JSON(http.StatusAccepted, out)
}

func InitUDB(ctx *gin.Context) {
	uid, ok := ctx.GetQuery("cid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	config, err := db.NewDBConfig(uid)
	if err != nil {
		log.Debug("Error in generating config of DB: ", uid)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	db, err := config.Connect()
	if err != nil {
		log.Debug("Error in Connect using config: ", config)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	_, err = db.InitUserDB()
	if err != nil {
		log.Debug("Error in InitDB: ", db)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "created"})
}

func CheckStatus(ctx *gin.Context) {
	bid, ok := ctx.GetQuery("bid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	globalDB := db.GetInstance().Db
	row := globalDB.QueryRow("SELECT status from basket.baskets where bid = $1", bid)
	var status string
	err := row.Scan(&status)
	if err != nil {
		log.Error("Error in read in status from DB: ", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
	ctx.JSON(http.StatusOK, gin.H{"Status": status})
}

func CreateNewBasket(ctx *gin.Context) {
	uid, ok := ctx.GetQuery("cid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	bid, ok := ctx.GetQuery("bid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	config, err := db.NewDBConfig(uid)
	if err != nil {
		log.Debug("Error in generating new config")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	db, err := config.Connect()
	if err != nil {
		log.Debug("Error in generating new DB Connection")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	_, err = db.CreateNewBasket(bid)
	if err != nil {
		log.Debug("Error in generating new basket table")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	_, err = db.DB.Exec("UPDATE baskets set status = 'Analysis Started' where bid = $1", bid)
	if err != nil {
		log.Debug("Error in status update")
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "created"})
}

func LoadOP(ctx *gin.Context) {
	uid, ok := ctx.GetQuery("cid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	bid, ok := ctx.GetQuery("bid")
	if !ok {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	config, err := db.NewDBConfig(uid)
	if err != nil {
		log.Debug("Error in generating new config")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	pdb, err := config.Connect()
	if err != nil {
		log.Debug("Error in generating new DB Connection")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	_, err = pdb.LoadOutput(bid)
	if err != nil {
		log.Debug("Error in generating new basket table")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	globalDB := db.GetInstance().Db
	_, err = globalDB.Exec("UPDATE basket.baskets set status = 'OG' where bid = $1", bid)
	if err != nil {
		log.Debug("Error in status update")
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "loaded"})

}

package main

import (
	"fmt"
	"time"

	"github.com/stock-jarvis/OutGen/db"
	"github.com/stock-jarvis/OutGen/metadata"
)

func main() {
	dbconfig, _ := db.NewDBConfig("TEST123", "/home/hakaitech/sandbox/src/OutGen/TEST123.db")
	var sdb db.DB
	var err error
	sdb, err = dbconfig.Connect()
	if err != nil && err.Error() == "DB Doesn't exist yet" {
		sdb, err = dbconfig.Create()
		if err != nil {
			return
		}
	}
	st := time.Now()
	sdb.CreateNewBasket("BasketTest2")
	sdb.LoadOutput("BasketTest2")
	fmt.Println(time.Since(st))
	md := sdb.FetchMetadata("BasketTest2")
	fmt.Println(metadata.TotalProfit(md))
	wd, ld, wp, lp := metadata.WinLossStats(md)
	fmt.Println("Win Percent/Loss Percent", wp, lp, "\n Avg ProfitWinLoss")
	fmt.Println(metadata.AvgProfitWinLoss(wd, ld, md))
	fmt.Println("WinStreak", metadata.CalcWStreak(md))
	fmt.Println("LossStreak", metadata.CalcLStreak(md))
	// fmt.Println(metadata.GetCumulativeProfit(md))
	fmt.Println(metadata.MDD(md))
}

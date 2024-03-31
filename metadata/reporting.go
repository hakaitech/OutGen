package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stock-jarvis/OutGen/db"
)

type Report struct {
	MetaData struct {
		TotalProfit          float64
		AverageDailyProfit   float64
		AverageMonthlyProfit float64
		AverageYearlyProfit  float64
		WinP                 float64
		LossP                float64
		WinDays              []int32
		LossDays             []int32
		AvgProfitWinDays     float32
		AvgLossLossDays      float32
		MaxDrawDown          float64
		MaxRecoveryTime      int32
		RTMDDR               float64
		MaxWinStreak         int
		MaxLossStreak        int
		Expectancy           float64
	}
	PositionData []db.PositionData
	ChartingArrs struct {
		CumulativeProfits []float64
		TradeDates        []int32
		DailyProfits      []float64
		DayWiseProfits    map[time.Weekday]float64
		MDs               []db.Metadata
	}
}

func GenerateReport(uid, bid string) Report {
	dbconfig, _ := db.NewDBConfig(uid)
	pdb, err := dbconfig.Connect()
	if err != nil {
		log.Error("Error in connecting to UserDB ", err, uid, bid)
		return Report{}
	}
	if s, err := pdb.LoadOutput(bid); err != nil || !s {
		log.Error("Error in loading output: ", err)
		return Report{}
	}
	// log.Info("Begining Metadata calculations")
	name := bid
	md, pd := pdb.FetchMetadata(name)
	tpnl, adpnl, ampnl, aypnl := TotalProfit(md)
	wd, ld, wp, lp := WinLossStats(md)
	wprof, lprof, _ := AvgProfitWinLoss(wd, ld, md)
	winStreak := CalcWStreak(md)
	lossStreak := CalcLStreak(md)
	cp := GetCumulativeProfit(md)
	mdd, mrp, rmddr := MDD(md, aypnl)
	// log.Debug("MD ready, starting report creation")
	report := Report{
		MetaData: struct {
			TotalProfit          float64
			AverageDailyProfit   float64
			AverageMonthlyProfit float64
			AverageYearlyProfit  float64
			WinP                 float64
			LossP                float64
			WinDays              []int32
			LossDays             []int32
			AvgProfitWinDays     float32
			AvgLossLossDays      float32
			MaxDrawDown          float64
			MaxRecoveryTime      int32
			RTMDDR               float64
			MaxWinStreak         int
			MaxLossStreak        int
			Expectancy           float64
		}{
			TotalProfit:          tpnl,
			AverageDailyProfit:   adpnl,
			AverageMonthlyProfit: ampnl,
			AverageYearlyProfit:  aypnl,
			WinP:                 wp,
			LossP:                lp,
			WinDays:              wd,
			LossDays:             ld,
			AvgProfitWinDays:     wprof,
			AvgLossLossDays:      lprof,
			MaxDrawDown:          mdd,
			MaxRecoveryTime:      mrp,
			RTMDDR:               rmddr,
			MaxWinStreak:         winStreak,
			MaxLossStreak:        lossStreak,
			Expectancy:           Expectancy(pd),
		},
		PositionData: pd,
		ChartingArrs: struct {
			CumulativeProfits []float64
			TradeDates        []int32
			DailyProfits      []float64
			DayWiseProfits    map[time.Weekday]float64
			MDs               []db.Metadata
		}{
			CumulativeProfits: cp,
			TradeDates:        GetTradeDays(md),
			DailyProfits:      GetDailyProfits(md),
			DayWiseProfits:    GetDayWiseProfit(md),
			MDs:               md,
		},
	}
	globalDB := db.GetInstance().Db
	_, err = globalDB.Exec("UPDATE basket.baskets set status = 'RG' where bid = $1 ", bid)
	if err != nil {
		log.Debug("Error in status update")
	}
	data, err := json.Marshal(report)
	if err != nil {
		log.Error("Error in marshalling report. ", err)
	}
	file, err := os.Create(fmt.Sprintf("/Data/Reports/%s.json", bid))
	if err != nil {
		log.Error(err)
		// log.Debug(data)
	}
	file.Write(data)
	file.Sync()

	return report
}

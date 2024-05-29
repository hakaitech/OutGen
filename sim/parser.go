package sim

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math"
	"sort"

	"github.com/stock-jarvis/OutGen/pkg"
	"github.com/stock-jarvis/OutGen/pkg/db"
)

type positionData struct {
	Ticker    string  `json:"ticker"`
	BS        string  `json:"bs"`
	Strike    int64   `json:"strike"`
	PosType   string  `json:"pos_type"`
	OptType   string  `json:"opt_type"`
	Expiry    string  `json:"expiry"`
	Lots      int64   `json:"lots"`
	LotSize   int64   `json:"lot_size"`
	PnL       float64 `json:"pnl"`
	ExitLTP   float64 `json:"entry_ltp"`
	EntryLTP  float64 `json:"exit_ltp"`
	Exited    int     `json:"exited"`
	EntryDate string  `json:"entry_date"`
	ExitDate  string  `json:"exit_date"`
	EntryTime string  `json:"entry_time"`
	ExitTime  string  `json:"exit_time"`
}

type Report struct {
	MetaData struct {
		TotalTrades        int
		TotalProfit        float64
		MaxDrawDown        float64
		MaxRecoveryTime    int32
		RTMDDR             float64
		MaxWinStreak       int
		MaxLossStreak      int
		Expectancy         float64
		AvgProfitOnWinDays float64
		AvgLossOnLossDays  float64
		AvgDayProfit       float64
		MaxProfit          float64
		MaxLoss            float64
		WinPercent         float64
		LossPercentage     float64
	}
	PositionData []positionData
	ChartingArrs struct {
		CumulativeProfits []float64
		TradeDates        []int32
	}
}

func SimReportRunner(name, uid string) (Report, map[string][]positionData) {
	data, err := parsePositionData(name, uid)
	if err != nil {
		log.Println(err)
		return Report{}, nil
	}
	sortedData, exitedPositions, nonexitedPosition, unrealizedProfit, realizedProfit := pollPositionData(data)
	mdd, mrp := MDD(sortedData)
	expectancy := Expectancy(sortedData)
	wStreak, wTrades, avgProfWTrades, maxProfit := CalcWStreak(sortedData)
	lStreak, lTrades, avgLossLTrades, maxLoss := CalcLStreak(sortedData)
	cumulativeProfit := CumulativeProfit(sortedData)
	newReport := Report{
		MetaData: struct {
			TotalTrades        int
			TotalProfit        float64
			MaxDrawDown        float64
			MaxRecoveryTime    int32
			RTMDDR             float64
			MaxWinStreak       int
			MaxLossStreak      int
			Expectancy         float64
			AvgProfitOnWinDays float64
			AvgLossOnLossDays  float64
			AvgDayProfit       float64
			MaxProfit          float64
			MaxLoss            float64
			WinPercent         float64
			LossPercentage     float64
		}{
			TotalProfit:        realizedProfit + unrealizedProfit,
			MaxDrawDown:        mdd,
			MaxRecoveryTime:    int32(mrp),
			RTMDDR:             mdd,
			MaxWinStreak:       wStreak,
			MaxLossStreak:      lStreak,
			Expectancy:         expectancy,
			TotalTrades:        lTrades + wTrades,
			AvgProfitOnWinDays: avgProfWTrades,
			AvgLossOnLossDays:  avgLossLTrades,
			MaxProfit:          maxProfit,
			MaxLoss:            maxLoss,
			WinPercent:         (float64(wTrades) / (float64(lTrades) + float64(wTrades))) * 100,
			LossPercentage:     (float64(lTrades) / (float64(lTrades) + float64(wTrades))) * 100,
		},
		PositionData: sortedData,
		ChartingArrs: struct {
			CumulativeProfits []float64
			TradeDates        []int32
		}{
			CumulativeProfits: cumulativeProfit,
		},
	}
	return newReport, map[string][]positionData{
		"unexitedPositions": nonexitedPosition,
		"exitedPositions":   exitedPositions,
	}
}

func parsePositionData(name, uid string) ([]positionData, error) {
	db := db.GetInstance().Db
	query := `
		SELECT data from tenants.simulations where cid = $1 AND name = $2
	`
	row := db.QueryRow(query, uid, name)
	var data string
	if err := row.Scan(&data); err != nil {
		return nil, err
	}
	pData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	var positionsData []positionData
	if err := json.Unmarshal(pData, &positionsData); err != nil {
		return nil, err
	}
	// fmt.Println(positionsData)
	return positionsData, nil
}

// returns in the order sortedData,exitedPositions, nonExitedPositions, realizedProfit, unrealizedProfit
func pollPositionData(data []positionData) ([]positionData, []positionData, []positionData, float64, float64) {
	var (
		exitedPositions    []positionData
		nonExitedPositions []positionData
		realizedProfit     float64
		unrealizedProfit   float64
		// totalPnL           []float64 //this will be the total pnl based off the array that is sorted by exit
	)

	sort.Slice(data, func(i, j int) bool {
		ts1, _ := pkg.GetTimestamp(data[i].ExitDate, data[i].ExitTime)
		ts2, _ := pkg.GetTimestamp(data[j].ExitDate, data[j].ExitTime)
		return ts1 < ts2
	})
	//split positions into exited and non-exited
	for _, position := range data {
		if position.Exited == 1 {
			exitedPositions = append(exitedPositions, position)
			realizedProfit += position.PnL
		} else {
			nonExitedPositions = append(nonExitedPositions, position)
			unrealizedProfit += position.PnL
		}
	}
	return data, exitedPositions, nonExitedPositions, realizedProfit, unrealizedProfit
	// TotalProfitLoss := realizedProfit + unrealizedProfit
	// mdd, mrp := MDD(data)
	// cp := CumulativeProfit(data)
}

func CumulativeProfit(data []positionData) []float64 {
	cp := make([]float64, len(data))
	for pos, pd := range data {
		if pos == 0 {
			cp[pos] = pd.PnL
			continue
		}
		cp[pos] = cp[pos-1] + pd.PnL
	}
	return cp
}

func MDD(data []positionData) (float64, float64) {
	mddpos := [][2]int{}
	for i := 0; i < len(data); {
		if data[i].PnL < 0 {
			pos := [2]int{i, 0}
			tempSum := data[i].PnL
			for j := i + 1; j < len(data); j++ {
				tempSum += data[j].PnL
				if tempSum >= 0 {
					pos[1] = j
					i = j + 1
					mddpos = append(mddpos, pos)
					break
				}
				if j == len(data)-1 {
					pos[1] = j
					mddpos = append(mddpos, pos)
					i = len(data)
					break
				}
			}
		} else {
			i += 1
		}
	}
	//we have all drawdown arrays.
	mdd := float64(0)
	mrp := int64(0)
	for _, ddp := range mddpos {
		ts0, _ := pkg.GetTimestamp(data[ddp[0]].ExitDate, data[ddp[0]].ExitTime)
		ts1, _ := pkg.GetTimestamp(data[ddp[1]].ExitDate, data[ddp[1]].ExitTime)
		rp := ts1 - ts0
		if rp > mrp {
			mrp = rp
		}
		dd := 0.0
		cump := []float64{}
		for i := ddp[0]; i <= ddp[1]; i++ {
			dd += float64(data[i].PnL)
			cump = append(cump, dd)
		}
		local_mdd := cump[0]
		for _, p := range cump {
			if p < local_mdd {
				local_mdd = p
			}
		}
		if local_mdd < mdd {
			mdd = local_mdd
		}
	}

	return mdd, float64(math.Floor(float64(mrp) / (60 * 60 * 24)))
}

func Expectancy(data []positionData) float64 {
	profitSum := 0.0
	profitCount := 0
	lossSum := 0.0
	lossCount := 0
	for _, pos := range data {
		if pos.PnL >= 0 {
			profitCount += 1
			profitSum += float64(pos.PnL)
		} else {
			lossCount += 1
			lossSum += float64(pos.PnL)
		}
	}
	lossSum = math.Abs(lossSum)
	avgProf := profitSum / float64(profitCount)
	avgLoss := lossSum / float64(lossCount)
	e1 := avgProf / avgLoss
	e2 := float64(profitCount) / float64(profitCount+lossCount)
	e3 := e1 * e2
	e4 := float64(lossCount) / float64(profitCount+lossCount)
	e5 := e3 - e4
	return e5
}

func CalcWStreak(data []positionData) (int, int, float64, float64) {
	maxStreak := 0
	streak := 0
	profit := 0.0
	winTrades := 0
	maxProfit := 0.0
	for _, m := range data {
		if m.PnL >= 0 {
			streak += 1
			profit += m.PnL
			winTrades += 1
			if m.PnL > maxProfit {
				maxProfit = m.PnL
			}
			if streak > maxStreak {
				maxStreak = streak
			}
		} else {
			streak = 0
		}
	}
	return maxStreak, winTrades, profit / float64(winTrades), maxProfit
}

func CalcLStreak(data []positionData) (int, int, float64, float64) {
	maxStreak := 0
	streak := 0
	loss := 0.0
	lossTrades := 0
	maxLoss := 0.0
	for _, m := range data {
		if m.PnL < 0 {
			streak += 1
			loss += m.PnL
			lossTrades += 1
			if m.PnL < maxLoss {
				maxLoss = m.PnL
			}
			if streak > maxStreak {
				maxStreak = streak
			}
		} else {
			streak = 0
		}
	}
	return maxStreak, lossTrades, loss / float64(lossTrades), maxLoss
}

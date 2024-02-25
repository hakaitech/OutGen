package metadata

import (
	"fmt"
	"sync"
	"time"

	"github.com/stock-jarvis/OutGen/db"
)

// TotalProfit calculates TotalPnL for all days, daily average and monthly average
// Out: pnl,daily average pnl,monthly average pnl,yearly average pnl
func TotalProfit(metadata []db.Metadata) (float64, float64, float64, float64) {
	var pnl float64
	for _, data := range metadata {
		pnl += float64(data.Result)
	}
	sd := metadata[0].Date
	ed := metadata[len(metadata)-1].Date
	std := time.Unix(int64(sd), 0)
	etd := time.Unix(int64(ed), 0)
	diff := etd.Sub(std).Abs().Hours()
	diff = (diff / 24) / (float64(365) / float64(12))
	return pnl, pnl / float64(len(metadata)), pnl / diff, (pnl / diff) * 12
}

// WinLossStats calculates WinDays,LossDays,WinPercentage and LossPercentage
// Out: WinDays array, LossDays array in UnixTS , WinPercentage and LossPercentage
func WinLossStats(metadata []db.Metadata) ([]int32, []int32, float64, float64) {
	var (
		WinDays  = []int32{}
		LossDays = []int32{}
	)
	for _, meta := range metadata {
		if meta.Result >= 0 {
			WinDays = append(WinDays, meta.Date)
		} else {
			LossDays = append(LossDays, meta.Date)
		}
	}
	WinP := (float64(len(WinDays)) / float64(len(metadata)) * 100)
	LossP := (float64(len(LossDays)) / float64(len(metadata)) * 100)
	return WinDays, LossDays, WinP, LossP
}

// AvgProfitWinLoss calculates average profit of win days and average loss of loss days
func AvgProfitWinLoss(WinDays, LossDays []int32, metadata []db.Metadata) (float32, float32, error) {
	var (
		wg       sync.WaitGroup
		APW, APL float32
		err      error
	)
	wg.Add(2)
	go func(metadata []db.Metadata, days []int32) {
		defer wg.Done()
		var pnl float32
		for _, dt := range days {
			pnl += GetProfitsFromMeta(metadata, dt)
		}
		APW = pnl / float32(len(days))
	}(metadata, WinDays)
	go func(metadata []db.Metadata, days []int32) {
		defer wg.Done()
		var pnl float32
		for _, dt := range days {
			pnl += GetProfitsFromMeta(metadata, dt)
		}
		APL = pnl / float32(len(days))
	}(metadata, LossDays)
	wg.Wait()
	if err != nil {
		return -1, -1, err
	}
	return APW, APL, nil
}

func GetProfitsFromMeta(m []db.Metadata, d int32) float32 {
	for _, meta := range m {
		if meta.Date == d {
			return meta.Result
		}
	}
	return -1
}

func CalcWStreak(metadata []db.Metadata) int {
	maxStreak := 0
	streak := 0
	for _, m := range metadata {
		if m.Result >= 0 {
			streak += 1
			if streak > maxStreak {
				maxStreak = streak
			}
		} else {
			streak = 0
		}
	}
	return maxStreak
}

func CalcLStreak(metadata []db.Metadata) int {
	maxStreak := 0
	streak := 0
	for _, m := range metadata {
		if m.Result < 0 {
			streak += 1
			if streak > maxStreak {
				maxStreak = streak
			}
		} else {
			streak = 0
		}
	}
	return maxStreak
}

func GetCumulativeProfit(metadata []db.Metadata) []float64 {
	output := make([]float64, len(metadata))
	for pos, m := range metadata {
		if pos == 0 {
			output[0] = float64(m.Result)
		} else {
			output[pos] = output[pos-1] + float64(m.Result)
		}
	}
	return output
}

func MDD(metadata []db.Metadata, ayp float64) (float64, int32, float64) {
	mddpos := [][2]int{}
	for i := 0; i < len(metadata); {
		if metadata[i].Result < 0 {
			fmt.Println(metadata[i].Result)
			pos := [2]int{i, 0}
			tempsum := metadata[i].Result
			for j := i + 1; j < len(metadata); j++ {
				tempsum += metadata[j].Result
				if tempsum >= 0 {
					pos[1] = j
					i = j + 1
					mddpos = append(mddpos, pos)
					break
				}
				if j == len(metadata)-1 {
					pos[1] = j
					mddpos = append(mddpos, pos)
					i = len(metadata)
					break
				}
			}
		} else {
			i += 1
		}
	}
	mdd := float64(0)
	mrp := int32(0)
	for _, ddp := range mddpos {
		rp := metadata[ddp[1]].Date - metadata[ddp[0]].Date
		if rp > mrp {
			mrp = rp
		}
		dd := 0.0
		cump := []float64{}
		for i := ddp[0]; i <= ddp[1]; i++ {
			dd += float64(metadata[i].Result)
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

	return mdd, mrp / (60 * 60 * 24), ayp / mdd
}

func Expectancy() {

}

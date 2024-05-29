package db

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Metadata struct {
	Date    int32
	Trigger int32
	Etime   int32
	Entry   float32
	Xtime   int32
	Exit    float32
	Result  float32
}

type PositionData struct {
	Date     int64
	Expiry   int64
	Strike   int32
	Opt_Type string
	Trigger  int32
	Etime    int64
	Entry    float32
	Xtime    int64
	Exit     float32
	Result   float32
}

func (db *DB) FetchMetadata(bid string) ([]Metadata, []PositionData) {
	query := "SELECT date, trigger, etime,entry,xtime,exit,result from %s where position = 'BASKET' order by date ASC"
	query2 := "SELECT date,position,trigger,etime,entry,xtime,exit,result from %s where position NOT LIKE 'BASKET' order by date ASC"
	rows, err := db.DB.Query(fmt.Sprintf(query, bid))
	if err != nil {
		log.Error("Error in querying Userdb: ", err, (fmt.Sprintf(query, bid)), db)
		return nil, nil
	}
	rows2, err := db.DB.Query(fmt.Sprintf(query2, bid))
	if err != nil {
		log.Error("Error in querying Userdb: ", err, (fmt.Sprintf(query2, bid)), db)
		return nil, nil
	}
	metadata := []Metadata{}
	posData := []PositionData{}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for rows.Next() {
			var (
				date    string
				trigger int32
				etime   int32
				entry   float32
				xtime   int32
				exit    float32
				result  float32
			)
			err := rows.Scan(&date, &trigger, &etime, &entry, &xtime, &exit, &result)
			if err != nil {
				log.Error("Error in scanning data md: ", err)
				// log.Debug(date, trigger, etime, entry, xtime, exit, result)
				return
			}
			if len(date) < 6 {
				date = fmt.Sprintf("0%s", date)
			}
			dstr, err := time.Parse("020106", date)
			if err != nil {
				log.Error("error in ts conversion: ", err)
			}
			dte := int32(dstr.Unix())
			metadata = append(metadata, Metadata{
				Date:    dte,
				Trigger: trigger,
				Etime:   dte + etime,
				Entry:   entry,
				Xtime:   dte + xtime,
				Exit:    exit,
				Result:  result,
			})
		}
		wg.Done()
	}()

	go func() {
		for rows2.Next() {
			var (
				date     string
				position string
				trigger  int32
				etime    int64
				entry    float32
				xtime    int64
				exit     float32
				result   float32
			)
			err := rows2.Scan(&date, &position, &trigger, &etime, &entry, &xtime, &exit, &result)
			if err != nil {
				log.Error("Error in scanning data pd: ", err)
				// log.Debug(date, position, trigger, etime, entry, xtime, exit, result)
				return
			}
			if len(date) < 6 {
				date = fmt.Sprintf("0%s", date)
			}
			dstr, err := time.Parse("020106", date)
			if err != nil {
				log.Error("error in ts conversion: ", err)
			}
			pos_parts := strings.Split(position, "-")
			exp, err := time.Parse("020106", pos_parts[1])
			if err != nil {
				log.Error("Error in Parsing time", err)
				return
			}
			var strike int64
			if pos_parts[0] == "FUT" {
				strike = -1
			} else {
				strike, _ = strconv.ParseInt(pos_parts[2], 10, 64)
			}

			dte := dstr.Unix()
			posData = append(posData, PositionData{
				Date:     dte,
				Expiry:   exp.Unix(),
				Opt_Type: pos_parts[0],
				Strike:   int32(strike / 100),
				Trigger:  trigger,
				Etime:    dte + etime,
				Entry:    entry,
				Xtime:    dte + xtime,
				Exit:     exit,
				Result:   result,
			})
		}
		wg.Done()
	}()
	wg.Wait()
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Date < metadata[j].Date
	})
	sort.Slice(metadata, func(i, j int) bool {
		return posData[i].Date <= posData[j].Date
	})

	return metadata, posData
}

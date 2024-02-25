package db

import (
	"fmt"
	"sort"
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

func (db *DB) FetchMetadata(bid string) []Metadata {
	query := "SELECT date, trigger, etime,entry,xtime,exit,result from %s where position = 'BASKET'"
	rows, err := db.DB.Query(fmt.Sprintf(query, bid))
	if err != nil {
		log.Error("Error in querying Userdb: ", err)
		return nil
	}
	metadata := []Metadata{}
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
			log.Error("Error in scanning data: ", err)
			return nil
		}
		if len(date) < 6 {
			date = fmt.Sprintf("0%s", date)
		}
		dstr, err := time.Parse("020106", date)
		if err != nil {
			log.Error("error in ts conversion: ", err)
		}
		metadata = append(metadata, Metadata{
			Date:    int32(dstr.Unix()),
			Trigger: trigger,
			Etime:   etime,
			Entry:   entry,
			Xtime:   xtime,
			Exit:    exit,
			Result:  result,
		})
	}
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Date < metadata[j].Date
	})
	return metadata
}

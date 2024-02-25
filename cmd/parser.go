package cmd

import (
	"encoding/csv"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func ParseFile(path string) (total [][]string, Metadata [][]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		log.Error("Error in opening output file: ", path, ": ", err)
		return
	}
	reader := csv.NewReader(f)
	total = [][]string{}
	Metadata = [][]string{}
	defer f.Close()
	for {
		rec, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Error("error in read: ", err)
		}
		if rec[1] == "BASKET" {
			Metadata = append(Metadata, rec)
		}
		total = append(total, rec)
	}

	return
}

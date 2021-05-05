package downloader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Download is ...
func Download(url string, path string) (bool, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode >= 400 {
		err := errors.New(url + " : " + response.Status)
		return false, err
	}

	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	fp, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	fp.WriteString(string(body))

	return true, nil
}

// DownloadReport is ...
type DownloadReport struct {
	DownloadedList []string
	ErrorList      []string
}

// ToJSON is ...
func (report DownloadReport) ToJSON() string {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

// Save is ...
func (report DownloadReport) Save(datadir string) (bool, error) {
	time, _ := time.Now().MarshalText()

	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		if err2 := os.Mkdir(datadir, os.FileMode(0777)); err2 != nil {
			log.Print("Failed Save Mkdir!!")
			log.Fatal(err)
		}
	}

	filePath := filepath.Join(datadir, "report_"+string(time)+".json")
	fp, err := os.Create(filePath)
	if err != nil {
		return false, err
	}
	defer fp.Close()
	fp.WriteString(report.ToJSON())

	return true, nil
}

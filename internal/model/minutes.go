package model

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Speaker は議事録に出現する話者を表現するための構造体です。
type Speaker struct {
	Label string
	Person Person
	ResolutionScore float64
}

// Speach is ...
type Speach struct {
	Speaker *Speaker
	Talks   []string
}

// Minutes is ...
type Minutes struct {
	Title             string
	WorkingGroup      string
	SpeachCount       int
	WorkingGroupOrder string
	WorkingGroupID    string
	Date              string
	Venue             string
	Topics            []string
	Speakers          map[string]*Speaker
	Speaches          []*Speach
}

// ToJSON は、Minutes型のデータをJSON形式の文字列として返すメソッドです。
func (m Minutes) ToJSON() string {
	jsondata, _ := json.MarshalIndent(m, "", "    ")
	return string(jsondata)
}

// ParseMinutesFromFile is a method for parsing minutes from a file
func ParseMinutesFromFile(fileName string) Minutes {
	brtag := regexp.MustCompile(`(?m)<br\/>`)
	speakertag := regexp.MustCompile(`(?m)^【(.+?)】`)
	const QUERY = "div#contentsMain h2:contains('議事録') ~ p"

	file, _ := ioutil.ReadFile(fileName)
	reader := strings.NewReader(string(file))
	doc, _ := goquery.NewDocumentFromReader(reader)

	minutes := Minutes{
		Title:    doc.Find("h1").Text(),
		Speaches: []*Speach{},
		Speakers: map[string]*Speaker{},
	}

	wginfotag := regexp.MustCompile(`no([0-9][0-9])wg([0-9][0-9][0-9])-.+htm`)
	wginfo := wginfotag.FindStringSubmatch(fileName)
	if len(wginfo) >= 3 {
		minutes.WorkingGroupOrder = wginfo[1]
		minutes.WorkingGroupID = wginfo[2]
	}

	minutes.WorkingGroup = strings.Split(minutes.Title, "　")[0]

	currentSpeach := new(Speach)

	doc.Find(QUERY).Each(func(index int, s *goquery.Selection) {
		html, _ := s.Html()
		elements := brtag.Split(html, -1)
		for _, s := range elements {
			speakerElements := speakertag.FindAllStringSubmatch(strings.TrimSpace(s), -1)
			talk := speakertag.ReplaceAllString(strings.TrimSpace(s), "")
			if len(speakerElements) > 0 {
				if len(currentSpeach.Talks) > 0 {
					minutes.Speaches = append(minutes.Speaches, currentSpeach)
				}
				currentSpeach = new(Speach)
				speaker, exists := minutes.Speakers[speakerElements[0][1]];

				if !exists {
				    speaker = &Speaker{
						Label: speakerElements[0][1],
					}
					minutes.Speakers[speaker.Label] = speaker
				}

				currentSpeach.Speaker = speaker
			}

			if len(talk) > 0 {
				currentSpeach.Talks = append(currentSpeach.Talks, talk)
			}
		}
	})

	minutes.Speaches = append(minutes.Speaches, currentSpeach)

	minutes.SpeachCount = len(minutes.Speaches)

	return minutes
}

// ParseMinutesFromPDF2Html は、AcrobatでPDFからHtmlに変換したファイルをパースするルーチンです。
func ParseMinutesFromPDF2Html(fileName string) Minutes {

	brtag := regexp.MustCompile(`(?m)<br\/>`)
	speakertag := regexp.MustCompile(`(?m)^【(.+?)】`)
	const QUERY = "p"

	file, _ := ioutil.ReadFile(fileName)
	reader := strings.NewReader(string(file))
	doc, _ := goquery.NewDocumentFromReader(reader)

	kaigiTitle := ""
	kaigitag := regexp.MustCompile(`.+ワーキンググループ.+[ 0-9　０-９]+回.+`)
	doc.Find("p:nth-child(-n+3)").Each(func(index int, s *goquery.Selection) {
		khtml, _ := s.Html()
		if kaigitag.MatchString(khtml) {
			kaigiTitle = khtml
		} else {
			kaigiTitle = kaigiTitle + ""
		}
	})

	minutes := Minutes{
		Title:    kaigiTitle,
		Speaches: []*Speach{},
		Speakers: map[string]*Speaker{},
	}

	wginfotag := regexp.MustCompile(`no([0-9][0-9])wg([0-9][0-9][0-9])-.+htm`)
	wginfo := wginfotag.FindStringSubmatch(fileName)
	if len(wginfo) >= 3 {
		minutes.WorkingGroupOrder = wginfo[1]
		minutes.WorkingGroupID = wginfo[2]
	}

	minutes.WorkingGroup = strings.Split(minutes.Title, "（")[0]

	//発話中のhtmlタグを除去するんだな…（Acrobatが勝手にアンダーラインとかも再現しちゃうので）
	reptag := regexp.MustCompile(`<("[^"]*"|'[^']*'|[^'">])*>`)
	//行の途中でページを跨いじゃってることを検知する正規表現
	kutentag := regexp.MustCompile(`[^。）―─]$`)

	currentSpeach := new(Speach)

	speakerDefined := false
	prevTalk := ""

	doc.Find(QUERY).Each(func(index int, s *goquery.Selection) {
		html, _ := s.Html()
		elements := brtag.Split(html, -1)
		for _, s := range elements {
			speakerElements := speakertag.FindAllStringSubmatch(strings.TrimSpace(s), -1)
			talk := speakertag.ReplaceAllString(strings.TrimSpace(s), "")
			talk = reptag.ReplaceAllString(talk, "")
			if len(speakerElements) > 0 {
				if len(currentSpeach.Talks) > 0 {
					minutes.Speaches = append(minutes.Speaches, currentSpeach)
				}
				currentSpeach = new(Speach)
				speaker, exists := minutes.Speakers[speakerElements[0][1]];

				if !exists {
				    speaker = &Speaker{
						Label: speakerElements[0][1],
					}
					minutes.Speakers[speaker.Label] = speaker
				}

				currentSpeach.Speaker = speaker
				speakerDefined = true
			}

			if len(talk) > 0 {
				if speakerDefined {
					//改ページ位置に跨って文中で分離してしまっている箇所をさがす
					bunmatsu := kutentag.FindString(talk)
					if bunmatsu != "" {
						//log.Print("FF Found!!: " + bunmatsu)
						prevTalk = talk
					} else {
						talk = prevTalk + talk
						currentSpeach.Talks = append(currentSpeach.Talks, talk)
						prevTalk = ""
					}

				}
			}
		}
	})

	minutes.Speaches = append(minutes.Speaches, currentSpeach)

	minutes.SpeachCount = len(minutes.Speaches)

	return minutes

}

// MinutesArray はMinutes構造体の配列です
type MinutesArray []Minutes

// ExportAsJSON は、MinutesArray型のデータをJSONファイルとしてエクスポートするメソッドです。
func (minutesArray MinutesArray) ExportAsJSON(outputdir string) {
	json2data, _ := json.MarshalIndent(minutesArray, "", "    ")
	filePath := filepath.Join(outputdir, "all.json")
	fp2, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fp2.Close()
	fp2.WriteString(string(json2data))
}

//ExportSpeakersAsCSV は、CSV形式で発話者リストを書き出すメソッドです。
func (minutesArray MinutesArray) ExportSpeakersAsCSV(outputdir string) {
	// O_WRONLY:書き込みモード開く, O_CREATE:無かったらファイルを作成
	filePath3 := filepath.Join(outputdir, "all_speaker.csv")
	file3, err := os.OpenFile(filePath3, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer file3.Close()

	err = file3.Truncate(0) // ファイルを空っぽにする(2回目以降用)
	if err != nil {
		log.Fatal(err)
	}

	writer := csv.NewWriter(transform.NewWriter(file3, japanese.ShiftJIS.NewEncoder()))
	writer.Write([]string{"WorkingGroupOrder", "WorkingGroupID", "Title", "Speaker.Label", "Speaker.ResolutionScore", "Person.ID", "Person.Label", "Person.Name", "Person.Role", "Person.Affiliation"})
	for _, v := range minutesArray {
		for _, sp := range v.Speakers {
		    score := strconv.FormatFloat(sp.ResolutionScore, 'g', -1, 32)
			writer.Write([]string{v.WorkingGroupOrder, v.WorkingGroupID, v.Title, sp.Label, score, sp.Person.ID, sp.Person.Label, sp.Person.Name, sp.Person.Role, sp.Person.Affiliation})
		}
	}

	writer.Flush()
}

// ExportAsKH は、KH Coder読み込み用のファイルをエクスポートするメソッドです。
func (minutesArray MinutesArray) ExportAsKH(wgList WorkingGroupList, outputdir string) (bool, error) {
	lines := []string{}

	crWgOrder := ""
	for _, v := range minutesArray {
		if crWgOrder != v.WorkingGroupOrder {
			wg, exists := wgList["no"+v.WorkingGroupOrder]
			if (exists) {
				lines = append(lines, "<h1>"+wg.Name+"</h1>\n")
			} else {
				return false, errors.New(v.WorkingGroupOrder + "という番号のWGが見つかりません。")
			}
		}
		crWgOrder = v.WorkingGroupOrder

		if (len(v.Title) > 0) {
			lines = append(lines, "<h2>"+v.Title+"</h2>\n")
		}

		for _, sp := range v.Speaches {
			if sp.Speaker == nil {
				sp.Speaker = &Speaker{}
			}
			lines = append(lines, "<h3>"+sp.Speaker.Label+"</h3>\n")
			for _, tk := range sp.Talks {
				lines = append(lines, tk+"\n")
			}
		}
	}

	filenameKH := filepath.Join(outputdir, "all_khcoder.txt")
	fileKH, err := os.Create(filenameKH)
	if err != nil {
		log.Fatal(err)
	}
	defer fileKH.Close()

	for _, line := range lines {
		_, err := fileKH.WriteString(line)
		if err != nil {
			log.Fatal(err)
		}
	}

	return true, nil
}

// ImportMinutesArrayFromHTML は、複数のHTMLファイルを読み込んで MinutesArray を作成する関数です。
func ImportMinutesArrayFromHTML(baseDirs []string, outputDir string) MinutesArray {
	var minutesArray MinutesArray
	var pdfFlag bool

	for num, baseDir := range baseDirs {

		if num == 0 {
			pdfFlag = false
		} else {
			pdfFlag = true
		}

		files, err := ioutil.ReadDir(baseDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {

			var m Minutes
			if pdfFlag {
				fmt.Println("Processing PDF2Html: " + file.Name())
				m = ParseMinutesFromPDF2Html(baseDir + "/" + file.Name())
			} else {
				fmt.Println("Processing: " + file.Name())
				m = ParseMinutesFromFile(baseDir + "/" + file.Name())
			}

			filePath := filepath.Join(outputDir, file.Name()+".json")
			fp, err := os.Create(filePath)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			fp.WriteString(string(m.ToJSON()))

			minutesArray = append(minutesArray, m)

		}

	}

	return minutesArray
}

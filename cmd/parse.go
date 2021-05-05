package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/tsunekawa/meroku/internal/model"
)

// ParseCmd は、議事録をダウンロードするためのコマンド関数です。
func ParseCmd(args []string) {

	var rootDir string
	var outputRootDir string
	var baseDirs []string
	var withMemberlistFlag bool

	defaultDir := "./data/example"
	defaultOutputDir := filepath.Join(defaultDir, "json")

	// コマンドラインオプションの設定
	fs := flag.NewFlagSet("parse", flag.ExitOnError)
	fs.StringVar(&rootDir, "dir", defaultDir, "読み込み元のディレクトリ")
	fs.StringVar(&outputRootDir, "out", defaultOutputDir, "保存先のディレクトリ")
	fs.BoolVar(&withMemberlistFlag, "memberlist", false, "名簿ページもパースする")
	fs.Parse(args)

	// ディレクトリがあればインポート対象に加える
	if _, err := os.Stat(filepath.Join(rootDir, "html")); err == nil {
		baseDirs = append(baseDirs, filepath.Join(rootDir, "html"))
	}
	if _, err := os.Stat(filepath.Join(rootDir, "html_from_pdf")); err == nil {
		baseDirs = append(baseDirs, filepath.Join(rootDir, "html_from_pdf"))
	}
	
	outputdir := filepath.Join(outputRootDir, "output_"+time.Now().Format("2006-01-02T150405"))

	//ダウンローダーで出力したワーキンググループリストを読み込み
	wgList := model.ImportWorkingGroupList(filepath.Join(rootDir, "working-groups.json"))

	// 出力するフォルダを作成する
	if _, err := os.Stat(outputdir); os.IsNotExist(err) {
		if err2 := os.Mkdir(outputdir, os.FileMode(0777)); err2 != nil {
			log.Print("Failed output Mkdir!!")
			log.Fatal(err)
		}
	}

	minutesArray := model.ImportMinutesArrayFromHTML(baseDirs, outputdir)

	//名簿のパースと出力(--memberlistオプション指定時のみ実行)
	if withMemberlistFlag {
		memberListDir := filepath.Join(rootDir, "html", "memberlist")
		if _, err := os.Stat(memberListDir); err == nil {
			os.MkdirAll(filepath.Join(outputdir, "memberlist"), 0777)

			files, err := filepath.Glob(filepath.Join(memberListDir, "*.htm*"))
			if err != nil {
				log.Fatal(err)
			}

			memberListMap := make(map[string]*model.MemberList)
			WGNOPATTERN := regexp.MustCompile("no([0-9]{2})")

			for _, file := range files  {
				wgno := WGNOPATTERN.FindAllStringSubmatch(file, -1)[0][1]

				memberlist, err := model.LoadMemberListFromHTML(file)
				if err != nil {
					log.Fatal(err)
				}
				memberListMap[wgno] = &memberlist

				filePath := filepath.Join(outputdir, "memberlist", filepath.Base(file[:len(file)-len(filepath.Ext(file))]) + ".json")

				fp, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)
				if err != nil {
					log.Fatal(err)
				}
				defer fp.Close()
				fp.WriteString(string(memberlist.ToJSON()))
			}

			resolutedMinutesArray := model.MinutesArray{}
			for _, minutes := range minutesArray {
				for _, speaker := range minutes.Speakers {
					memberlist, exists := memberListMap[minutes.WorkingGroupOrder]
					if exists {
						person, sims, err := memberlist.Resolve(speaker.Label)
						if err != nil {
							log.Println(err)
						} else {
							speaker.Person = *person
							speaker.ResolutionScore = sims[0].Score
							minutes.Speakers[speaker.Label] = speaker
							log.Println("名寄せ：" + speaker.Label + "(" +speaker.Person.ID + ")")
						}
					}
				}
				resolutedMinutesArray = append(resolutedMinutesArray,  minutes)
			}
			minutesArray = resolutedMinutesArray
		}
	}

	fmt.Println("Output All Combined File.")
	minutesArray.ExportAsJSON(outputdir)

	//CSVで発話者リストを書き出し（動作検証用）
	fmt.Println("Output Speaker CSV File.")
	minutesArray.ExportSpeakersAsCSV(outputdir)

	fmt.Println("Output KHCoder Source File.")
	minutesArray.ExportAsKH(wgList, outputdir)

}

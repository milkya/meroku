package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/tsunekawa/meroku/internal/downloader"
	"github.com/tsunekawa/meroku/internal/model"
)

// DownloadCmd は、議事録をダウンロードするためのコマンド関数です。
func DownloadCmd(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)

	var wgID string
	var allFlag bool
	var withMemberlistFlag bool
	var downloaddir string
	var report downloader.DownloadReport

	defaultDir := filepath.Join("./data", "download_"+time.Now().Format("2006-01-02T150405"))

	// コマンドラインオプションの設定
	fs.StringVar(&downloaddir, "dir", defaultDir, "保存先のディレクトリ")
	fs.StringVar(&wgID, "wgid", "", "ワーキンググループの番号")
	fs.BoolVar(&allFlag, "all", false, "すべてのワーキンググループをダウンロードする")
	fs.BoolVar(&withMemberlistFlag, "memberlist", false, "名簿ページもダウンロードする")
	fs.Parse(args)

	// HTMLをダウンロードするフォルダを作成する
	if _, err := os.Stat(downloaddir); os.IsNotExist(err) {
		if err2 := os.Mkdir(downloaddir, os.FileMode(0777)); err2 != nil {
			log.Print("Failed main Mkdir!!")
			log.Fatal(err)
		}
	}

	// ワーキンググループの一覧を取得
	workingGroups := model.GetWorkingGroups()

	// ワーキンググループ情報をJSONファイルとして保存
	data, err := json.MarshalIndent(workingGroups, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	filePath := filepath.Join(downloaddir, "working-groups.json")
	fp, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	fp.WriteString(string(data))

	// ワーキンググループの一覧から指定した番号のワーキンググループを取得
	downloadTargets := []*model.WorkingGroup{}
	if allFlag {
		for _, item := range workingGroups {
			downloadTargets = append(downloadTargets, item)
		}
	} else if wgID != "" {
		// コマンドラインからワーキンググループ番号を受け取る
		isValidwgID, _ := regexp.MatchString(`\d{3}`, wgID)
		if !isValidwgID {
			err := errors.New(wgID + "はワーキンググループ番号でありません。")
			log.Fatal(err)
		}

		var wg *model.WorkingGroup
		wgExists := false
		for _, item := range workingGroups {
			if item.ID == wgID {
				wg = item
				wgExists = true
			}
		}

		if !wgExists {
			err := errors.New(wgID + "という番号のワーキンググループはありません。")
			log.Fatal(err)
		}

		downloadTargets = append(downloadTargets, wg)
	}

	// 指定したワーキンググループの議事録を保存する
	for _, wg := range downloadTargets {

		if (withMemberlistFlag) {
			report, err := wg.DownloadMemberListAll(downloaddir)
			if err != nil {
				log.Printf("WARN: ワーキンググループ「%v」(%v) の名簿を取得できませんでした。\n", wg.Name, wg.ID)
				log.Println(err)
			}
			report.Save(downloaddir)
		}

		report, err = wg.DownloadMinutesAll(downloaddir)
		if err != nil {
			log.Printf("WARN: ワーキンググループ「%v」(%v) の議事録を取得できませんでした。\n", wg.Name, wg.ID)
			log.Println(err)
		}
		report.Save(downloaddir)
	}

}

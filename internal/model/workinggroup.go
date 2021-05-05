package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/tsunekawa/meroku/internal/downloader"
)

// WorkingGroup は、中央教育審議会のワーキンググループを表す構造体です。
type WorkingGroup struct {
	Order          string
	ID             string
	Name           string
	URL            string
	MinutesListURL string
	MinutesURLs    []string
	MemberListURLs []string
}

// GetIDFromURL は、ワーキンググループ情報のURLからIDを抽出するメソッドです。実行するとIDメンバーに結果が格納されます。
func (wg *WorkingGroup) GetIDFromURL() string {
	reg := regexp.MustCompile(`chukyo3/(\d{3})/index.htm`)
	if len(wg.URL) <= 0 {
		wg.ID = ""
	}

	results := reg.FindAllStringSubmatch(wg.URL, -1)

	if len(results) > 0 {
		wg.ID = results[0][1]
	} else {
		wg.ID = ""
	}

	return wg.ID
}

// GetMinutesListURL は、議事録一覧ページのURLをワーキンググループのページから抽出するメソッドです。実行すると MinutesListURLメンバーに値が格納されます。
func (wg *WorkingGroup) GetMinutesListURL() (string, error) {
	doc, err := goquery.NewDocument(wg.URL)
	if err != nil {
		return "", err
	}

	nodes := doc.Find("a:contains('これまでの議事要旨・議事録・配付資料の一覧はこちら')")
	if nodes.Length() <= 0 {
		err := errors.New("議事録一覧のリンクがありません : " + wg.URL)
		return "", err
	}
	gijiListLink := nodes.First()
	gijiListHref, _ := gijiListLink.Attr("href")

	baseURL, _ := url.Parse(wg.URL)
	gijiListURL := toAbsURL(baseURL, gijiListHref)

	wg.MinutesListURL = gijiListURL

	return wg.MinutesListURL, nil
}

// GetMinutesList は、MinuteListURLから議事録のURLの一覧を取得し、MinuteURLs に配列として格納する
func (wg *WorkingGroup) GetMinutesList() ([]string, error) {
	gijiList := []string{}

	minutesURL, err := wg.GetMinutesListURL()
	if err != nil {
		return gijiList, err
	}

	doc, err := goquery.NewDocument(minutesURL)
	if err != nil {
		log.Fatalf("%v : WG %v", err, wg.ID)
	}
	doc.Find("a:contains('議事録')").Each(func(idx int, node *goquery.Selection) {
		if node.Text() == "議事録" {
			href, _ := node.Attr("href")

			baseURL, _ := url.Parse(minutesURL)
			gijiURL := toAbsURL(baseURL, href)

			gijiList = append(gijiList, gijiURL)
		}
	})

	wg.MinutesURLs = gijiList

	return gijiList, nil
}

//GetMemberListURLs は、名簿ページ一覧のURLをワーキンググループのページから抽出するメソッドです。実行すると MemberListURLメンバーに値が格納されます。
func (wg *WorkingGroup) GetMemberListURLs() (memberListURLs []string, err error) {

	doc, err := goquery.NewDocument(wg.URL)
	if err != nil {
		return memberListURLs, err
	}

	nodes := doc.Find("a:contains('委員名簿')")
	if nodes.Length() <= 0 {
		err := errors.New("名簿のリンクがありません : " + wg.URL)
		return memberListURLs, err
	}

	nodes.Each(func(idx int, node *goquery.Selection) {
		href, _ := node.Attr("href")

		baseURL, _ := url.Parse(wg.URL)
		memberListURL := toAbsURL(baseURL, href)

		memberListURLs = append(memberListURLs, memberListURL)
	})

	wg.MemberListURLs = memberListURLs

	return wg.MemberListURLs, nil
}

// DownloadMinutesAll は、ワーキンググループの議事録を一括ダウンロードしてHTMLファイルとして保存するメソッドです。
func (wg WorkingGroup) DownloadMinutesAll(datadir string) (downloader.DownloadReport, error) {
	downloadedURLs := []string{}
	errorURLs := []string{}
	minutesList, err := wg.GetMinutesList()
	if err != nil {
		return downloader.DownloadReport{}, err
	}

	for _, minutesURL := range minutesList {
		fileName := wg.Order + "wg" + wg.ID + "-" + regexp.MustCompile(`[^/]+$`).FindString(minutesURL)
		//dir := filepath.Join(datadir, wg.Order)
		dir := filepath.Join(datadir, "html")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err2 := os.Mkdir(dir, os.FileMode(0777)); err2 != nil {
				log.Print("Failed DownloadMinutesAll Mkdir!!")
				log.Fatal(err)
			}
		}
		filePath := dir + "/" + fileName
		success, _ := downloader.Download(minutesURL, filePath)

		if success {
			downloadedURLs = append(downloadedURLs, minutesURL)
		} else {
			errorURLs = append(errorURLs, minutesURL)
		}
	}
	return downloader.DownloadReport{
		DownloadedList: downloadedURLs,
		ErrorList:      errorURLs,
	}, nil
}

// DownloadMemberListAll は、特定のワーキンググループの名簿ページをダウンロードするメソッドです。
func (wg WorkingGroup) DownloadMemberListAll(datadir string) (downloader.DownloadReport, error) {
	downloadedURLs := []string{}
	errorURLs := []string{}
	memberListList, err := wg.GetMemberListURLs()
	if err != nil {
		return downloader.DownloadReport{}, err
	}

	for _, memberListURL := range memberListList {
		fileName := wg.Order + "wg" + wg.ID + "-" + regexp.MustCompile(`[^/]+$`).FindString(memberListURL)
		//dir := filepath.Join(datadir, wg.Order)
		dir := filepath.Join(datadir, "html", "memberlist")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err2 := os.MkdirAll(dir, os.FileMode(0777)); err2 != nil {
				log.Print("Failed DownloadMinutesAll Mkdir!!")
				log.Fatal(err)
			}
		}
		filePath := dir + "/" + fileName
		success, _ := downloader.Download(memberListURL, filePath)

		if success {
			downloadedURLs = append(downloadedURLs, memberListURL)
		} else {
			errorURLs = append(errorURLs, memberListURL)
		}
	}



	return downloader.DownloadReport{
		DownloadedList: downloadedURLs,
		ErrorList:      errorURLs,
	}, err
}

// GetWorkingGroups は、中央教育審議会のページからワーキンググループの一覧情報を抽出し、配列して返却するメソッドです。
func GetWorkingGroups() map[string]*WorkingGroup {
	const rootURL = "https://www.mext.go.jp/b_menu/shingi/chukyo/chukyo3/index.htm"
	doc, _ := goquery.NewDocument(rootURL)

	workingGroups := map[string]*WorkingGroup{}

	doc.Find(".shingi_block ul li a").Each(func(idx int, node *goquery.Selection) {
		wg := new(WorkingGroup)
		wg.Name = node.Text()

		href, exists := node.Attr("href")
		if exists {
			wg.URL = "https://www.mext.go.jp" + href
		}

		wg.GetIDFromURL()
		_, err := wg.GetMinutesList()
		if err != nil {
			log.Printf("WARN: 議事録一覧の取得失敗 : 「%v」(%v)\n", wg.Name, wg.ID)
			log.Println(err)
		}

		_, err = wg.GetMemberListURLs()
		if err != nil {
			log.Printf("WARN: 名簿一覧の取得失敗 : 「%v」(%v)\n", wg.Name, wg.ID)
			log.Println(err)
		}

		dispOrder := fmt.Sprintf("no%02d", idx)
		wg.Order = dispOrder

		workingGroups[dispOrder] = wg
	})

	return workingGroups
}

// toAbsURL はベースURLと相対URLから絶対URLを返す関数です
func toAbsURL(baseurl *url.URL, weburl string) string {
	relurl, err := url.Parse(weburl)
	if err != nil {
		return ""
	}
	absurl := baseurl.ResolveReference(relurl)
	return absurl.String()
}

// WorkingGroupList はWorkingGroupのスライス
type WorkingGroupList map[string]WorkingGroup

//ImportWorkingGroupList は downloaderが出力した working_groups.json を読み込むための関数です。
func ImportWorkingGroupList(importFilePath string) WorkingGroupList {
	raw, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var wgList WorkingGroupList
	json.Unmarshal(raw, &wgList)

	return wgList
}

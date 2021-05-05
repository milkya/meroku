package model

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/masatana/go-textdistance"
)

// Person は、議事録や名簿に現れる人物を表現するための構造体です。
type Person struct {
	ID          string
	Label       string
	Name        string
	Affiliation string
	Role        string
}

// NormarizeLabel は、Personの属性からラベルを生成するためのメソッドです。
func (person Person) NormarizeLabel() (label string) {
	return strings.Join([]string{person.Name, person.Role}, "")
}

// MemberList は、ワーキング・グループを構成する委員の名簿を表す構造体です。
type MemberList struct {
	WorkingGroup *WorkingGroup
	Members      []*Person
}

// ToJSON は、MemberList型のデータをJSON形式の文字列として返すメソッドです。
func (m MemberList) ToJSON() string {
	jsondata, _ := json.MarshalIndent(m, "", "    ")
	return string(jsondata)
}


// Similarity は、Personの間の類似度を表現する構造体です。
type Similarity struct {
	Target *Person
	Score  float64
}

// SimilarityArray は、Similarity を格納する配列です。
type SimilarityArray []Similarity

// Len は、SimilarityArrayの要素数を返します。
// sort.Sort 対応用です。
func (simArray SimilarityArray) Len() int {
	return len(simArray)
}

// Less は、Similarity 同士の大小を比較するメソッドです。
// sort.Sort 対応用です。
func (simArray SimilarityArray) Less(i, j int) bool {
	return simArray[i].Score <= simArray[j].Score
}

// Swap は、SimilarityArray 内の要素の位置をスワップするメソッドです。
// sort.Sort 対応用です。
func (simArray SimilarityArray) Swap(i, j int) {
	simArray[i], simArray[j] = simArray[j], simArray[i]
}

// Resolve は、引数として与えられた話者のラベルに対応する Person を推測して返すメソッドです。推測にあたっては、members に格納されているPersonのラベルを正規化し、引数との編集距離を求めることで類似度を算出しています。
// 編集距離は、現在ジャロ・ウィンクラー距離を使用しています。
func (m MemberList) Resolve(nameLabel string) (person *Person, sims []Similarity, err error) {
	var member *Person
	var memberLabel string
	var score float64
	var similarityArray SimilarityArray

	for _, member = range m.Members {
		memberLabel = member.NormarizeLabel()
		if []rune(nameLabel)[0] != []rune(memberLabel)[0] {
			score = 0.0
		} else {
			score = textdistance.JaroWinklerDistance(memberLabel, nameLabel)
		}
		similarityArray = append(similarityArray, Similarity{Target: member, Score: score})
	}

	if len(similarityArray) <= 0 {
		person = nil
		err = errors.New("名寄せ失敗: " + nameLabel)
	} else {
		sort.Sort(sort.Reverse(similarityArray))
		score = similarityArray[0].Score

		if score > 0.0 {
			person = similarityArray[0].Target
		} else {
			person = nil
			err = errors.New("名寄せ失敗: " + nameLabel)
		}
	}

	return person, similarityArray, err
}

// ParseMemberListFromHTML は、名簿のHTMLファイルをパースしてMemberList を返すメソッドです。
func ParseMemberListFromHTML(reader io.Reader) (m MemberList, err error) {
	const defaultRole = "委員"
	const query = "#contentsMain table tr"
	spacePattern := regexp.MustCompile("[\\s　]+")

	m = MemberList{}

	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	document.Find(query).Each(func(idx int, selection *goquery.Selection) {
		member := Person{}

		// 委員のIDとしてUUIDを生成する
		uuid, err := uuid.NewRandom()
		if err != nil {
			log.Fatal(err)
		}
		member.ID = uuid.String()

		member.Role = selection.Find("th").First().Text()

		if len(member.Role) <= 0 {
			member.Role = defaultRole
		}

		node := selection.Find("td")

		member.Name = spacePattern.ReplaceAllString(node.First().Text(), "")
		member.Affiliation = node.Next().Text()
		member.Label = member.NormarizeLabel()

		m.Members = append(m.Members, &member)
	})

	return m, err
}

// LoadMemberListFromHTML は、引数として与えられたファイルパスから名簿HTMLファイルを開いてパースし、MemberListを返すメソッドです。
func LoadMemberListFromHTML(filepath string) (memberList MemberList, err error) {
	var reader *os.File
	reader, err = os.Open(filepath)
	defer reader.Close()

	memberList, err = ParseMemberListFromHTML(reader)
	if err != nil {
		log.Fatal(err)
	}

	return memberList, err
}

// LoadMemberListFromURL は、引数として与えられたURLから名簿HTMLファイルを取得し、名簿をパースするメソッドです。
func LoadMemberListFromURL(url string) (memberList MemberList, err error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return memberList, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return memberList, err
	}

	if response.StatusCode >= 400 {
		err := errors.New(url + " : " + response.Status)
		return memberList, err
	}

	memberList, err = ParseMemberListFromHTML(response.Body)

	return memberList, err
}
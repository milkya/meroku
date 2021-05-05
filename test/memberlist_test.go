package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tsunekawa/meroku/internal/model"
)

func ExampleParseMemberListFromHTML() {
	baseDir := "../data/example/memberlist"
	filepath := filepath.Join(baseDir, "example01.htm")

	reader, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	memberList, err := model.ParseMemberListFromHTML(reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", len(memberList.Members))
	//Output: 4
}

func ExampleLoadMemberListFromHTML() {
	baseDir := "../data/example/memberlist"
	filepath := filepath.Join(baseDir, "example01.htm")

	memberList, err := model.LoadMemberListFromHTML(filepath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", len(memberList.Members))
	//Output: 4
}

func ExampleLoadMemberListFromURL() {
	url := "https://www.mext.go.jp/b_menu/shingi/chukyo/chukyo3/meibo/1421854.htm"
	memberList, err := model.LoadMemberListFromURL(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", len(memberList.Members))
	//Output: 38
}

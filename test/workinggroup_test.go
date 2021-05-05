package main

import (
	"fmt"
	"log"

	"github.com/tsunekawa/meroku/internal/model"
)

func ExampleWorkingGroup_GetMemberListURLs_one() {
	workingGroup := new(model.WorkingGroup)
	workingGroup.URL = "https://www.mext.go.jp/b_menu/shingi/chukyo/chukyo3/057/index.htm"
	memberListURLs, err   := workingGroup.GetMemberListURLs()
	if (err != nil) {
		log.Fatal(err)
	}

	fmt.Printf("%v", memberListURLs[0])
	//Output: https://www.mext.go.jp/b_menu/shingi/chukyo/chukyo3/meibo/1363293.htm
}

// ワーキンググループが複数の名簿を持つ場合
func ExampleWorkingGroup_GetMemberListURLs_two() {
	workingGroup := new(model.WorkingGroup)
	workingGroup.URL = "https://www.mext.go.jp/b_menu/shingi/chukyo/chukyo3/074/index.htm"
	memberListURLs, err   := workingGroup.GetMemberListURLs()
	if (err != nil) {
		log.Fatal(err)
	}

	for _, url := range memberListURLs {
		fmt.Println(url)
	}

	// Unordered output:
	//https://www.mext.go.jp/b_menu/shingi/chukyo/chukyo3/meibo/1372229.htm
	//https://www.mext.go.jp/b_menu/shingi/chukyo/chukyo3/meibo/1366594.htm
}

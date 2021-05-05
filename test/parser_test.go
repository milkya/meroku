package main

import (
	"fmt"
	"log"

	"github.com/tsunekawa/meroku/internal/model"
)

func ExamplePerson_normarizeLabel() {
	person := model.Person{
		ID:          "test-0001",
		Label:       "山田花子らべる",
		Name:        "山田花子",
		Role:        "主査代理",
		Affiliation: "ラブラドール大学レトリーバー研究科教授",
	}

	fmt.Println(person.NormarizeLabel())
	// Output: 山田花子主査代理
}

func ExampleMemberList_resolve() {
	memberList := model.MemberList{}
	memberList.Members = []*model.Person{
		&(model.Person{
			ID:          "test-0001",
			Label:       "山田花子らべる",
			Name:        "山田花子",
			Role:        "主査代理",
			Affiliation: "ラブラドール大学レトリーバー研究科教授",
		}),
		&(model.Person{
			ID:          "test-0002",
			Label:       "鈴木一郎ラベル",
			Name:        "鈴木一郎",
			Role:        "委員",
			Affiliation: "ブルドッグ工科大学特任教授",
		}),
	}

	person, simArray, _ := memberList.Resolve("山田（花）主査代理")

	log.Printf("Similarity: ")
	for _, sim := range simArray {
		log.Printf("%v (%v)", sim.Target.Label, sim.Score)
	}

	fmt.Println(person.Label)
	// Output: 山田花子らべる
}

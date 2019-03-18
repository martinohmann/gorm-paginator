package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	paginator "github.com/martinohmann/gorm-paginator"
)

type Related struct {
	ID      int
	Model   *Model
	ModelID int `gorm:"foreignkey:ID;association_foreignkey:ModelID"`
	Name    string
}

type Model struct {
	ID      int
	Related []Related `gorm:"foreignkey:ModelID"`
}

func main() {
	db, err := gorm.Open("mysql", "root:root@tcp(mysql)/db?parseTime=true")
	if err != nil {
		panic(err)
	}

	var related []Related

	model := Model{ID: 1}

	options := []paginator.Option{
		paginator.WithPage(2),
		paginator.WithLimit(10),
		paginator.WithOrder("name DESC"),
	}

	res, err := paginator.PaginateRelated(db, &model, &related, "Related", options...)
	if err != nil {
		panic(err)
	}

	fmt.Printf("TotalRecords:   %d\n", res.TotalRecords)
	fmt.Printf("CurrentPage:    %d\n", res.CurrentPage)
	fmt.Printf("MaxPage:        %d\n", res.MaxPage)
	fmt.Printf("RecordsPerPage: %d\n", res.RecordsPerPage)
	fmt.Printf("IsFirstPage?:   %v\n", res.IsFirstPage())
	fmt.Printf("IsLastPage?:    %v\n", res.IsLastPage())

	for _, record := range res.Records.([]Related) {
		fmt.Printf("ID:   %d", record.ID)
		fmt.Printf("Name: %s", record.Name)
	}
}

package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	paginator "github.com/martinohmann/gorm-paginator"
)

type model struct {
	gorm.Model
	Name string
}

func main() {
	db, err := gorm.Open("mysql", "root:root@tcp(mysql)/db?parseTime=true")
	if err != nil {
		panic(err)
	}

	var m []model

	options := []paginator.Option{
		paginator.WithPage(2),
		paginator.WithLimit(10),
		paginator.WithOrder("name DESC"),
	}

	res, err := paginator.Paginate(db, &m, options...)
	if err != nil {
		panic(err)
	}

	fmt.Printf("TotalRecords:   %d\n", res.TotalRecords)
	fmt.Printf("CurrentPage:    %d\n", res.CurrentPage)
	fmt.Printf("MaxPage:        %d\n", res.MaxPage)
	fmt.Printf("RecordsPerPage: %d\n", res.RecordsPerPage)
	fmt.Printf("IsFirstPage?:   %v\n", res.IsFirstPage())
	fmt.Printf("IsLastPage?:    %v\n", res.IsLastPage())

	for _, record := range res.Records.([]model) {
		fmt.Printf("ID:   %d", record.ID)
		fmt.Printf("Name: %s", record.Name)
	}
}

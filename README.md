gorm-paginator
==============

[![Build Status](https://travis-ci.org/martinohmann/gorm-paginator.svg)](https://travis-ci.org/martinohmann/gorm-paginator)
[![codecov](https://codecov.io/gh/martinohmann/gorm-paginator/branch/master/graph/badge.svg)](https://codecov.io/gh/martinohmann/gorm-paginator)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinohmann/gorm-paginator)](https://goreportcard.com/report/github.com/martinohmann/gorm-paginator)
[![GoDoc](https://godoc.org/github.com/martinohmann/gorm-paginator?status.svg)](https://godoc.org/github.com/martinohmann/gorm-paginator)

A simple paginator for [gorm](https://github.com/jinzhu/gorm). Also supports direct pagination via http.Request query parameters.

Installation
------------

```sh
go get -u github.com/martinohmann/gorm-paginator
```

Usage
-----

### Basic usage

```go
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
```

### Pagination via http.Request query params

```go
package main

import (
	"encoding/json"
	"net/http"

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

	// Example pagination request: /model?page=2&order=name+DESC&limit=10
	//
	// Check the godoc for paginator.WithRequest and paginator.ParamNames to
	// see how to configure the parameter names.
	http.Handle("/model", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m []model

		res, err := paginator.Paginate(db, &m, paginator.WithRequest(r))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(res)
	}))

	http.ListenAndServe(":8080", nil)
}
```

### Pagination of associations

This works for has-many and many-to-many relations.

```go
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
```

License
-------

The source code of gorm-paginator is released under the MIT License. See the bundled
LICENSE file for details.

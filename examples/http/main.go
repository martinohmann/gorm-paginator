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

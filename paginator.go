// Package paginator provides a simple paginator implementation for gorm. It
// also supports configuring the paginator via http.Request query params.
package paginator

import (
	"github.com/jinzhu/gorm"
)

// DefaultLimit defines the default limit for paginated queries. This is a
// variable so that users can configure it at runtime.
var DefaultLimit = 20

// Paginator defines the interface for a paginator.
type Paginator interface {
	// Paginate takes a value as arguments and returns a paginated result
	// containing records of the value type.
	Paginate(interface{}) (*Result, error)
}

// paginator defines a paginator.
type paginator struct {
	db    *gorm.DB
	limit int
	page  int
	order []string
}

// countResult defines the result of the count query executed by the paginator.
type countResult struct {
	total int
	err   error
}

// Result defines a paginated result.
type Result struct {
	CurrentPage    int         `json:"currentPage"`
	MaxPage        int         `json:"maxPage"`
	RecordsPerPage int         `json:"recordsPerPage"`
	TotalRecords   int         `json:"totalRecords"`
	Records        interface{} `json:"records"`
}

// New create a new value of the Paginator type. It expects a gorm DB handle
// and pagination options.
//     var v []SomeModel
//     p := paginator.New(db, paginator.WithPage(2))
//     res, err := p.Paginate(&v)
func New(db *gorm.DB, options ...Option) Paginator {
	p := &paginator{
		db:    db,
		page:  1,
		limit: DefaultLimit,
		order: make([]string, 0),
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// Paginate is a convenience wrapper for the paginator.
//     var v []SomeModel
//     res, err := paginator.Paginate(db, &v, paginator.WithPage(2))
func Paginate(db *gorm.DB, value interface{}, options ...Option) (*Result, error) {
	return New(db, options...).Paginate(value)
}

// Paginate implements the Paginator interface.
func (p *paginator) Paginate(value interface{}) (*Result, error) {
	db := p.prepareDB()

	c := make(chan countResult, 1)

	go countRecords(db, value, c)

	err := db.Limit(p.limit).Offset(p.offset()).Find(value).Error

	if err != nil {
		<-c
		return nil, err
	}

	return p.result(value, <-c)
}

// prepareDB prepares the statement by adding the order clauses.
func (p *paginator) prepareDB() *gorm.DB {
	db := p.db

	for _, o := range p.order {
		db = db.Order(o)
	}

	return db
}

// offset computes the offset used for the paginated query.
func (p *paginator) offset() int {
	return (p.page - 1) * p.limit
}

// countRecords counts the result rows for given query and returns the result
// in the provided channel.
func countRecords(db *gorm.DB, value interface{}, c chan<- countResult) {
	var result countResult
	result.err = db.Model(value).Count(&result.total).Error
	c <- result
}

// result creates a new Result out of the retrieved value and the count query
// result.
func (p *paginator) result(value interface{}, c countResult) (*Result, error) {
	if c.err != nil {
		return nil, c.err
	}

	maxPageF := float64(c.total) / float64(p.limit)
	maxPage := int(maxPageF)

	if float64(maxPage) < maxPageF {
		maxPage++
	} else if maxPage == 0 {
		maxPage = 1
	}

	return &Result{
		TotalRecords:   c.total,
		Records:        value,
		CurrentPage:    p.page,
		RecordsPerPage: p.limit,
		MaxPage:        maxPage,
	}, nil
}

// IsLastPage returns true if the current page of the result is the last page.
func (r *Result) IsLastPage() bool {
	return r.CurrentPage >= r.MaxPage
}

// IsFirstPage returns true if the current page of the result is the first page.
func (r *Result) IsFirstPage() bool {
	return r.CurrentPage <= 1
}

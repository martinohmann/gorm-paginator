package paginator

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

type model struct {
	ID   int
	Name string
}

type Foo struct {
	ID    int
	Baz   *Baz
	BazID int   `gorm:"foreignkey:ID;association_foreignkey:BazID"`
	Bars  []Bar `gorm:"many2many:bars_foos"`
}

type Bar struct {
	ID   int
	Foos []Foo `gorm:"many2many:bars_foos"`
}

type Baz struct {
	ID   int
	Foos []Foo `gorm:"foreignkey:BazID"`
}

func createMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	matcher := sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual)
	mdb, mock, err := sqlmock.New(matcher)
	if err != nil {
		t.Fatalf("unexpected error which creating sqlmock.Sqlmock: %s", err)
	}

	mock.MatchExpectationsInOrder(false)

	db, err := gorm.Open("mysql", mdb)
	if err != nil {
		t.Fatalf("unexpected error while creating gorm.DB: %s", err)
	}

	return db, mock
}

func TestPaginate(t *testing.T) {
	cases := []struct {
		name           string
		pagedQuery     string
		totalRecords   int
		currentPage    int
		maxPage        int
		recordsPerPage int
		options        []Option
	}{
		{
			name:           "zero rows",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 0",
			totalRecords:   0,
			currentPage:    1,
			maxPage:        1,
			recordsPerPage: 20,
		},
		{
			name:           "less than full page",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 0",
			totalRecords:   7,
			currentPage:    1,
			maxPage:        1,
			recordsPerPage: 20,
		},
		{
			name:           "full page",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 0",
			totalRecords:   20,
			currentPage:    1,
			maxPage:        1,
			recordsPerPage: 20,
		},
		{
			name:           "two pages",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 0",
			totalRecords:   21,
			currentPage:    1,
			maxPage:        2,
			recordsPerPage: 20,
		},
		{
			name:           "with page option",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 20",
			totalRecords:   21,
			currentPage:    2,
			maxPage:        2,
			recordsPerPage: 20,
			options:        []Option{WithPage(2)},
		},
		{
			name:           "page option exceeds maxPage",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 40",
			totalRecords:   21,
			currentPage:    3,
			maxPage:        2,
			recordsPerPage: 20,
			options:        []Option{WithPage(3)},
		},
		{
			name:           "invalid page option",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 0",
			totalRecords:   21,
			currentPage:    1,
			maxPage:        2,
			recordsPerPage: 20,
			options:        []Option{WithPage(0)},
		},
		{
			name:           "with limit option",
			pagedQuery:     "SELECT * FROM `models` LIMIT 10 OFFSET 0",
			totalRecords:   21,
			currentPage:    1,
			maxPage:        3,
			recordsPerPage: 10,
			options:        []Option{WithLimit(10)},
		},
		{
			name:           "invalid limit option",
			pagedQuery:     "SELECT * FROM `models` LIMIT 20 OFFSET 0",
			totalRecords:   21,
			currentPage:    1,
			maxPage:        2,
			recordsPerPage: 20,
			options:        []Option{WithLimit(0)},
		},
		{
			name:           "with order option",
			pagedQuery:     "SELECT * FROM `models` ORDER BY name DESC LIMIT 20 OFFSET 0",
			totalRecords:   21,
			currentPage:    1,
			maxPage:        2,
			recordsPerPage: 20,
			options:        []Option{WithOrder("name DESC")},
		},
		{
			name:           "with multiple order options",
			pagedQuery:     "SELECT * FROM `models` ORDER BY name DESC,`id` LIMIT 20 OFFSET 0",
			totalRecords:   21,
			currentPage:    1,
			maxPage:        2,
			recordsPerPage: 20,
			options:        []Option{WithOrder("name DESC", "id")},
		},
		{
			name:           "with multiple options",
			pagedQuery:     "SELECT * FROM `models` ORDER BY name ASC LIMIT 2 OFFSET 8",
			totalRecords:   21,
			currentPage:    5,
			maxPage:        11,
			recordsPerPage: 2,
			options:        []Option{WithOrder("name ASC"), WithPage(5), WithLimit(2)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := createMockDB(t)

			rows := sqlmock.NewRows([]string{"id", "name"})
			mock.ExpectQuery(tc.pagedQuery).WillReturnRows(rows)

			countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(tc.totalRecords)
			mock.ExpectQuery("SELECT count(*) FROM `models`").WillReturnRows(countRows)

			var m []model

			res, err := Paginate(db, &m, tc.options...)
			if err != nil {
				t.Fatalf("unexpected error while calling Paginate: %s", err)
			}

			if res.TotalRecords != tc.totalRecords {
				t.Fatalf("expected res.TotalRecords of %d, got %d", tc.totalRecords, res.TotalRecords)
			}

			if res.CurrentPage != tc.currentPage {
				t.Fatalf("expected res.CurrentPage to be %d, got %d", tc.currentPage, res.CurrentPage)
			}

			if res.RecordsPerPage != tc.recordsPerPage {
				t.Fatalf("expected res.RecordsPerPage to be %d, got %d", tc.recordsPerPage, res.RecordsPerPage)
			}

			if res.MaxPage != tc.maxPage {
				t.Fatalf("expected res.MaxPage to be %d, got %d", tc.maxPage, res.MaxPage)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestQueryError(t *testing.T) {
	db, mock := createMockDB(t)

	expectedError := errors.New("some error")

	mock.ExpectQuery("SELECT * FROM `models` LIMIT 20 OFFSET 0").
		WillReturnError(expectedError)
	mock.ExpectQuery("SELECT count(*) FROM `models`").
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))

	var m []model

	_, err := Paginate(db, &m)
	if err == nil {
		t.Fatal("expected error while calling Paginate but got nil")
	}

	if err != expectedError {
		t.Fatalf("expected error %q while calling Paginate but got %q", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCountQueryError(t *testing.T) {
	db, mock := createMockDB(t)

	expectedError := errors.New("some error")

	mock.ExpectQuery("SELECT * FROM `models` LIMIT 20 OFFSET 0").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery("SELECT count(*) FROM `models`").
		WillReturnError(expectedError)

	var m []model

	_, err := Paginate(db, &m)
	if err == nil {
		t.Fatal("expected error while calling Paginate but got nil")
	}

	if err != expectedError {
		t.Fatalf("expected error %q while calling Paginate but got %q", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPaginateRelatedManyToMany(t *testing.T) {
	db, mock := createMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("SELECT `bars`.* FROM `bars` INNER JOIN `bars_foos` ON `bars_foos`.`bar_id` = `bars`.`id` WHERE (`bars_foos`.`foo_id` IN (?)) LIMIT 20 OFFSET 0").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(0)
	mock.ExpectQuery("SELECT count(*) FROM `bars` INNER JOIN `bars_foos` ON `bars_foos`.`bar_id` = `bars`.`id` WHERE (`bars_foos`.`foo_id` IN (?))").
		WillReturnRows(countRows)

	foo := Foo{ID: 1}

	var bars []Bar

	_, err := PaginateRelated(db, &foo, &bars, "Bars")
	if err != nil {
		t.Fatalf("unexpected error while calling Paginate: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPaginateRelatedHasMany(t *testing.T) {
	db, mock := createMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("SELECT * FROM `foos` WHERE (`baz_id` = ?) LIMIT 20 OFFSET 0").WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(0)
	mock.ExpectQuery("SELECT count(*) FROM `foos` WHERE (`baz_id` IN (?))").WillReturnRows(countRows)

	baz := Baz{ID: 1}

	var foos []Foo

	_, err := PaginateRelated(db, &baz, &foos, "Foos")
	if err != nil {
		t.Fatalf("unexpected error while calling Paginate: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRelatedQueryError(t *testing.T) {
	db, mock := createMockDB(t)

	expectedError := errors.New("some error")

	mock.ExpectQuery("SELECT * FROM `foos` WHERE (`baz_id` = ?) LIMIT 20 OFFSET 0").
		WillReturnError(expectedError)

	mock.ExpectQuery("SELECT count(*) FROM `foos` WHERE (`baz_id` IN (?))").
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))

	baz := Baz{ID: 1}

	var foos []Foo

	_, err := PaginateRelated(db, &baz, &foos, "Foos")
	if err == nil {
		t.Fatal("expected error while calling Paginate but got nil")
	}

	if err != expectedError {
		t.Fatalf("expected error %q while calling PaginateRelated but got %q", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRelatedCountQueryError(t *testing.T) {
	db, mock := createMockDB(t)

	expectedError := errors.New("some error")

	mock.ExpectQuery("SELECT * FROM `foos` WHERE (`baz_id` = ?) LIMIT 20 OFFSET 0").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery("SELECT count(*) FROM `foos` WHERE (`baz_id` IN (?))").
		WillReturnError(expectedError)

	baz := Baz{ID: 1}

	var foos []Foo

	_, err := PaginateRelated(db, &baz, &foos, "Foos")
	if err == nil {
		t.Fatal("expected error while calling Paginate but got nil")
	}

	if err != expectedError {
		t.Fatalf("expected error %q while calling PaginateRelated but got %q", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

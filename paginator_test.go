package paginator

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

type model struct {
	ID   int
	Name string
}

func createMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, error) {
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

	return db, mock, nil
}

func TestPaginate(t *testing.T) {
	db, mock, err := createMockDB(t)

	p := New(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "foo")
	rows2 := sqlmock.NewRows([]string{"count(*)"}).AddRow(45)
	mock.ExpectQuery("SELECT * FROM `models` LIMIT 20 OFFSET 0").WillReturnRows(rows)
	mock.ExpectQuery("SELECT count(*) FROM `models`").WillReturnRows(rows2)

	var m []model

	res, err := p.Paginate(&m)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if res.TotalRecords != 45 {
		t.Fatalf("expected total record count of %d, got %d", 45, res.TotalRecords)
	}

	if res.CurrentPage != 1 {
		t.Fatalf("expected current page to be %d, got %d", 1, res.CurrentPage)
	}

	if res.MaxPage != 3 {
		t.Fatalf("expected max page to be %d, got %d", 3, res.MaxPage)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithOrder(t *testing.T) {
	db, mock, err := createMockDB(t)

	p := New(db, WithOrder("name DESC", "id"))

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "foo")
	rows2 := sqlmock.NewRows([]string{"count(*)"}).AddRow(45)
	mock.ExpectQuery("SELECT * FROM `models` ORDER BY name DESC,`id` LIMIT 20 OFFSET 0").WillReturnRows(rows)
	mock.ExpectQuery("SELECT count(*) FROM `models`").WillReturnRows(rows2)

	var m []model

	_, err = p.Paginate(&m)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithLimit(t *testing.T) {
	db, mock, err := createMockDB(t)

	p := New(db, WithLimit(100))

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "foo")
	rows2 := sqlmock.NewRows([]string{"count(*)"}).AddRow(40)
	mock.ExpectQuery("SELECT * FROM `models` LIMIT 100 OFFSET 0").WillReturnRows(rows)
	mock.ExpectQuery("SELECT count(*) FROM `models`").WillReturnRows(rows2)

	var m []model

	res, err := p.Paginate(&m)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if res.MaxPage != 1 {
		t.Fatalf("expected max page to be %d, got %d", 1, res.MaxPage)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithPage(t *testing.T) {
	db, mock, err := createMockDB(t)

	p := New(db, WithPage(2))

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "foo")
	rows2 := sqlmock.NewRows([]string{"count(*)"}).AddRow(40)
	mock.ExpectQuery("SELECT * FROM `models` LIMIT 20 OFFSET 20").WillReturnRows(rows)
	mock.ExpectQuery("SELECT count(*) FROM `models`").WillReturnRows(rows2)

	var m []model

	res, err := p.Paginate(&m)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if res.TotalRecords != 40 {
		t.Fatalf("expected total record count of %d, got %d", 40, res.TotalRecords)
	}

	if res.CurrentPage != 2 {
		t.Fatalf("expected current page to be %d, got %d", 2, res.CurrentPage)
	}

	if res.MaxPage != 2 {
		t.Fatalf("expected max page to be %d, got %d", 2, res.MaxPage)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

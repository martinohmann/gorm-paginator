package paginator

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestWithRequest(t *testing.T) {
	r := &http.Request{
		URL: &url.URL{
			RawQuery: "page=2&order=name+desc,id&limit=100",
		},
	}

	p := &paginator{}

	WithRequest(r)(p)

	expectedOrder := []string{"name desc", "id"}

	if p.page != 2 {
		t.Fatalf("expected paginator.page to be 2 but got %d", p.page)
	}

	if p.limit != 100 {
		t.Fatalf("expected paginator.limit to be 100 but got %d", p.limit)
	}

	if !reflect.DeepEqual(expectedOrder, p.order) {
		t.Fatalf("expected paginator.order to be %#v but got %#v", expectedOrder, p.order)
	}
}

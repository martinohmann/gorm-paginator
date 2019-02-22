package paginator

import (
	"net/http"
	"strconv"
	"strings"
)

// Option defines the signature of a paginator option function.
type Option func(p *paginator)

// ParamNames defines a type to configure names of query parameters to use from
// a http.Request. If a field is set to the empty string, it will not be used.
type ParamNames struct {
	Page  string
	Limit string
	Order string
}

// DefaultParamNames specifies the query parameter names to use from
// a http.Request by default when the WithRequest option is uses. This can be
// overridden at runtime.
var DefaultParamNames = ParamNames{"page", "limit", "order"}

// WithPage configures the page of the paginator.
//     paginator.Paginate(db, &v, paginator.WithPage(2))
func WithPage(page int) Option {
	return func(p *paginator) {
		if page > 0 {
			p.page = page
		}
	}
}

// WithLimit configures the limit of the paginator.
//     paginator.Paginate(db, &v, paginator.WithLimit(10))
func WithLimit(limit int) Option {
	return func(p *paginator) {
		if limit > 0 {
			p.limit = limit
		}
	}
}

// WithOrder configures the order of the paginator.
//     paginator.Paginate(db, &v, paginator.WithOrder("name DESC", "id"))
func WithOrder(order ...string) Option {
	return func(p *paginator) {
		p.order = filterNonEmpty(order)
	}
}

// WithRequest configures the paginator from a *http.Request.
//     paginator.Paginate(db, &v, paginator.WithRequest(request))
//
//     paginator.Paginate(db, &v, paginator.WithRequest(request, paginator.ParamNames{
//         Page: "page",
//         Limit: "",       // Disable limit query param.
//         Order: "order",
//     }))
func WithRequest(r *http.Request, paramNames ...ParamNames) Option {
	params := DefaultParamNames

	if len(paramNames) > 0 {
		params = paramNames[0]
	}

	return func(p *paginator) {
		if value, ok := getQueryParam(r, params.Page); ok {
			if page, err := strconv.Atoi(value); err == nil {
				WithPage(page)(p)
			}
		}

		if value, ok := getQueryParam(r, params.Limit); ok {
			if limit, err := strconv.Atoi(value); err == nil {
				WithLimit(limit)(p)
			}
		}

		if value, ok := getQueryParam(r, params.Order); ok {
			if order := strings.TrimSpace(value); len(order) > 0 {
				WithOrder(strings.Split(order, ",")...)(p)
			}
		}
	}
}

// getQueryParam gets the first query param matching key from the request.
// Returns empty string of key or param value is empty. Second return value
// indicates wether the param was present in the query or not.
func getQueryParam(r *http.Request, key string) (string, bool) {
	if key == "" {
		return "", false
	}

	if values, ok := r.URL.Query()[key]; ok && len(values) > 0 {
		return values[0], true
	}

	return "", false
}

// filterNonEmpty filters out all elements that are either empty or contain
// solely whitespace characters.
func filterNonEmpty(elements []string) []string {
	nonEmpty := make([]string, 0)

	for _, el := range elements {
		if el = strings.TrimSpace(el); len(el) > 0 {
			nonEmpty = append(nonEmpty, el)
		}
	}

	return nonEmpty
}

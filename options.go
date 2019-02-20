package paginator

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Option defines the signature of a paginator option function.
type Option func(p *paginator)

// ParamNames defines a type to configure names of query parameters to use from
// the gin.Context.
type ParamNames struct {
	Page  string
	Limit string
	Order string
}

// defaultParamNames specifies the query parameter names to use from
// gin.Context if nothing is configured.
var defaultParamNames = ParamNames{"page", "limit", "order"}

// WithPage configures the page of the paginator.
func WithPage(page int) Option {
	return func(p *paginator) {
		if page > 0 {
			p.page = page
		}
	}
}

// WithLimit configures the limit of the paginator.
func WithLimit(limit int) Option {
	return func(p *paginator) {
		if limit > 0 {
			p.limit = limit
		}
	}
}

// WithOrder configures the order of the paginator.
func WithOrder(order ...string) Option {
	return func(p *paginator) {
		p.order = onlyNonEmpty(order)
	}
}

// WithGinContext configures the paginator for a gin.Context.
func WithGinContext(c *gin.Context, paramNames ...ParamNames) Option {
	params := defaultParamNames

	if len(paramNames) > 0 {
		params = paramNames[0]
	}

	return func(p *paginator) {
		if params.Page != "" {
			page, err := strconv.Atoi(c.Query(params.Page))
			if err == nil {
				WithPage(page)(p)
			}
		}

		if params.Page != "" {
			limit, err := strconv.Atoi(c.Query(params.Limit))
			if err == nil {
				WithLimit(limit)(p)
			}
		}

		if params.Order != "" {
			order := strings.TrimSpace(c.Query(params.Order))
			if len(order) > 0 {
				WithOrder(strings.Split(order, ",")...)(p)
			}
		}
	}
}

// onlyNonEmpty filters out all elements that are either empty or contain
// solely whitespace characters.
func onlyNonEmpty(elements []string) []string {
	nonEmpty := make([]string, 0)

	for _, el := range elements {
		el = strings.TrimSpace(el)
		if len(el) > 0 {
			nonEmpty = append(nonEmpty, el)
		}
	}

	return nonEmpty
}

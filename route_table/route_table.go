package route_table

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type RoTable struct {
	routes           []route
	echoString       http.HandlerFunc
	randDigit        http.HandlerFunc
	incrementCounter http.HandlerFunc
	headersString    http.HandlerFunc
	dataPost         http.HandlerFunc
}

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc // type funcAsField func(int, int) int
}

type ctxKey struct {
}

func (p *RoTable) newRoute(method, pattern string, handler http.HandlerFunc) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

func (p *RoTable) Init(
	echoString http.HandlerFunc,
	randDigit http.HandlerFunc,
	incrementCounter http.HandlerFunc,
	headersString http.HandlerFunc,
	dataPost http.HandlerFunc,
) {
	p.echoString = echoString
	p.randDigit = randDigit
	p.incrementCounter = incrementCounter
	p.randDigit = randDigit
	p.dataPost = dataPost

	p.routes = []route{
		p.newRoute("GET",  "/", p.echoString),
		p.newRoute("GET",  "/rand", p.randDigit),
		p.newRoute("GET",  "/inc", p.incrementCounter),
		p.newRoute("GET",  "/headers", p.headersString),
		p.newRoute("POST", "/data", p.dataPost),
	}
}

func (p *RoTable) Serve(w http.ResponseWriter, r *http.Request) {
	var allow []string
	for _, route := range p.routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				allow = append(allow, route.method)
				continue
			}
			ctxkey := ctxKey{}
			ctx := context.WithValue(r.Context(), ctxkey, matches[1:])
			if ctx == nil {
				fmt.Println("!!! ctx is nil")
			}
			if r == nil {
				fmt.Println("!!! r is nil")
			}
			route.handler(w, r.WithContext(ctx))

			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)

		return
	}
	http.NotFound(w, r)
}

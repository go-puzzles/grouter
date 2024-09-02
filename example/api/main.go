package main

import (
	"fmt"
	"net/http"

	"github.com/go-puzzles/prouter"
	"github.com/go-puzzles/puzzles/plog"
	"github.com/gorilla/mux"
)

type routers []prouter.Route

func (r routers) Routes() []prouter.Route {
	return r
}

var (
	myRouters = routers{prouter.NewRoute(http.MethodGet, "/test", prouter.HandleFunc(helloHandler))}
)

func helloHandler(ctx *prouter.Context) (prouter.Response, error) {
	name := ctx.Var("name")
	return prouter.SuccessResponse("hello world" + name), nil
}

func testMiddleware(ctx *prouter.Context) (prouter.Response, error) {
	fmt.Println(111)
	return nil, nil
}

type Data struct {
	Name string `json:"name" uri:"name"`
}

type Resp struct {
	Data string `json:"data"`
}

func bodyParseTestHandler(ctx *prouter.Context, data *Data) (*string, error) {
	fmt.Println("data", data)

	var resp *string
	resp = new(string)
	*resp = fmt.Sprintf("Hello, %s!", data.Name)

	if err := ctx.Session().Set("testsess", "test"); err != nil {
		return nil, err
	}
	return resp, nil
}

func parseUriHandler(ctx *prouter.Context, data *Data) (*string, error) {
	fmt.Println("data", data)

	var resp *string
	resp = new(string)
	*resp = fmt.Sprintf("Hello, %s!", data.Name)

	return resp, nil
}

func main() {
	prouter.SetMode(prouter.DebugMode)
	router := prouter.NewProuter()
	router.UseMiddleware(prouter.NewSessionMiddleware("testsession"))

	router.HandleRoute(http.MethodGet, "/hello/noheader/{name}", helloHandler)
	router.HandleRoute(http.MethodGet, "/hello/header/{name}", helloHandler, func(route *mux.Route) *mux.Route {
		return route.Headers("Content-Type", "application/json", "X-Requested-With", "XMLHttpRequest")
	})
	router.POST("/hello/parse", prouter.BodyParserHandleFunc(bodyParseTestHandler))
	router.GET("/hello/{name}", prouter.BodyParserHandleFunc(parseUriHandler))
	router.HandlerRouter(myRouters)
	router.Static("/static", "./content")

	group := router.Group("/group1")
	group.Use(testMiddleware)
	group.HandleRoute(http.MethodGet, "/hello2/{name}", helloHandler)

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	plog.PanicError(srv.ListenAndServe())
}

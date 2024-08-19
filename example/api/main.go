package main

import (
	"fmt"
	"net/http"

	"github.com/go-puzzles/plog"
	"github.com/go-puzzles/prouter"
	"github.com/gorilla/mux"
)

type routers []prouter.Route

func (r routers) Routes() []prouter.Route {
	return r
}

var (
	myRouters = routers{prouter.NewRoute(http.MethodGet, "/test", prouter.HandleFunc(helloHandler))}
)

func helloHandler(ctx *prouter.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (prouter.Response, error) {
	panic("test err")
	// return prouter.SuccessResponse("hello world"), nil
}

func testMiddleware(ctx *prouter.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (prouter.Response, error) {
	fmt.Println(111)
	return nil, nil
}

type Data struct {
	Name string `json:"name"`
}

type Resp struct {
	Data string `json:"data"`
}

func bodyParseTestHandler(ctx *prouter.Context, data *Data) (*string, error) {
	fmt.Println("data", data)

	var resp *string
	resp = new(string)
	*resp = fmt.Sprintf("Hello, %s!", data.Name)
	return resp, nil

	// return &Resp{Data: fmt.Sprintf("Hello, %s!", data.Name)}, nil
}

func main() {
	prouter.SetMode(prouter.DebugMode)
	router := prouter.NewProuter()

	router.HandleRoute(http.MethodGet, "/hello/{name}", prouter.HandleFunc(helloHandler), func(route *mux.Route) *mux.Route {
		return route.Headers("Content-Type", "application/json", "X-Requested-With", "XMLHttpRequest")
	})
	router.POST("/hello/parse", prouter.BodyParserHandleFunc(bodyParseTestHandler))
	router.HandlerRouter(myRouters)
	router.Static("/static", "./content")

	group := router.Group("/group1")
	group.UseMiddleware(prouter.HandleFunc(testMiddleware))
	group.HandleRoute(http.MethodGet, "/hello2/{name}", prouter.HandleFunc(helloHandler))

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	plog.PanicError(srv.ListenAndServe())
}

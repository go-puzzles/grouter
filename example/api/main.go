package main

import (
	"errors"
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
	return prouter.SuccessResponse("hello world" + name).SetCode(2100), nil
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

func panicRouterTest(ctx *prouter.Context) (prouter.Response, error) {
	panic("test panic")
}

func errorRouterTest(ctx *prouter.Context) (prouter.Response, error) {
	return nil, prouter.NewErr(5112, errors.New("test error"), "this is error message").SetResponseType(prouter.BadRequest)
}

type SuccessResponse struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

func (s *SuccessResponse) SetCode(code int) prouter.Response {
	s.Code = code
	return s
}

func (s *SuccessResponse) SetData(data any) prouter.Response {
	s.Data = data
	return s
}

func (s *SuccessResponse) SetMessage(_ string) prouter.Response {
	return s
}

func (s *SuccessResponse) GetCode() int {
	return s.Code
}

func (s *SuccessResponse) GetMessage() string {
	return "Success"
}

func (s *SuccessResponse) GetData() any {
	return s
}

func (s *SuccessResponse) GetError() error {
	return nil
}

func SuccessReturn(data any) *SuccessResponse {
	return &SuccessResponse{
		Code: http.StatusOK,
		Data: data,
	}
}

func main() {
	prouter.SetMode(prouter.DebugMode)
	router := prouter.NewProuter()
	router.UseMiddleware(prouter.NewSessionMiddleware("testsession"))

	router.HandleRoute(http.MethodGet, "/noheader/hello/{name}", helloHandler)
	router.HandleRoute(http.MethodGet, "/header/hello/{name}", helloHandler, func(route *mux.Route) *mux.Route {
		return route.Headers("X-Requested-With", "XMLHttpRequest")
	})
	router.POST("/hello/parse", prouter.BodyParser(bodyParseTestHandler))
	router.GET("/hello/{name}", prouter.BodyParser(parseUriHandler))
	router.GET("/panic", panicRouterTest)
	router.GET("/error", errorRouterTest)
	router.HandleRouter(myRouters)
	router.Static("/static", "./content")

	group := router.Group("/group1")
	group.Use(testMiddleware)
	group.HandleRoute(http.MethodGet, "/hello2/{name}", helloHandler)

	router.GET("/path", func(ctx *prouter.Context) (prouter.Response, error) {
		return prouter.SuccessResponse("path test").SetCode(2100), nil
	})

	router.GET("/path/hello", func(ctx *prouter.Context) (prouter.Response, error) {
		return SuccessReturn("path/hello test"), nil
		// return prouter.SuccessResponse("path/hello test").SetCode(2100), nil
	})

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	plog.PanicError(srv.ListenAndServe())
}

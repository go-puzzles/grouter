package prouter

import "net/http"

var mapCodeToStatus = func(code int) (statusCode int) {
	if http.StatusText(code) != "" {
		return code
	}
	
	switch {
	case code >= 5000 && code < 6000:
		statusCode = http.StatusInternalServerError
	case code >= 4000 && code < 5000:
		statusCode = http.StatusBadRequest
	case code >= 2000 && code < 3000:
		statusCode = http.StatusOK
	default:
		statusCode = http.StatusOK
	}
	
	return
}

func SetMapCodeToStatusFunc(fn func(code int) (statusCode int)) {
	mapCodeToStatus = fn
}

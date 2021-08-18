package webrouter

import (
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	writedHeader bool
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.writedHeader {
		rw.writedHeader = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

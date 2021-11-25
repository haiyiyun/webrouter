package webrouter

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
	"sync"
)

type ResponseWriter struct {
	mu sync.RWMutex
	http.ResponseWriter
	wroteHeader bool
	getResData  bool
	data        map[string]interface{}
	resData     []byte
}

func (rw *ResponseWriter) SetGetResData(getResData bool) {
	rw.mu.Lock()
	rw.getResData = getResData
	rw.mu.Unlock()
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if rw.getResData {
		rw.mu.Lock()
		buf := bytes.NewBuffer(rw.resData)
		buf.Write(b)
		rw.resData = buf.Bytes()
		rw.mu.Unlock()
	}

	return rw.ResponseWriter.Write(b)
}

func (rw *ResponseWriter) GetResData() []byte {
	return rw.resData
}

func (rw *ResponseWriter) initData() {
	rw.mu.Lock()
	if rw.data == nil {
		rw.data = map[string]interface{}{}
	}
	rw.mu.Unlock()
}

func (rw *ResponseWriter) SetData(key string, value interface{}) {
	rw.initData()

	rw.mu.Lock()
	rw.data[key] = value
	rw.mu.Unlock()
}

func (rw *ResponseWriter) GetData(key string) (interface{}, bool) {
	rw.initData()

	rw.mu.RLock()
	value, found := rw.data[key]
	rw.mu.RUnlock()

	return value, found
}

func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}

	return nil, nil, errors.New("response does not implement http.Hijacker")
}

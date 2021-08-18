package webrouter

import (
	"net/http"
	"time"
)

var (
	DefaultRouter = NewManager()
	DefaultServer = http.Server{
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}
)

func SetFilterPrefix(filterPrefix string) {
	DefaultRouter.SetFilterPrefix(filterPrefix)
}

func GetFilterPrefix() string {
	return DefaultRouter.GetFilterPrefix()
}

func SetDelimiterStyle(delimiterStyle byte) {
	DefaultRouter.SetDelimiterStyle(delimiterStyle)
}

func GetDelimiterStyle() byte {
	return DefaultRouter.GetDelimiterStyle()
}

func SetDefaultMethodName(methodName string) {
	DefaultRouter.SetDefaultMethodName(methodName)
}

func GetDefaultMethodName() string {
	return DefaultRouter.GetDefaultMethodName()
}

func SetHTTPMethodNamePrefix(methodNamePrefix string) {
	DefaultRouter.SetHTTPMethodNamePrefix(methodNamePrefix)
}

func GetHTTPMethodNamePrefix() string {
	return DefaultRouter.GetHTTPMethodNamePrefix()
}

func SetBeforeMethodName(methodName string) {
	DefaultRouter.SetBeforeMethodName(methodName)
}

func ClearBeforeMethodName() {
	DefaultRouter.ClearBeforeMethodName()
}

func GetBeforeMethodName() []string {
	return DefaultRouter.GetBeforeMethodName()
}

func SetAfterMethodName(methodName string) {
	DefaultRouter.SetAfterMethodName(methodName)
}

func ClearAfterMethodName() {
	DefaultRouter.ClearAfterMethodName()
}

func GetAfterMethodName() []string {
	return DefaultRouter.GetAfterMethodName()
}

func GetFilterMethodNameAndHTTPMethodName(methodName string) (string, string) {
	return DefaultRouter.GetFilterMethodNameAndHTTPMethodName(methodName)
}

func ResponseWriter(responseWriter http.ResponseWriter) {
	DefaultRouter.ResponseWriter(responseWriter)
}

func SetCloser(fn func()) {
	DefaultRouter.SetCloser(fn)
}

func Close() {
	DefaultRouter.Close()
}

func MakePattern(method string) string {
	return makePattern(method, DefaultRouter.GetDelimiterStyle())
}

func Registers() map[string]register {
	return DefaultRouter.Registers()
}

func Register(patternRoot string, i interface{}) {
	DefaultRouter.Register(patternRoot, i)
}

func Handle(pattern string, handler http.Handler) {
	DefaultRouter.ServeMux.Handle(pattern, handler)
}

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	DefaultRouter.ServeMux.HandleFunc(pattern, handler)
}

func Handler(req *http.Request) (h http.Handler, pattern string) {
	return DefaultRouter.ServeMux.Handler(req)
}

func NotFoundHandler(error string) {
	DefaultRouter.NotFoundHandler(error)
}

func NotFoundHtmlHandler(error string) {
	DefaultRouter.NotFoundHtmlHandler(error)
}

func ServeHTTP(w http.ResponseWriter, req *http.Request) {
	DefaultRouter.ServeHTTP(w, req)
}

func Injector(name, follower string, priority uint, handler func(http.ResponseWriter, *http.Request) bool) {
	DefaultRouter.Injector(name, follower, priority, handler)
}

func Releasor(name, leader string, lag uint, handler func(http.ResponseWriter, *http.Request) bool) {
	DefaultRouter.Releasor(name, leader, lag, handler)
}

func ListenAndServe(addr string, handler http.Handler) error {
	DefaultServer.Addr = addr
	if handler == nil {
		DefaultServer.Handler = DefaultRouter
	} else {
		DefaultServer.Handler = handler
	}

	return DefaultServer.ListenAndServe()
}

func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	DefaultServer.Addr = addr
	if handler == nil {
		DefaultServer.Handler = DefaultRouter
	} else {
		DefaultServer.Handler = handler
	}

	return DefaultServer.ListenAndServeTLS(certFile, keyFile)
}

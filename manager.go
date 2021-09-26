package webrouter

import (
	"net/http"
	"reflect"
	"strings"
)

type register struct {
	I interface{}
	reflect.Type
	reflect.Value
}

type Manager struct {
	*http.ServeMux
	injections           []injector
	releases             []releasor
	notFoundHandle       http.Handler
	filterPrefix         string
	delimiterStyle       byte
	defaultMethodName    string
	httpMethodNamePrefix string
	beforeMethodName     []string
	afterMethodName      []string
	registers            map[string]register
	responseWriter       http.ResponseWriter
	closer               []func()
}

type filterHTTPMethod struct {
	httpMethod string
	rcvm       reflect.Value
}

func NewManager() *Manager {
	rm := &Manager{
		ServeMux:             http.NewServeMux(),
		injections:           []injector{},
		releases:             []releasor{},
		filterPrefix:         "Route",
		delimiterStyle:       '-',
		defaultMethodName:    "Default",
		httpMethodNamePrefix: "Route_",
		beforeMethodName:     []string{},
		afterMethodName:      []string{},
		registers:            map[string]register{},
	}

	return rm
}

func (rm *Manager) hasSameMethod(methods []string, method string) bool {
	for _, mtd := range methods {
		if mtd == method {
			return true
		}
	}

	return false
}

//if filterPrefix value is '@' that mean not to filter, but it is has hidden danger, so you kown what to do.
func (rm *Manager) SetFilterPrefix(filterPrefix string) *Manager {
	if filterPrefix == "@" {
		rm.filterPrefix = ""
	} else if filterPrefix != "" {
		rm.filterPrefix = filterPrefix
	}

	return rm
}

func (rm *Manager) GetFilterPrefix() string {
	return rm.filterPrefix
}

func (rm *Manager) SetDelimiterStyle(delimiterStyle byte) *Manager {
	if delimiterStyle > 0 {
		rm.delimiterStyle = delimiterStyle
	}

	return rm
}

func (rm *Manager) GetDelimiterStyle() byte {
	return rm.delimiterStyle
}

func (rm *Manager) SetDefaultMethodName(methodName string) *Manager {
	if methodName != "" {
		rm.defaultMethodName = methodName
	}

	return rm
}

func (rm *Manager) GetDefaultMethodName() string {
	return rm.defaultMethodName
}

func (rm *Manager) SetHTTPMethodNamePrefix(methodNamePrefix string) *Manager {
	if methodNamePrefix != "" {
		rm.httpMethodNamePrefix = methodNamePrefix
	}

	return rm
}

func (rm *Manager) GetHTTPMethodNamePrefix() string {
	return rm.httpMethodNamePrefix
}

func (rm *Manager) SetBeforeMethodName(methodName string) *Manager {
	if !rm.hasSameMethod(rm.beforeMethodName, methodName) {
		rm.beforeMethodName = append(rm.beforeMethodName, methodName)
	}

	return rm
}

func (rm *Manager) ClearBeforeMethodName() *Manager {
	rm.beforeMethodName = []string{}

	return rm
}

func (rm *Manager) GetBeforeMethodName() []string {
	return rm.beforeMethodName
}

func (rm *Manager) SetAfterMethodName(methodName string) *Manager {
	if !rm.hasSameMethod(rm.afterMethodName, methodName) {
		rm.afterMethodName = append(rm.afterMethodName, methodName)
	}

	return rm
}

func (rm *Manager) ClearAfterMethodName() *Manager {
	rm.afterMethodName = []string{}

	return rm
}

func (rm *Manager) GetAfterMethodName() []string {
	return rm.afterMethodName
}

func (rm *Manager) ResponseWriter(writer http.ResponseWriter) {
	rm.responseWriter = writer
}

func (rm *Manager) SetCloser(fn func()) {
	rm.closer = append(rm.closer, fn)
}

func (rm *Manager) Close() {
	for _, close := range rm.closer {
		close()
	}
}

func (rm *Manager) Injector(name, follower string, priority uint, handler func(http.ResponseWriter, *http.Request) bool) {
	if hasSameInjector(rm.injections, name) {
		panic("multiple registrations injector for " + name)
	}

	rm.injections = append(rm.injections, injector{
		name:     name,
		follower: follower,
		priority: int(priority),
		h:        handler,
	})

	rm.injections = sortInjector(rm.injections)
}

func (rm *Manager) Releasor(name, leader string, lag uint, handler func(http.ResponseWriter, *http.Request) bool) {
	if hasSameReleasor(rm.releases, name) {
		panic("multiple registrations releasor for " + name)
	}

	rm.releases = append(rm.releases, releasor{
		name:   name,
		leader: leader,
		lag:    int(lag),
		h:      handler,
	})

	rm.releases = sortReleasor(rm.releases)
}

func (rm *Manager) Registers() map[string]register {
	regs := map[string]register{}
	for k, v := range rm.registers {
		regs[k] = v
	}

	return regs
}

/*
Priority:

1. [<beforeMethodName>_method] | [beforeMethodName]

2. [method]

3. [http_<method>_method]

4. [<afterMethodName>_method] | [afterMethodName]
*/
func (rm *Manager) Register(patternRoot string, i interface{}) {
	rm.registers[patternRoot] = register{
		I:     i,
		Type:  reflect.TypeOf(i),
		Value: reflect.ValueOf(i),
	}

	//使用反射的值来make新的值
	rcvi := reflect.New(rm.registers[patternRoot].Type).Elem()
	rcvi.Set(rm.registers[patternRoot].Value)
	rcti := rcvi.Type()

	filterPrefix := rm.filterPrefix
	delimiterStyle := rm.delimiterStyle

	rcvhm := rm.findHTTPMethod(rcti, rcvi)

	handlePattern := map[string]bool{}

	for i := 0; i < rcti.NumMethod(); i++ {
		mName := rcti.Method(i).Name
		if filterMethodName, _ := rm.GetFilterMethodNameAndHTTPMethodName(mName); filterMethodName != "" {
			pattern := patternRoot
			if filterMethodName != rm.defaultMethodName {
				pattern += makePattern(filterMethodName, delimiterStyle)
			}

			if !handlePattern[pattern] {
				handlePattern[pattern] = true

				var rcvmbs, rcvmas []reflect.Value
				for _, beforeMethodName := range rm.beforeMethodName {
					beforeMethodNamePrefix := beforeMethodName + "_" + filterMethodName
					if _, ok := rcti.MethodByName(beforeMethodNamePrefix); ok {
						rcvmbs = append(rcvmbs, rcvi.MethodByName(beforeMethodNamePrefix))
					} else {
						if _, hasBeforeMethodName := rcti.MethodByName(beforeMethodName); hasBeforeMethodName {
							rcvmbs = append(rcvmbs, rcvi.MethodByName(beforeMethodName))
						}
					}
				}

				for _, afterMethodName := range rm.afterMethodName {
					afterMethodNamePrefix := afterMethodName + "_" + filterMethodName
					if _, found := rcti.MethodByName(afterMethodNamePrefix); found {
						rcvmas = append(rcvmas, rcvi.MethodByName(afterMethodNamePrefix))
					} else {
						if _, found := rcti.MethodByName(afterMethodName); found {
							rcvmas = append(rcvmas, rcvi.MethodByName(afterMethodName))
						}
					}
				}

				rm.ServeMux.Handle(pattern, rm.makeHandler(rcvi.MethodByName(filterPrefix+filterMethodName), rcvmbs, rcvhm[filterMethodName], rcvmas))
			}
		}
	}
}

func (rm *Manager) GetFilterMethodNameAndHTTPMethodName(methodName string) (filterMethodName, httpMethodName string) {
	filterPrefix := rm.filterPrefix
	httpMethodNamePrefix := rm.httpMethodNamePrefix
	if filterPrefix == httpMethodNamePrefix {
		if strings.HasPrefix(methodName, filterPrefix) {
			filterMethodName = methodName[len(filterPrefix):]
			if mpos := strings.Index(filterMethodName, "_"); mpos != -1 {
				filterMethodName = filterMethodName[mpos+1:]
				httpMethodName = filterMethodName[:mpos]
			}
		}
	} else {
		if strings.HasPrefix(methodName, httpMethodNamePrefix) {
			hmname := methodName[len(httpMethodNamePrefix):]
			if mpos := strings.Index(hmname, "_"); mpos != -1 {
				filterMethodName = hmname[mpos+1:]
				httpMethodName = hmname[:mpos]
			}
		}

		if filterMethodName == "" && strings.HasPrefix(methodName, filterPrefix) {
			filterMethodName = methodName[len(filterPrefix):]
		}
	}

	return
}

func (rm *Manager) findHTTPMethod(rcti reflect.Type, rcvi reflect.Value) map[string][]filterHTTPMethod {
	rcvhm := make(map[string][]filterHTTPMethod)
	for i := 0; i < rcti.NumMethod(); i++ {
		mName := rcti.Method(i).Name
		objMethod, httpMethod := rm.GetFilterMethodNameAndHTTPMethodName(mName)
		if httpMethod != "" {
			rcvhm[objMethod] = append(rcvhm[objMethod], filterHTTPMethod{
				httpMethod: httpMethod,
				rcvm:       rcvi.Method(i),
			})
		}
	}

	return rcvhm
}

/*
NewPattern => new[delimiterStyle]pattern
delimiterStyle => -
NewPattern => new-pattern
*/
func makePattern(method string, delimiterStyle byte) string {
	var c byte
	bl := byte('a' - 'A')
	l := len(method)
	pattern := make([]byte, 0, l+8)
	for i := 0; i < l; i++ {
		c = method[i]
		if c >= 'A' && c <= 'Z' {
			c += bl
			if i > 0 {
				pattern = append(pattern, delimiterStyle)
			}
		}

		pattern = append(pattern, c)
	}

	return string(pattern)
}

/*
workflow:

1. [<beforeMethodName>_method] | [beforeMethodName]

2. [method]

3. [http_<method>_method]

4. [<afterMethodName>_method] | [afterMethodName]

router.Xxx Can return one result of bool, if result is ture mean return func
*/
func (rm *Manager) makeHandler(rcvm reflect.Value, rcvmbs []reflect.Value, rcvhm []filterHTTPMethod, rcvmas []reflect.Value) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rcvw, rcvr := reflect.ValueOf(w), reflect.ValueOf(req)

		hfm := filterHTTPMethod{}
		for _, fm := range rcvhm {
			if fm.httpMethod != "" && fm.httpMethod == strings.ToUpper(req.Method) {
				hfm = fm
				break
			}
		}

		if rcvm.IsValid() || hfm.httpMethod != "" {
			for _, rcvmb := range rcvmbs {
				if arv := callMethod(rcvmb, rcvw, rcvr); len(arv) > 0 && arv[0].Bool() {
					return
				}
			}

			if rcvm.IsValid() {
				if arv := callMethod(rcvm, rcvw, rcvr); len(arv) > 0 && arv[0].Bool() {
					return
				}
			}

			if hfm.httpMethod != "" {
				if arv := callMethod(hfm.rcvm, rcvw, rcvr); len(arv) > 0 && arv[0].Bool() {
					return
				}
			}

			for _, rcvma := range rcvmas {
				if arv := callMethod(rcvma, rcvw, rcvr); len(arv) > 0 && arv[0].Bool() {
					return
				}
			}
		} else {
			if rm.notFoundHandle != nil {
				rm.notFoundHandle.ServeHTTP(w, req)
			} else {
				http.NotFound(w, req)
			}
		}
	})
}

func callMethod(rcvm, rcvw, rcvr reflect.Value) (arv []reflect.Value) {
	mt := rcvm.Type()
	mtni := mt.NumIn()
	switch mtni {
	case 1:
		if mt.In(0) == rcvr.Type() {
			arv = rcvm.Call([]reflect.Value{rcvr})
		} else {
			arv = rcvm.Call([]reflect.Value{rcvw})
		}
	case 2:
		if mt.In(0) == rcvr.Type() {
			arv = rcvm.Call([]reflect.Value{rcvr, rcvw})
		} else {
			arv = rcvm.Call([]reflect.Value{rcvw, rcvr})
		}
	default:
		arv = rcvm.Call([]reflect.Value{})
	}

	return
}

func (rm *Manager) NotFoundHandler(errstr string) {
	rm.notFoundHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Error(w, errstr, http.StatusNotFound)
	})
}

func (rm *Manager) NotFoundHtmlHandler(errstr string) {
	rm.notFoundHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Error(w, errstr, http.StatusNotFound, CtHTMLHeader)
	})
}

func (rm *Manager) Handler(r *http.Request) (h http.Handler, pattern string) {
	h, pattern = rm.ServeMux.Handler(r)
	if pattern == "" && rm.notFoundHandle != nil {
		h = rm.notFoundHandle
	}

	return
}

//processing order: injector > handler > releasor
func (rm *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rm.responseWriter = &responseWriter{
		ResponseWriter: w,
	}

	for _, injection := range rm.injections {
		if abort := injection.h(rm.responseWriter, r); abort {
			return
		}
	}

	h, _ := rm.Handler(r)
	h.ServeHTTP(rm.responseWriter, r)

	for _, release := range rm.releases {
		if abort := release.h(rm.responseWriter, r); abort {
			return
		}
	}
}

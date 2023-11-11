package sigeaserver2

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

type AjaxObject interface {
	Build(builder *AjaxHandler)
}

type AjaxMethod struct {
	name   string
	params []string
}

type AjaxHandler struct {
	objectName  string
	ajaxObject  AjaxObject
	ajaxMethods []AjaxMethod
}

type AjaxMethodHandler struct {
	objectName string
	ajaxObject AjaxObject
	methodInfo AjaxMethod
	method     reflect.Value
}

func extractParams(h *AjaxMethodHandler, r *http.Request) ([]reflect.Value, error) {
	params := []reflect.Value{}
	for n := 0; n < h.method.Type().NumIn(); n++ {
		if r == nil {
			log.Printf("**** R is NULL\n")
			return params, fmt.Errorf("**** R is NULL\n")
		}
		if n >= len(h.methodInfo.params) {
			log.Printf("**** N out of range [%v] [%v] [%v] [%v]\n", h, n, h.methodInfo.params, h.method.Type().NumIn())
			return params, fmt.Errorf("**** N out of range [%v] [%v] [%v] [%v]\n", h, n, h.methodInfo.params, h.method.Type().NumIn())
		}
		pr := r.FormValue(h.methodInfo.params[n])
		tp := h.method.Type().In(n)

		var param reflect.Value
		if tp.String() == "int" {
			intValue, _ := strconv.Atoi(pr)
			param = reflect.ValueOf(intValue)
		} else if tp.String() == "string" {
			param = reflect.ValueOf(pr)
		} else if tp.String() == "float32" {
			f64, _ := strconv.ParseFloat(pr, 32)
			param = reflect.ValueOf(float32(f64))
		} else if tp.String() == "float64" {
			f64, _ := strconv.ParseFloat(pr, 64)
			param = reflect.ValueOf(float64(f64))
		} else if tp.String() == "[]string" {
			var par []string
			json.Unmarshal([]byte(pr), &par)
			param = reflect.ValueOf(par)
		} else if tp.String() == "[]int" {
			var par []int
			json.Unmarshal([]byte(pr), &par)
			param = reflect.ValueOf(par)
		} else {
			return params, fmt.Errorf("Tipo desconhecido [%v]", tp.Name())
		}
		params = append(params, param)
	}
	return params, nil
}

func (h *AjaxMethodHandler) handleAjaxCall(w http.ResponseWriter, r *http.Request) {
	var rr map[string]interface{}
	rr = make(map[string]interface{})
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Add("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	pp := make(map[string]interface{})

	params, err := extractParams(h, r)
	log.Printf("Call to method [%v]\n", h.methodInfo.name)
	for n, p := range params {
		log.Printf("  Param value [%v] [%v]\n", h.methodInfo.params[n], p)
		pp[h.methodInfo.params[n]] = r.FormValue(h.methodInfo.params[n])
	}
	if err == nil {
		result := h.method.Call(params)
		var jsonResult interface{}
		jsonResult = result[0].Interface()

		rr["ok"] = true
		rr["result"] = jsonResult
	} else {
		rr["ok"] = false
		rr["result"] = fmt.Sprintf("%v", err)
	}
	json, _ := json.Marshal(rr)
	w.Write(json)

	if ajaxCallRegister != nil {
		callLog := CallLog{MethodName: h.methodInfo.name, Ip: r.RemoteAddr, CallParams: pp, CallResult: string(json)}
		ajaxCallRegister.RegisterCallLog(callLog)
	}
}

func (s *SigeaServer) AddAjaxObject(objectName string, ajaxObject AjaxObject) {
	ajaxMethods := []AjaxMethod{}
	handler := AjaxHandler{objectName, ajaxObject, ajaxMethods}
	ajaxObject.Build(&handler)
	s.ajaxHandlers = append(s.ajaxHandlers, handler)
}

func (h *AjaxHandler) AddMethod(methodName string, params ...string) {
	h.ajaxMethods = append(h.ajaxMethods, AjaxMethod{methodName, params})
}

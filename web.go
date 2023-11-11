package sigeaserver2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type WebHandler struct {
	pattern string
	handler func(http.ResponseWriter, *http.Request)
}

type SigeaServerManager struct {
	s *SigeaServer
}

func (ssm *SigeaServerManager) Set(methodName, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	ssm.s.webHandlers = append(ssm.s.webHandlers, WebHandler{pattern, handler})
}

func (ssm *SigeaServerManager) GetParserAndLog(request *http.Request, response http.ResponseWriter) reqParserAndLog {
	InsertAllowOrigin(response, request)
	parameters := make(map[string]interface{})
	resp := reqParserAndLog{response, request, parameters}
	if err := json.NewDecoder(request.Body).Decode(&resp.parameters); err != nil {
		fmt.Println(err.Error())
	}
	return resp
}

//func (s *SigeaServer) AddWebHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
//	s.webHandlers = append(s.webHandlers, WebHandler{pattern, handler})
//}

func (s *SigeaServer) AddWebHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.webHandlers = append(s.webHandlers, WebHandler{pattern, handler})
}

func (s *SigeaServer) AddWebHandlerSsl(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.webHandlersSsl = append(s.webHandlersSsl, WebHandler{pattern, handler})
}

type reqParserAndLog struct {
	response   http.ResponseWriter
	request    *http.Request
	parameters map[string]interface{}
}

func (r *reqParserAndLog) ParamInt(nomeParam string) (int, error) {
	if val, ok := r.parameters[nomeParam]; ok {
		return int(val.(float64)), nil
	} else {
		return 0, fmt.Errorf("Valor não encontrado")
	}
}

func (r *reqParserAndLog) ParamObject(nomeParam string) (interface{}, error) {
	if val, ok := r.parameters[nomeParam]; ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("Valor não encontrado")
	}
}

func (r *reqParserAndLog) ParamFloat(nome string) (float64, error) {
	if val, ok := r.parameters[nome]; ok {
		return val.(float64), nil
	} else {
		return 0, fmt.Errorf("Valor não encontrado")
	}
}

func (r *reqParserAndLog) ParamStruct(nome string, m interface{}) error {
	if val, ok := r.parameters[nome]; ok {
		bytes, err := json.Marshal(val.(map[string]interface{}))
		if err != nil {
			return err
		}
		return json.Unmarshal(bytes, m)
	} else {
		return fmt.Errorf("Valor não encontrado")
	}
}

func (r *reqParserAndLog) ParamArrayStruct(nome string, m interface{}) error {
	if val, ok := r.parameters[nome]; ok {
		t := val.([]interface{})
		mapa := make([]map[string]interface{}, 0)
		for _, v := range t {
			mapa = append(mapa, v.(map[string]interface{}))
		}
		bytes, err := json.Marshal(mapa)
		if err != nil {
			return err
		}
		return json.Unmarshal(bytes, m)
	} else {
		return fmt.Errorf("Valor não encontrado")
	}
}

func (r *reqParserAndLog) ParamBool(nomeParam string) (bool, error) {
	value := r.parameters[nomeParam].(bool)
	return value, nil
}

func (r *reqParserAndLog) ParamString(nomeParam string) string {
	value := r.parameters[nomeParam].(string)
	fmt.Printf("  %v : [%v]\n", nomeParam, value)
	return value
}

func (r *reqParserAndLog) ParamTime(nomeParam, layout string) (time.Time, error) {
	value, ok := r.parameters[nomeParam].(string)
	if !ok {
		return time.Time{}, fmt.Errorf("Valor não encontrado")
	}
	fmt.Printf("  %v : [%v]\n", nomeParam, value)
	return time.Parse(layout, value)
}

func (r *reqParserAndLog) WriteJsonAndLog(resp interface{}, err error) {
	writeResponseJson(r.response, resp, err)
}

func InsertAllowOrigin(response http.ResponseWriter, request *http.Request) {
	origin := request.Header.Get("Origin")
	if origin != "" {
		response.Header().Add("Access-Control-Allow-Origin", origin)
	}
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func writeResponseJson(response http.ResponseWriter, resp interface{}, err error) {
	rr := make(map[string]interface{})
	if err == nil {
		rr["ok"] = true
		rr["result"] = resp
	} else {
		rr["ok"] = false
		rr["msg"] = err.Error()
	}
	js, err := json.Marshal(rr)
	if err != nil {
		writeResponseJson(response, nil, err)
		return
	}
	response.Write(js)
	fmt.Println(string(js))
	response.Write([]byte("\n"))
}

func obtemParamFloat(request *http.Request, nomeParam string) (float64, error) {
	return strconv.ParseFloat(request.FormValue(nomeParam), 64)
}

package sigeaserver2

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"text/template"

	"github.com/gorilla/mux"
)

func (s *SigeaServer) runTemplateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	arquivo, _ := vars["arquivo"]

	fmt.Println("arquivo ", arquivo)
	fmt.Println("vars ", vars)

	fileName := "web/" + arquivo + ".html"
	file, _ := ioutil.ReadFile(fileName)

	tmpl, err := template.New(arquivo).Parse(string(file))

	if err != nil {
		http.Error(w, fmt.Sprintf("Template nao encontrado [%v] [%v]", arquivo, fileName), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	methodName := "RunTemplate" + makeFirstUpperCase(arquivo)
	mm := reflect.ValueOf(s.templateHandler).MethodByName(methodName)
	if mm.IsValid() {
		p := make(map[string]string)
		for k, v := range r.URL.Query() {
			p[k] = v[0]
		}
		params := []reflect.Value{reflect.ValueOf(p)}

		log.Println("params ", params)

		rr := mm.Call(params)

		log.Println("resultado ", rr)

		ctx := rr[0].Interface()
		err := rr[1].Interface()

		if err == nil {
			w.Header().Set("Content-Type", "text/html")
			tmpl.Execute(w, ctx)
		} else {
			http.Error(w, fmt.Sprintf("Erro executando metodo %v %v", methodName, err), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, fmt.Sprintf("Metodo nao encontrado %v", methodName), http.StatusInternalServerError)
		return
	}
}

//func (s *SigeaServer) AddTemplateHandler(handler interface{}) {
//s.templateHandler = handler
//s.AddWebHandler("/v/{arquivo}.html", s.runTemplateHandler)
//}

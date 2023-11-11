package sigeaserver2

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type SigeaServer struct {
	webHandlers     []WebHandler
	webHandlersSsl  []WebHandler
	ajaxHandlers    []AjaxHandler
	webSocketInfos  []WebSocketInfo
	tcpHandlers     []HandleConn
	templateHandler interface{}
	serveStatic     bool
	useCustomPath   bool
	httpsPort       int
	certPem         string
	certKey         string
	router          *mux.Router
}

func New() *SigeaServer {
	wHandlers := []WebHandler{}
	wHandlersSsl := []WebHandler{}
	aHandlers := []AjaxHandler{}
	sHandlers := []WebSocketInfo{}
	tHandlers := []HandleConn{}
	return &SigeaServer{wHandlers, wHandlersSsl, aHandlers, sHandlers, tHandlers, nil, false, false, 0, "", "", mux.NewRouter()}
}

func (s *SigeaServer) ServeStatic() {
	s.serveStatic = true
}

func (s *SigeaServer) CustomAjaxPath() {
	s.useCustomPath = true
}

func (s *SigeaServer) runTcpServer(port int) {
	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(port))

	for _, h := range s.tcpHandlers {
		fmt.Printf("Handler name [%v]\n", h.Name())
	}

	fmt.Printf("Server tcp started on port %v ...\n", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handleConn(Connection{conn})
	}
}

func makeFirstUpperCase(s string) string {
	if len(s) < 2 {
		return strings.ToUpper(s)
	}

	bts := []byte(s)

	lc := bytes.ToUpper([]byte{bts[0]})
	rest := bts[1:]

	return string(bytes.Join([][]byte{lc, rest}, nil))
}

func (s *SigeaServer) HttpsConfig(httpsPort int, certPem string, certKey string) {
	s.httpsPort = httpsPort
	s.certPem = certPem
	s.certKey = certKey
}

func (s *SigeaServer) runHttpsServer(r *mux.Router) {
	for _, h := range s.webHandlersSsl {
		r.HandleFunc(h.pattern, h.handler)
	}
	err := http.ListenAndServeTLS(":"+strconv.Itoa(s.httpsPort), s.certPem, s.certKey, nil)
	fmt.Printf("Servidor https iniciado na porta %v\n", s.httpsPort)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *SigeaServer) Start(httpPort int, netPort int) {
	r := s.router

	for _, h := range s.webHandlers {
		r.HandleFunc(h.pattern, h.handler)
	}

	for _, h := range s.ajaxHandlers {
		for _, m := range h.ajaxMethods {
			var prefix string
			if s.useCustomPath {
				prefix = ""
			} else {
				prefix = "/ajax"
			}
			pattern := fmt.Sprintf("%v/%v/%v", prefix, h.objectName, m.name)
			mm := reflect.ValueOf(h.ajaxObject).MethodByName(makeFirstUpperCase(m.name))
			if mm.IsValid() {
				mh := AjaxMethodHandler{h.objectName, h.ajaxObject, m, mm}
				r.HandleFunc(pattern, mh.handleAjaxCall)
				fmt.Printf("Criando handler ajax [%v] [%v] [%v]\n", m.name, pattern, mm)
			} else {
				panic(fmt.Errorf("Erro criando handler [%v] [%v]", m.name, pattern))
			}
		}
	}

	for _, h := range s.webSocketInfos {
		http.HandleFunc("/ws/"+h.name, h.handleWebSocket)
		fmt.Printf("Criando handler websocket [/ws/%v]\n", h.name)
	}

	if s.serveStatic {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	}

	http.Handle("/", r)

	if netPort != 0 {
		go s.runTcpServer(netPort)
	}

	if s.httpsPort != 0 {
		go s.runHttpsServer(r)
	}

	fmt.Printf("Servidor http iniciado na porta %v\n", httpPort)
	http.ListenAndServe(":"+strconv.Itoa(httpPort), nil)
}

func (server *SigeaServer) Router() *mux.Router {
	return server.router
}

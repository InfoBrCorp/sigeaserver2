package sigeaserver2

var ajaxCallRegister CallRegister = nil

type CallLog struct {
	MethodName string
	CallParams map[string]interface{}
	Ip         string
	CallResult string
}

type CallRegister interface {
	RegisterCallLog(callLog CallLog)
}

func RegisterAjaxLogCallback(callRegister CallRegister) {
	ajaxCallRegister = callRegister
}

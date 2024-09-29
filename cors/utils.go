package cors

import "github.com/BryanMwangi/pine"

func ParseMethod(method string) bool {
	switch method {
	case pine.MethodGet:
		return true
	case pine.MethodPost:
		return true
	case pine.MethodPut:
		return true
	case pine.MethodPatch:
		return true
	case pine.MethodDelete:
		return true
	default:
		return false
	}
}

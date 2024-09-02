package pine

import (
	"reflect"
	"runtime"
)

func getFunctionName(fn interface{}) string {
	pc := reflect.ValueOf(fn).Pointer()
	function := runtime.FuncForPC(pc)
	if function == nil {
		return "unknown"
	}
	return function.Name()
}

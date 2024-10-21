package pine

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
)

var (
	ErrParse      = errors.New("bind: cannot parse")
	ErrConvert    = errors.New("bind: cannot convert")
	ErrType       = errors.New("bind: unexpected type")
	ErrPtr        = errors.New("bind: destination must be a pointer")
	ErrValidation = errors.New("bind: validation failed")
)

// BindJSON binds the request body to the given interface.
// You can use this to validate the request body without adding further logic
// to your handlers.
//
// Tested with nested JSON objects and arrays
// If you notice any errors, please open an issue on Github and I will fix it right away
func (c *Ctx) BindJSON(v interface{}) error {
	err := json.NewDecoder(c.Request.Body).Decode(v)
	if err != nil {
		return ErrParse
	}
	return bindData(v)
}

// BindParam binds the specified parameter value of a request.
func (c *Ctx) BindParam(key string, v interface{}) error {
	param := c.Params(key)
	if param == "" {
		return ErrValidation
	}
	return bind(param, v)
}

// BindQuery binds the specified query value of a request.
func (c *Ctx) BindQuery(key string, v interface{}) error {
	param := c.Query(key)
	if param == "" {
		return ErrValidation
	}

	return bind(param, v)
}

// Internal helper function to validate the bind
// requires the input in this case is the key of the param or query
// and the destination is the value of the param or query
// you can also specify the type of the destination
func bind(input string, destination interface{}) error {
	// reflect the type and value of the destination
	typ := reflect.TypeOf(destination)
	val := reflect.ValueOf(destination)

	if typ.Kind() != reflect.Ptr {
		return ErrPtr
	}

	// Dereference pointer type to assign value
	val = reflect.Indirect(val)

	switch val.Kind() {
	case reflect.String:
		val.SetString(input)
	case reflect.Int, reflect.Int64:
		parsed, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return ErrConvert
		}
		val.SetInt(parsed)
	case reflect.Float64, reflect.Float32:
		parsed, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return ErrConvert
		}
		val.SetFloat(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(input)
		if err != nil {
			return ErrConvert
		}
		val.SetBool(parsed)
	default:
		return ErrType
	}
	return nil
}

// Called to the bind of the JSON body
// A future revision of this will be implemented to handle forms and XML bodies
// but the logic will pretty much be the same
func bindData(destination interface{}) error {
	v := reflect.ValueOf(destination)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// we can check if the value is a struct or a slice
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if isZeroValue(field) {
				return ErrValidation
			}
		}
	}
	if v.Kind() == reflect.Slice {
		length := v.Len()
		for i := 0; i < length; i++ {
			if isZeroValue(v.Index(i)) {
				return ErrValidation
			}
		}
	}
	return nil
}

// Internal helper function to check if the value is zero
// Returns true if the value is zero and hence handled as an error since
// the unpacked value is not set
func isZeroValue(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int64, reflect.Float64:
		return val.Int() == 0 || val.Float() == 0.0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Slice, reflect.Array:
		// For slices and arrays, check each element
		if val.Len() == 0 {
			return true
		}
		for i := 0; i < val.Len(); i++ {
			if isZeroValue(val.Index(i)) {
				return true
			}
		}
		return false
	case reflect.Map:
		// Maps should be non-nil and have at least one entry
		return val.Len() == 0 || val.IsNil()
	case reflect.Struct:
		// For nested structs, recursively bind their fields
		return bindData(val.Addr().Interface()) != nil
	case reflect.Ptr:
		// For pointers, check if it's nil or dereference it and check its value
		if val.IsNil() {
			return true
		}
		return isZeroValue(val.Elem())
	default:
		// For other types, treat as zero value
		return false
	}
}

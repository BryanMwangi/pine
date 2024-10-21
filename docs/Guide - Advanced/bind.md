---
sidebar_position: 3
---

# Bind

You may have seen the `Bind` example in the landing page or in [the basics section](/docs/Guide%20-%20Basics/ctx.md#bindjson). This is just a very simple helper function that validates that some of the properties of the request body, params or query match the expected type you specify.

I had to implement my own way of validating the fields used by `BindJSON`, `BindParam` and `BindQuery` and if you find yourself needing more sophisticated methods of implementing validations, here are some packages that you can use:

- [github.com/go-playground/validator](https://github.com/go-playground/validator)
- [github.com/go-ozzo/ozzo-validation](https://github.com/go-ozzo/ozzo-validation)

Feel free to use whatever package you wish.

:::caution Experimental
This API and its features are experimental and may be updated soon.
:::

## BindJSON

BindJSON binds the request body to the given interface. You can use this method to validate the the request body before processing the request.

```go
func (c *Ctx) BindJSON(v interface{}) error
```

Let us see how it works under the hood.

First, we use the [json.NewDecoder](https://pkg.go.dev/encoding/json#NewDecoder) to decode the request body into the given interface. Then we call an internal function `bindData` that takes the resulting output and checks for non-zero values. If some of the values are zero or empty strings, it would mean that the request body did not satisfy the requirements of the specified interface.

### bindData

```go
func bindData(destination interface{}) error
```

Here is the full implementation of bindData:

```go
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
```

The function leverages Go's in built [reflect](https://pkg.go.dev/reflect) package to inspect and validate the data in the interface.

1. Handling Pointers:
   The function accepts `destination` which could be of any data type as an argument.First it checks if the destination is a pointer. It then [dereferences](https://stackoverflow.com/a/27084475/22594767) the pointer to access the underlying value which in this case will be obtained by calling the `v.Elem()`

2. Handling Structs and Slices:
   As of now, we handle JSON objects or arrays as defined in the [json standard](https://www.json.org/json-en.html). How we do that:

- For JSON objects, in this case, the equivalent of a struct, we loop over all the fields of the struct and check if they are non-zero.

- For JSON arrays, we loop over all the elements of the array or slice and check if they are non-zero.

### isZeroValue

Let us demystify the isZeroValue function. It is used to check if the value is zero or empty string. It returns a boolean value; true if the value is zero or empty string, false otherwise.

```go
func isZeroValue(val reflect.Value) bool
```

Within isZerorValue, if there are nested structs or slices, we recursively call the function to check if the value is zero or empty string.

```go
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
```

There is no known recursion depth limit as of now, however, if you encounter such a scenario, please open an issue and I will look into it.

:::caution Known Limitation
Currently, BindJSON only supports JSON objects and arrays. If you wish to support other data types, please open an issue and I will look into it.
:::

## BindParam

Validates the request params and binds them to the given interface.

```go
func (c *Ctx) BindParam(key string, v interface{}) error
```

Under the hood it utilizes the `bind` function to validate the request params.

## BindQuery

Validates the request query and binds them to the given interface.

```go
func (c *Ctx) BindQuery(key string, v interface{}) error
```

Under the hood it utilizes the `bind` function to validate the request query.

### bind

The bind function is used to validate the request query or params. It requires the input in this case is the key of the param or query and the destination is the value of the param or query.

```go
func bind(input string, destination interface{}) error
```

Here is the full implementation of the bind function:

```go
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
```

Here are the steps that bind takes:

1. Reflection of the destination:
   The function leverages Go's in built [reflect](https://pkg.go.dev/reflect) package to inspect and validate the data in the interface. It accepts `destination` which could be of any data type as an argument.

2. Check for pointer:
   The function checks if `destination` is a pointer. This is important because it can only assign values to pointers that can be [dereferenced](https://stackoverflow.com/a/27084475/22594767) to the underlying types.
   If the destination is not a pointer, it returns an error.

3. Dereference pointer:
   Dereferencing the pointer is done by calling the `reflect.Indirect` function. This makes sure it works with the actual underlying type.

4. Check for supported types:
   The function checks if the underlying type of the destination is one of the following:

- String
- Int
- Int64
- Float64
- Float32
- Bool

If the underlying type is not one of the above, it returns an error. More types will be added in the future.

5. Parse the input:
   After checking for the supported types, the function parses the input string into the underlying type. This is done using the `strconv` package.

6. Assign the value:
   Finally, the function assigns the parsed value to the destination.

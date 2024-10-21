---
sidebar_position: 5
---

# Cookie

In this section, we will try and go into detail of how cookies are handled in Pine.

Here is how the cookie struct looks like:

```go
// cookie struct that defines the structure of a cookie
type Cookie struct {
    //Name of the cookie
    //
    //Required field
    Name string

    //what data is stored in the cookie
    //
    //Required field
    Value string

    //determines the path in which the cookie is supposed to be used on
    //you can set this to "/" so that every request will contain the cookie
    Path string

    //This allows the browser to associate your cookie with a specific domain
    //when set to example.com cookies from example.com will always be sent
    //with every request to example.com
    Domain string

    //Determines the specific time the cookie expires
    //Max age is more prefered than expires
    Expires time.Time

    //Also sets the expiry date and you can use a string here instead
    RawExpires string

    //Max-Age field of the cookie determines how long the cookie
    // stays in the browser before expiring
    //if you want the cookies to expire immediately such as when a user logs out
    //you can set this to -1
    //
    //accepts int value which should be the time in milliseconds you want
    //the cookie to be stored in the browser
    MaxAge int

    //A boolean value that determines whether cookies will be sent over https
    //or http only.
    //
    //Default is false and http can also send the cookies
    Secure bool

    //determines whether requests over http only can send the cookie
    HttpOnly bool

    //Cookies from the same domain can only be used on the specified domain
    //Eg: cookies from app.example.com can only be used by app.example.com
    //if you want all domains associated with example.com you can set this to
    //*.example.com
    //Now both app.example.com or dev.example.com can use the same cookie
    //
    //Options include the following:
    // 0 - SameSite=Lax
    // 1 - SameSite=Strict
    // 2 - SameSite=None
    //It will alwas default to Lax
    SameSite SameSite

    //All cookie data in string format. You do not need to set this
    //Pine can handle it for you
    Raw bool

    //Pine will also take care of this
    Unparsed []string
}

```

## How to use

In a real world scenario, you would not need all these fields to simply set a cookie to the response instead, the main fields that most browsers would use are:

```go
// cookie struct that defines the structure of a cookie
type Cookie struct {
    Name string
    Value string
    Expires time.Time
    Domain string
    Secure bool
    HttpOnly bool
}
```

## Example

An example of how you could set a cookie in Pine is:

```go
func main() {
    // Create a new cookie
	app := pine.New()

    app.Get("/hello", func(c *pine.Ctx) error{
        cookie := pine.Cookie{
            Name: "session",
            Value: data, // data can be a string such as a JWT token
            Expires: time.Now().Add(time.Hour * 24 * 7),
            Domain: "example.com",
            Secure: true,
            HttpOnly: true,
        }
        return c.SetCookie(cookie).SendString("Hello World")
    })
}
```

You can also use `MaxAge` other than `Expires` however, make sure to set it to unix time milliseconds such that

```go
// expires in 7 days
expires :=time.Now().Add(time.Hour * 24 * 7)

// the same time but in unix milli will be 86400000 milliseconds from now
maxAge := 86400000

// or
maxAge := time.Now().Add(time.Hour * 24 * 7).UnixMilli()
```

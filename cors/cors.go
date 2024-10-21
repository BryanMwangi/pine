package cors

import (
	"net/http"
	"strings"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/logger"
)

type Config struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If set to "*", all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	//
	// Default value is []string{"*"}
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use with
	// cross-domain requests.
	//
	// Default value is simple methods ["GET", "POST", "PUT", "PATCH", "DELETE"]
	// This is the same as in the server.config.DefaultMethods
	AllowedMethods []string

	// AllowedHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	//
	// If the special "*" value is present in the list, all headers will be allowed.
	// Default value is "Content-Type, Authorization"
	AllowedHeaders string

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	//
	// Default value is ""
	ExposedHeaders string

	// AllowedCredentials refers to whether the request can include user credentials
	// like cookies, HTTP authentication or client side SSL certificates.
	//
	// Default value is false
	AllowCredentials bool

	// MaxAge referes to how long the results of a preflight request can be cached
	// by the browser. This is always in seconds.
	//
	// Default value is 0, i.e. the browser does not cache the result.
	// if set to 0, max-age is set to 5 seconds which is the default value set
	// by most browsers.
	MaxAge int
}

func New(config ...Config) pine.Middleware {
	var setConfig Config
	cfg := Config{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   pine.DefaultMethods,
		AllowedHeaders:   "Content-Type, Authorization",
		ExposedHeaders:   "",
		MaxAge:           0,
		AllowCredentials: false,
	}

	if len(config) > 0 {
		setConfig = config[0]
		// Overwrite the default Allowed Origins with the user Allowed Origins
		if setConfig.AllowedOrigins != nil && setConfig.AllowedOrigins[0] != "*" {
			for _, origin := range setConfig.AllowedOrigins {
				origin = strings.TrimSpace(origin)
				origin = strings.ToLower(origin)
				if origin == "" {
					logger.RuntimeError("invalid origin: " + origin)
					continue
				}
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, origin)
			}
		}

		// Overwrite the default Allowed Methods with the user Allowed Methods
		if setConfig.AllowedMethods != nil {
			for _, method := range setConfig.AllowedMethods {
				method = strings.TrimSpace(method)
				method = strings.ToUpper(method)
				// we check if the method is valid
				valid := ParseMethod(method)
				if !valid {
					logger.RuntimeError("invalid method: " + method)
					continue
				}
				cfg.AllowedMethods = append(cfg.AllowedMethods, method)
			}
			if len(cfg.AllowedMethods) == 0 {
				cfg.AllowedMethods = pine.DefaultMethods
			}
		}

		// Overwrite the default Allowed Headers with the user Allowed Headers
		if setConfig.AllowedHeaders != "" {
			setConfig.AllowedHeaders = strings.TrimSpace(setConfig.AllowedHeaders)
			setConfig.AllowedHeaders = strings.ToLower(setConfig.AllowedHeaders)
			cfg.AllowedHeaders = setConfig.AllowedHeaders
		}

		// Overwrite the default Exposed Headers with the user Exposed Headers
		if setConfig.ExposedHeaders != "" {
			cfg.ExposedHeaders = setConfig.ExposedHeaders
		}

		if setConfig.AllowCredentials {
			cfg.AllowCredentials = setConfig.AllowCredentials
		}

		// Overwrite the default MaxAge with the user MaxAge
		if setConfig.MaxAge != 0 {
			cfg.MaxAge = setConfig.MaxAge
		}
	}

	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			// preflight request
			if c.Method == http.MethodOptions {
				c = SetCors(c, cfg)
				return c.SendStatus(http.StatusNoContent)
			}
			c = SetCors(c, cfg)
			return next(c)
		}
	}
}

func SetCors(c *pine.Ctx, cfg Config) *pine.Ctx {
	allowedOrigins := strings.Join(cfg.AllowedOrigins, ",")
	allowedMethods := strings.Join(cfg.AllowedMethods, ",")
	exposeHeaders := strings.TrimSpace(cfg.ExposedHeaders)
	allowHeaders := strings.TrimSpace(cfg.AllowedHeaders)

	c.Set("Access-Control-Allow-Origin", allowedOrigins)
	c.Set("Access-Control-Allow-Methods", allowedMethods)
	c.Set("Access-Control-Allow-Headers", allowHeaders)
	c.Set("Access-Control-Expose-Headers", exposeHeaders)
	c.Set("Access-Control-Allow-Credentials", cfg.AllowCredentials)
	c.Set("Access-Control-Max-Age", cfg.MaxAge)
	return c
}

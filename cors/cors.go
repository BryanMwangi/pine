package cors

import (
	"strings"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/logger"
)

type Config struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	//
	// Default value is ["*"]
	AllowedOrigins []string

	// AllowOriginFunc is a custom function to validate the origin. It take the origin
	// as argument and returns true if allowed or false otherwise. If this option is
	// set, the content of AllowedOrigins is ignored.
	//
	// Default value is nil
	AllowOriginFunc func(origin string) bool

	// AllowedMethods is a list of methods the client is allowed to use with
	// cross-domain requests.
	//
	// Default value is simple methods ["GET", "POST", "PUT", "PATCH", "DELETE"]
	AllowedMethods []string

	// AllowedHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	// If the special "*" value is present in the list, all headers will be allowed.
	// Default value is "" but "Origin" is always appended to the list.
	AllowedHeaders string

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	//
	// Default value is ""
	ExposedHeaders string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached
	// Default value is 0, i.e. the browser does not cache the result.
	// if set to 0, max-age is set to 5 seconds which is the default value set
	// by the browsers.
	MaxAge int
}

var defaultConfig = Config{
	AllowedOrigins:  []string{"*"},
	AllowOriginFunc: nil,
	AllowedMethods: []string{
		pine.MethodGet,
		pine.MethodPost,
		pine.MethodPut,
		pine.MethodPatch,
		pine.MethodDelete,
	},
	AllowedHeaders: "",
	ExposedHeaders: "",
	MaxAge:         0,
}

func New(config ...Config) pine.Middleware {
	var setConfig Config
	var cfg Config

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
		} else {
			cfg.AllowedOrigins = defaultConfig.AllowedOrigins
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
				cfg.AllowedMethods = defaultConfig.AllowedMethods
			}
		} else {
			cfg.AllowedMethods = defaultConfig.AllowedMethods
		}

		if setConfig.AllowOriginFunc != nil {
			cfg.AllowOriginFunc = setConfig.AllowOriginFunc
		} else {
			cfg.AllowOriginFunc = defaultConfig.AllowOriginFunc
		}

		// Overwrite the default Allowed Headers with the user Allowed Headers
		if setConfig.AllowedHeaders != "" {
			setConfig.AllowedHeaders = strings.TrimSpace(setConfig.AllowedHeaders)
			setConfig.AllowedHeaders = strings.ToLower(setConfig.AllowedHeaders)
			cfg.AllowedHeaders = setConfig.AllowedHeaders
		} else {
			cfg.AllowedHeaders = defaultConfig.AllowedHeaders
		}

		// Overwrite the default Exposed Headers with the user Exposed Headers
		if setConfig.ExposedHeaders != "" {
			cfg.ExposedHeaders = setConfig.ExposedHeaders
		} else {
			cfg.ExposedHeaders = defaultConfig.ExposedHeaders
		}

		// Overwrite the default MaxAge with the user MaxAge
		if setConfig.MaxAge != 0 {
			cfg.MaxAge = setConfig.MaxAge
		} else {
			cfg.MaxAge = defaultConfig.MaxAge
		}
	} else {
		cfg = defaultConfig
	}

	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			c = SetCors(c, cfg)
			return next(c)
		}
	}
}

func SetCors(c *pine.Ctx, cfg Config) *pine.Ctx {
	allowedOrigins := strings.Join(cfg.AllowedOrigins, ",")
	allowedMethods := strings.Join(cfg.AllowedMethods, ",")

	c.Set("Access-Control-Allow-Origin", allowedOrigins)
	c.Set("Access-Control-Allow-Methods", allowedMethods)
	c.Set("Access-Control-Allow-Headers", cfg.AllowedHeaders)
	c.Set("Access-Control-Expose-Headers", cfg.ExposedHeaders)
	c.Set("Access-Control-Max-Age", cfg.MaxAge)
	return c
}

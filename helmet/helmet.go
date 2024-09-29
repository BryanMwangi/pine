package helmet

import (
	"fmt"

	"github.com/BryanMwangi/pine"
)

type Config struct {
	// XSSProtection
	// Optional. Default value "0".
	XSSProtection string

	// ContentTypeNosniff
	// Optional. Default value "nosniff".
	ContentTypeNosniff string

	// XFrameOptions
	// Optional. Default value "SAMEORIGIN".
	// Possible values: "SAMEORIGIN", "DENY", "ALLOW-FROM uri"
	XFrameOptions string

	// HSTSMaxAge
	// Optional. Default value 0.
	HSTSMaxAge int

	// HSTSExcludeSubdomains
	// Optional. Default value false.
	HSTSExcludeSubdomains bool

	// ContentSecurityPolicy
	// Optional. Default value "".
	ContentSecurityPolicy string

	// CSPReportOnly
	// Optional. Default value false.
	CSPReportOnly bool

	// HSTSPreloadEnabled
	// Optional. Default value false.
	HSTSPreloadEnabled bool

	// ReferrerPolicy
	// Optional. Default value "ReferrerPolicy".
	ReferrerPolicy string

	// Permissions-Policy
	// Optional. Default value "".
	PermissionPolicy string

	// Cross-Origin-Embedder-Policy
	// Optional. Default value "require-corp".
	CrossOriginEmbedderPolicy string

	// Cross-Origin-Opener-Policy
	// Optional. Default value "same-origin".
	CrossOriginOpenerPolicy string

	// Cross-Origin-Resource-Policy
	// Optional. Default value "same-origin".
	CrossOriginResourcePolicy string

	// Origin-Agent-Cluster
	// Optional. Default value "?1".
	OriginAgentCluster string

	// X-DNS-Prefetch-Control
	// Optional. Default value "off".
	XDNSPrefetchControl string

	// X-Download-Options
	// Optional. Default value "noopen".
	XDownloadOptions string

	// X-Permitted-Cross-Domain-Policies
	// Optional. Default value "none".
	XPermittedCrossDomain string
}

var defaultConfig = Config{
	XSSProtection:             "0",
	ContentTypeNosniff:        "nosniff",
	XFrameOptions:             "SAMEORIGIN",
	HSTSMaxAge:                0,
	HSTSExcludeSubdomains:     false,
	ContentSecurityPolicy:     "",
	CSPReportOnly:             false,
	HSTSPreloadEnabled:        false,
	ReferrerPolicy:            "ReferrerPolicy",
	PermissionPolicy:          "",
	CrossOriginEmbedderPolicy: "require-corp",
	CrossOriginOpenerPolicy:   "same-origin",
	CrossOriginResourcePolicy: "same-origin",
	OriginAgentCluster:        "?1",
	XDNSPrefetchControl:       "off",
	XDownloadOptions:          "noopen",
	XPermittedCrossDomain:     "none",
}

func New(config ...Config) pine.Middleware {
	var setConfig Config
	var cfg Config
	if len(config) > 0 {
		setConfig = config[0]
		// Overwrite the default config with the user config
		if setConfig.XSSProtection != "" {
			cfg.XSSProtection = setConfig.XSSProtection
		} else {
			cfg.XSSProtection = defaultConfig.XSSProtection
		}

		if setConfig.ContentTypeNosniff != "" {
			cfg.ContentTypeNosniff = setConfig.ContentTypeNosniff
		} else {
			cfg.ContentTypeNosniff = defaultConfig.ContentTypeNosniff
		}

		if setConfig.XFrameOptions != "" {
			cfg.XFrameOptions = setConfig.XFrameOptions
		} else {
			cfg.XFrameOptions = defaultConfig.XFrameOptions
		}

		if setConfig.HSTSMaxAge != 0 {
			cfg.HSTSMaxAge = setConfig.HSTSMaxAge
		} else {
			cfg.HSTSMaxAge = defaultConfig.HSTSMaxAge
		}

		if setConfig.HSTSExcludeSubdomains {
			cfg.HSTSExcludeSubdomains = setConfig.HSTSExcludeSubdomains
		} else {
			cfg.HSTSExcludeSubdomains = defaultConfig.HSTSExcludeSubdomains
		}

		if setConfig.ContentSecurityPolicy != "" {
			cfg.ContentSecurityPolicy = setConfig.ContentSecurityPolicy
		} else {
			cfg.ContentSecurityPolicy = defaultConfig.ContentSecurityPolicy
		}

		if setConfig.CSPReportOnly {
			cfg.CSPReportOnly = setConfig.CSPReportOnly
		} else {
			cfg.CSPReportOnly = defaultConfig.CSPReportOnly
		}

		if setConfig.HSTSPreloadEnabled {
			cfg.HSTSPreloadEnabled = setConfig.HSTSPreloadEnabled
		} else {
			cfg.HSTSPreloadEnabled = defaultConfig.HSTSPreloadEnabled
		}

		if setConfig.ReferrerPolicy != "" {
			cfg.ReferrerPolicy = setConfig.ReferrerPolicy
		} else {
			cfg.ReferrerPolicy = defaultConfig.ReferrerPolicy
		}

		if setConfig.PermissionPolicy != "" {
			cfg.PermissionPolicy = setConfig.PermissionPolicy
		} else {
			cfg.PermissionPolicy = defaultConfig.PermissionPolicy
		}

		if setConfig.CrossOriginEmbedderPolicy != "" {
			cfg.CrossOriginEmbedderPolicy = setConfig.CrossOriginEmbedderPolicy
		} else {
			cfg.CrossOriginEmbedderPolicy = defaultConfig.CrossOriginEmbedderPolicy
		}

		if setConfig.CrossOriginOpenerPolicy != "" {
			cfg.CrossOriginOpenerPolicy = setConfig.CrossOriginOpenerPolicy
		} else {
			cfg.CrossOriginOpenerPolicy = defaultConfig.CrossOriginOpenerPolicy
		}

		if setConfig.CrossOriginResourcePolicy != "" {
			cfg.CrossOriginResourcePolicy = setConfig.CrossOriginResourcePolicy
		} else {
			cfg.CrossOriginResourcePolicy = defaultConfig.CrossOriginResourcePolicy
		}

		if setConfig.OriginAgentCluster != "" {
			cfg.OriginAgentCluster = setConfig.OriginAgentCluster
		} else {
			cfg.OriginAgentCluster = defaultConfig.OriginAgentCluster
		}

		if setConfig.XDNSPrefetchControl != "" {
			cfg.XDNSPrefetchControl = setConfig.XDNSPrefetchControl
		} else {
			cfg.XDNSPrefetchControl = defaultConfig.XDNSPrefetchControl
		}

		if setConfig.XDownloadOptions != "" {
			cfg.XDownloadOptions = setConfig.XDownloadOptions
		} else {
			cfg.XDownloadOptions = defaultConfig.XDownloadOptions
		}

		if setConfig.XPermittedCrossDomain != "" {
			cfg.XPermittedCrossDomain = setConfig.XPermittedCrossDomain
		} else {
			cfg.XPermittedCrossDomain = defaultConfig.XPermittedCrossDomain
		}
	} else {
		cfg = defaultConfig
	}

	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			c = SetHelmet(c, cfg)
			return next(c)
		}
	}
}

func SetHelmet(c *pine.Ctx, cfg Config) *pine.Ctx {
	if cfg.XSSProtection != "" {
		c.Set("X-XSS-Protection", cfg.XSSProtection)
	}
	if cfg.ContentTypeNosniff != "" {
		c.Set("X-Content-Type-Options", cfg.ContentTypeNosniff)
	}
	if cfg.XFrameOptions != "" {
		c.Set("X-Frame-Options", cfg.XFrameOptions)
	}
	if cfg.HSTSMaxAge != 0 {
		c.Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge))
	}
	if cfg.HSTSExcludeSubdomains {
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
	if cfg.ContentSecurityPolicy != "" {
		c.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
	}
	if cfg.CSPReportOnly {
		c.Set("Content-Security-Policy-Report-Only", cfg.ContentSecurityPolicy)
	}
	// will investigate later
	// if cfg.HSTSPreloadEnabled {
	// 	c.Set("Public-Key-Pins", "max-age=31536000; pin-sha256=\"")
	// }
	if cfg.ReferrerPolicy != "" {
		c.Set("Referrer-Policy", cfg.ReferrerPolicy)
	}
	if cfg.PermissionPolicy != "" {
		c.Set("Permissions-Policy", cfg.PermissionPolicy)
	}
	if cfg.CrossOriginEmbedderPolicy != "" {
		c.Set("Cross-Origin-Embedder-Policy", cfg.CrossOriginEmbedderPolicy)
	}
	if cfg.CrossOriginOpenerPolicy != "" {
		c.Set("Cross-Origin-Opener-Policy", cfg.CrossOriginOpenerPolicy)
	}
	if cfg.CrossOriginResourcePolicy != "" {
		c.Set("Cross-Origin-Resource-Policy", cfg.CrossOriginResourcePolicy)
	}
	if cfg.OriginAgentCluster != "" {
		c.Set("Origin-Agent-Cluster", cfg.OriginAgentCluster)
	}
	if cfg.XDNSPrefetchControl != "" {
		c.Set("X-DNS-Prefetch-Control", cfg.XDNSPrefetchControl)
	}
	if cfg.XDownloadOptions != "" {
		c.Set("X-Download-Options", cfg.XDownloadOptions)
	}
	if cfg.XPermittedCrossDomain != "" {
		c.Set("X-Permitted-Cross-Domain-Policies", cfg.XPermittedCrossDomain)
	}
	return c
}

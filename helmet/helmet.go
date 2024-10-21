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
	HSTSIncludeSubdomains bool

	// ContentSecurityPolicy
	// Optional. Default value "".
	ContentSecurityPolicy string

	// ReferrerPolicy
	// Optional. Default value "ReferrerPolicy".
	ReferrerPolicy string

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

func New(config ...Config) pine.Middleware {
	cfg := Config{
		XSSProtection:             "0",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "SAMEORIGIN",
		HSTSMaxAge:                0,
		HSTSIncludeSubdomains:     false,
		ContentSecurityPolicy:     "",
		ReferrerPolicy:            "ReferrerPolicy",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
		OriginAgentCluster:        "?1",
		XDNSPrefetchControl:       "off",
		XDownloadOptions:          "noopen",
		XPermittedCrossDomain:     "none",
	}
	if len(config) > 0 {
		useConfig := config[0]
		// Overwrite the default config with the user config
		if useConfig.XSSProtection != "" {
			cfg.XSSProtection = useConfig.XSSProtection
		}

		if useConfig.ContentTypeNosniff != "" {
			cfg.ContentTypeNosniff = useConfig.ContentTypeNosniff
		}

		if useConfig.XFrameOptions != "" {
			cfg.XFrameOptions = useConfig.XFrameOptions
		}
		if useConfig.HSTSMaxAge != 0 {
			cfg.HSTSMaxAge = useConfig.HSTSMaxAge
		}

		if useConfig.HSTSIncludeSubdomains {
			cfg.HSTSIncludeSubdomains = useConfig.HSTSIncludeSubdomains
		}
		if useConfig.ContentSecurityPolicy != "" {
			cfg.ContentSecurityPolicy = useConfig.ContentSecurityPolicy
		}

		if useConfig.ReferrerPolicy != "" {
			cfg.ReferrerPolicy = useConfig.ReferrerPolicy
		}

		if useConfig.CrossOriginEmbedderPolicy != "" {
			cfg.CrossOriginEmbedderPolicy = useConfig.CrossOriginEmbedderPolicy
		}

		if useConfig.CrossOriginOpenerPolicy != "" {
			cfg.CrossOriginOpenerPolicy = useConfig.CrossOriginOpenerPolicy
		}

		if useConfig.CrossOriginResourcePolicy != "" {
			cfg.CrossOriginResourcePolicy = useConfig.CrossOriginResourcePolicy
		}

		if useConfig.OriginAgentCluster != "" {
			cfg.OriginAgentCluster = useConfig.OriginAgentCluster
		}

		if useConfig.XDNSPrefetchControl != "" {
			cfg.XDNSPrefetchControl = useConfig.XDNSPrefetchControl
		}

		if useConfig.XDownloadOptions != "" {
			cfg.XDownloadOptions = useConfig.XDownloadOptions
		}

		if useConfig.XPermittedCrossDomain != "" {
			cfg.XPermittedCrossDomain = useConfig.XPermittedCrossDomain
		}
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
	if cfg.HSTSIncludeSubdomains {
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
	if cfg.ContentSecurityPolicy != "" {
		c.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
	}
	if cfg.ReferrerPolicy != "" {
		c.Set("Referrer-Policy", cfg.ReferrerPolicy)
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

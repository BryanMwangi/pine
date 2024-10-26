// Pine's implementation of rate limiting
// This is a middleware that can be used to rate limit requests to your Pine application

// Note: Rate limiter may reduce the overall performance of your application as it relies
// heavily on the Pine's in memory cache for lookups. Please keep this in mind when using
// rate limiter

package limiter

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/cache"
)

type Config struct {
	// Defines the maximum number of requests a client can make within a specified time
	// window
	//
	// Default: 5
	MaxRequests int

	// Defines the time window between which a client is allowed to make a request.
	// If the client makes more than MaxRequests requests within this time window,
	// the client will be blocked until the time window has passed.
	//
	// Default: 1 second
	Window time.Duration

	// Defines the handler that will be called when a client is blocked or rate limits
	// are exceeded.
	//
	// Default: returns a 429 status code
	Handler pine.Handler

	// Defines whether or not to show the rate limit headers in the response.
	//
	// Default: true
	ShowHeader bool

	// Defines the function that will be used to generate the key for the rate limit.
	// You can use the IP address of the client or you can use the user id of the user.
	//
	// Default: returns the IP address of the client
	KeyGen func(c *pine.Ctx) string

	// Defines a list of IP addresses or user ids that are allowed to make requests.
	// Any IP address defined in the whitelist will be allowed to make requests beyond
	// the rate limit.
	//
	// Default: []string{}
	Whitelist []string

	// Defines a list of IP addresses or user ids that are not allowed to make requests.
	// IP addresses in this list will be blocked whether or not the rate limit is defined.
	//
	// Default: []string{}
	Blacklist []string

	// Internal map for blacklist look up. I used a map instead of a slice because
	// it is faster to check if a key exists in a map than in a slice
	internalBlacklist map[string]struct{}

	// Internal map for fast whitelist look up.
	internalWhitelist map[string]struct{}

	// Defines the store that will be used to store the rate limit data.
	// This is an internal field and you should not need to change it or define it.
	store *cache.Cache
}

// This is the structure of the rate limit data stored in the cache
// This is for internal use and you should not need to change it
type entry struct {
	mu        sync.Mutex
	key       string
	count     int
	reset     time.Time
	remaining int
}

// more information about source for this headers can be found here https://www.ietf.org/archive/id/draft-polli-ratelimit-headers-02.html
const (
	xrateLimitLimit     = "X-RateLimit-Limit"
	xrateLimitRemaining = "X-RateLimit-Remaining"
	xrateLimitReset     = "X-RateLimit-Reset"
)

var (
	ErrBlacklist = errors.New("blacklisted")
)

func defaultHandler(c *pine.Ctx) error {
	return c.SendStatus(429)
}

func defaultKeyGen(c *pine.Ctx) string {
	return c.IP()
}

func New(config ...Config) pine.Middleware {
	cfg := Config{
		MaxRequests:       5,
		Window:            1 * time.Second,
		ShowHeader:        true,
		KeyGen:            defaultKeyGen,
		Whitelist:         []string{},
		Blacklist:         []string{},
		internalWhitelist: make(map[string]struct{}),
		internalBlacklist: make(map[string]struct{}),
		Handler:           defaultHandler,
	}

	// We check if the user has provided any configuration
	// First config is always used as default
	if len(config) > 0 {
		userConfig := config[0]
		if userConfig.MaxRequests != 0 {
			cfg.MaxRequests = userConfig.MaxRequests
		}
		if userConfig.Window != 0 {
			cfg.Window = userConfig.Window
		}
		if userConfig.ShowHeader {
			cfg.ShowHeader = userConfig.ShowHeader
		}
		if userConfig.KeyGen != nil {
			cfg.KeyGen = userConfig.KeyGen
		}
		if userConfig.Whitelist != nil {
			for _, w := range userConfig.Whitelist {
				cfg.internalWhitelist[w] = struct{}{}
			}
		}
		if userConfig.Blacklist != nil {
			for _, b := range userConfig.Blacklist {
				cfg.internalBlacklist[b] = struct{}{}
			}
		}
		if userConfig.Handler != nil {
			cfg.Handler = userConfig.Handler
		}
	}
	cfg.store = cache.New(cfg.Window)

	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			// process the rate limit checker
			e, err := cfg.process(c)

			if err == ErrBlacklist {
				c.Set(xrateLimitLimit, 0)
				c.Set(xrateLimitRemaining, 0)
				c.Set(xrateLimitReset, 0)
				return cfg.Handler(c)
			}
			if e == nil {
				return next(c)
			}
			if cfg.ShowHeader {
				c.Set(xrateLimitLimit, cfg.MaxRequests)
				c.Set(xrateLimitRemaining, e.remaining)
				c.Set(xrateLimitReset, e.reset.Format(http.TimeFormat))
			}
			if e.remaining == 0 {
				return cfg.Handler(c)
			}
			return next(c)
		}
	}
}

func (cfg *Config) process(c *pine.Ctx) (*entry, error) {
	// generate the key. You can use the IP address of the client
	// or you can use the user id of the user
	key := cfg.KeyGen(c)

	if cfg.Whitelist != nil {
		if _, whitelist := cfg.internalWhitelist[key]; whitelist {
			return nil, nil
		}
	}

	if cfg.Blacklist != nil {
		if _, blacklisted := cfg.internalBlacklist[key]; blacklisted {
			return nil, ErrBlacklist
		}
	}

	// store is memory safe and thread safe
	ent := cfg.store.Get(key)

	// if the entry is not found in the cache, we create a new entry
	if ent == nil {
		e := &entry{
			key:       key,
			count:     1,
			reset:     time.Now().Add(cfg.Window),
			remaining: cfg.MaxRequests,
		}
		cfg.store.Set(key, e)
		return e, nil
	}
	// we convert the entry to the rate limit entry
	e := ent.(*entry)

	e.mu.Lock()
	defer e.mu.Unlock()

	// rate limit is exceeded
	if e.remaining == 0 {
		return e, nil
	}
	// reduce the remaining requests
	e.remaining--

	// update the cache with the new rate limit entry
	cfg.store.Set(key, e)
	return e, nil
}

package cache

import (
	"sync"
	"time"
)

// Cache is a simple in memory cache that stores data in a map in memory.
// The cache is not persistent, so it will be lost when the application is restarted.
//
// For the sake of speed and simplicity, try to store only necessary data in the cache
// to reduce the memory footprint and improve performance.
type Cache struct {
	mu      sync.RWMutex
	data    map[string]keyVal // the data stored in the cache
	c       time.Duration     // frequency of checking for expired data
	running bool              //condition to check if the cache is running
	cond    *sync.Cond        // condition to signal the cache to start
}

type keyVal struct {
	data interface{} // the data of the item stored in the cache
	exp  int64       // expiry date of the item which is in unix milliseconds
}

// Use this function to create a new cache
//
// You can opt out of specifying the reset time and by default it will be set to 1 second
// Reset time is the time between each check for expired data
func New(reset ...time.Duration) *Cache {
	if len(reset) == 0 {
		reset = []time.Duration{1 * time.Second}
	}
	cache := &Cache{
		data:    make(map[string]keyVal),
		c:       reset[0],
		running: false,
	}
	cache.cond = sync.NewCond(&cache.mu)
	// starts the cache instance
	go cache.start()
	return cache
}

// Sets a new item to the cache specifying the key and data to store
//
// You can opt out of specifying the time to live (ttl) and by default
// the cache will use the value specified when creating the cache using the New function
//
// This will also start the cache if there was no items in the cache before.
func (c *Cache) Set(key string, data interface{}, ttl ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(ttl) == 0 || ttl[0] == 0 {
		ttl = []time.Duration{c.c}
	}
	c.data[key] = keyVal{
		data: data,
		exp:  time.Now().Add(ttl[0]).Unix(),
	}

	if !c.running {
		c.running = true
		c.cond.Signal()
	}
}

// Gets the data from the cache using the key. If the data is not found, it returns nil
func (c *Cache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	if !ok || val.exp < time.Now().Unix() {
		return nil
	}
	return val.data
}

// checks if the data exists in the cache using the key
//
// If you call this method and immediately afterwards call the Get method within
// the window of time that the data in the cache is set to expire, sometimes exists will return
// true but Get will return nil
//
// To avoid this, you can call the Get method directly and check if the value returned is
// nil or not
func (c *Cache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.data[key]
	return ok
}

// deletes the data from the cache using the key
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// clears all the data in the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]keyVal)
}

// Starts the cache
// This is called automatically when the cache is created
func (c *Cache) start() {
	ticker := time.NewTicker(c.c)
	defer ticker.Stop()
	for {
		c.mu.Lock()
		// reduce cpu usage by checking if the cache is running. Saves cpu cycles
		for !c.running {
			c.cond.Wait()
		}
		c.mu.Unlock()

		<-ticker.C

		c.mu.Lock()
		// current time of checking the cache
		now := time.Now().Unix()
		for k, v := range c.data {
			// remove expired data
			if v.exp < now {
				delete(c.data, k)
			}
		}

		c.running = len(c.data) > 0
		c.mu.Unlock()
	}
}

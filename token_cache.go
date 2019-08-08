package quic

import (
	"container/list"
	"sync"

	"github.com/lucas-clemente/quic-go/internal/utils"
)

type singleOriginTokenCache struct {
	tokens []*ClientToken
	len    int
	p      int
}

func newSingleOriginTokenCache(size int) *singleOriginTokenCache {
	return &singleOriginTokenCache{tokens: make([]*ClientToken, size)}
}

func (c *singleOriginTokenCache) Add(token *ClientToken) {
	c.tokens[c.p] = token
	c.p = c.index(c.p + 1)
	c.len = utils.Min(c.len+1, len(c.tokens))
}

func (c *singleOriginTokenCache) Get() *ClientToken {
	c.p = c.index(c.p - 1)
	token := c.tokens[c.p]
	c.tokens[c.p] = nil
	c.len = utils.Max(c.len-1, 0)
	return token
}

func (c *singleOriginTokenCache) Len() int {
	return c.len
}

func (c *singleOriginTokenCache) index(i int) int {
	mod := len(c.tokens)
	return (i + mod) % mod
}

type lruTokenCacheEntry struct {
	key   string
	cache *singleOriginTokenCache
}

type lruTokenCache struct {
	mutex sync.Mutex

	m                map[string]*list.Element
	q                *list.List
	capacity         int
	singleOriginSize int
}

var _ TokenCache = &lruTokenCache{}

// NewLRUTokenCache creates a new LRU cache for tokens received by the client.
// maxOrigins specifies how many origins this cache is saving tokens for.
// tokensPerOrigin specifies the maximum number of tokens per origin.
func NewLRUTokenCache(maxOrigins, tokensPerOrigin int) TokenCache {
	return &lruTokenCache{
		m:                make(map[string]*list.Element),
		q:                list.New(),
		capacity:         maxOrigins,
		singleOriginSize: tokensPerOrigin,
	}
}

func (c *lruTokenCache) Put(key string, token *ClientToken) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if el, ok := c.m[key]; ok {
		entry := el.Value.(*lruTokenCacheEntry)
		entry.cache.Add(token)
		c.q.MoveToFront(el)
		return
	}

	if c.q.Len() < c.capacity {
		entry := &lruTokenCacheEntry{
			key:   key,
			cache: newSingleOriginTokenCache(c.singleOriginSize),
		}
		entry.cache.Add(token)
		c.m[key] = c.q.PushFront(entry)
		return
	}

	elem := c.q.Back()
	entry := elem.Value.(*lruTokenCacheEntry)
	delete(c.m, entry.key)
	entry.key = key
	entry.cache = newSingleOriginTokenCache(c.singleOriginSize)
	entry.cache.Add(token)
	c.q.MoveToFront(elem)
	c.m[key] = elem
}

func (c *lruTokenCache) Get(key string) (*ClientToken, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var token *ClientToken
	if el, ok := c.m[key]; ok {
		c.q.MoveToFront(el)
		cache := el.Value.(*lruTokenCacheEntry).cache
		token = cache.Get()
		if cache.Len() == 0 {
			c.q.Remove(el)
			delete(c.m, key)
		}
	}
	return token, token != nil
}

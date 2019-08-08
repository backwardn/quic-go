package quic

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Token Cache", func() {
	var c TokenCache

	BeforeEach(func() {
		c = NewLRUTokenCache(3, 4)
	})

	mockToken := func(num int) *ClientToken {
		return &ClientToken{data: []byte(fmt.Sprintf("%d", num))}
	}

	Context("for a single origin", func() {
		const origin = "localhost"

		expectToken := func(token *ClientToken) {
			t, ok := c.Get(origin)
			ExpectWithOffset(1, ok).To(BeTrue())
			ExpectWithOffset(1, t).To(Equal(token))
		}

		expectNoToken := func() {
			_, ok := c.Get(origin)
			ExpectWithOffset(1, ok).To(BeFalse())
		}

		It("adds and gets tokens", func() {
			c.Put(origin, mockToken(1))
			c.Put(origin, mockToken(2))
			expectToken(mockToken(2))
			expectToken(mockToken(1))
			expectNoToken()
		})

		It("overwrites old tokens", func() {
			c.Put(origin, mockToken(1))
			c.Put(origin, mockToken(2))
			c.Put(origin, mockToken(3))
			c.Put(origin, mockToken(4))
			c.Put(origin, mockToken(5))
			expectToken(mockToken(5))
			expectToken(mockToken(4))
			expectToken(mockToken(3))
			expectToken(mockToken(2))
			expectNoToken()
		})

		It("continues after getting a token", func() {
			c.Put(origin, mockToken(1))
			c.Put(origin, mockToken(2))
			c.Put(origin, mockToken(3))
			expectToken(mockToken(3))
			c.Put(origin, mockToken(4))
			c.Put(origin, mockToken(5))
			expectToken(mockToken(5))
			expectToken(mockToken(4))
			expectToken(mockToken(2))
			expectToken(mockToken(1))
			expectNoToken()
		})
	})

	Context("for multiple origins", func() {
		expectToken := func(origin string, token *ClientToken) {
			t, ok := c.Get(origin)
			ExpectWithOffset(1, ok).To(BeTrue())
			ExpectWithOffset(1, t).To(Equal(token))
		}

		expectNoToken := func(origin string) {
			_, ok := c.Get(origin)
			ExpectWithOffset(1, ok).To(BeFalse())
		}

		It("adds and gets tokens", func() {
			c.Put("host1", mockToken(1))
			c.Put("host2", mockToken(2))
			expectToken("host1", mockToken(1))
			expectNoToken("host1")
			expectToken("host2", mockToken(2))
			expectNoToken("host2")
		})

		It("evicts old entries", func() {
			c.Put("host1", mockToken(1))
			c.Put("host2", mockToken(2))
			c.Put("host3", mockToken(3))
			c.Put("host4", mockToken(4))
			expectNoToken("host1")
			expectToken("host2", mockToken(2))
			expectToken("host3", mockToken(3))
			expectToken("host4", mockToken(4))
		})

		It("moves old entries to the front, when new tokens are added", func() {
			c.Put("host1", mockToken(1))
			c.Put("host2", mockToken(2))
			c.Put("host3", mockToken(3))
			c.Put("host1", mockToken(11))
			// make sure one is evicted
			c.Put("host4", mockToken(4))
			expectNoToken("host2")
			expectToken("host1", mockToken(11))
			expectToken("host1", mockToken(1))
			expectToken("host3", mockToken(3))
			expectToken("host4", mockToken(4))
		})

		It("deletes hosts that are empty", func() {
			c.Put("host1", mockToken(1))
			c.Put("host2", mockToken(2))
			c.Put("host3", mockToken(3))
			expectToken("host2", mockToken(2))
			expectNoToken("host2")
			// host2 is now empty and should have been deleted, making space for host4
			c.Put("host4", mockToken(4))
			expectToken("host1", mockToken(1))
			expectToken("host3", mockToken(3))
			expectToken("host4", mockToken(4))
		})
	})
})

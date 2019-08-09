package quic

import (
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

type connectionIDManager struct {
	queue utils.ConnectionIDEntryList

	deletedPriorTo uint64
}

func (c *connectionIDManager) Add(id uint64, connID protocol.ConnectionID, token [16]byte) {
	if id < c.deletedPriorTo {
		return
	}
	for el := c.queue.Front(); el != nil; el = el.Next() {
		if el.Value.ID == id { // duplicate frame
			return
		}
		if el.Value.ID > id {
			c.queue.InsertBefore(utils.ConnectionIDEntry{
				ID:                  id,
				ConnectionID:        connID,
				StatelessResetToken: token,
			}, el)
			return
		}
	}
}

func (c *connectionIDManager) DeletePriorTo(id uint64) {
	if id < c.deletedPriorTo {
		return
	}
	var next *utils.ConnectionIDEntryElement
	for el := c.queue.Front(); el != nil; el = next {
		next = el.Next()
		if el.Value.ID >= id {
			return
		}
		c.queue.Remove(el)
	}
}

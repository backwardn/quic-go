package utils

import (
	"github.com/lucas-clemente/quic-go/internal/protocol"
)

type ConnectionIDEntry struct {
	ID                  uint64
	ConnectionID        protocol.ConnectionID
	StatelessResetToken [16]byte
}

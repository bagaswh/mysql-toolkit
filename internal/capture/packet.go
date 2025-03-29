package capture

type MySQLConnectionLifecycle byte

// Does not support ReplicationMode as of now
const (
	// PhaseUnknown is the default phase when the program
	// sees CommandPhase packets but never seen CommandPhase packets.
	//
	// Of course, being a packet capture, it is a very possible scenario.
	PhaseUnknown MySQLConnectionLifecycle = iota

	// ConnectionPhase
	//
	// See https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_lifecycle.html
	PhaseConnection

	// CommandPhase
	//
	// See https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_lifecycl
	PhaseCommand
)

// Really, I only save 1 byte vs storing it as 4-byte int.
// But I love 1 byte. I really do. And you should too.
type PayloadLength [3]byte

func (l PayloadLength) Int() int32 {
	return int32(l[0]) | int32(l[1])<<8 | int32(l[2])<<16
}

type MySQLCommonPacket struct {
	PayloadLength PayloadLength
	SequenceID    byte
}

// ---
// Server packet
// ---
type MySQLServerPacket struct{}

// ---
// Client packet
// ---

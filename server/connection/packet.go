package connection

import (
	"encoding/binary"
	"fmt"
	"log"
)

// general packet offsets
type packetOff struct {
	start int
	len   int
}

// packetOff is a int couple
// packetOff.start is the offset from the base of the raw buffer
// packetOff.len is the length of the data
// if it's == 0 then it means that the length has to be passed as a parameter to parseMember and that the type of the member is []byte

// packet types
const (
	PacketIn int = iota
	PacketOut
	PacketBroadcast
	PacketNewConnection
)

// general packet offsets
var (
	pgTypeOff packetOff = packetOff{start: 0, len: 8}
	pgLenOff  packetOff = packetOff{start: 8, len: 8}
)

// specific packet offsets
// packet in
var (
	piReceiverId packetOff = packetOff{start: 16, len: 8}
	piContents   packetOff = packetOff{start: 24, len: 0}
)

// packet out
var (
	poSenderId packetOff = packetOff{start: 16, len: 8}
	poContents packetOff = packetOff{start: 24, len: 0}
)

// packet new connection
var (
	pnContents packetOff = packetOff{start: 16, len: 0}
)

type Packet struct {
	Id         int
	SenderId   int
	ReceiverId int
	Type       int
	Len        int
	Contents   []byte

	// these members are not ment to be sent
	TargetConnectionIdx int
	TargetHandlerIdx    int
}

// if offset.len == 0 then the type of value has to be []byte (will use len(value.([]byte))) as len
func addRawMember(b []byte, offset packetOff, value any) error {
	switch offset.len {
	case 0:
		eval, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("offset.len == 0 && type(value) != []byte (value has to be a slice of bytes!)")
		}
		copy(b[offset.start:offset.start+len(eval)], eval)
	case 2:
		rval := make([]byte, 2)
		ival := value.(uint16)
		binary.LittleEndian.PutUint16(rval, ival)
		copy(b[offset.start:offset.start+2], rval)
	case 4:
		rval := make([]byte, 4)
		ival := value.(uint32)
		binary.LittleEndian.PutUint32(rval, ival)
		copy(b[offset.start:offset.start+4], rval)
	case 8:
		rval := make([]byte, 8)
		ival := value.(uint64)
		binary.LittleEndian.PutUint64(rval, ival)
		copy(b[offset.start:offset.start+8], rval)
	default:
		eval := value.([]byte)
		copy(b[offset.start:offset.start+offset.len], eval)
	}

	return nil
}

func getRawLen(ls ...int) int {
	sum := 0
	for i := 0; i < len(ls); i++ {
		sum += ls[i]
	}
	return sum
}

func (pkt *Packet) convertToRawData() ([]byte, error) {
	switch pkt.Type {
	case PacketIn:
		return nil, fmt.Errorf("cannot send PacketIn packets")
	case PacketBroadcast:
		return nil, fmt.Errorf("cannot send PacketBroadcast packets")
	case PacketOut:
		b := make([]byte, getRawLen(pgTypeOff.len, pgLenOff.len, poSenderId.len, len(pkt.Contents)))
		addRawMember(b, pgTypeOff, uint64(pkt.Type))
		addRawMember(b, pgLenOff, uint64(len(b)))
		addRawMember(b, poSenderId, uint64(pkt.SenderId))
		addRawMember(b, poContents, pkt.Contents)
		return b, nil
	case PacketNewConnection:
		b := make([]byte, getRawLen(pgTypeOff.len, pgLenOff.len, len(pkt.Contents)))
		addRawMember(b, pgTypeOff, uint64(pkt.Type))
		addRawMember(b, pgLenOff, uint64(len(b)))
		addRawMember(b, pnContents, pkt.Contents)
		return b, nil
	}
	return nil, fmt.Errorf("Packet type not recognised")
}

func parseMember(b []byte, offset packetOff, optLen int) (any, error) {
	if (offset.len == 0 && offset.start+optLen > len(b)) || offset.start+offset.len > len(b) {
		return nil, fmt.Errorf("the packet does not contain the specified member")
	}

	switch offset.len {
	case 0:
		return b[offset.start : offset.start+optLen], nil
	case 2:
		return binary.LittleEndian.Uint16(b[offset.start:]), nil
	case 4:
		return binary.LittleEndian.Uint32(b[offset.start:]), nil
	case 8:
		return binary.LittleEndian.Uint64(b[offset.start:]), nil
	default:
		return b[offset.start : offset.start+offset.len], nil
	}
}

func (pkt *Packet) parseRawData(b []byte) error {
	if val, err := parseMember(b, pgTypeOff, 0); err != nil {
		log.Println(err)
	} else {
		pkt.Type = int(val.(uint64))
	}

	if val, err := parseMember(b, pgLenOff, 0); err != nil {
		log.Println(err)
	} else {
		pkt.Len = int(val.(uint64))
	}

	// if the length indicated in the packet is not the real length of the packet
	if pkt.Len != len(b) {
		return fmt.Errorf("found a packet with packet.Len != len(rawBuffer)! (packet id: %d)", pkt.Id)
	}

	switch pkt.Type {
	case PacketIn:
		if val, err := parseMember(b, piReceiverId, 0); err != nil {
			return err
		} else {
			pkt.ReceiverId = int(val.(uint64))
		}
		if val, err := parseMember(b, piContents, len(b[piContents.start:])); err != nil {
			return err
		} else {
			pkt.Contents = val.([]byte)
		}

	case PacketBroadcast:
		log.Printf("received a broadcast packet (packet broadcasting not implemented)")
	case PacketOut:
		if val, err := parseMember(b, poSenderId, 0); err != nil {
			return err
		} else {
			pkt.SenderId = int(val.(uint64))
		}
		if val, err := parseMember(b, poContents, len(b[poContents.start:])); err != nil {
			return err
		} else {
			pkt.Contents = val.([]byte)
		}
	case PacketNewConnection:
		if val, err := parseMember(b, pnContents, len(b[pnContents.start:])); err != nil {
			return err
		} else {
			pkt.Contents = val.([]byte)
		}
	}

	return nil
}

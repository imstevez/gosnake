package base

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"unsafe"
)

const headerSize = int(unsafe.Sizeof(childHeader{}))

type childHeader struct {
	Serial uint64
	Total  uint32
	Index  uint32
}

type child struct {
	header  *childHeader
	payload []byte
}

func encodeChild(ch *child) (enc []byte) {
	var headerBuf bytes.Buffer
	err := binary.Write(&headerBuf, binary.BigEndian, ch.header)
	if err != nil {
		panic(err)
	}
	enc = make([]byte, headerSize+len(ch.payload))
	copy(enc[:headerSize], headerBuf.Bytes())
	copy(enc[headerSize:], ch.payload)
	return
}

func decodeChild(enc []byte) (ch *child) {
	_ = enc[headerSize-1]
	ch = &child{
		header: &childHeader{},
	}
	headerReader := bytes.NewReader(enc[:headerSize])
	err := binary.Read(headerReader, binary.BigEndian, ch.header)
	if err != nil {
		panic(err)
	}
	ch.payload = enc[headerSize:]
	return
}

type PacketWriter interface {
	WriteTo(p []byte, addr net.Addr) (n int, err error)
}

type PacketSplitSender struct {
	payloadSize uint32
	serials     sync.Map
	writer      PacketWriter
}

func (pss *PacketSplitSender) serialOf(addr net.Addr) (serial uint64) {
	key := addr.String()
	actual, loaded := pss.serials.LoadOrStore(key, &serial)
	if !loaded {
		return
	}
	serial = atomic.AddUint64(actual.(*uint64), 1)
	return
}

func NewPacketSplitSender(childSize int, writer PacketWriter) *PacketSplitSender {
	payloadSize := uint32(childSize - headerSize)
	if payloadSize < 1 {
		panic("child size is to small")
	}
	return &PacketSplitSender{
		payloadSize: payloadSize,
		serials:     sync.Map{},
		writer:      writer,
	}
}

func (pss *PacketSplitSender) Send(data []byte, addr net.Addr) (n int, err error) {
	header := &childHeader{}
	header.Serial = pss.serialOf(addr)
	dataSize := uint32(len(data))
	header.Total = dataSize / pss.payloadSize
	if dataSize%pss.payloadSize != 0 {
		header.Total++
	}
	ch := &child{
		header: header,
	}
	for header.Index = 0; header.Index < header.Total; header.Index++ {
		from := header.Index * pss.payloadSize
		to := from + pss.payloadSize
		if to > dataSize {
			to = dataSize
		}
		ch.payload = data[from:to]
		cn, err := pss.writer.WriteTo(
			encodeChild(ch), addr,
		)
		n += cn
		if err != nil {
			return
		}
	}
	return
}

type PacketReader interface {
	ReadFrom(p []byte) (n int, addr net.Addr, err error)
}

type receiveBuffer struct {
	serial   uint64
	total    uint32
	received uint32
	payloads [][]byte
}

type PacketSplitReceiver struct {
	childSize uint32
	buffers   map[string]*receiveBuffer
	reader    PacketReader
}

func NewPacketSplitReceiver(childSize uint32, reader PacketReader) *PacketSplitReceiver {
	return &PacketSplitReceiver{
		childSize: childSize,
		buffers:   make(map[string]*receiveBuffer),
		reader:    reader,
	}
}

func (psr *PacketSplitReceiver) Recv() (packet []byte, addr net.Addr, err error) {
	n := 0
	for {
		buf := make([]byte, psr.childSize)
		n, addr, err = psr.reader.ReadFrom(buf)
		if err != nil {
			return
		}
		ch := decodeChild(buf[:n])
		recvBuf, ok := psr.buffers[addr.String()]
		if !ok || recvBuf.serial < ch.header.Serial {
			// fist receive or skip prev uncompleted packet
			payloads := make([][]byte, ch.header.Total)
			payloads[ch.header.Index] = ch.payload
			recvBuf = &receiveBuffer{
				serial:   ch.header.Serial,
				total:    ch.header.Total,
				received: 1,
				payloads: payloads,
			}
		} else if recvBuf.serial == ch.header.Serial {
			// fill current packet
			recvBuf.received += 1
			recvBuf.payloads[ch.header.Index] = ch.payload
		} else {
			// skip prev packet child
			continue
		}
		// return completed packet
		if recvBuf.received == recvBuf.total {
			packet = bytes.Join(recvBuf.payloads, []byte{})
			return
		}
	}
}

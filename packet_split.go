package gosnake

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync/atomic"
)

const childHeaderSize = 16

type childPackageHeader struct {
	serialNumber uint64
	number       uint32
	total        uint32
}

func encodeChildPackage(header *childPackageHeader, data []byte) []byte {
	buf := make([]byte, childHeaderSize)
	binary.BigEndian.PutUint64(buf[0:8], header.serialNumber)
	binary.BigEndian.PutUint32(buf[8:12], header.number)
	binary.BigEndian.PutUint32(buf[12:16], header.total)
	data = append(buf, data...)
	return data
}

func decodeChildPackage(data []byte) (*childPackageHeader, []byte) {
	header := &childPackageHeader{}
	header.serialNumber = binary.BigEndian.Uint64(data[:8])
	header.number = binary.BigEndian.Uint32(data[8:12])
	header.total = binary.BigEndian.Uint32(data[12:16])
	return header, data[16:]
}

type SplitDataSender struct {
	childDataSize uint32
	serialNumber  uint64
}

func NewSplitDataSender(childSize uint32) *SplitDataSender {
	size := childSize - childHeaderSize
	if size <= 0 {
		panic("child package size is too small")
	}
	return &SplitDataSender{
		childDataSize: size,
	}
}

func (sds *SplitDataSender) SendDataWithUDP(data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	datal := uint32(len(data))
	childrenNum := datal / sds.childDataSize
	if datal%sds.childDataSize != 0 {
		childrenNum += 1
	}
	serialNum := atomic.AddUint64(&sds.serialNumber, 1)
	for i := uint32(0); i < childrenNum; i++ {
		s := i * sds.childDataSize
		e := s + sds.childDataSize
		if e > datal {
			e = datal
		}
		header := &childPackageHeader{
			serialNumber: serialNum,
			number:       i,
			total:        childrenNum,
		}
		childData := encodeChildPackage(header, data[s:e])
		fmt.Printf("[S] %s %d\n", addr, len(childData))
		conn.WriteToUDP(childData, addr)
	}
}

type SplitDataReciever struct {
	childSize        uint32
	buffer           [][]byte
	serialNumber     uint64
	totalChildren    uint32
	receivedChildren uint32
}

func NewSplitDataReceiver(childSize uint32, childNum uint32) *SplitDataReciever {
	return &SplitDataReciever{
		childSize: childSize,
		buffer:    make([][]byte, childNum),
	}
}

func (sdr *SplitDataReciever) ReceiveDataFromUDP(conn *net.UDPConn) []byte {
	bufferSize := uint32(len(sdr.buffer))
	for {
		// read data
		buf := make([]byte, sdr.childSize)
		n, err := conn.Read(buf)
		if n == 0 || err != nil {
			return nil
		}

		// decode and drop the invalid child
		header, data := decodeChildPackage(buf[:n])
		if header.number >= bufferSize {
			continue
		}

		// drop the outdated package
		if header.serialNumber < sdr.serialNumber {
			continue
		}

		// skip the uncompleted package
		if header.serialNumber > sdr.serialNumber {
			sdr.serialNumber = header.serialNumber
			sdr.totalChildren = header.total
			sdr.receivedChildren = 0
		}

		sdr.buffer[header.number] = data
		sdr.receivedChildren += 1
		if sdr.receivedChildren == sdr.totalChildren {
			return bytes.Join(sdr.buffer[:sdr.totalChildren], []byte{})
		}
	}
}

package gosnake

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sync/atomic"
)

var sendNum uint64

const PackagePayloadSize = 435

type Package struct {
	ID     uint64
	Number int
	Total  int
	Data   []byte
}

func encodePackage(pac *Package) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(pac)
	return buf.Bytes()
}

func decodePackage(data []byte) *Package {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	pac := &Package{}
	decoder.Decode(pac)
	return pac
}

func SendData(data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	num := len(data) / PackagePayloadSize
	if len(data)%PackagePayloadSize != 0 {
		num += 1
	}
	snum := atomic.AddUint64(&sendNum, 1)
	for i := 0; i < num; i++ {
		s := i * PackagePayloadSize
		e := s + PackagePayloadSize
		if e > len(data) {
			e = len(data)
		}
		data := encodePackage(&Package{
			ID:     snum,
			Number: i,
			Total:  num,
			Data:   data[s:e],
		})
		if len(data) == 0 {
			break
		}
		fmt.Println(snum, len(data))
		conn.WriteToUDP(data, addr)
	}
}

const maxPackageNum = 10

var (
	packagesBuf         = make([][]byte, maxPackageNum)
	currentPackageID    uint64
	currentPackageTotal int
	receivedNums        int
)

func ReceiveData(data []byte) []byte {
	pac := decodePackage(data)
	if pac == nil || pac.Number >= maxPackageNum {
		return nil
	}
	if pac.ID > currentPackageID {
		currentPackageID = pac.ID
		currentPackageTotal = pac.Total
		receivedNums = 0
	}
	packagesBuf[pac.Number] = pac.Data
	receivedNums += 1
	if receivedNums == currentPackageTotal {
		return bytes.Join(packagesBuf[:currentPackageTotal], []byte{})
	}
	return nil
}

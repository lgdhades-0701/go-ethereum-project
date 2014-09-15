package ethutil

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

type Bytes []byte

func (self Bytes) String() string {
	return string(self)
}

func DeleteFromByteSlice(s [][]byte, hash []byte) [][]byte {
	for i, h := range s {
		if bytes.Compare(h, hash) == 0 {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}

// Number to bytes
//
// Returns the number in bytes with the specified base
func NumberToBytes(num interface{}, bits int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		fmt.Println("NumberToBytes failed:", err)
	}

	return buf.Bytes()[buf.Len()-(bits/8):]
}

// Bytes to number
//
// Attempts to cast a byte slice to a unsigned integer
func BytesToNumber(b []byte) uint64 {
	var number uint64

	// Make sure the buffer is 64bits
	data := make([]byte, 8)
	data = append(data[:len(b)], b...)

	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &number)
	if err != nil {
		fmt.Println("BytesToNumber failed:", err)
	}

	return number
}

// Read variable int
//
// Read a variable length number in big endian byte order
func ReadVarInt(buff []byte) (ret uint64) {
	switch l := len(buff); {
	case l > 4:
		d := LeftPadBytes(buff, 8)
		binary.Read(bytes.NewReader(d), binary.BigEndian, &ret)
	case l > 2:
		var num uint32
		d := LeftPadBytes(buff, 4)
		binary.Read(bytes.NewReader(d), binary.BigEndian, &num)
		ret = uint64(num)
	case l > 1:
		var num uint16
		d := LeftPadBytes(buff, 2)
		binary.Read(bytes.NewReader(d), binary.BigEndian, &num)
		ret = uint64(num)
	default:
		var num uint8
		binary.Read(bytes.NewReader(buff), binary.BigEndian, &num)
		ret = uint64(num)
	}

	return
}

// Binary length
//
// Returns the true binary length of the given number
func BinaryLength(num int) int {
	if num == 0 {
		return 0
	}

	return 1 + BinaryLength(num>>8)
}

// Copy bytes
//
// Returns an exact copy of the provided bytes
func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

func IsHex(str string) bool {
	l := len(str)
	return l >= 4 && l%2 == 0 && str[0:2] == "0x"
}

func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)

	return h
}

func StringToByteFunc(str string, cb func(str string) []byte) (ret []byte) {
	if len(str) > 1 && str[0:2] == "0x" && !strings.Contains(str, "\n") {
		ret = Hex2Bytes(str[2:])
	} else {
		ret = cb(str)
	}

	return
}

func FormatData(data string) []byte {
	if len(data) == 0 {
		return nil
	}
	// Simple stupid
	d := new(big.Int)
	if data[0:1] == "\"" && data[len(data)-1:] == "\"" {
		return RightPadBytes([]byte(data[1:len(data)-1]), 32)
	} else if len(data) > 1 && data[:2] == "0x" {
		d.SetBytes(Hex2Bytes(data[2:]))
	} else {
		d.SetString(data, 0)
	}

	return BigToBytes(d, 256)
}

func ParseData(data ...interface{}) (ret []byte) {
	for _, item := range data {
		switch t := item.(type) {
		case string:
			var str []byte
			if IsHex(t) {
				str = Hex2Bytes(t[2:])
			} else {
				str = []byte(t)
			}

			ret = append(ret, RightPadBytes(str, 32)...)
		case []byte:
			ret = append(ret, LeftPadBytes(t, 32)...)
		}
	}

	return
}

func RightPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[0:len(slice)], slice)

	return padded
}

func LeftPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func LeftPadString(str string, l int) string {
	if l < len(str) {
		return str
	}

	zeros := Bytes2Hex(make([]byte, (l-len(str))/2))

	return zeros + str

}

func RightPadString(str string, l int) string {
	if l < len(str) {
		return str
	}

	zeros := Bytes2Hex(make([]byte, (l-len(str))/2))

	return str + zeros

}

func Address(slice []byte) (addr []byte) {
	if len(slice) < 20 {
		addr = LeftPadBytes(slice, 20)
	} else if len(slice) > 20 {
		addr = slice[len(slice)-20:]
	} else {
		addr = slice
	}

	addr = CopyBytes(addr)

	return
}

func ByteSliceToInterface(slice [][]byte) (ret []interface{}) {
	for _, i := range slice {
		ret = append(ret, i)
	}

	return
}

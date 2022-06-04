package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

// 将int64转化为byte数组的工具函数
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	// use binary library to convert to hex as BigEndian
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// 字符串反转
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

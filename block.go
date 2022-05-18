package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"strconv"
	"time"
)

// 仅包含公链的核心结构
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func (b *Block) SetHash() {
	// 将timestamp按照十进制的形式保存为byte数组
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	// 将其余的byte数组合成一个
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	// 计算byte数组对应的sha256 hash
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// 新建创始区块
func NewGenesisBlock() *Block {
	return NewBlock("Genesis xiaoyang block", []byte{})
}

func NewBlock(data string, PrevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), PrevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)

	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	return block
}

// serialize the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// deserialize the block
func Deserialize(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

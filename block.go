package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

// 仅包含公链的核心结构
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	// 将每笔交易取出
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	// 计算所有hash结果拼接起来的sha256
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// 新建创始区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func NewBlock(transcations []*Transaction, PrevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transcations, PrevBlockHash, []byte{}, 0}
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

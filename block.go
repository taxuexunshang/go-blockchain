package main

import (
	"bytes"
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

// 将hash的计算方法改为默克尔树
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
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

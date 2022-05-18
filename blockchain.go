package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

// 将区块链数据存储的地址抽成常量
const dbFile = "blockchain.db"
const blocksBucket = "blocks"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	// select the last block hash from dbFile
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		// 这里通过key"l" 获取区块链当前最后一块的数据
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// new a block by data and lastHash
	NewBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(NewBlock.Hash, NewBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), NewBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = NewBlock.Hash
		return nil
	})
}

// 创建区块链
func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// 对blocks这个bucket进行查询
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			// 若未查到有对应的bucket，则认定目前区块链数据库中没有数据
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

//创建迭代器 用于遍历区块链的数据
type BlockchainIntertor struct {
	currentHash []byte
	db          *bolt.DB
}

// 根据传入的区块链对象 构建区块链的迭代器
func (bc *Blockchain) Iterator() *BlockchainIntertor {
	bci := &BlockchainIntertor{bc.tip, bc.db}

	return bci
}

// 通过当前区块的数据 实现区块链的反向遍历
func (i *BlockchainIntertor) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = Deserialize(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// 将区块链当前区块的前一个区块的hash传入
	i.currentHash = block.PrevBlockHash

	return block
}

package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

// 将区块链数据存储的地址抽成常量
const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Xiao Yang Coin will be issued on May 28, 2022"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// 将判断区块链数据库是否存在的逻辑抽离
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewBlockChain() *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		// 遍历整条区块链
		block := bci.Next()

		// 遍历区块链的每一条交易
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			// 由于整个是从后向前遍历的 所以扫描的顺序是先判断输出 再判断输入
			// 并且根据UTXO的特性，一个transaction中不会出现指向同一个地址的多个输出,故能够通过以下方式找到
		outputs:
			// 遍历交易的输出,检查输出有没有作为另一笔交易的输入即可
			// 遍历区块的输出
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					// 该交易已存在输入集合中 则证明其以用于交易
					for _, spentOut := range spentTXOs[txID] {
						// 匹配到一个使用过的输出则跳过
						if spentOut == outIdx {
							continue outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				// 若匹配到一个未使用过的输出 则记录下当前交易 证明当前交易中存在未使用的输出
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				// 将所有交易的输入TranscationID,vout放入
				for _, in := range tx.Vin {
					// 遍历所有input
					inTxID := hex.EncodeToString(in.Txid)
					// 将当前用户input的Vout存储起来
					// 存储input的Transcation ID及Vout
					// 存放在spentTXOs集合中的输出是一定花费过的
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)

				}
			}
		}
		// 区块链到头 则终止循环
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

// 创建区块链,主要负责创世块的挖矿,首次铸币交易,区块链的持久化
func CreateBlockChain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// 创建创世块 创世块为铸币交易
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		// 首先创建bucket
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		// 将标记最后一个区块链的标记放入桶中
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

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

// 实现交易区块的挖矿
func (bc *Blockchain) MineBlock(transcations []*Transaction) *Block {
	var lastHash []byte

	for _, tx := range transcations {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	// 查找当前区块链中最后一个块的hash
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transcations, lastHash)

	// TODO: 不知道为什么 这个东西会报错 待解决bug
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		// 将区块链当前指向的最后一块更新
		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// 查找包含未使用输出的交易
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspetTXs []Transaction
	// 存储的是交易ID-
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		// 遍历整条区块链
		block := bci.Next()

		// 遍历区块链的每一条交易
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			// 由于整个是从后向前遍历的 所以扫描的顺序是先判断输出 再判断输入
			// 并且根据UTXO的特性，一个transaction中不会出现指向同一个地址的多个输出,故能够通过以下方式找到
		outputs:
			// 遍历交易的输出,检查输出有没有作为另一笔交易的输入即可
			// 遍历区块的输出
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					// 该交易已存在输入集合中 则证明其以用于交易
					for _, spentOut := range spentTXOs[txID] {
						// 匹配到一个使用过的输出则跳过
						if spentOut == outIdx {
							continue outputs
						}
					}
				}

				// 若匹配到一个未使用过的输出 则记录下当前交易 证明当前交易中存在未使用的输出
				if out.IsLockedWithKey(pubKeyHash) {
					unspetTXs = append(unspetTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				// 将所有交易的输入TranscationID,vout放入
				for _, in := range tx.Vin {
					// 遍历所有input
					if in.UseKey(pubKeyHash) {
						// 将Txid编码
						inTxID := hex.EncodeToString(in.Txid)
						// 将当前用户input的Vout存储起来
						// 存储input的Transcation ID及Vout
						// 存放在spentTXOs集合中的输出是一定花费过的
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		// 区块链到头 则终止循环
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspetTXs
}

// 先定义验证当前交易是否合法的函数
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	// 得到包含未使用Outputs的交易
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

// 找到对应ID的交易
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// 对交易进行签名的方法
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// 对交易进行验证
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

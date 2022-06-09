package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

// 实现UTXO缓存
type UTXOset struct {
	Blockchain *Blockchain
}

// 统计所有UTXO的总数并返回
func (u UTXOset) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return counter
}

// UTXO集合的初始化方法 若其存在则先将其删除
func (u UTXOset) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		// 将所有未花费的UTXO都放入set中
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

// 找到UTXO中未花费的输出,统计金额总数，并且返回ID及output中对应的索引集合
func (u UTXOset) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0

	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		// 这里的Cursor可以理解为索引
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIndex, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIndex)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// 查找所有未花费的UTXO
func (u UTXOset) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

// 用于生成区块后UTXO集的更新
func (u UTXOset) Update(block *Block) {
	db := u.Blockchain.db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				// 处理输入
				// 从bucket中取出所有本次交易输入对应的输出
				for _, vin := range tx.Vin {
					updateOuts := TXOutputs{}
					// 从set中取出所有输入对应的UTXO
					outsBytes := b.Get(vin.Txid)
					// 反序列化
					outs := DeserializeOutputs(outsBytes)

					// 遍历交易中的output
					for outIndex, out := range outs.Outputs {
						// 取出当前交易中没有使用的UTXO
						if outIndex != vin.Vout {
							// 如果当前的输入
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}

					// 如果恰好使用完了 直接从UTXO中移除这个输出对应的所有即可
					if len(updateOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						// 否则更新为仅包含当前未使用的UTXO
						err := b.Put(vin.Txid, updateOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}

				// 处理输出
				// 将新的输出放入UTXO集即可
				newOutputs := TXOutputs{}
				for _, out := range tx.Vout {
					newOutputs.Outputs = append(newOutputs.Outputs, out)
				}
				err := b.Put(tx.ID, newOutputs.Serialize())
				if err != nil {
					log.Panic(err)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

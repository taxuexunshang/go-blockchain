package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// 设置补贴（出块奖励）
const subsidy = 10

// 按照bitcoin论文中的模型定义一个transcation
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// 判断当前交易是否为铸币交易
func (tx Transaction) IsCoinbase() bool {
	// 在此区块链原型中 将铸币交易的Txid设置为空，并且将Vout设置为-1 并且只有一个输入
	// 通过以上特征判断交易是否为铸币交易
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// SetID 方法将transcation序列化后的hash作为当前交易的ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	// 通过gob的Encoder能够直接对数据结构进行序列化
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// 创建一个铸币交易 在公链区块链中 铸币交易是不可取代的一种交易
func NewCoinbaseTX(to, data string) *Transaction {
	// 如果没有指定铸币交易的data
	// 则默认将铸币交易的data设置为奖励 to
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// 检查地址是否发生了交易
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// 检查当前的输出是否能够被该用户解锁
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

type TXInput struct {
	// 存储其来源的Txid
	Txid []byte
	// 来源Txid中对应的索引
	Vout      int
	ScriptSig string
}

func NewUTXOTransction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// 验证输入的币是否足够支付输出
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// 构造输入的list
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// 构造输出的list
	outputs = append(outputs, TXOutput{amount, to})
	// 当支付的UTXO 大于其需要使用的UTXO时
	if acc > amount {
		// 增加一个找零输出
		outputs = append(outputs, TXOutput{acc - amount, from}) // a change
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx

}

package main

import "bytes"

type TXInput struct {
	// 存储其来源的Txid
	Txid []byte
	// 来源Txid中对应的索引
	Vout int
	// 签名
	Signature []byte
	// 公钥
	PubKey []byte
}

// 检验提供的公钥hash是否用当前交易的公钥生成
func (in *TXInput) UseKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

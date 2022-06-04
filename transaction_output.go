package main

import "bytes"

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// 使用公钥hash对output签名
func (out *TXOutput) Lock(address []byte) {
	// 地址解码后中间部分即为公钥hash
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	// 使用公钥hash锁定输出
	out.PubKeyHash = pubKeyHash
}

// 检查当前的TXOutput是否是由当前的公钥锁定的
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// 新建一个Output transcation
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	// 通过lock方法填充公钥hash
	txo.Lock([]byte(address))

	return txo
}

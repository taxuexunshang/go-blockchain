package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const targetBits = 24

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}
	return pow
}

// convert all blockchain data to byte array
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.Data,
		IntToHex(pow.block.Timestamp),
		IntToHex(int64(targetBits)),
		IntToHex(int64(nonce)),
	}, []byte{})

	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	MaxNonce := math.MaxInt64
	nonce := 0
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	for nonce < MaxNonce {
		// 拼接nonce对应的byte数据
		data := pow.prepareData(nonce)
		// 计算其sha256
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		// 将此对应的sha256转化为大整数
		hashInt.SetBytes(hash[:])
		// 只要比pow.target小则是一个合法的hash
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")
	return nonce, hash[:]
}

// 验证当前的工作量证明是否有效 应用于挖矿时的验证 而非矿工的验证
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	// 目前的工作量证明的验证是求块中所有数据的hash再判断其是否符合目标
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

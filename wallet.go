package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

// 一个钱包存储一对公私钥
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionPayload := append([]byte{version}, pubKeyHash...)
	checkSum := checkSum(versionPayload)

	fullPayload := append(versionPayload, checkSum...)
	address := Base58Encode(fullPayload)

	return address
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160

}

// 获取指定内容的校验码
func checkSum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func ValidateAddress(address string) bool {
	// 分离出来原始的校验码 之后通过已有信息重新计算校验码
	pubKeyHash := Base58Encode([]byte(address))
	// 分离出校验码
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	// 分离出版本号
	version := pubKeyHash[0]
	// 分离原始公钥hash
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	// 重新计算校验码
	targetChecksum := checkSum(append([]byte{version}, pubKeyHash...))
	// 验证校验码是否合法
	return bytes.Equal(actualChecksum, targetChecksum)
}

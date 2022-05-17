package main

import "fmt"

func main() {
	// 这个地方会存在一个坑 在通过命令运行的时候要将所有涉及调用的文件都加上 或者编译之后再运行
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to XiaoLi")
	bc.AddBlock("Send 2 more BTC to XiaoYang")

	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}

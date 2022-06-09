package main

import "crypto/sha256"

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

type MerkleTree struct {
	RootNode *MerkleNode
}

// 创建单棵merkleTree的逻辑
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// 如果当前的data不是2的倍数 则复制最后一个节点 使其为2的倍数
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	// 构造所有Merkle单独的节点
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}
		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

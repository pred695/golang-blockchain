package Blockchain

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left *MerkleNode, right *MerkleNode, data []byte) *MerkleNode {
	var mNode MerkleNode = MerkleNode{}

	if left == nil && right == nil {
		var hash [32]byte = sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		var prevHashes []byte = append(left.Data, right.Data...)
		var hash [32]byte = sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data) % 2 != 0{
		data = append(data, data[len(data)-1])	//concatenate the duplicate of last element of the slice to the end of the slice so as to make the length of the elements even
	}

	for _, info := range data {
		var node MerkleNode = MerkleNode{Data: info, Left: nil, Right: nil}
		nodes = append(nodes, node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j]/*Left*/, &nodes[j+1]/*Right*/, nil/*data*/)	//initialising the parent node
			newLevel = append(newLevel, *node)	//appending the parent node to the new level
		}

		nodes = newLevel	//updating the nodes slice with the new level(parent nodes right now are child nodes to the next level parent nodes)
	}

	var mTree MerkleTree = MerkleTree{RootNode: &nodes[0]}

	return &mTree
}

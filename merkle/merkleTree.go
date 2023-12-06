package merkle

import (
	"context"
	"fmt"
	"math/big"

	merkletree "github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/memory"
)

// Sparse MT
func main() {

	ctx := context.Background()

	// Tree storage
	store := memory.NewMemoryStorage()

	// Generate a new MerkleTree with 32 levels
	mt, _ := merkletree.NewMerkleTree(ctx, store, 32)

	// Add a leaf to the tree with index 1 and value 10
	index1 := big.NewInt(1)
	value1 := big.NewInt(10)
	mt.Add(ctx, index1, value1)

	// Add another leaf to the tree
	index2 := big.NewInt(2)
	value2 := big.NewInt(15)
	mt.Add(ctx, index2, value2)

	// Proof of membership of a leaf with index 1
	proofExist, value, _ := mt.GenerateProof(ctx, index1, mt.Root())

	fmt.Println("Proof of membership:", proofExist.Existence)
	fmt.Println("Value corresponding to the queried index:", value)

	// Proof of non-membership of a leaf with index 4
	proofNotExist, _, _ := mt.GenerateProof(ctx, big.NewInt(4), mt.Root())

	fmt.Println("Proof of membership:", proofNotExist.Existence)
}

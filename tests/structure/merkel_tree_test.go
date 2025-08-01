package test

import (
	"bytes"
	"crypto/sha256"
	. "fortuna/structure"
	"math/big"
	"testing"
)

type testContent string

func (t testContent) CalculateHash() ([]byte, error) {
	sum := sha256.Sum256([]byte(t))
	return sum[:], nil
}
func (t testContent) Equals(other Content) (bool, error) {
	o, ok := other.(testContent)
	if !ok {
		return false, nil
	}
	return t == o, nil
}

func sortAppendTest(sort bool, a, b []byte) []byte {
	if !sort {
		return append(a, b...)
	}
	var aBig, bBig big.Int
	aBig.SetBytes(a)
	bBig.SetBytes(b)
	if aBig.Cmp(&bBig) == -1 {
		return append(a, b...)
	}
	return append(b, a...)
}

func computeRootFromPath(leafHash []byte, siblings [][]byte, indexes []int64, sort bool) []byte {
	cur := leafHash
	for i, sib := range siblings {
		var combined []byte
		if sort {
			combined = sortAppendTest(true, cur, sib)
		} else {
			if indexes[i] == 1 {
				combined = append(cur, sib...)
			} else {
				combined = append(sib, cur...)
			}
		}
		h := sha256.Sum256(combined)
		cur = h[:]
	}
	return cur
}

func TestNewTreeAndVerify(t *testing.T) {
	cs := []Content{
		testContent("a"),
		testContent("b"),
		testContent("c"),
		testContent("d"),
	}
	mt, err := NewTree(cs)
	if err != nil {
		t.Fatalf("NewTree error: %v", err)
	}

	ok, err := mt.VerifyTree()
	if err != nil {
		t.Fatalf("VerifyTree error: %v", err)
	}
	if !ok {
		t.Fatalf("VerifyTree = false, want true")
	}

	if len(mt.MerkleRoot()) == 0 {
		t.Fatalf("MerkleRoot empty")
	}

	mt2, err := NewTree(cs)
	if err != nil {
		t.Fatalf("NewTree error: %v", err)
	}
	if !bytes.Equal(mt.MerkleRoot(), mt2.MerkleRoot()) {
		t.Fatalf("roots differ for identical content")
	}
}

func TestOddLeavesDuplication(t *testing.T) {
	cs := []Content{
		testContent("x"),
		testContent("y"),
		testContent("z"),
	}
	mt, err := NewTree(cs)
	if err != nil {
		t.Fatalf("NewTree error: %v", err)
	}
	if len(mt.Leafs)%2 != 0 {
		t.Fatalf("expected even number of leafs due to duplication; got %d", len(mt.Leafs))
	}

	ok, err := mt.VerifyTree()
	if err != nil {
		t.Fatalf("VerifyTree error: %v", err)
	}
	if !ok {
		t.Fatalf("VerifyTree = false, want true")
	}
}

func TestVerifyContent(t *testing.T) {
	cs := []Content{
		testContent("a"),
		testContent("b"),
		testContent("c"),
	}
	mt, err := NewTree(cs)
	if err != nil {
		t.Fatalf("NewTree error: %v", err)
	}

	ok, err := mt.VerifyContent(testContent("b"))
	if err != nil {
		t.Fatalf("VerifyContent error: %v", err)
	}
	if !ok {
		t.Fatalf("VerifyContent(b) = false, want true")
	}

	ok, err = mt.VerifyContent(testContent("not-exist"))
	if err != nil {
		t.Fatalf("VerifyContent error: %v", err)
	}
	if ok {
		t.Fatalf("VerifyContent(non-exist) = true, want false")
	}
}

func TestGetMerklePath_Unsorted(t *testing.T) {
	cs := []Content{
		testContent("a"),
		testContent("b"),
		testContent("c"),
		testContent("d"),
		testContent("e"),
	}
	mt, err := NewTree(cs)
	if err != nil {
		t.Fatalf("NewTree error: %v", err)
	}

	target := testContent("c")
	leafHash, _ := target.CalculateHash()

	path, indexes, err := mt.GetMerklePath(target)
	if err != nil {
		t.Fatalf("GetMerklePath error: %v", err)
	}
	if len(path) == 0 {
		t.Fatalf("empty merkle path")
	}
	root := computeRootFromPath(leafHash, path, indexes, false)
	if !bytes.Equal(root, mt.MerkleRoot()) {
		t.Fatalf("reconstructed root mismatch (unsorted)")
	}
}

func TestGetMerklePath_Sorted(t *testing.T) {
	cs := []Content{
		testContent("k1"),
		testContent("k2"),
		testContent("k3"),
		testContent("k4"),
	}

	mt, err := NewTreeWithHashStrategySorted(cs, sha256.New, true)
	if err != nil {
		t.Fatalf("NewTreeWithHashStrategySorted error: %v", err)
	}

	target := testContent("k3")
	leafHash, _ := target.CalculateHash()

	path, indexes, err := mt.GetMerklePath(target)
	if err != nil {
		t.Fatalf("GetMerklePath error: %v", err)
	}
	if len(path) == 0 {
		t.Fatalf("empty merkle path")
	}
	_ = indexes

	root := computeRootFromPath(leafHash, path, indexes, true)
	if !bytes.Equal(root, mt.MerkleRoot()) {
		t.Fatalf("reconstructed root mismatch (sorted)")
	}
}

func TestRebuildTree(t *testing.T) {
	cs := []Content{
		testContent("a"),
		testContent("b"),
		testContent("c"),
		testContent("d"),
	}
	mt, err := NewTree(cs)
	if err != nil {
		t.Fatalf("NewTree error: %v", err)
	}
	root1 := append([]byte(nil), mt.MerkleRoot()...)

	if err := mt.RebuildTree(); err != nil {
		t.Fatalf("RebuildTree error: %v", err)
	}
	root2 := mt.MerkleRoot()
	if !bytes.Equal(root1, root2) {
		t.Fatalf("root changed after RebuildTree with same content")
	}

	cs2 := []Content{
		testContent("a"),
		testContent("b"),
		testContent("c"),
		testContent("X"),
	}
	if err := mt.RebuildTreeWith(cs2); err != nil {
		t.Fatalf("RebuildTreeWith error: %v", err)
	}
	if bytes.Equal(root1, mt.MerkleRoot()) {
		t.Fatalf("root did not change after RebuildTreeWith different content")
	}

	ok, err := mt.VerifyTree()
	if err != nil {
		t.Fatalf("VerifyTree error: %v", err)
	}
	if !ok {
		t.Fatalf("VerifyTree(false) after rebuild with new content")
	}
}

func TestNewTree_EmptyError(t *testing.T) {
	_, err := NewTree(nil)
	if err == nil {
		t.Fatalf("expected error for empty content, got nil")
	}
}

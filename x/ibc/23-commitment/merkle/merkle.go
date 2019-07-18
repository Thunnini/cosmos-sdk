package merkle

import (
	"errors"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const merkleKind = "merkle"

// merkle.Proof implementation of Proof
// Applied on SDK-based IBC implementation
var _ commitment.Root = Root{}

type Root struct {
	Hash []byte
}

func NewRoot(hash []byte) Root {
	return Root{
		Hash: hash,
	}
}

func (Root) CommitmentKind() string {
	return merkleKind
}

var _ commitment.Path = Path{}

type Path struct {
	KeyPath   [][]byte
	KeyPrefix []byte
}

func NewPath(keypath [][]byte, keyprefix []byte) Path {
	return Path{
		KeyPath:   keypath,
		KeyPrefix: keyprefix,
	}
}

func (Path) CommitmentKind() string {
	return merkleKind
}

func (path Path) Pathify(key []byte) (res []byte) {
	res = make([]byte, len(path.KeyPrefix)+len(key))
	copy(res, path.KeyPrefix)
	copy(res[len(path.KeyPrefix):], key)
	return
}

var _ commitment.Proof = Proof{}

type Proof struct {
	Proof *merkle.Proof
	Key   []byte
}

func (Proof) CommitmentKind() string {
	return merkleKind
}

func (proof Proof) GetKey() []byte {
	return proof.Key
}

func (proof Proof) Verify(croot commitment.Root, cpath commitment.Path, value []byte) error {
	root, ok := croot.(Root)
	if !ok {
		return errors.New("invalid commitment root type")
	}

	path, ok := cpath.(Path)
	if !ok {
		return errors.New("invalid commitment path type")
	}

	keypath := merkle.KeyPath{}
	for _, key := range path.KeyPath {
		keypath = keypath.AppendKey(key, merkle.KeyEncodingHex)
	}
	// KeyPrefix is not appended, we assume that the proof.Key already contains it
	keypath = keypath.AppendKey(proof.Key, merkle.KeyEncodingHex)

	// Hard coded for now
	runtime := rootmulti.DefaultProofRuntime()

	if value != nil {
		return runtime.VerifyValue(proof.Proof, root.Hash, keypath.String(), value)
	}
	return runtime.VerifyAbsence(proof.Proof, root.Hash, keypath.String())
}

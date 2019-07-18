package tendermint

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

const chainid = "testchain"

func defaultComponents(storename string) (sdk.StoreKey, sdk.Context, stypes.CommitMultiStore, *codec.Codec) {
	key := sdk.NewKVStoreKey(storename)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := codec.New()
	return key, ctx, cms, cdc
}

type Node struct {
	PrevValset MockValidators
	Valset     MockValidators

	Cms   sdk.CommitMultiStore
	Key   sdk.StoreKey
	Store sdk.KVStore

	Commits []tmtypes.SignedHeader

	Path merkle.Path
}

func NewNode(valset MockValidators, path merkle.Path) *Node {
	key, ctx, cms, _ := defaultComponents(string(path.KeyPath[0]))

	return &Node{
		Valset:  valset,
		Cms:     cms,
		Key:     key,
		Store:   ctx.KVStore(key),
		Commits: nil,
		Path:    path,
	}
}

func (node *Node) Last() tmtypes.SignedHeader {
	if len(node.Commits) == 0 {
		return tmtypes.SignedHeader{}
	}
	return node.Commits[len(node.Commits)-1]
}

func (node *Node) Commit() tendermint.Header {
	valsethash := node.Valset.ValidatorSet().Hash()
	nextvalset := node.Valset.Mutate()
	nextvalsethash := nextvalset.ValidatorSet().Hash()
	commitid := node.Cms.Commit()

	header := tmtypes.Header{
		ChainID: chainid,
		Height:  int64(len(node.Commits) + 1),
		LastBlockID: tmtypes.BlockID{
			Hash: node.Last().Header.Hash(),
		},

		ValidatorsHash:     valsethash,
		NextValidatorsHash: nextvalsethash,
		AppHash:            commitid.Hash,
	}

	commit := node.Valset.Sign(header)

	node.PrevValset = node.Valset
	node.Valset = nextvalset
	node.Commits = append(node.Commits, commit)

	return tendermint.Header{
		SignedHeader:     commit,
		ValidatorSet:     node.PrevValset.ValidatorSet(),
		NextValidatorSet: node.Valset.ValidatorSet(),
	}
}

func (node *Node) LastStateVerifier() *Verifier {
	return NewVerifier(node.Last(), node.Valset, node.Root())
}

func (node *Node) Root() merkle.Root {
	return merkle.NewRoot(node.Last().AppHash)

}

func (node *Node) Context() sdk.Context {
	return sdk.NewContext(node.Cms, abci.Header{}, false, log.NewNopLogger())
}

type Verifier struct {
	client.ConsensusState
}

func NewVerifier(header tmtypes.SignedHeader, nextvalset MockValidators, root merkle.Root) *Verifier {
	return &Verifier{
		tendermint.ConsensusState{
			ChainID:          chainid,
			Height:           uint64(header.Height),
			Root:             merkle.NewRoot(header.AppHash),
			NextValidatorSet: nextvalset.ValidatorSet(),
		},
	}
}

func (v *Verifier) Validate(header tendermint.Header, valset, nextvalset MockValidators) error {
	newcs, err := v.ConsensusState.Validate(header)
	if err != nil {
		return err
	}
	v.ConsensusState = newcs.(tendermint.ConsensusState)

	return nil
}

func (node *Node) Query(t *testing.T, k []byte) ([]byte, commitment.Proof) {
	resp := node.Cms.(stypes.Queryable).Query(abci.RequestQuery{
		Path:  "/" + string(node.Path.KeyPath[0]) + "/key",
		Data:  k,
		Prove: true,
	})
	require.Equal(t, uint32(0), resp.Code)
	proof := merkle.Proof{
		Key:   k,
		Proof: resp.Proof,
	}
	return resp.Value, proof
}

func (node *Node) Set(k, value []byte) {
	node.Store.Set(node.Path.Key(k), value)
}

// nolint:deadcode,unused
func testProof(t *testing.T) {
	node := NewNode(NewMockValidators(100, 10), merkle.NewPath([][]byte{[]byte("1")}, []byte{0x00, 0x01}))

	node.Commit()

	kvps := cmn.KVPairs{}
	for h := 0; h < 20; h++ {
		for i := 0; i < 100; i++ {
			k := make([]byte, 32)
			v := make([]byte, 32)
			_, err := rand.Read(k)
			require.NoError(t, err)
			_, err = rand.Read(v)
			require.NoError(t, err)
			kvps = append(kvps, cmn.KVPair{Key: k, Value: v})
			node.Set(k, v)
		}

		header := node.Commit()
		proofs := []commitment.Proof{}
		root := merkle.NewRoot(header.AppHash)
		for _, kvp := range kvps {
			v, p := node.Query(t, node.Path.Key(kvp.Key))

			require.Equal(t, kvp.Value, v)
			proofs = append(proofs, p)
		}
		cstore, err := commitment.NewStore(root, node.Path, proofs)
		require.NoError(t, err)

		for _, kvp := range kvps {
			require.True(t, cstore.Prove(kvp.Key, kvp.Value))
		}
	}
}

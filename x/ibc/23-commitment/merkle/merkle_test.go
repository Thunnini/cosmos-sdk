package merkle

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func defaultComponents() (sdk.StoreKey, sdk.Context, types.CommitMultiStore, *codec.Codec) {
	key := sdk.NewKVStoreKey("test")
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

func commit(cms types.CommitMultiStore) Root {
	cid := cms.Commit()
	return NewRoot(cid.Hash)
}

func queryMultiStore(t *testing.T, path Path, cms types.CommitMultiStore, key, value []byte) Proof {
	code, val, proof := path.QueryMultiStore(cms, key)
	require.Equal(t, uint32(0), code)
	require.Equal(t, value, val)
	return proof
}

func TestStore(t *testing.T) {
	k, ctx, cms, _ := defaultComponents()
	kvstore := ctx.KVStore(k)
	path := Path{KeyPath: [][]byte{[]byte("test")}, KeyPrefix: []byte{0x01, 0x03, 0x05}}

	kvstore.Set(path.Key([]byte("hello")), []byte("world"))
	kvstore.Set(path.Key([]byte("merkle")), []byte("tree"))
	kvstore.Set(path.Key([]byte("block")), []byte("chain"))

	root := commit(cms)

	c1, v1, p1 := path.QueryMultiStore(cms, []byte("hello"))
	require.Equal(t, uint32(0), c1)
	require.Equal(t, []byte("world"), v1)
	c2, v2, p2 := path.QueryMultiStore(cms, []byte("merkle"))
	require.Equal(t, uint32(0), c2)
	require.Equal(t, []byte("tree"), v2)
	c3, v3, p3 := path.QueryMultiStore(cms, []byte("block"))
	require.Equal(t, uint32(0), c3)
	require.Equal(t, []byte("chain"), v3)

	cstore, err := commitment.NewStore(root, path, []commitment.Proof{p1, p2, p3})
	require.NoError(t, err)

	require.True(t, cstore.Prove([]byte("hello"), []byte("world")))
	require.True(t, cstore.Prove([]byte("merkle"), []byte("tree")))
	require.True(t, cstore.Prove([]byte("block"), []byte("chain")))

	kvstore.Set(path.Key([]byte("12345")), []byte("67890"))
	kvstore.Set(path.Key([]byte("qwerty")), []byte("zxcv"))
	kvstore.Set(path.Key([]byte("hello")), []byte("dlrow"))

	root = commit(cms)

	c1, v1, p1 = path.QueryMultiStore(cms, []byte("12345"))
	require.Equal(t, uint32(0), c1)
	require.Equal(t, []byte("67890"), v1)
	c2, v2, p2 = path.QueryMultiStore(cms, []byte("qwerty"))
	require.Equal(t, uint32(0), c2)
	require.Equal(t, []byte("zxcv"), v2)
	c3, v3, p3 = path.QueryMultiStore(cms, []byte("hello"))
	require.Equal(t, uint32(0), c3)
	require.Equal(t, []byte("dlrow"), v3)
	c4, v4, p4 := path.QueryMultiStore(cms, []byte("merkle"))
	require.Equal(t, uint32(0), c4)
	require.Equal(t, []byte("tree"), v4)

	cstore, err = commitment.NewStore(root, path, []commitment.Proof{p1, p2, p3, p4})
	require.NoError(t, err)

	require.True(t, cstore.Prove([]byte("12345"), []byte("67890")))
	require.True(t, cstore.Prove([]byte("qwerty"), []byte("zxcv")))
	require.True(t, cstore.Prove([]byte("hello"), []byte("dlrow")))
	require.True(t, cstore.Prove([]byte("merkle"), []byte("tree")))
}

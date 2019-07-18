package connection

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// CLIObject stores the key for each object fields
type CLIObject struct {
	ID            string
	ConnectionKey []byte
	AvailableKey  []byte
	KindKey       []byte

	Client client.CLIObject

	StoreName string
	Cdc       *codec.Codec
}

func (man Manager) CLIObject(path merkle.Path, id string, clientid string) CLIObject {
	obj := man.object(id)
	return CLIObject{
		ID:            obj.id,
		ConnectionKey: obj.connection.Key(),
		AvailableKey:  obj.available.Key(),
		KindKey:       obj.kind.Key(),

		Client: man.client.CLIObject(path, clientid),

		StoreName: man.protocol.StoreName(),
		Cdc:       obj.connection.Cdc(),
	}
}

func (obj CLIObject) query(ctx context.CLIContext, key []byte, ptr interface{}) (merkle.Proof, error) {
	/*resp, err := ctx.QueryABCI(abci.RequestQuery{
		Path: "/store/" + obj.StoreName + "/key",
		Data: key,
	})
	if err != nil {
		return merkle.Proof{}, err
	}
	proof := merkle.Proof{
		Key:   key,
		Proof: resp.Proof,
	}
	err = obj.Cdc.UnmarshalBinaryBare(resp.Value, ptr)*/
	return merkle.Proof{}, nil

}

func (obj CLIObject) Connection(ctx context.CLIContext) (res Connection, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.ConnectionKey, &res)
	return
}

func (obj CLIObject) Available(ctx context.CLIContext) (res bool, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.AvailableKey, &res)
	return
}

func (obj CLIObject) Kind(ctx context.CLIContext) (res string, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.KindKey, &res)
	return
}

type CLIHandshakeObject struct {
	CLIObject

	StateKey              []byte
	CounterpartyClientKey []byte
	TimeoutKey            []byte
}

func (man Handshaker) CLIObject(path merkle.Path, id string, clientid string) CLIHandshakeObject {
	obj := man.object(man.man.object(id))
	return CLIHandshakeObject{
		CLIObject: man.man.CLIObject(path, id, clientid),

		StateKey:              obj.state.Key(),
		CounterpartyClientKey: obj.counterpartyClient.Key(),
		TimeoutKey:            obj.nextTimeout.Key(),
	}
}

func (obj CLIHandshakeObject) State(ctx context.CLIContext, root merkle.Root) (res byte, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.StateKey, &res)
	return
}

func (obj CLIHandshakeObject) CounterpartyClient(ctx context.CLIContext, root merkle.Root) (res string, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.CounterpartyClientKey, &res)
	return
}

func (obj CLIHandshakeObject) Timeout(ctx context.CLIContext, root merkle.Root) (res uint64, proof merkle.Proof, err error) {
	proof, err = obj.query(ctx, obj.TimeoutKey, &res)
	return
}

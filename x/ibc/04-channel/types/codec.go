package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

// SubModuleCdc defines the IBC channel codec.
var SubModuleCdc = codec.New()

// RegisterCodec registers all the necessary types and interfaces for the
// IBC channel.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.PacketI)(nil), nil)
	cdc.RegisterConcrete(Packet{}, "ibc/channel/Packet", nil)
	cdc.RegisterConcrete(OpaquePacket{}, "ibc/channel/OpaquePacket", nil)

	cdc.RegisterConcrete(MsgChannelOpenInit{}, "ibc/channel/MsgChannelOpenInit", nil)
	cdc.RegisterConcrete(MsgChannelOpenTry{}, "ibc/channel/MsgChannelOpenTry", nil)
	cdc.RegisterConcrete(MsgChannelOpenAck{}, "ibc/channel/MsgChannelOpenAck", nil)
	cdc.RegisterConcrete(MsgChannelOpenConfirm{}, "ibc/channel/MsgChannelOpenConfirm", nil)
	cdc.RegisterConcrete(MsgChannelCloseInit{}, "ibc/channel/MsgChannelCloseInit", nil)
	cdc.RegisterConcrete(MsgChannelCloseConfirm{}, "ibc/channel/MsgChannelCloseConfirm", nil)
}

func init() {
	RegisterCodec(SubModuleCdc)
}
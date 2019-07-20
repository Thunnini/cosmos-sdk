package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	swap "github.com/cosmos/cosmos-sdk/x/uniswap"

	"github.com/spf13/cobra"
)

func GetSwapTxCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:  "swap [asset] [target_denom]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// parse coins
			asset, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()

			msg := swap.NewMsgSwap(from, asset, args[1])
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

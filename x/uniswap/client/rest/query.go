package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	swap "github.com/cosmos/cosmos-sdk/x/uniswap"

	"github.com/gorilla/mux"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	registerQueryRoutes(cliCtx, r, cdc)
}

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc(
		"/swap/estimate/{sender}/{asset}/{targetDenom}",
		swapEstimateHandlerFn(cliCtx, cdc),
	).Methods("GET")
}

func swapEstimateHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assetStr := vars["asset"]
		asset, err := sdk.ParseCoin(assetStr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		senderStr := vars["sender"]
		sender, err := sdk.AccAddressFromBech32(senderStr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := swap.NewMsgSwap(sender, asset, vars["targetDenom"])
		stdTx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{
			auth.StdSignature{},
		}, "")
		bz, err := cdc.MarshalBinaryLengthPrefixed(stdTx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		simulated, err := cliCtx.Query("app/simulate", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		result := sdk.Result{}
		err = cdc.UnmarshalBinaryLengthPrefixed(simulated, &result)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var res struct {
			Coin string `json:"coin"`
		}

		if result.Code != 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, result.Log)
			return
		}

		for _, tag := range result.Tags {
			if string(tag.Key) == "swap" {
				coin, err := sdk.ParseCoin(string(tag.Value))
				if err != nil {
					rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
					return
				}
				res.Coin = coin.String()
			}
		}

		if res.Coin == "" {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "Unkwon issue")
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

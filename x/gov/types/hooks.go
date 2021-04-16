package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple governance hooks, all hook functions are run in array sequence
type MultiGovHooks []GovHooks

func NewMultiGovHooks(hooks ...GovHooks) MultiGovHooks {
	return hooks
}

func (h MultiGovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	for i := range h {
		h[i].AfterProposalSubmission(ctx, proposalID)
	}
}

func (h MultiGovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositAmount sdk.Coins) {
	for i := range h {
		h[i].AfterProposalDeposit(ctx, proposalID, depositAmount)
	}
}

func (h MultiGovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, option VoteOption) {
	for i := range h {
		h[i].AfterProposalVote(ctx, proposalID, voterAddr, option)
	}
}
func (h MultiGovHooks) AfterProposalInactive(ctx sdk.Context, proposalID uint64) {
	for i := range h {
		h[i].AfterProposalInactive(ctx, proposalID)
	}
}
func (h MultiGovHooks) AfterProposalActive(ctx sdk.Context, proposalID uint64) {
	for i := range h {
		h[i].AfterProposalActive(ctx, proposalID)
	}
}

package cmds

import (
	"context"

	currencycmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	currencyprocessor "github.com/ProtoconNet/mitum-currency/v3/operation/processor"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-sto/operation/kyc"
	"github.com/ProtoconNet/mitum-sto/operation/sto"

	"github.com/ProtoconNet/mitum-sto/operation/processor"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/ps"
)

var PNameOperationProcessorsMap = ps.Name("mitum-sto-operation-processors-map")

type processorInfo struct {
	hint      hint.Hint
	processor types.GetNewProcessor
}

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database
	var opr *currencyprocessor.OperationProcessor
	var set *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		currencycmds.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &set,
	); err != nil {
		return pctx, err
	}

	err := opr.SetCheckDuplicationFunc(processor.CheckDuplication)
	if err != nil {
		return pctx, err
	}
	err = opr.SetGetNewProcessorFunc(processor.GetNewProcessor)
	if err != nil {
		return pctx, err
	}

	ps := []processorInfo{
		{sto.AuthorizeOperatorsHint, sto.NewAuthorizeOperatorsProcessor()},
		{sto.CreateSecurityTokensHint, sto.NewCreateSecurityTokensProcessor()},
		{sto.IssueSecurityTokensHint, sto.NewIssueSecurityTokensProcessor()},
		{sto.RedeemTokensHint, sto.NewRedeemTokensProcessor()},
		{sto.RevokeOperatorsHint, sto.NewRevokeOperatorsProcessor()},
		{sto.SetDocumentHint, sto.NewSetDocumentProcessor()},
		{sto.TransferSecurityTokensPartitionHint, sto.NewTransferSecurityTokensPartitionProcessor()},
		{kyc.AddControllersHint, kyc.NewAddControllersProcessor()},
		{kyc.AddCustomersHint, kyc.NewAddCustomersProcessor()},
		{kyc.CreateKYCServiceHint, kyc.NewCreateKYCServiceProcessor()},
		{kyc.RemoveControllersHint, kyc.NewRemoveControllersProcessor()},
		{kyc.UpdateCustomersHint, kyc.NewUpdateCustomersProcessor()},
	}

	for _, p := range ps {
		if err := opr.SetProcessor(p.hint, p.processor); err != nil {
			return pctx, err
		}

		if err := set.Add(p.hint, func(height base.Height) (base.OperationProcessor, error) {
			return opr.New(
				height,
				db.State,
				nil,
				nil,
			)
		}); err != nil {
			return pctx, err
		}
	}

	var f currencycmds.ProposalOperationFactHintFunc = IsSupportedProposalOperationFactHintFunc

	pctx = context.WithValue(pctx, currencycmds.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, set) //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, currencycmds.ProposalOperationFactHintContextKey, f)

	return pctx, nil
}

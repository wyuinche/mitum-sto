package sto

import (
	"context"
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/v2/currency"
	"github.com/ProtoconNet/mitum-currency/v2/currency"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var revokeOperatorsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RevokeOperatorsItemProcessor)
	},
}

var revokeOperatorsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RevokeOperatorsProcessor)
	},
}

func (RevokeOperators) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type RevokeOperatorsItemProcessor struct {
	h            util.Hash
	sender       base.Address
	item         RevokeOperatorsItem
	operators    *[]base.Address
	tokenHolders *[]base.Address
}

func (ipp *RevokeOperatorsItemProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) error {
	it := ipp.item

	if err := checkExistsState(extensioncurrency.StateKeyContractAccount(it.Contract()), getStateFunc); err != nil {
		return err
	}

	if err := checkExistsState(StateKeySTODesign(it.Contract(), it.STO()), getStateFunc); err != nil {
		return err
	}

	if err := checkExistsState(StateKeyPartitionBalance(it.Contract(), it.STO(), it.Partition()), getStateFunc); err != nil {
		return err
	}

	for i, ad := range *ipp.operators {
		if ad.Equal(it.Operator()) {
			break
		}

		if i == len(*ipp.operators)-1 {
			return errors.Errorf("operator not in tokenholder operators, %q", it.Operator())
		}
	}

	for i, ad := range *ipp.tokenHolders {
		if ad.Equal(ipp.sender) {
			break
		}

		if i == len(*ipp.tokenHolders)-1 {
			return errors.Errorf("sender not in operator tokenholders, %q", ipp.sender)
		}
	}

	if err := checkExistsState(currency.StateKeyCurrencyDesign(it.Currency()), getStateFunc); err != nil {
		return err
	}

	return nil
}

func (ipp *RevokeOperatorsItemProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	sts := make([]base.StateMergeValue, 1)

	it := ipp.item

	for i, ad := range *ipp.operators {
		if ad.Equal(it.Operator()) {
			if i < len(*ipp.operators)-1 {
				copy((*ipp.operators)[i:], (*ipp.operators)[i+1:])
			}
			*ipp.operators = (*ipp.operators)[:len(*ipp.operators)-1]
			break
		}

		if i == len(*ipp.operators)-1 {
			return nil, errors.Errorf("operator not in tokenholder operators, %q", it.Operator())
		}
	}

	holders := *ipp.tokenHolders
	for i, ad := range holders {
		if ad.Equal(ipp.sender) {
			if i < len(holders)-1 {
				copy((holders)[i:], (holders)[i+1:])
			}
			holders = (holders)[:len(holders)-1]
			break
		}

		if i == len(holders)-1 {
			return nil, errors.Errorf("sender not in operator tokenholders, %q", ipp.sender)
		}
	}

	sts[0] = NewStateMergeValue(
		StateKeyOperatorTokenHolders(it.Contract(), it.STO(), it.Operator()),
		NewOperatorTokenHoldersStateValue(holders),
	)

	return sts, nil
}

func (ipp *RevokeOperatorsItemProcessor) Close() error {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = RevokeOperatorsItem{}
	ipp.operators = nil
	ipp.tokenHolders = nil

	revokeOperatorsItemProcessorPool.Put(ipp)

	return nil
}

type RevokeOperatorsProcessor struct {
	*base.BaseOperationProcessor
}

func NewRevokeOperatorsProcessor() extensioncurrency.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new RevokeOperatorsProcessor")

		nopp := revokeOperatorsProcessorPool.Get()
		opp, ok := nopp.(*RevokeOperatorsProcessor)
		if !ok {
			return nil, e(nil, "expected RevokeOperatorsProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e(err, "")
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *RevokeOperatorsProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess RevokeOperators")

	fact, ok := op.Fact().(RevokeOperatorsFact)
	if !ok {
		return ctx, nil, e(nil, "expected RevokeOperatorsFact, not %T", op.Fact())
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, nil, e(err, "")
	}

	if err := checkExistsState(currency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
	}

	if err := checkNotExistsState(extensioncurrency.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("contract account cannot set its operators, %q", fact.Sender()), nil
	}

	if err := checkFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	operators := map[string]*[]base.Address{}
	holders := map[string]*[]base.Address{}

	for _, it := range fact.Items() {
		var ops, hds []base.Address

		k := StateKeyTokenHolderPartitionOperators(it.Contract(), it.STO(), fact.sender, it.Partition())
		if _, found := operators[k]; !found {
			switch st, found, err := getStateFunc(k); {
			case err != nil:
				return nil, base.NewBaseOperationProcessReasonError("failed to find tokenholder partition operators, %s", k), nil
			case found:
				ops, err = StateTokenHolderPartitionOperatorsValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partition operators, %s", k), nil
				}
			default:
				return nil, base.NewBaseOperationProcessReasonError("tokenholder partition operators not in state, %q", k), nil
			}
			operators[k] = &ops
		}

		k = StateKeyOperatorTokenHolders(it.Contract(), it.STO(), it.Operator())
		if _, found := holders[k]; !found {
			switch st, found, err := getStateFunc(k); {
			case err != nil:
				return nil, base.NewBaseOperationProcessReasonError("failed to find operator tokenholders, %s", k), nil
			case found:
				hds, err = StateOperatorTokenHoldersValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError("failed to get operator tokenholders, %s", k), nil
				}
			default:
				return nil, base.NewBaseOperationProcessReasonError("operator tokenholders not in state, %q", k), nil
			}
			holders[k] = &hds
		}
	}

	for _, it := range fact.Items() {
		ip := revokeOperatorsItemProcessorPool.Get()
		ipc, ok := ip.(*RevokeOperatorsItemProcessor)
		if !ok {
			return nil, nil, e(nil, "expected RevokeOperatorsItemProcessor, not %T", ip)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it
		ipc.operators = operators[StateKeyTokenHolderPartitionOperators(it.Contract(), it.STO(), fact.sender, it.Partition())]
		ipc.tokenHolders = holders[StateKeyOperatorTokenHolders(it.Contract(), it.STO(), it.Operator())]

		if err := ipc.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("fail to preprocess RevokeOperatorsItem: %w", err), nil
		}

		ipc.Close()
	}

	return ctx, nil, nil
}

func (opp *RevokeOperatorsProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringErrorFunc("failed to process RevokeOperators")

	fact, ok := op.Fact().(RevokeOperatorsFact)
	if !ok {
		return nil, nil, e(nil, "expected RevokeOperatorsFact, not %T", op.Fact())
	}

	var sts []base.StateMergeValue // nolint:prealloc

	operators := map[string]*[]base.Address{}
	holders := map[string]*[]base.Address{}

	for _, it := range fact.Items() {
		var ops, hds []base.Address

		k := StateKeyTokenHolderPartitionOperators(it.Contract(), it.STO(), fact.sender, it.Partition())
		if _, found := operators[k]; !found {
			switch st, found, err := getStateFunc(k); {
			case err != nil:
				return nil, base.NewBaseOperationProcessReasonError("failed to find tokenholder partition operators, %s", k), nil
			case found:
				ops, err = StateTokenHolderPartitionOperatorsValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partition operators, %s", k), nil
				}
			default:
				return nil, base.NewBaseOperationProcessReasonError("tokenholder partition operators not in state, %q", k), nil
			}
			operators[k] = &ops
		}

		k = StateKeyOperatorTokenHolders(it.Contract(), it.STO(), it.Operator())
		if _, found := holders[k]; !found {
			switch st, found, err := getStateFunc(k); {
			case err != nil:
				return nil, base.NewBaseOperationProcessReasonError("failed to find operator tokenholders, %s", k), nil
			case found:
				hds, err = StateOperatorTokenHoldersValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError("failed to get operator tokenholders, %s", k), nil
				}
			default:
				return nil, base.NewBaseOperationProcessReasonError("operator tokenholders not in state, %q", k), nil
			}

			holders[k] = &hds
		}
	}

	ipcs := make([]*RevokeOperatorsItemProcessor, len(fact.items))
	for i, it := range fact.Items() {
		ip := revokeOperatorsItemProcessorPool.Get()
		ipc, ok := ip.(*RevokeOperatorsItemProcessor)
		if !ok {
			return nil, nil, e(nil, "expected RevokeOperatorsItemProcessor, not %T", ip)
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it
		ipc.operators = operators[StateKeyTokenHolderPartitionOperators(it.Contract(), it.STO(), fact.sender, it.Partition())]
		ipc.tokenHolders = holders[StateKeyOperatorTokenHolders(it.Contract(), it.STO(), it.Operator())]

		s, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process RevokeOperatorsItem: %w", err), nil
		}
		sts = append(sts, s...)

		ipcs[i] = ipc
	}

	for k, v := range operators {
		sts = append(sts, NewStateMergeValue(
			k,
			NewTokenHolderPartitionOperatorsStateValue(*v),
		))
	}

	for _, ipc := range ipcs {
		ipc.Close()
	}

	fitems := fact.Items()
	items := make([]STOItem, len(fitems))
	for i := range fact.Items() {
		items[i] = fitems[i]
	}

	required, err := calculateSTOItemsFee(getStateFunc, items)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to calculate fee: %w", err), nil
	}
	sb, err := currency.CheckEnoughBalance(fact.sender, required, getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check enough balance: %w", err), nil
	}

	for i := range sb {
		v, ok := sb[i].Value().(currency.BalanceStateValue)
		if !ok {
			return nil, nil, e(nil, "expected BalanceStateValue, not %T", sb[i].Value())
		}
		stv := currency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(required[i][0])))
		sts = append(sts, currency.NewBalanceStateMergeValue(sb[i].Key(), stv))
	}

	return sts, nil, nil
}

func (opp *RevokeOperatorsProcessor) Close() error {
	revokeOperatorsProcessorPool.Put(opp)

	return nil
}

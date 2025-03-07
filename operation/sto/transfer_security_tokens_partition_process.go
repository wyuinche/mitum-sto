package sto

import (
	"context"
	"math/big"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	currencyoperation "github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	currencystate "github.com/ProtoconNet/mitum-currency/v3/state"
	currency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	extensioncurrency "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	stostate "github.com/ProtoconNet/mitum-sto/state/sto"
	stotypes "github.com/ProtoconNet/mitum-sto/types/sto"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var transferSecurityTokensPartitionItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransferSecurityTokensPartitionItemProcessor)
	},
}

var transferSecurityTokensPartitionProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransferSecurityTokensPartitionProcessor)
	},
}

func (TransferSecurityTokensPartition) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type TransferSecurityTokensPartitionItemProcessor struct {
	h          util.Hash
	sender     base.Address
	item       TransferSecurityTokensPartitionItem
	partitions map[string][]stotypes.Partition
	balances   map[string]common.Big
}

func (ipp *TransferSecurityTokensPartitionItemProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) error {
	it := ipp.item

	if err := currencystate.CheckExistsState(extensioncurrency.StateKeyContractAccount(it.Contract()), getStateFunc); err != nil {
		return err
	}

	if err := currencystate.CheckExistsState(currency.StateKeyAccount(it.TokenHolder()), getStateFunc); err != nil {
		return err
	}

	if err := currencystate.CheckNotExistsState(extensioncurrency.StateKeyContractAccount(it.TokenHolder()), getStateFunc); err != nil {
		return err
	}

	if err := currencystate.CheckExistsState(currency.StateKeyAccount(it.Receiver()), getStateFunc); err != nil {
		return err
	}

	if err := currencystate.CheckNotExistsState(extensioncurrency.StateKeyContractAccount(it.Receiver()), getStateFunc); err != nil {
		return err
	}

	partitions := ipp.partitions[stostate.StateKeyTokenHolderPartitions(it.Contract(), it.STO(), it.TokenHolder())]
	if len(partitions) == 0 {
		return errors.Errorf("empty tokenholder partitions, %s-%s-%s", it.Contract(), it.STO(), it.TokenHolder())
	}

	for i, p := range partitions {
		if p == it.Partition() {
			break
		}

		if i == len(partitions)-1 {
			return errors.Errorf("partition not in tokenholder partitions, %s-%s-%s, %q", it.Contract(), it.STO(), it.TokenHolder(), it.Partition())
		}
	}

	st, err := currencystate.ExistsState(stostate.StateKeyDesign(it.Contract(), it.STO()), "key of sto design", getStateFunc)
	if err != nil {
		return err
	}

	design, err := stostate.StateDesignValue(st)
	if err != nil {
		return err
	}

	policy := design.Policy()

	if !it.TokenHolder().Equal(ipp.sender) {
		controllers := policy.Controllers()
		isController, isOperator := false, false

		for _, con := range controllers {
			if con.Equal(ipp.sender) {
				isController = true
				break
			}
		}

		if !isController {
			st, err := currencystate.ExistsState(stostate.StateKeyTokenHolderPartitionOperators(it.Contract(), it.STO(), it.TokenHolder(), it.Partition()), "key of tokenholder partition operators", getStateFunc)
			if err != nil {
				return err
			}

			operators, err := stostate.StateTokenHolderPartitionOperatorsValue(st)
			if err != nil {
				return err
			}

			for _, op := range operators {
				if op.Equal(ipp.sender) {
					isOperator = true
					break
				}
			}
		}

		if !(isController || isOperator) {
			return errors.Errorf("sender is neither controller nor operator, %s, %q", it.Partition(), ipp.sender)
		}
	}

	gn := new(big.Int)
	gn.SetUint64(design.Granularity())

	if mod := common.NewBigFromBigInt(new(big.Int)).Mod(it.Amount().Int, gn); common.NewBigFromBigInt(mod).OverZero() {
		return errors.Errorf("amount unit does not comply with sto granularity rule, %q, %q", it.Amount(), design.Granularity())
	}

	if err := currencystate.CheckExistsState(currency.StateKeyCurrencyDesign(it.Currency()), getStateFunc); err != nil {
		return err
	}

	return nil
}

func (ipp *TransferSecurityTokensPartitionItemProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	it := ipp.item

	partitionsKey := stostate.StateKeyTokenHolderPartitions(it.Contract(), it.STO(), it.TokenHolder())
	balanceKey := stostate.StateKeyTokenHolderPartitionBalance(it.Contract(), it.STO(), it.TokenHolder(), it.Partition())

	receiverPartitionsKey := stostate.StateKeyTokenHolderPartitions(it.Contract(), it.STO(), it.Receiver())
	receiverBalanceKey := stostate.StateKeyTokenHolderPartitionBalance(it.Contract(), it.STO(), it.Receiver(), it.Partition())

	balance := ipp.balances[balanceKey]
	partitions := ipp.partitions[partitionsKey]

	receiverBalance := ipp.balances[receiverBalanceKey]
	receiverPartitions := ipp.partitions[receiverPartitionsKey]

	balance = balance.Sub(it.Amount())
	receiverBalance = receiverBalance.Add(it.Amount())

	sts := []base.StateMergeValue{}

	if !balance.OverZero() {
		for i, p := range partitions {
			if p == it.Partition() {
				if i < len(partitions)-1 {
					copy(partitions[i:], partitions[i+1:])
				}
				partitions = partitions[:len(partitions)-1]
			}
		}

		opk := stostate.StateKeyTokenHolderPartitionOperators(it.Contract(), it.STO(), it.TokenHolder(), it.Partition())

		var operators []base.Address
		switch st, found, err := getStateFunc(opk); {
		case err != nil:
			return nil, err
		case found:
			operators, err = stostate.StateTokenHolderPartitionOperatorsValue(st)
			if err != nil {
				return nil, err
			}
		default:
			operators = []base.Address{}
		}

		sts = append(sts, currencystate.NewStateMergeValue(
			opk, stostate.NewTokenHolderPartitionOperatorsStateValue([]base.Address{}),
		))

		for _, op := range operators {
			thk := stostate.StateKeyOperatorTokenHolders(it.Contract(), it.STO(), op, it.Partition())

			st, err := currencystate.ExistsState(thk, "key of operator tokenholders", getStateFunc)
			if err != nil {
				return nil, err
			}

			holders, err := stostate.StateOperatorTokenHoldersValue(st)
			if err != nil {
				return nil, err
			}

			for i, th := range holders {
				if th.Equal(it.TokenHolder()) {
					if i < len(holders)-1 {
						copy(holders[i:], holders[i+1:])
					}
					holders = holders[:len(holders)-1]
				}
			}

			sts = append(sts, currencystate.NewStateMergeValue(
				thk, stostate.NewOperatorTokenHoldersStateValue(holders),
			))
		}
	}

	if len(receiverPartitions) == 0 {
		receiverPartitions = append(receiverPartitions, it.Partition())
	} else {
		for i, p := range receiverPartitions {
			if p == it.Partition() {
				break
			}

			if i == len(receiverPartitions)-1 {
				receiverPartitions = append(receiverPartitions, it.Partition())
			}
		}
	}

	ipp.partitions[partitionsKey] = partitions
	ipp.partitions[receiverPartitionsKey] = receiverPartitions
	ipp.balances[balanceKey] = balance
	ipp.balances[receiverBalanceKey] = receiverBalance

	return sts, nil
}

func (ipp *TransferSecurityTokensPartitionItemProcessor) Close() error {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = TransferSecurityTokensPartitionItem{}
	ipp.balances = nil
	ipp.partitions = nil

	transferSecurityTokensPartitionItemProcessorPool.Put(ipp)

	return nil
}

type TransferSecurityTokensPartitionProcessor struct {
	*base.BaseOperationProcessor
}

func NewTransferSecurityTokensPartitionProcessor() currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new TransferSecurityTokensPartitionProcessor")

		nopp := transferSecurityTokensPartitionProcessorPool.Get()
		opp, ok := nopp.(*TransferSecurityTokensPartitionProcessor)
		if !ok {
			return nil, e.Wrap(errors.Errorf("expected TransferSecurityTokensPartitionProcessor, not %T", nopp))
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *TransferSecurityTokensPartitionProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess TransferSecurityTokensPartition")

	fact, ok := op.Fact().(TransferSecurityTokensPartitionFact)
	if !ok {
		return ctx, nil, e.Wrap(errors.Errorf("expected TransferSecurityTokensPartitionFact, not %T", op.Fact()))
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, nil, e.Wrap(err)
	}

	if err := currencystate.CheckExistsState(currency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
	}

	if err := currencystate.CheckNotExistsState(extensioncurrency.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("contract account cannot issue security tokens, %q: %w", fact.Sender(), err), nil
	}

	if err := currencystate.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	partitions := map[string][]stotypes.Partition{}

	for _, it := range fact.Items() {
		k := stostate.StateKeyTokenHolderPartitions(it.Contract(), it.STO(), it.TokenHolder())

		if _, found := partitions[k]; !found {
			pts, err := stostate.ExistsTokenHolderPartitions(it.Contract(), it.STO(), it.TokenHolder(), getStateFunc)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partitions value, %q: %w", k, err), nil
			}

			partitions[k] = pts
		}
	}

	for _, it := range fact.Items() {
		ip := transferSecurityTokensPartitionItemProcessorPool.Get()
		ipc, ok := ip.(*TransferSecurityTokensPartitionItemProcessor)
		if !ok {
			return nil, nil, e.Wrap(errors.Errorf("expected TransferSecurityTokensPartitionItemProcessor, not %T", ip))
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it
		ipc.partitions = partitions
		ipc.balances = nil

		if err := ipc.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("fail to preprocess TransferSecurityTokensPartitionItem: %w", err), nil
		}

		ipc.Close()
	}

	if err := checkEnoughTokenHolderBalance(getStateFunc, fact.Items()); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("not enough tokenholder partition balance: %w", err), nil
	}

	return ctx, nil, nil
}

func (opp *TransferSecurityTokensPartitionProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process TransferSecurityTokensPartition")

	fact, ok := op.Fact().(TransferSecurityTokensPartitionFact)
	if !ok {
		return nil, nil, e.Wrap(errors.Errorf("expected TransferSecurityTokensPartitionFact, not %T", op.Fact()))
	}

	partitions := map[string][]stotypes.Partition{}
	balances := map[string]common.Big{}

	for _, it := range fact.Items() {
		k := stostate.StateKeyTokenHolderPartitions(it.Contract(), it.STO(), it.TokenHolder())

		if _, found := partitions[k]; !found {
			pts, err := stostate.ExistsTokenHolderPartitions(it.Contract(), it.STO(), it.TokenHolder(), getStateFunc)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partitions value, %q: %w", k, err), nil
			}

			partitions[k] = pts
		}

		k = stostate.StateKeyTokenHolderPartitionBalance(it.Contract(), it.STO(), it.TokenHolder(), it.Partition())

		if _, found := balances[k]; !found {
			balance, err := stostate.ExistsTokenHolderPartitionBalance(it.Contract(), it.STO(), it.TokenHolder(), it.Partition(), getStateFunc)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partition balance value, %q: %w", k, err), nil
			}

			balances[k] = balance
		}

		k = stostate.StateKeyTokenHolderPartitions(it.Contract(), it.STO(), it.Receiver())

		if _, found := partitions[k]; !found {
			var pts []stotypes.Partition

			switch st, found, err := getStateFunc(k); {
			case err != nil:
				return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partitions, %q: %w", k, err), nil
			case found:
				pts, err = stostate.StateTokenHolderPartitionsValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partitions value, %q: %w", k, err), nil
				}
			default:
				pts = []stotypes.Partition{}
			}

			partitions[k] = pts
		}

		k = stostate.StateKeyTokenHolderPartitionBalance(it.Contract(), it.STO(), it.Receiver(), it.Partition())

		if _, found := balances[k]; !found {
			var am common.Big

			switch st, found, err := getStateFunc(k); {
			case err != nil:
				return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partition balance, %q: %w", k, err), nil
			case found:
				am, err = stostate.StateTokenHolderPartitionBalanceValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError("failed to get tokenholder partition balance value, %q: %w", k, err), nil
				}
			default:
				am = common.ZeroBig
			}

			balances[k] = am
		}
	}

	var sts []base.StateMergeValue // nolint:prealloc

	ipcs := make([]*TransferSecurityTokensPartitionItemProcessor, len(fact.Items()))
	for i, it := range fact.Items() {
		ip := transferSecurityTokensPartitionItemProcessorPool.Get()
		ipc, ok := ip.(*TransferSecurityTokensPartitionItemProcessor)
		if !ok {
			return nil, nil, e.Wrap(errors.Errorf("expected TransferSecurityTokensPartitionItemProcessor, not %T", ip))
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it
		ipc.partitions = partitions
		ipc.balances = balances

		s, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process TransferSecurityTokensPartitionItem: %w", err), nil
		}
		sts = append(sts, s...)

		ipcs[i] = ipc
	}

	for k, v := range partitions {
		sts = append(sts, currencystate.NewStateMergeValue(k, stostate.NewTokenHolderPartitionsStateValue(v)))
	}

	for _, it := range fact.Items() {
		k := stostate.StateKeyTokenHolderPartitionBalance(it.Contract(), it.STO(), it.TokenHolder(), it.Partition())
		sts = append(sts, currencystate.NewStateMergeValue(k, stostate.NewTokenHolderPartitionBalanceStateValue(balances[k], it.Partition())))

		k = stostate.StateKeyTokenHolderPartitionBalance(it.Contract(), it.STO(), it.Receiver(), it.Partition())
		sts = append(sts, currencystate.NewStateMergeValue(k, stostate.NewTokenHolderPartitionBalanceStateValue(balances[k], it.Partition())))
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
	sb, err := currencyoperation.CheckEnoughBalance(fact.sender, required, getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check enough balance: %w", err), nil
	}

	for i := range sb {
		v, ok := sb[i].Value().(currency.BalanceStateValue)
		if !ok {
			return nil, nil, e.Wrap(errors.Errorf("expected BalanceStateValue, not %T", sb[i].Value()))
		}
		stv := currency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(required[i][0])))
		sts = append(sts, currencystate.NewStateMergeValue(sb[i].Key(), stv))
	}

	return sts, nil, nil
}

func (opp *TransferSecurityTokensPartitionProcessor) Close() error {
	transferSecurityTokensPartitionProcessorPool.Put(opp)

	return nil
}

func checkEnoughTokenHolderBalance(getStateFunc base.GetStateFunc, items []TransferSecurityTokensPartitionItem) error {
	balances := map[string]common.Big{}
	amounts := map[string]common.Big{}

	for _, it := range items {
		k := stostate.StateKeyTokenHolderPartitionBalance(it.Contract(), it.STO(), it.TokenHolder(), it.Partition())

		if _, found := balances[k]; found {
			amounts[k] = amounts[k].Add(it.Amount())
			continue
		}

		balance, err := stostate.ExistsTokenHolderPartitionBalance(it.Contract(), it.STO(), it.TokenHolder(), it.Partition(), getStateFunc)
		if err != nil {
			return err
		}

		balances[k] = balance
		amounts[k] = it.Amount()
	}

	for k, balance := range balances {
		if balance.Compare(amounts[k]) < 0 {
			return errors.Errorf("tokenholder partition balance not over total amounts, %q, %q < %q", k, balance, amounts[k])
		}
	}

	return nil
}

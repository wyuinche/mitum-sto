package kyc

import (
	"context"
	"sync"

	currencyoperation "github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	currencystate "github.com/ProtoconNet/mitum-currency/v3/state"
	currency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	extensioncurrency "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	kycstate "github.com/ProtoconNet/mitum-sto/state/kyc"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var addCustomersItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(AddCustomersItemProcessor)
	},
}

var addCustomersProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(AddCustomersProcessor)
	},
}

func (AddCustomers) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type AddCustomersItemProcessor struct {
	h      util.Hash
	sender base.Address
	item   AddCustomersItem
}

func (ipp *AddCustomersItemProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) error {
	it := ipp.item

	st, err := currencystate.ExistsState(extensioncurrency.StateKeyContractAccount(it.Contract()), "key of contract account", getStateFunc)
	if err != nil {
		return err
	}

	ca, err := extensioncurrency.StateContractAccountValue(st)
	if err != nil {
		return err
	}

	if !ca.Owner().Equal(ipp.sender) {
		policy, err := kycstate.ExistsPolicy(it.Contract(), it.KYC(), getStateFunc)
		if err != nil {
			return err
		}

		controllers := policy.Controllers()
		if len(controllers) == 0 {
			return errors.Errorf("not contract account owner neither its controller, %s-%s", it.Contract(), it.KYC())
		}

		for i, con := range controllers {
			if con.Equal(ipp.sender) {
				break
			}

			if i == len(controllers)-1 {
				return errors.Errorf("not contract account owner neither its controller, %s-%s", it.Contract(), it.KYC())
			}
		}
	}

	if err := currencystate.CheckNotExistsState(kycstate.StateKeyCustomer(it.Contract(), it.KYC(), it.Customer()), getStateFunc); err != nil {
		return err
	}

	if err := currencystate.CheckExistsState(currency.StateKeyCurrencyDesign(it.Currency()), getStateFunc); err != nil {
		return err
	}

	return nil
}

func (ipp *AddCustomersItemProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	it := ipp.item

	v := currencystate.NewStateMergeValue(
		kycstate.StateKeyCustomer(it.Contract(), it.KYC(), it.Customer()),
		kycstate.NewCustomerStateValue(kycstate.Status(it.Status())),
	)

	sts := []base.StateMergeValue{v}

	return sts, nil
}

func (ipp *AddCustomersItemProcessor) Close() error {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = AddCustomersItem{}

	addCustomersItemProcessorPool.Put(ipp)

	return nil
}

type AddCustomersProcessor struct {
	*base.BaseOperationProcessor
}

func NewAddCustomersProcessor() currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new AddCustomersProcessor")

		nopp := addCustomersProcessorPool.Get()
		opp, ok := nopp.(*AddCustomersProcessor)
		if !ok {
			return nil, e.Wrap(errors.Errorf("expected AddCustomersProcessor, not %T", nopp))
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

func (opp *AddCustomersProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess AddCustomers")

	fact, ok := op.Fact().(AddCustomersFact)
	if !ok {
		return ctx, nil, e.Wrap(errors.Errorf("expected AddCustomersFact, not %T", op.Fact()))
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, nil, e.Wrap(err)
	}

	if err := currencystate.CheckExistsState(currency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
	}

	if err := currencystate.CheckNotExistsState(extensioncurrency.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("contract account cannot add customer status, %q: %w", fact.Sender(), err), nil
	}

	if err := currencystate.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	for _, it := range fact.Items() {
		ip := addCustomersItemProcessorPool.Get()
		ipc, ok := ip.(*AddCustomersItemProcessor)
		if !ok {
			return nil, nil, e.Wrap(errors.Errorf("expected AddCustomersItemProcessor, not %T", ip))
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it

		if err := ipc.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to preprocess AddCustomersItem: %w", err), nil
		}

		ipc.Close()
	}

	return ctx, nil, nil
}

func (opp *AddCustomersProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process AddCustomers")

	fact, ok := op.Fact().(AddCustomersFact)
	if !ok {
		return nil, nil, e.Wrap(errors.Errorf("expected AddCustomersFact, not %T", op.Fact()))
	}

	var sts []base.StateMergeValue // nolint:prealloc

	for _, it := range fact.Items() {
		ip := addCustomersItemProcessorPool.Get()
		ipc, ok := ip.(*AddCustomersItemProcessor)
		if !ok {
			return nil, nil, e.Wrap(errors.Errorf("expected AddCustomersItemProcessor, not %T", ip))
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it

		st, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process AddCustomersItem: %w", err), nil
		}

		sts = append(sts, st...)

		ipc.Close()
	}

	fitems := fact.Items()
	items := make([]KYCItem, len(fitems))
	for i := range fact.Items() {
		items[i] = fitems[i]
	}

	required, err := calculateKYCItemsFee(getStateFunc, items)
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

func (opp *AddCustomersProcessor) Close() error {
	addCustomersProcessorPool.Put(opp)

	return nil
}

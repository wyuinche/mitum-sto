package cmds

import (
	"context"

	"github.com/pkg/errors"

	currencycmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	"github.com/ProtoconNet/mitum-sto/operation/sto"
	"github.com/ProtoconNet/mitum2/base"
)

type RevokeOperatorsCommand struct {
	BaseCommand
	currencycmds.OperationFlags
	Sender    currencycmds.AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract  currencycmds.AddressFlag    `arg:"" name:"contract" help:"contract account address" required:"true"`
	STO       currencycmds.ContractIDFlag `arg:"" name:"sto-id" help:"sto id" required:"true"`
	Operator  currencycmds.AddressFlag    `arg:"" name:"operator" help:"operator" required:"true"`
	Partition PartitionFlag               `arg:"" name:"partition" help:"default partition" required:"true"`
	Currency  currencycmds.CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	sender    base.Address
	contract  base.Address
	operator  base.Address
}

func NewRevokeOperatorsCommand() RevokeOperatorsCommand {
	cmd := NewBaseCommand()
	return RevokeOperatorsCommand{
		BaseCommand: *cmd,
	}
}

func (cmd *RevokeOperatorsCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	encs = cmd.Encoders
	enc = cmd.Encoder

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	currencycmds.PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *RevokeOperatorsCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	sender, err := cmd.Sender.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	}
	cmd.sender = sender

	contract, err := cmd.Contract.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid contract account format, %q", cmd.Contract.String())
	}
	cmd.contract = contract

	operator, err := cmd.Operator.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid operator account format, %q", cmd.Operator.String())
	}
	cmd.operator = operator

	return nil
}

func (cmd *RevokeOperatorsCommand) createOperation() (base.Operation, error) { // nolint:dupl
	var items []sto.RevokeOperatorsItem

	item := sto.NewRevokeOperatorsItem(
		cmd.contract,
		cmd.STO.ID,
		cmd.operator,
		cmd.Partition.Partition,
		cmd.Currency.CID,
	)
	if err := item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := sto.NewRevokeOperatorsFact([]byte(cmd.Token), cmd.sender, items)

	op, err := sto.NewRevokeOperators(fact)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke operators operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke operators operation")
	}

	return op, nil
}

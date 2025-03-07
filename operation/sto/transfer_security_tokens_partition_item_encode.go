package sto

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	stotypes "github.com/ProtoconNet/mitum-sto/types/sto"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *TransferSecurityTokensPartitionItem) unpack(enc encoder.Encoder, ht hint.Hint, ca, sto, th, rc, p, am, cid string) error {
	e := util.StringError("failed to unmarshal TransferSecurityTokensPartitionItem")

	it.BaseHinter = hint.NewBaseHinter(ht)
	it.stoID = currencytypes.ContractID(sto)
	it.partition = stotypes.Partition(p)
	it.currency = currencytypes.CurrencyID(cid)

	switch a, err := base.DecodeAddress(ca, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.contract = a
	}

	switch a, err := base.DecodeAddress(th, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.tokenholder = a
	}

	switch a, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.receiver = a
	}

	amount, err := common.NewBigFromString(am)
	if err != nil {
		return e.Wrap(err)
	}
	it.amount = amount

	return nil
}

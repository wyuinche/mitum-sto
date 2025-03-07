package sto

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type TransferSecurityTokensPartitionFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Owner base.Address                          `json:"sender"`
	Items []TransferSecurityTokensPartitionItem `json:"items"`
}

func (fact TransferSecurityTokensPartitionFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(TransferSecurityTokensPartitionFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Owner:                 fact.sender,
		Items:                 fact.items,
	})
}

type TransferSecurityTokensPartitionFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Owner string          `json:"sender"`
	Items json.RawMessage `json:"items"`
}

func (fact *TransferSecurityTokensPartitionFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of TransferSecurityTokensPartitionFact")

	var uf TransferSecurityTokensPartitionFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Owner, uf.Items)
}

type TransferSecurityTokensPartitionMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op TransferSecurityTokensPartition) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(TransferSecurityTokensPartitionMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *TransferSecurityTokensPartition) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of TransferSecurityTokensPartition")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}

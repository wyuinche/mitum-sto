package kyc

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type UpdateCustomersFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Owner base.Address          `json:"sender"`
	Items []UpdateCustomersItem `json:"items"`
}

func (fact UpdateCustomersFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(UpdateCustomersFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Owner:                 fact.sender,
		Items:                 fact.items,
	})
}

type UpdateCustomersFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Owner string          `json:"sender"`
	Items json.RawMessage `json:"items"`
}

func (fact *UpdateCustomersFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of UpdateCustomersFact")

	var uf UpdateCustomersFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Owner, uf.Items)
}

type UpdateCustomersMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op UpdateCustomers) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(UpdateCustomersMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *UpdateCustomers) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of UpdateCustomers")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}

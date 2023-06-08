package kyc

import (
	"encoding/json"

	currencybase "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type RemoveControllersFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Owner base.Address            `json:"sender"`
	Items []RemoveControllersItem `json:"items"`
}

func (fact RemoveControllersFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RemoveControllersFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Owner:                 fact.sender,
		Items:                 fact.items,
	})
}

type RemoveControllersFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Owner string          `json:"sender"`
	Items json.RawMessage `json:"items"`
}

func (fact *RemoveControllersFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of RemoveControllersFact")

	var uf RemoveControllersFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Owner, uf.Items)
}

type RemoveControllersMarshaler struct {
	currencybase.BaseOperationJSONMarshaler
}

func (op RemoveControllers) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RemoveControllersMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *RemoveControllers) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of RemoveControllers")

	var ubo currencybase.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e(err, "")
	}

	op.BaseOperation = ubo

	return nil
}

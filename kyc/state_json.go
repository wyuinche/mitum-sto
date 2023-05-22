package kyc

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type DesignStateValueJSONMarshaler struct {
	hint.BaseHinter
	KYC Design `json:"kyc"`
}

func (de DesignStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(DesignStateValueJSONMarshaler{
		BaseHinter: de.BaseHinter,
		KYC:        de.Design,
	})
}

type DesignStateValueJSONUnmarshaler struct {
	KYC json.RawMessage `json:"kyc"`
}

func (de *DesignStateValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of DesignStateValue")

	var u DesignStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	var design Design

	if err := design.DecodeJSON(u.KYC, enc); err != nil {
		return e(err, "")
	}

	de.Design = design

	return nil
}

type CustomerStateValueJSONMarshaler struct {
	hint.BaseHinter
	Status Status `json:"status"`
}

func (cm CustomerStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CustomerStateValueJSONMarshaler{
		BaseHinter: cm.BaseHinter,
		Status:     cm.status,
	})
}

type CustomerStateValueJSONUnmarshaler struct {
	Status bool `json:"status"`
}

func (cm *CustomerStateValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of CustomerStateValue")

	var u CustomerStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}
	cm.status = Status(u.Status)

	return nil
}

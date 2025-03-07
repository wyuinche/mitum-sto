package sto // nolint:dupl

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it IssueSecurityTokensItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     it.Hint().String(),
			"contract":  it.contract,
			"stoid":     it.stoID,
			"receiver":  it.receiver,
			"amount":    it.amount.String(),
			"partition": it.partition,
			"currency":  it.currency,
		},
	)
}

type IssueSecurityTokensItemBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	Contract  string `bson:"contract"`
	STO       string `bson:"stoid"`
	Receiver  string `bson:"receiver"`
	Amount    string `bson:"amount"`
	Partition string `bson:"partition"`
	Currency  string `bson:"currency"`
}

func (it *IssueSecurityTokensItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of IssueSecurityTokensItem")

	var uit IssueSecurityTokensItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, ht, uit.Contract, uit.STO, uit.Receiver, uit.Amount, uit.Partition, uit.Currency)
}

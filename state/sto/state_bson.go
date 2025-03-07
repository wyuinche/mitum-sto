package sto

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	stotypes "github.com/ProtoconNet/mitum-sto/types/sto"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (de DesignStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": de.Hint().String(),
			"sto":   de.Design,
		},
	)
}

type DesignStateValueBSONUnmarshaler struct {
	Hint string   `bson:"_hint"`
	STO  bson.Raw `bson:"sto"`
}

func (de *DesignStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of DesignStateValue")

	var u DesignStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	de.BaseHinter = hint.NewBaseHinter(ht)

	var design stotypes.Design
	if err := design.DecodeBSON(u.STO, enc); err != nil {
		return e.Wrap(err)
	}

	de.Design = design

	return nil
}

func (p PartitionBalanceStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  p.Hint().String(),
			"amount": p.Amount.String(),
		},
	)
}

type PartitionBalanceStateValueBSONUnmarshaler struct {
	Hint   string `bson:"_hint"`
	Amount string `bson:"amount"`
}

func (de *PartitionBalanceStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of PartitionBalanceStateValue")

	var u PartitionBalanceStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	de.BaseHinter = hint.NewBaseHinter(ht)

	big, err := common.NewBigFromString(u.Amount)
	if err != nil {
		return err
	}
	de.Amount = big

	return nil
}

func (p TokenHolderPartitionsStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      p.Hint().String(),
			"partitions": p.Partitions,
		},
	)
}

type TokenHolderPartitionsStateValueBSONUnmarshaler struct {
	Hint       string   `bson:"_hint"`
	Partitions []string `bson:"partitions"`
}

func (p *TokenHolderPartitionsStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of TokenHolderPartitionsStateValue")

	var u TokenHolderPartitionsStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	p.BaseHinter = hint.NewBaseHinter(ht)

	partitions := make([]stotypes.Partition, len(u.Partitions))
	for i, s := range u.Partitions {
		partitions[i] = stotypes.Partition(s)
	}

	p.Partitions = partitions

	return nil
}

func (p TokenHolderPartitionBalanceStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      p.Hint().String(),
			"amount":     p.Amount.String(),
			"partitions": p.Partition,
		},
	)
}

type TokenHolderPartitionBalanceStateValueBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	Amount    string `bson:"amount"`
	Partition string `bson:"partition"`
}

func (p *TokenHolderPartitionBalanceStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of TokenHolderPartitionBalanceStateValue")

	var u TokenHolderPartitionBalanceStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	p.BaseHinter = hint.NewBaseHinter(ht)

	big, err := common.NewBigFromString(u.Amount)
	if err != nil {
		return e.Wrap(err)
	}
	p.Amount = big

	p.Partition = stotypes.Partition(u.Partition)

	return nil
}

func (ops TokenHolderPartitionOperatorsStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     ops.Hint().String(),
			"operators": ops.Operators,
		},
	)
}

type TokenHolderPartitionOperatorsStateValueBSONUnmarshaler struct {
	Hint      string   `bson:"_hint"`
	Operators []string `bson:"operators"`
}

func (ops *TokenHolderPartitionOperatorsStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of TokenHolderPartitionOperatorsStateValue")

	var u TokenHolderPartitionOperatorsStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	ops.BaseHinter = hint.NewBaseHinter(ht)

	operators := make([]base.Address, len(u.Operators))
	for i := range u.Operators {
		a, err := base.DecodeAddress(u.Operators[i], enc)
		if err != nil {
			return e.Wrap(err)
		}
		operators[i] = a
	}
	ops.Operators = operators

	return nil
}

func (oth OperatorTokenHoldersStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":        oth.Hint().String(),
			"tokenholders": oth.TokenHolders,
		},
	)
}

type OperatorTokenHoldersStateValueBSONUnmarshaler struct {
	Hint         string   `bson:"_hint"`
	TokenHolders []string `bson:"tokenholders"`
}

func (oth *OperatorTokenHoldersStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of OperatorTokenHoldersStateValue")

	var u OperatorTokenHoldersStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	oth.BaseHinter = hint.NewBaseHinter(ht)

	holders := make([]base.Address, len(u.TokenHolders))
	for i := range u.TokenHolders {
		a, err := base.DecodeAddress(u.TokenHolders[i], enc)
		if err != nil {
			return e.Wrap(err)
		}
		holders[i] = a
	}
	oth.TokenHolders = holders

	return nil
}

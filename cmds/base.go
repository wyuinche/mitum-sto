package cmds

import (
	"context"
	"io"
	"os"

	currencycmds "github.com/ProtoconNet/mitum-currency/v3/cmds"

	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/ProtoconNet/mitum2/util/ps"
	"github.com/rs/zerolog"
)

type BaseCommand struct {
	Encoder  *jsonenc.Encoder
	Encoders *encoder.Encoders
	Log      *zerolog.Logger
	Out      io.Writer `kong:"-"`
}

func NewBaseCommand() *BaseCommand {
	return &BaseCommand{
		Out: os.Stdout,
	}
}

func (cmd *BaseCommand) prepare(pctx context.Context) (context.Context, error) {
	cmd.Out = os.Stdout
	pps := ps.NewPS("cmd")

	_ = pps.
		AddOK(launch.PNameEncoder, currencycmds.PEncoder, nil)

	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)

	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return pctx, err
	}

	cmd.Log = log.Log()

	pctx, err := pps.Run(pctx) //revive:disable-line:modifies-parameter
	if err != nil {
		return pctx, err
	}

	return pctx, util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &cmd.Encoders,
		launch.EncoderContextKey, &cmd.Encoder,
	)
}

func PAddHinters(ctx context.Context) (context.Context, error) {
	e := util.StringError("add hinters")

	var enc encoder.Encoder
	if err := util.LoadFromContextOK(ctx, launch.EncoderContextKey, &enc); err != nil {
		return ctx, e.Wrap(err)
	}
	var benc encoder.Encoder
	if err := util.LoadFromContextOK(ctx, currencycmds.BEncoderContextKey, &benc); err != nil {
		return ctx, e.Wrap(err)
	}

	if err := LoadHinters(enc); err != nil {
		return ctx, e.Wrap(err)
	}
	if err := LoadHinters(benc); err != nil {
		return ctx, e.Wrap(err)
	}

	return ctx, nil
}

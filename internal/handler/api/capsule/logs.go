package capsule

import (
	"context"
	"io"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) Logs(ctx context.Context, req *connect.Request[capsule.LogsRequest], stream *connect.ServerStream[capsule.LogsResponse]) error {
	it, err := h.cs.Logs(ctx, req.Msg.GetCapsuleId(), req.Msg.GetInstanceId(), req.Msg.GetFollow())
	if err != nil {
		return err
	}

	defer it.Close()

	for {
		log, err := it.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if err := stream.Send(&capsule.LogsResponse{
			Log: log,
		},
		); err != nil {
			return err
		}
	}
}

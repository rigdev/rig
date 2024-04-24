package pipeline

import (
	"context"

	connect "connectrpc.com/connect"
	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
)

func (h *handler) WatchObjectStatus(
	ctx context.Context,
	req *connect.Request[apipipeline.WatchObjectStatusRequest],
	stream *connect.ServerStream[apipipeline.WatchObjectStatusResponse],
) error {
	c := make(chan *apipipeline.ObjectStatusChange, 128)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		for change := range c {
			if err := stream.Send(&apipipeline.WatchObjectStatusResponse{
				Change: change,
			}); err != nil {
				h.logger.Error(err, "error sending object status")
				return
			}
		}
	}()

	return h.objectstatus.Watch(ctx, req.Msg.GetNamespace(), c)
}

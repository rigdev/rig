package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

// DownloadObject implements storageconnect.ServiceHandler
func (h *Handler) DownloadObject(ctx context.Context, req *connect.Request[storage.DownloadObjectRequest], stream *connect.ServerStream[storage.DownloadObjectResponse]) error {
	reader, err := h.ss.DownloadObject(ctx, req.Msg.Bucket, req.Msg.Path)
	if err != nil {
		return err
	}
	defer reader.Close()
	buf := make([]byte, 64*1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
		if err := stream.Send(&storage.DownloadObjectResponse{
			Chunk: buf[:n],
		}); err != nil {
			return err
		}
	}
	return nil
}

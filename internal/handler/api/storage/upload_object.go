package storage

import (
	context "context"
	"io"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/errors"
)

// UploadObject implements storageconnect.ServiceHandler
func (h *Handler) UploadObject(ctx context.Context, stream *connect.ClientStream[storage.UploadObjectRequest]) (*connect.Response[storage.UploadObjectResponse], error) {
	if stream.Receive() {
		switch v := stream.Msg().GetRequest().(type) {
		case *storage.UploadObjectRequest_Metadata_:
			r, w := io.Pipe()
			go func() {
				for stream.Receive() {
					switch v := stream.Msg().GetRequest().(type) {
					case *storage.UploadObjectRequest_Chunk:
						_, err := w.Write(v.Chunk)
						if err != nil {
							w.CloseWithError(err)
							return
						}
					default:
						w.CloseWithError(errors.InvalidArgumentErrorf("invalid request"))
						return
					}
				}
				if err := stream.Err(); err != nil {
					w.CloseWithError(err)
					return
				} else {
					w.Close()
				}
			}()
			_, _, err := h.ss.UploadObject(ctx, r, v.Metadata)
			if err != nil {
				return nil, err
			}
			return &connect.Response[storage.UploadObjectResponse]{}, nil
		default:
			return nil, errors.InvalidArgumentErrorf("invalid request")
		}
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}

	return nil, errors.InvalidArgumentErrorf("invalid request")
}

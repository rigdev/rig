package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
)

// Logs implements cluster.Gateway.
func (c *Client) Logs(ctx context.Context, capsuleName string, instanceID string, follow bool) (iterator.Iterator[*capsule.Log], error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	req := c.cs.CoreV1().
		Pods(projectID.String()).
		GetLogs(instanceID, &v1.PodLogOptions{
			Container:  capsuleName,
			Timestamps: true,
			Follow:     follow,
		})
	rc, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get log stream: %w", err)
	}

	p := iterator.NewProducer[*capsule.Log]()
	w := &logsWriter{p: p}

	go func() {
		_, err := io.Copy(w, rc)
		p.Error(err)
	}()

	return p, nil
}

type logsWriter struct {
	p *iterator.Producer[*capsule.Log]
}

func (w *logsWriter) Write(bs []byte) (int, error) {
	l := &capsule.Log{
		Message: &capsule.LogMessage{},
	}
	index := bytes.IndexByte(bs, ' ')
	if index > 0 {
		// TODO: check if this parsing works
		if ts, err := time.Parse(time.RFC3339Nano, string(bs[:index])); err == nil {
			l.Timestamp = timestamppb.New(ts)
		}
	}

	// Note that when returning from `Write`, the buffer may no longer be referenced -> dup.
	out := bytes.Clone(bs[index+1:])
	l.Message.Message = &capsule.LogMessage_Stdout{Stdout: out}
	if err := w.p.Value(l); err != nil {
		return 0, err
	}

	return len(bs), nil
}

package storage

import (
	"context"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var excludeList = []string{
	".git",
	".DS_Store",
}

func (c Cmd) cp(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	rawFrom := args[0]
	rawTo := args[1]

	pw := progress.NewWriter()
	pw.SetOutputWriter(cmd.OutOrStderr())
	pw.SetStyle(progress.StyleCircle)
	pw.SetNumTrackersExpected(3)
	go pw.Render()

	if isRigUri(rawFrom) {
		bucket, prefix, err := parseRigUri(rawFrom)
		if err != nil {
			return err
		}
		results := []*storage.ListObjectsResponse_Result{}
		base := ""
		if prefix != "" {
			base = filepath.Base(prefix)
		}
		if !strings.Contains(base, ".") {
			res, err := c.Rig.Storage().ListObjects(ctx, &connect.Request[storage.ListObjectsRequest]{
				Msg: &storage.ListObjectsRequest{
					Bucket:    bucket,
					Prefix:    prefix,
					Recursive: storageRecursive,
				},
			})
			if err != nil {
				return err
			}
			results = res.Msg.GetResults()
		} else {
			toBase := filepath.Base(rawTo)
			if !strings.Contains(toBase, ".") {
				rawTo = path.Join(rawTo, base)
			}

			res, err := c.Rig.Storage().GetObject(ctx, &connect.Request[storage.GetObjectRequest]{
				Msg: &storage.GetObjectRequest{
					Bucket: bucket,
					Path:   prefix,
				},
			})
			if err != nil {
				return err
			}
			results = append(results, &storage.ListObjectsResponse_Result{
				Result: &storage.ListObjectsResponse_Result_Object{
					Object: res.Msg.GetObject(),
				},
			})
		}
		if isRigUri(rawTo) {
			// Copy.
			dstBucket, dstPrefix, err := parseRigUri(rawTo)
			if err != nil {
				return err
			}
			var p int64 = 3
			sem := semaphore.NewWeighted(p)
			for _, o := range results {
				obj := o.GetObject()
				if obj == nil {
					continue
				}
				p := strings.TrimPrefix(obj.GetPath(), prefix)

				sem.Acquire(ctx, 1)

				go func() {
					t := &progress.Tracker{
						Message: obj.GetPath(),
						Units:   progress.UnitsBytes,
						Total:   int64(obj.GetSize()),
					}
					pw.AppendTracker(t)

					if _, err := c.Rig.Storage().CopyObject(ctx, &connect.Request[storage.CopyObjectRequest]{
						Msg: &storage.CopyObjectRequest{
							FromBucket: bucket,
							FromPath:   obj.GetPath(),
							ToBucket:   dstBucket,
							ToPath:     path.Join(dstPrefix, p),
						},
					}); err != nil {
						log.Fatal(err)
					}
					t.Increment(t.Total)
					sem.Release(1)
				}()
			}
			sem.Acquire(ctx, p)
			pw.Stop()
			return nil
		} else {
			// Download.
			var p int64 = 3
			sem := semaphore.NewWeighted(p)
			for _, o := range results {
				obj := o.GetObject()
				if obj == nil {
					continue
				}
				p := strings.TrimPrefix(obj.GetPath(), prefix)

				sem.Acquire(ctx, 1)
				go func() {
					t := &progress.Tracker{
						Message: obj.GetPath(),
						Units:   progress.UnitsBytes,
						Total:   int64(obj.GetSize()),
					}
					pw.AppendTracker(t)

					if err := c.downloadFile(ctx, cmd, t, bucket, path.Join(rawTo, p)); err != nil {
						log.Fatal(err)
					}

					sem.Release(1)
				}()
			}

			sem.Acquire(ctx, p)
			pw.Stop()
			return nil
		}
	} else if isRigUri(rawTo) {
		// Upload.
		bucket, prefix, err := parseRigUri(rawTo)
		if err != nil {
			return err
		}

		itFiles := iterator.NewProducer[string]()

		go func() {
			defer itFiles.Done()
			if err := filepath.WalkDir(rawFrom, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if regexp.MustCompile(strings.Join(excludeList, "|")).Match([]byte(path)) {
					return nil
				}
				if path != rawFrom && !storageRecursive && d.IsDir() {
					return filepath.SkipDir
				}
				if d.Type().IsRegular() {
					itFiles.Value(path)
				}
				return nil
			}); err != nil {
				itFiles.Error(err)
			}
		}()

		it := iterator.Map[string](itFiles, func(filePath string) (*progress.Tracker, error) {
			t := &progress.Tracker{
				Message: filePath,
				Units:   progress.UnitsBytes,
			}
			pw.AppendTracker(t)
			return t, nil
		})

		defer it.Close()

		var p int64 = 3

		sem := semaphore.NewWeighted(p)
		for {
			t, err := it.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			p := strings.TrimPrefix(t.Message, rawFrom)

			sem.Acquire(ctx, 1)

			go func() {
				if err := c.uploadFile(ctx, cmd, t, bucket, path.Join(prefix, p)); err != nil {
					log.Fatal(err)
				}

				sem.Release(1)
			}()
		}

		sem.Acquire(ctx, p)
		pw.Stop()
		return nil
	} else {
		return errors.InvalidArgumentErrorf("one of `from` and `to` must be a storage path")
	}
}

func (c Cmd) uploadFile(ctx context.Context, cmd *cobra.Command, t *progress.Tracker, bucket, path string) error {
	from := t.Message

	f, err := os.Open(from)
	if err != nil {
		return err
	}

	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		return err
	}

	size := s.Size()

	t.UpdateTotal(size)

	mimeData := make([]byte, 512)
	n, err := f.Read(mimeData)
	if err != nil {
		return err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	m := &storage.UploadObjectRequest_Metadata{
		Bucket:      bucket,
		Path:        path,
		Size:        uint64(size),
		ContentType: http.DetectContentType(mimeData[:n]),
	}

	// Upload.
	cc := c.Rig.Storage().UploadObject(ctx)
	if err := cc.Send(&storage.UploadObjectRequest{Request: &storage.UploadObjectRequest_Metadata_{Metadata: m}}); err != nil {
		return err
	}

	buffer := make([]byte, 64*1024)
	for {
		n, err := f.Read(buffer)
		if err == io.EOF {
			_, err := cc.CloseAndReceive()
			if err != nil {
				return err
			}

			return nil
		} else if err != nil {
			return err
		}

		if err := cc.Send(&storage.UploadObjectRequest{Request: &storage.UploadObjectRequest_Chunk{Chunk: buffer[:n]}}); err != nil {
			return err
		}

		t.Increment(int64(n))
	}
}

func (c Cmd) downloadFile(ctx context.Context, cmd *cobra.Command, t *progress.Tracker, bucket, path string) error {
	from := t.Message
	// Create the directories if they don't exist.
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	// Download.
	cc, err := c.Rig.Storage().DownloadObject(ctx, &connect.Request[storage.DownloadObjectRequest]{
		Msg: &storage.DownloadObjectRequest{
			Bucket: bucket,
			Path:   from,
		},
	})
	if err != nil {
		return err
	}
	defer cc.Close()

	for cc.Receive() {
		res := cc.Msg()
		n, err := f.Write(res.GetChunk())
		if err != nil {
			return err
		}
		t.Increment(int64(n))
	}
	// For some reason the EOF error does not match io.EOF, but instead is unknown at just says unknown: EOF
	if cc.Err() == io.EOF {
		return nil
	}
	if errors.IsUnknown(cc.Err()) {
		return nil
	} else if cc.Err() != nil {
		return cc.Err()
	} else {
		return nil
	}
}

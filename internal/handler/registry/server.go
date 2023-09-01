package registry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	"github.com/distribution/distribution/v3/registry/storage/driver"
	"github.com/distribution/distribution/v3/registry/storage/driver/factory"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	storage_gateway "github.com/rigdev/rig/internal/gateway/storage"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	cfg    config.Config
	logger *zap.Logger
	sg     storage_gateway.Gateway
}

func NewServer(lc fx.Lifecycle, cfg config.Config, logger *zap.Logger, sg storage_gateway.Gateway) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		logger: logger.Named("registry").WithOptions(zap.IncreaseLevel(zap.WarnLevel)),
		sg:     sg,
	}

	// create the deafult registry bucket
	if _, err := s.sg.GetBucket(context.Background(), "registry"); err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		if _, err := s.sg.CreateBucket(context.Background(), "registry", ""); err != nil {
			return nil, err
		}
	}

	lc.Append(fx.StartStopHook(s.Start, s.Stop))
	factory.Register("rig", s)
	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	regCfg := &configuration.Configuration{
		Storage: configuration.Storage{
			"rig": configuration.Parameters{},
		},
	}
	regCfg.HTTP.Addr = fmt.Sprint(":", s.cfg.Registry.Port)
	regCfg.Log.Level = configuration.Loglevel("error")
	regCfg.Log.AccessLog.Disabled = true

	r, err := registry.NewRegistry(ctx, regCfg)
	if err != nil {
		return err
	}

	go func() {
		if err := r.ListenAndServe(); err != nil {
			s.logger.Fatal("error serving registry", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}

func (s *Server) Create(parameters map[string]interface{}) (driver.StorageDriver, error) {
	return &storageDriver{
		sg:     s.sg,
		logger: s.logger.With(zap.String("bucket", "registry")),
		bucket: "registry",
	}, nil
}

type storageDriver struct {
	sg     storage_gateway.Gateway
	logger *zap.Logger
	bucket string
}

// Name returns the human-readable "name" of the driver, useful in error
// messages and logging. By convention, this will just be the registration
// name, but drivers may provide other information here.
func (d *storageDriver) Name() string { return "rig" }

func (d *storageDriver) GetContent(ctx context.Context, path string) ([]byte, error) {
	d.logger.Debug("get content", zap.String("path", path))
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)
	r, err := d.sg.DownloadObject(ctx, d.bucket, path)
	if errors.IsNotFound(err) {
		d.logger.Info("content not found", zap.String("path", path))
		return nil, driver.PathNotFoundError{Path: path}
	} else if err != nil {
		return nil, err
	}

	defer r.Close()

	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (d *storageDriver) PutContent(ctx context.Context, path string, content []byte) error {
	d.logger.Debug("put content", zap.String("path", path), zap.Int("size", len(content)))
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)
	_, _, err := d.sg.UploadObject(ctx, bytes.NewReader(content), int64(len(content)), d.bucket, path, "")
	if err != nil {
		return err
	}

	return nil
}

func (d *storageDriver) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	d.logger.Debug("new reader", zap.String("path", path), zap.Int64("offset", offset))
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)

	r, err := d.sg.DownloadObject(ctx, d.bucket, path)
	if err != nil {
		return nil, err
	}

	if _, err := r.Seek(offset, 0); err != nil {
		return nil, err
	}

	return r, err
}

type fileWriter struct {
	ctx    context.Context
	cancel context.CancelFunc
	bucket string
	path   string
	sg     storage_gateway.Gateway
	size   int64
	done   <-chan struct{}

	io.WriteCloser
}

func (w *fileWriter) Write(bs []byte) (int, error) {
	w.size += int64(len(bs))
	return w.WriteCloser.Write(bs)
}

func (w *fileWriter) Close() error {
	if err := w.WriteCloser.Close(); err != nil {
		return err
	}

	<-w.ctx.Done()

	return nil
}

// Size returns the number of bytes written to this FileWriter.
func (w *fileWriter) Size() int64 {
	return w.size
}

// Cancel removes any written content from this FileWriter.
func (w *fileWriter) Cancel(ctx context.Context) error {
	return w.deleteParts(ctx)
}

// Commit flushes all content written to this FileWriter and makes it
// available for future calls to StorageDriver.GetContent and
// StorageDriver.Reader.
func (w *fileWriter) Commit() error {
	ls, err := w.getParts(w.ctx)
	if err != nil {
		return err
	}

	var srcs []string
	for _, l := range ls {
		if l.Size > 0 {
			srcs = append(srcs, l.GetPath())
		}
	}

	if err := w.sg.ComposeObject(w.ctx, w.bucket, w.path, srcs...); err != nil {
		return err
	}

	return w.deleteObjects(w.ctx, ls)
}

func (w *fileWriter) getParts(ctx context.Context) ([]*storage.Object, error) {
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)
	_, it, err := w.sg.ListObjects(ctx, w.bucket, "", w.path+"/", "", "", false, 1024)
	if err != nil {
		return nil, err
	}

	return iterator.Collect(iterator.Map(it, func(f *storage.ListObjectsResponse_Result) (*storage.Object, error) {
		switch v := f.GetResult().(type) {
		case *storage.ListObjectsResponse_Result_Object:
			return v.Object, nil
		default:
			return nil, errors.InvalidArgumentErrorf("invalid resource in storage")
		}
	}))
}

func (w *fileWriter) deleteParts(ctx context.Context) error {
	ls, err := w.getParts(ctx)
	if err != nil {
		return err
	}

	return w.deleteObjects(ctx, ls)
}

func (w *fileWriter) deleteObjects(ctx context.Context, ls []*storage.Object) error {
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)

	for _, l := range ls {
		if err := w.sg.DeleteObject(ctx, w.bucket, l.GetPath()); err != nil {
			return err
		}
	}

	return nil
}

func (d *storageDriver) Writer(ctx context.Context, contentPath string, append bool) (driver.FileWriter, error) {
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)

	r, w := io.Pipe()
	fw := &fileWriter{
		bucket:      d.bucket,
		path:        contentPath,
		sg:          d.sg,
		WriteCloser: w,
	}

	index := 0
	if append {
		ls, err := fw.getParts(ctx)
		if err != nil {
			return nil, err
		}
		for _, l := range ls {
			fw.size += int64(l.GetSize())
			index++
		}
	}

	fw.ctx, fw.cancel = context.WithCancel(auth.WithProjectID(context.Background(), auth.RigProjectID))

	go func() {
		defer fw.cancel()

		// Each part is suffixed by an incremental ordered index.
		partID := fmt.Sprintf("%07d", index)
		partPath := path.Join(contentPath, partID)
		_, _, err := d.sg.UploadObject(fw.ctx, r, -1, d.bucket, partPath, "")
		if err != nil {
			w.CloseWithError(err)
		}
	}()

	return fw, nil
}

type fileInfo struct {
	path    string
	size    int64
	modTime time.Time
}

func (fi *fileInfo) Path() string {
	return fi.path
}

// Size returns current length in bytes of the file. The return value can
// be used to write to the end of the file at path. The value is
// meaningless if IsDir returns true.
func (fi *fileInfo) Size() int64 {
	return fi.size
}

// ModTime returns the modification time for the file. For backends that
// don't have a modification time, the creation time should be returned.
func (fi *fileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns true if the path is a directory.
func (fi *fileInfo) IsDir() bool {
	return false
}

func (d *storageDriver) Stat(ctx context.Context, path string) (driver.FileInfo, error) {
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)
	o, err := d.sg.GetObject(ctx, d.bucket, path)
	if errors.IsNotFound(err) {
		return nil, driver.PathNotFoundError{Path: path}
	} else if err != nil {
		return nil, err
	}

	return &fileInfo{
		path:    o.GetPath(),
		size:    int64(o.GetSize()),
		modTime: o.LastModified.AsTime(),
	}, nil
}

func (d *storageDriver) List(ctx context.Context, folderPath string) ([]string, error) {
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)

	_, it, err := d.sg.ListObjects(ctx, d.bucket, "", path.Join(folderPath, ""), "", "", false, 1024)
	if err != nil {
		return nil, err
	}

	l, err := iterator.Collect(iterator.Map(it, func(f *storage.ListObjectsResponse_Result) (string, error) {
		switch v := f.GetResult().(type) {
		case *storage.ListObjectsResponse_Result_Object:
			return v.Object.Path, nil
		case *storage.ListObjectsResponse_Result_Folder:
			return v.Folder, nil
		default:
			return "", errors.InvalidArgumentErrorf("invalid resource in storage")
		}
	}))

	return l, err
}

func (d *storageDriver) Move(ctx context.Context, sourcePath string, destPath string) error {
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)

	if err := d.sg.CopyObject(ctx, d.bucket, destPath, d.bucket, sourcePath); err != nil {
		return err
	}

	return d.sg.DeleteObject(ctx, d.bucket, sourcePath)
}

func (d *storageDriver) Delete(ctx context.Context, path string) error {
	ctx = auth.WithProjectID(ctx, auth.RigProjectID)
	return d.sg.DeleteObject(ctx, d.bucket, path)
}

func (d *storageDriver) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	return "", driver.ErrUnsupportedMethod{}
}

func (d *storageDriver) Walk(ctx context.Context, path string, f driver.WalkFn) error {
	return driver.ErrUnsupportedMethod{}
}

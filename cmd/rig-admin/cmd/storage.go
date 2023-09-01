package cmd

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/uuid"
	storage_service "github.com/rigdev/rig/internal/service/storage"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var recursive bool

func init() {
	storage := &cobra.Command{
		Use: "storage",
	}

	createBucket := &cobra.Command{
		Use:  "create-bucket <name> <provider-name> <region> <provider-id>",
		RunE: register(StorageCreateBucket),
	}
	storage.AddCommand(createBucket)

	getBucket := &cobra.Command{
		Use:  "get-bucket <name>",
		RunE: register(StorageGetBucket),
	}
	storage.AddCommand(getBucket)

	deleteBucket := &cobra.Command{
		Use:  "delete-bucket <name>",
		RunE: register(StorageDeleteBucket),
	}
	storage.AddCommand(deleteBucket)

	listBuckets := &cobra.Command{
		Use:  "list-buckets",
		RunE: register(StorageListBuckets),
	}
	storage.AddCommand(listBuckets)

	getObject := &cobra.Command{
		Use:  "get-object <bucket> <path>",
		RunE: register(StorageGetObject),
	}
	storage.AddCommand(getObject)

	listObject := &cobra.Command{
		Use:  "list-objects <bucket>",
		RunE: register(StorageListObjects),
	}
	listObject.Flags().BoolVarP(&recursive, "recursive", "r", false, "recursive")
	storage.AddCommand(listObject)

	deleteObject := &cobra.Command{
		Use:  "delete-object <bucket> <path>",
		RunE: register(StorageDeleteObject),
	}
	storage.AddCommand(deleteObject)

	copyObject := &cobra.Command{
		Use:  "copy-object <src-bucket> <src-path> <dst-bucket> <dst-path>",
		RunE: register(StorageCopyObject),
	}

	storage.AddCommand(copyObject)

	uploadObject := &cobra.Command{
		Use:  "upload <bucket> <path> <file>",
		RunE: register(StorageUploadFile),
	}
	storage.AddCommand(uploadObject)

	downloadObject := &cobra.Command{
		Use:  "download <bucket> <path> <file>",
		RunE: register(StorageDownloadFile),
	}
	storage.AddCommand(downloadObject)

	rootCmd.AddCommand(storage)
}

func StorageCreateBucket(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	if bucketName == "" {
		return errors.New("name is required")
	}

	providerBucketName := args[1]
	if providerBucketName == "" {
		return errors.New("provider name is required")
	}

	bucketRegion := args[2]
	if bucketName == "" {
		return errors.New("region is required")
	}
	providerID, err := uuid.Parse(args[3])
	if err != nil {
		return err
	}
	err = ss.CreateBucket(ctx, bucketName, providerBucketName, bucketRegion, providerID)
	if err != nil {
		return err
	}
	logger.Info("created bucket", zap.String("name", bucketName))
	return nil
}

func StorageGetBucket(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	if bucketName == "" {
		return errors.New("name is required")
	}
	bucket, err := ss.GetBucket(ctx, bucketName)
	if err != nil {
		return err
	}
	logger.Info("got bucket", zap.String("name", bucket.Name))
	return nil
}

func StorageDeleteBucket(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	if bucketName == "" {
		return errors.New("name is required")
	}
	err := ss.DeleteBucket(ctx, bucketName)
	if err != nil {
		return err
	}
	logger.Info("deleted bucket", zap.String("name", bucketName))
	return nil
}

func StorageListBuckets(ctx context.Context, cmd *cobra.Command, ss *storage_service.Service, logger *zap.Logger) error {
	it, err := ss.ListBuckets(ctx)
	if err != nil {
		return err
	}
	for {
		bucket, err := it.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		logger.Info("got bucket", zap.String("name", bucket.Name))
	}
	return nil
}

func StorageGetObject(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	path := args[1]
	if bucketName == "" {
		return errors.New("name is required")
	}
	if path == "" {
		return errors.New("path is required")
	}
	object, err := ss.GetObject(ctx, bucketName, path)
	if err != nil {
		return err
	}
	logger.Info("got object", zap.String("name", object.GetPath()), zap.Int("size", int(object.GetSize())), zap.String("etag", object.GetEtag()))
	return nil
}

func StorageDeleteObject(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	path := args[1]
	if bucketName == "" {
		return errors.New("name is required")
	}
	if path == "" {
		return errors.New("path is required")
	}
	err := ss.DeleteObject(ctx, bucketName, path)
	if err != nil {
		return err
	}
	logger.Info("deleted object", zap.String("name", path))
	return nil
}

func StorageCopyObject(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	path := args[1]
	dstBucketName := args[2]
	dstObjectpath := args[3]
	if bucketName == "" {
		return errors.New("name is required")
	}
	if path == "" {
		return errors.New("path is required")
	}
	if dstBucketName == "" {
		return errors.New("dstbucket is required")
	}
	if dstObjectpath == "" {
		return errors.New("dstpath is required")
	}
	err := ss.CopyObject(ctx, bucketName, path, dstBucketName, dstObjectpath)
	if err != nil {
		return err
	}
	logger.Info("copied object from" + bucketName + ":" + path + " to " + dstBucketName + ":" + dstObjectpath)
	return nil
}

func StorageListObjects(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]

	if bucketName == "" {
		return errors.New("name is required")
	}
	_, it, err := ss.ListObjects(ctx, bucketName, "", "/", "", "", recursive, 100)
	if err != nil {
		return err
	}
	for {
		object, err := it.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		switch v := object.GetResult().(type) {
		case *storage.ListObjectsResponse_Result_Folder:
			logger.Info("got folder", zap.String("name", v.Folder))
		case *storage.ListObjectsResponse_Result_Object:
			logger.Info("got object", zap.String("name", v.Object.GetPath()), zap.Int("size", int(v.Object.GetSize())), zap.String("etag", v.Object.GetEtag()))
		}
	}
	return nil
}

func StorageUploadFile(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	path := args[1]
	filePath := args[2]

	if bucketName == "" {
		return errors.New("name is required")
	}
	if path == "" {
		return errors.New("path is required")
	}
	if filePath == "" {
		return errors.New("file is required")
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	metadata := &storage.UploadObjectRequest_Metadata{
		Bucket: bucketName,
		Path:   path,
		Size:   uint64(fileStat.Size()),
	}
	_, _, err = ss.UploadObject(ctx, file, metadata)
	if err != nil {
		return err
	}
	logger.Info("uploaded file", zap.String("name", path), zap.Int64("size", fileStat.Size()))
	return nil
}

func StorageDownloadFile(ctx context.Context, cmd *cobra.Command, args []string, ss *storage_service.Service, logger *zap.Logger) error {
	bucketName := args[0]
	path := args[1]
	filePath := args[2]

	if bucketName == "" {
		return errors.New("name is required")
	}
	if path == "" {
		return errors.New("path is required")
	}
	if filePath == "" {
		return errors.New("file is required")
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	object, err := ss.DownloadObject(ctx, bucketName, path)
	if err != nil {
		return err
	}
	written, err := io.Copy(file, object)
	if err != nil {
		return err
	}
	logger.Info("downloaded file", zap.String("name", path), zap.Int64("size", written))
	return nil
}

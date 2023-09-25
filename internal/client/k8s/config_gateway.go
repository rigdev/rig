package k8s

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type configGateway struct {
	logger  *zap.Logger
	restCfg *rest.Config
	cc      client.Client
}

func newConfigGateway(logger *zap.Logger, restCfg *rest.Config, cs *kubernetes.Clientset) *configGateway {
	scheme := runtime.NewScheme()
	sm := runtime.NewSchemeBuilder(v1.AddToScheme, v1alpha1.AddToScheme)
	utilruntime.Must(sm.AddToScheme(scheme))

	cc, err := client.New(restCfg, client.Options{
		Scheme: scheme,
	})
	utilruntime.Must(err)

	return &configGateway{
		logger:  logger,
		restCfg: restCfg,
		cc:      cc,
	}
}

func (g *configGateway) CreateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error {
	if err := g.cc.Create(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: cfg.Namespace}}); err != nil {
		return checkError(err)
	}

	if err := g.cc.Create(ctx, cfg); err != nil {
		return checkError(err)
	}

	return nil
}

func (g *configGateway) UpdateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error {
	if err := g.cc.Update(ctx, cfg); err != nil {
		return checkError(err)
	}

	return nil
}

func (g *configGateway) ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*v1alpha1.Capsule], int64, error) {
	ls, err := getList(ctx, pagination, g.cc, &v1alpha1.CapsuleList{})
	if err != nil {
		return nil, 0, err
	}

	return toIterator(ctx, pagination, ls, ls.Items)
}

func (g *configGateway) GetCapsuleConfig(ctx context.Context, capsuleID string) (*v1alpha1.Capsule, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	res := &v1alpha1.Capsule{}
	if err := g.cc.Get(ctx, client.ObjectKey{Name: capsuleID, Namespace: projectID}, res); err != nil {
		return nil, checkError(err)
	}

	return res, nil
}

func (g *configGateway) DeleteCapsuleConfig(ctx context.Context, capsuleID string) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	g.logger.Debug("delete capsule", zap.String("name", capsuleID), zap.String("namespace", projectID))

	if err := g.cc.Delete(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsuleID,
			Namespace: projectID,
		},
	}); err != nil {
		return checkError(err)
	}

	return nil
}

func (g *configGateway) SetEnvironmentVariables(ctx context.Context, capsuleID string, envs map[string]string) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	envFile := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsuleID,
			Namespace: projectID,
		},
		Data: envs,
	}

	return g.SetFile(ctx, capsuleID, envFile)
}

func (g *configGateway) GetEnvironmentVariables(ctx context.Context, capsuleID string) (map[string]string, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	cf, err := g.GetFile(ctx, capsuleID, capsuleID, projectID)
	if err != nil {
		return nil, err
	}

	return cf.Data, nil
}

func (g *configGateway) SetEnvironmentVariable(ctx context.Context, capsuleID, name, value string) error {
	return errors.UnimplementedErrorf("unimplemented SetEnvironmentVariable")
}

func (g *configGateway) GetEnvironmentVariable(ctx context.Context, capsuleID, name string) (value string, ok bool, err error) {
	return "", false, errors.UnimplementedErrorf("unimplemented GetEnvironmentVariable")
}

func (g *configGateway) DeleteEnvironmentVariable(ctx context.Context, capsuleID, name string) error {
	return errors.UnimplementedErrorf("unimplemented DeleteEnvironmentVariable")
}

func (g *configGateway) GetFile(ctx context.Context, capsuleID, name, namespace string) (*v1.ConfigMap, error) {
	res := &v1.ConfigMap{}
	if err := g.cc.Get(ctx, client.ObjectKey{Name: capsuleID, Namespace: namespace}, res); err != nil {
		return nil, checkError(err)
	}

	return res, nil
}

func (g *configGateway) SetFile(ctx context.Context, capsuleID string, file *v1.ConfigMap) error {
	return g.upsert(ctx, capsuleID, file)
}

func (g *configGateway) ListFiles(ctx context.Context, capsuleID string, pagination *model.Pagination) (iterator.Iterator[*v1.ConfigMap], int64, error) {
	ls, err := getList(ctx, pagination, g.cc, &v1.ConfigMapList{})
	if err != nil {
		return nil, 0, err
	}

	return toIterator(ctx, pagination, ls, ls.Items)
}

func (g *configGateway) DeleteFile(ctx context.Context, capsuleID, name, namespace string) error {
	g.logger.Debug("delete file", zap.String("name", name), zap.String("namespace", namespace))
	if err := g.cc.Delete(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}); err != nil {
		return checkError(err)
	}

	return nil
}

func (g *configGateway) GetSecret(ctx context.Context, capsuleID, name, namespace string) (*v1.Secret, error) {
	res := &v1.Secret{}
	if err := g.cc.Get(ctx, client.ObjectKey{Name: capsuleID, Namespace: namespace}, res); err != nil {
		return nil, checkError(err)
	}

	return res, nil
}

func (g *configGateway) SetSecret(ctx context.Context, capsuleID string, secret *v1.Secret) error {
	return g.upsert(ctx, capsuleID, secret)
}

func (g *configGateway) ListSecrets(ctx context.Context, capsuleID string, pagination *model.Pagination) (iterator.Iterator[*v1.Secret], int64, error) {
	ls, err := getList(ctx, pagination, g.cc, &v1.SecretList{})
	if err != nil {
		return nil, 0, err
	}

	return toIterator(ctx, pagination, ls, ls.Items)
}

func (g *configGateway) DeleteSecret(ctx context.Context, capsuleID, name, namespace string) error {
	g.logger.Debug("delete secret", zap.String("name", name), zap.String("namespace", namespace))
	if err := g.cc.Delete(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}); err != nil {
		return checkError(err)
	}

	return nil
}

func (g *configGateway) upsert(ctx context.Context, capsuleID string, file client.Object) error {
	f := file.DeepCopyObject().(client.Object)
	if err := checkError(g.cc.Get(ctx, client.ObjectKeyFromObject(file), f)); errors.IsNotFound(err) {
		return checkError(g.cc.Create(ctx, file))
	} else if err != nil {
		return err
	}

	return checkError(g.cc.Update(ctx, file))
}

func getList[L client.ObjectList](ctx context.Context, pagination *model.Pagination, cc client.Client, l L) (L, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return l, err
	}

	if err := cc.List(ctx, l, client.InNamespace(projectID)); err != nil {
		return l, checkError(err)
	}
	return l, nil
}

func toIterator[T any](ctx context.Context, pagination *model.Pagination, ol client.ObjectList, ls []T) (iterator.Iterator[*T], int64, error) {
	var ts []*T
	for _, i := range ls {
		v := i
		ts = append(ts, &v)
	}

	var c int64 = int64(len(ls))
	if ol.GetRemainingItemCount() != nil {
		c += *ol.GetRemainingItemCount()
	}

	return iterator.FromList[*T](ts), c, nil
}

func checkError(err error) error {
	switch apierrors.ReasonForError(err) {
	case metav1.StatusReasonNotFound:
		return errors.NotFoundErrorf("%v", err.Error())
	case metav1.StatusReasonAlreadyExists:
		return errors.AlreadyExistsErrorf("%v", err.Error())
	default:
		return err
	}
}

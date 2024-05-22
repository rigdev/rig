package obj

import (
	"slices"

	"github.com/gobwas/glob"
	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewComparison(from, to client.Object, scheme *runtime.Scheme) *Comparison {
	return &Comparison{
		From:   from,
		To:     to,
		scheme: scheme,
	}
}

type Comparison struct {
	From client.Object
	To   client.Object

	scheme *runtime.Scheme

	filters     []FilterFunc
	removeDiffs []string
}

type FilterFunc func(client.Object)

func (c *Comparison) AddFilter(f FilterFunc) {
	c.filters = append(c.filters, f)
}

func (c *Comparison) AddRemoveDiffs(keys ...string) {
	c.removeDiffs = append(c.removeDiffs, keys...)
}

func RemoveAnnotationsFilter(names ...string) FilterFunc {
	return func(co client.Object) {
		annotations := co.GetAnnotations()
		if annotations == nil {
			return
		}

		for _, name := range names {
			delete(annotations, name)
		}

		co.SetAnnotations(annotations)
	}
}

type Diff struct {
	From      client.Object
	To        client.Object
	FromBytes []byte
	ToBytes   []byte
	Report    *dyff.Report
}

func (c *Comparison) ComputeDiff() (*Diff, error) {
	from, err := c.normalize(c.From)
	if err != nil {
		return nil, err
	}

	to, err := c.normalize(c.To)
	if err != nil {
		return nil, err
	}

	if from != nil && to != nil {
		if from.GetObjectKind().GroupVersionKind() != to.GetObjectKind().GroupVersionKind() {
			return nil, errors.InvalidArgumentErrorf("from and to objects must be same type for diff")
		}
	}

	fromBytes, err := Encode(from, c.scheme)
	if err != nil {
		return nil, err
	}

	toBytes, err := Encode(to, c.scheme)
	if err != nil {
		return nil, err
	}

	report, err := compare(fromBytes, toBytes)
	if err != nil {
		return nil, err
	}

	for _, rm := range c.removeDiffs {
		g, err := glob.Compile(rm, '.')
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(report.Diffs); i++ {
			d := report.Diffs[i]
			if d.Path == nil {
				continue
			}

			if g.Match(d.Path.ToDotStyle()) {
				report.Diffs = slices.Delete(report.Diffs, i, i+1)
				i--
			}
		}
	}

	return &Diff{
		From:      from,
		To:        to,
		FromBytes: fromBytes,
		ToBytes:   toBytes,
		Report:    report,
	}, nil
}

func compare(from, to []byte) (*dyff.Report, error) {
	if len(from) == 0 {
		from = []byte("{}")
	}
	if len(to) == 0 {
		to = []byte("{}")
	}

	fromNodes, err := ytbx.LoadYAMLDocuments(from)
	if err != nil {
		return nil, err
	}
	fromFile := ytbx.InputFile{
		Location:  "from",
		Documents: fromNodes,
	}
	toNodes, err := ytbx.LoadYAMLDocuments(to)
	if err != nil {
		return nil, err
	}
	toFile := ytbx.InputFile{
		Location:  "to",
		Documents: toNodes,
	}

	r, err := dyff.CompareInputFiles(fromFile, toFile, dyff.KubernetesEntityDetection(false))
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *Comparison) normalize(co client.Object) (client.Object, error) {
	if co == nil {
		return nil, nil
	}

	co = co.DeepCopyObject().(client.Object)

	if co.GetObjectKind().GroupVersionKind().Empty() {
		gvks, _, err := c.scheme.ObjectKinds(co)
		if err != nil {
			return nil, err
		}

		co.GetObjectKind().SetGroupVersionKind(gvks[0])
	}

	co.SetManagedFields(nil)
	co.SetCreationTimestamp(v1.Time{})
	co.SetGeneration(0)
	co.SetResourceVersion("")
	co.SetOwnerReferences(nil)
	co.SetUID("")

	for _, f := range c.filters {
		f(co)
	}

	return co, nil
}

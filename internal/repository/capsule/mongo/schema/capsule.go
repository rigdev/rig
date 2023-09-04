package schema

import (
	"fmt"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/gen/go/rollout"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

type Capsule struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	CapsuleID uuid.UUID `bson:"capsule_id" json:"capsule_id"`
	Name      string    `bson:"name,omitempty" json:"name,omitempty"`
	Data      []byte    `bson:"data,omitempty" json:"data,omitempty"`
}

type Rollout struct {
	ProjectID   uuid.UUID  `bson:"project_id" json:"project_id"`
	CapsuleID   uuid.UUID  `bson:"capsule_id" json:"capsule_id"`
	RolloutID   uint64     `bson:"rollout_id" json:"rollout_id"`
	Version     uint64     `bson:"version" json:"version"`
	ScheduledAt *time.Time `bson:"scheduled_at,omitempty" json:"scheduled_at,omitempty"`
	Config      []byte     `bson:"config,omitempty" json:"config,omitempty"`
	Status      []byte     `bson:"status,omitempty" json:"status,omitempty"`
}

type CapsuleMetric struct {
	ProjectID  uuid.UUID `bson:"project_id" json:"project_id"`
	Timestamp  time.Time `bson:"timestamp" json:"timestamp"`
	CapsuleID  uuid.UUID `bson:"capsule_id" json:"capsule_id"`
	InstanceID string    `bson:"instance_id" json:"instance_id"`
	Data       []byte    `bson:"data" json:"data"`
}

func (c Capsule) ToProto() (*capsule.Capsule, error) {
	p := &capsule.Capsule{}
	if err := proto.Unmarshal(c.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func CapsuleFromProto(projectID uuid.UUID, p *capsule.Capsule) (Capsule, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return Capsule{}, err
	}

	return Capsule{
		ProjectID: projectID,
		CapsuleID: uuid.UUID(p.GetCapsuleId()),
		Name:      p.GetName(),
		Data:      bs,
	}, nil
}

func (r Rollout) ConfigToProto() (*capsule.RolloutConfig, error) {
	p := &capsule.RolloutConfig{}
	if err := proto.Unmarshal(r.Config, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (r Rollout) StatusToProto() (*rollout.Status, error) {
	p := &rollout.Status{}
	if err := proto.Unmarshal(r.Status, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (r Rollout) ToProto() (*capsule.Rollout, error) {
	c, err := r.ConfigToProto()
	if err != nil {
		return nil, err
	}

	s, err := r.StatusToProto()
	if err != nil {
		return nil, err
	}

	return &capsule.Rollout{
		RolloutId: r.RolloutID,
		Config:    c,
		Status:    s.GetStatus(),
	}, nil
}

func RolloutFromProto(projectID, capsuleID uuid.UUID, rolloutID, version uint64, rc *capsule.RolloutConfig, rs *rollout.Status) (Rollout, error) {
	bsCfg, err := proto.Marshal(rc)
	if err != nil {
		return Rollout{}, err
	}

	bsSta, err := proto.Marshal(rs)
	if err != nil {
		return Rollout{}, err
	}

	r := Rollout{
		ProjectID: projectID,
		CapsuleID: capsuleID,
		RolloutID: rolloutID,
		Version:   version,
		Config:    bsCfg,
		Status:    bsSta,
	}

	if ts := rs.GetScheduledAt().AsTime(); ts.Unix() != 0 {
		r.ScheduledAt = &ts
	}

	return r, nil
}

func RolloutStatusFromProto(version uint64, rs *rollout.Status) (bson.M, error) {
	bs, err := proto.Marshal(rs)
	if err != nil {
		return nil, err
	}

	set := bson.M{
		"version": version + 1,
		"status":  bs,
	}
	unset := bson.M{}

	if ts := rs.GetScheduledAt().AsTime(); ts.Unix() != 0 {
		set["scheduled_at"] = ts
	} else {
		unset["scheduled_at"] = 1
	}

	return bson.M{
		"$set":   set,
		"$unset": unset,
	}, nil
}

func MetricFromProto(projectID uuid.UUID, p *capsule.InstanceMetrics) (CapsuleMetric, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return CapsuleMetric{}, fmt.Errorf("could not marshal metric to proto: %w", err)
	}

	return CapsuleMetric{
		ProjectID:  projectID,
		Timestamp:  p.GetMainContainer().Timestamp.AsTime(),
		CapsuleID:  uuid.UUID(p.GetCapsuleId()),
		InstanceID: p.GetInstanceId(),
		Data:       bs,
	}, nil
}

func (m CapsuleMetric) ToProto() (*capsule.InstanceMetrics, error) {
	var p capsule.InstanceMetrics
	if err := proto.Unmarshal(m.Data, &p); err != nil {
		return nil, fmt.Errorf("could not unmarshal metric data to proto: %w", err)
	}
	return &p, nil
}

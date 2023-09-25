package repository

import (
	"errors"
	"fmt"

	"github.com/rigdev/rig/internal/config"
	capsule_mongo "github.com/rigdev/rig/internal/repository/capsule/mongo"
	cluster_config_mongo "github.com/rigdev/rig/internal/repository/cluster_config/mongo"
	group_mongo "github.com/rigdev/rig/internal/repository/group/mongo"
	project_mongo "github.com/rigdev/rig/internal/repository/project/mongo"
	secret_mongo "github.com/rigdev/rig/internal/repository/secret/mongo"
	service_account_mongo "github.com/rigdev/rig/internal/repository/service_account/mongo"
	session_mongo "github.com/rigdev/rig/internal/repository/session/mongo"
	user_mongo "github.com/rigdev/rig/internal/repository/user/mongo"
	verification_code_mongo "github.com/rigdev/rig/internal/repository/verification_code/mongo"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"repository",
	fx.Provide(
		NewGroup,
		NewProjects,
		NewSession,
		NewUser,
		NewVerificationCode,
		NewServiceAccount,
		NewCapsule,
		NewSecret,
		NewClusterConfig,
	),
)

type params struct {
	fx.In

	Cfg    config.Config
	Logger *zap.Logger

	MongoClient *mongo.Client `optional:"true"`
	Postgres    *bun.DB       `optional:"true"`
}

func NewGroup(p params) (Group, error) {
	s := p.Cfg.Repository.Group.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return group_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("group", s)
	}
}

func NewProjects(p params) (Project, error) {
	s := p.Cfg.Repository.Project.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return project_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("project", s)
	}
}

func NewSession(p params) (Session, error) {
	s := p.Cfg.Repository.Session.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return session_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("session", s)
	}
}

func NewUser(p params) (User, error) {
	s := p.Cfg.Repository.User.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return user_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("user", s)
	}
}

func NewVerificationCode(p params) (VerificationCode, error) {
	s := p.Cfg.Repository.VerificationCode.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return verification_code_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("verification_code", s)
	}
}

func NewServiceAccount(p params) (ServiceAccount, error) {
	s := p.Cfg.Repository.ServiceAccount.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return service_account_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("service_account", s)
	}
}

func NewCapsule(p params) (Capsule, error) {
	s := p.Cfg.Repository.Capsule.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return capsule_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("capsule", s)
	}
}

func NewSecret(p params) (Secret, error) {
	s := p.Cfg.Repository.Secret.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return secret_mongo.NewRepository(p.MongoClient, p.Cfg.Repository.Secret.MongoDB.Key)
	default:
		return nil, errInvalidStore("secret", s)
	}
}

func NewClusterConfig(p params) (ClusterConfig, error) {
	s := p.Cfg.Repository.ClusterConfig.Store
	switch s {
	case storeTypeMongoDB:
		if p.MongoClient == nil {
			return nil, errNoMongoDBClient
		}
		return cluster_config_mongo.NewRepository(p.MongoClient)
	default:
		return nil, errInvalidStore("secret", s)
	}
}

var errNoMongoDBClient = errors.New("no mongo client configured, RIG_CLIENT_MONGO_HOST environment variable missing")

func errInvalidStore(name, actual string) error {
	return fmt.Errorf("invalid %s store '%s'", name, actual)
}

var storeTypeMongoDB = "mongodb"

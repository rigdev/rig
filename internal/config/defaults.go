package config

import "go.uber.org/zap/zapcore"

func newDefault() Config {
	return Config{
		Management: Management{
			Port:      4747,
			PublicURL: "",
			Telemetry: ManagementTelemetry{
				Enabled: true,
			},
		},

		Auth: Auth{
			JWT: AuthJWT{
				Secret:             "",
				CertificateFile:    "",
				CertificateKeyFile: "",
			},
		},

		Client: Client{
			Postgres: ClientPostgres{
				User:     "",
				Password: "",
				Host:     "",
			},
			Mongo: ClientMongo{
				Host: "",
			},
			Minio: ClientMinio{
				Host:   "",
				Secure: false,
			},
			Docker: ClientDocker{
				Host: "",
			},
			Mailjet: ClientMailjet{
				From:      "",
				APIKey:    "",
				SecretKey: "",
			},
		},

		Repository: Repository{
			Storage:        defaultRepositoryStore(),
			Capsule:        defaultRepositoryStore(),
			Database:       defaultRepositoryStore(),
			ServiceAccount: defaultRepositoryStore(),
			Group:          defaultRepositoryStore(),
			Project:        defaultRepositoryStore(),
			Secret: RepositoryStoreSecret{
				Store: defaultRepositoryStore().Store,
				MongoDB: RepositoryStoreSecretMongoDB{
					Key: "",
				},
			},
			Session:          defaultRepositoryStore(),
			User:             defaultRepositoryStore(),
			VerificationCode: defaultRepositoryStore(),
		},

		OAuth: OAuth{
			Google: OAuthClientCredentials{
				ClientID:     "",
				ClientSecret: "",
			},
			Github: OAuthClientCredentials{
				ClientID:     "",
				ClientSecret: "",
			},
			Facebook: OAuthClientCredentials{
				ClientID:     "",
				ClientSecret: "",
			},
		},

		Cluster: Cluster{
			Type: ClusterTypeDocker,
		},

		Email: Email{
			Type: EmailTypeNoEmail,
		},

		Registry: Registry{
			Enabled:  false,
			Port:     5001,
			LogLevel: zapcore.InfoLevel,
		},
	}
}

func defaultRepositoryStore() RepositoryStore {
	return RepositoryStore{Store: "mongodb"}
}

package v1alpha1

import (
	"github.com/rigdev/rig/pkg/ptr"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// OperatorConfig is the Schema for the configs API
// +kubebuilder:object:root=true
type OperatorConfig struct {
	metav1.TypeMeta `json:",inline"`

	// WebhooksEnabled set wether or not webhooks should be enabled. When
	// enabled a certificate should be mounted at the webhook server
	// certificate path. Defaults to true if omitted.
	WebhooksEnabled *bool `json:"webhooksEnabled,omitempty"`

	// DevModeEnabled enables verbose logs and changes the logging format to be
	// more human readable.
	DevModeEnabled bool `json:"devModeEnabled,omitempty"`

	// LeaderElectionEnabled enables leader election when running multiple
	// instances of the operator.
	LeaderElectionEnabled *bool `json:"leaderElectionEnabled,omitempty"`

	// Certmanager holds configuration for how the operator should create
	// certificates for ingress resources.
	Certmanager *CertManagerConfig `json:"certManager,omitempty"`

	// Ingress holds the configuration for ingress resources created by the
	// operator.
	Ingress IngressConfig `json:"ingress,omitempty"`
}

type CertManagerConfig struct {
	// ClusterIssuer to use for issueing ingress certificates
	ClusterIssuer string `json:"clusterIssuer"`

	// CreateCertificateResources specifies wether to create Certificate
	// resources. If this is not enabled we will use ingress annotations. This
	// is handy in environments where the ingress-shim isen't enabled.
	CreateCertificateResources bool `json:"createCertificateResources,omitempty"`
}

type IngressConfig struct {
	// Annotations for all ingress resources created.
	Annotations map[string]string `json:"annotations"`

	// ClassName specifies the default ingress class to use for all ingress
	// resources created.
	ClassName string `json:"className"`
}

func (c *OperatorConfig) Default() {
	if c.WebhooksEnabled == nil {
		c.WebhooksEnabled = ptr.New(true)
	}
	if c.LeaderElectionEnabled == nil {
		c.LeaderElectionEnabled = ptr.New(true)
	}
	if c.Ingress.Annotations == nil {
		c.Ingress.Annotations = map[string]string{}
	}
}

func init() {
	SchemeBuilder.Register(&OperatorConfig{})
	SchemeBuilder.Register(&PlatformConfig{})
}

// OperatorConfig is the Schema for the configs API
// +kubebuilder:object:root=true
type PlatformConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Port sets the port the platform should listen on
	Port int `json:"port,omitempty"`

	// PublicUrl sets the public url for the platform. This is used for
	// generating urls for the platform when using oauth2.
	PublicURL string `json:"publicUrl,omitempty"`

	TelemetryEnabled bool `json:"telemetryEnabled,omitempty"`

	Auth Auth `json:"auth,omitempty"`

	// Client holds configuration for clients used in the platform.
	Client Client `json:"client,omitempty"`

	// Repository specifies the type of db to use along with secret key
	Repository Repository `json:"repository,omitempty"`

	// OAuth holds configuration for oauth2 clients, namely google, github and facebook.
	OAuth OAuth `json:"oauth,omitempty"`

	Cluster Cluster `json:"cluster,omitempty"`

	// Email holds configuration for sending emails. Either using mailjet or using SMTP
	Email Email `json:"email,omitempty"`

	// Loggin holds information about the granularity of logging
	Logging Logging `json:"logging,omitempty"`
}

type Auth struct {
	Secret             string `json:"secret,omitempty"`
	CertificateFile    string `json:"certificateFile,omitempty"`
	CertificateKeyFile string `json:"certificateKeyFile,omitempty"`
}

type Client struct {
	// Postgres holds configuration for the postgres client.
	Postgres ClientPostgres `json:"postgres,omitempty"`

	// Mongo holds configuration for the Mongo client.
	Mongo ClientMongo `json:"mongo,omitempty"`

	// Docker sets the host for the Docker client.
	Docker ClientDocker `json:"docker,omitempty"`

	// Mailjet sets the API key and secret for the Mailjet client.
	Mailjet ClientMailjet `json:"mailjet,omitempty"`

	// SMTP sets the host, port, username and password for the SMTP client.
	SMTP ClientSMTP `json:"smtp,omitempty"`

	// Operator sets the base url for the Operator client.
	Operator ClientOperator `json:"operator,omitempty"`
}

type Logging struct {
	// DevModeEnabled enables verbose logs and changes the logging format to be
	// more human readable.
	DevMode bool `json:"devMode,omitempty"`

	// Level sets the granularity of logging
	Level zapcore.Level `json:"level,omitempty"`
}

type ClientPostgres struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`

	// Database in the postgres server to use
	Database string `json:"database,omitempty"`

	// Use SSL when connecting to the postgres server
	Insecure bool `json:"insecure,omitempty"`
}

type ClientMongo struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	// Host of the mongo server. This is both the host and port.
	Host string `json:"host,omitempty"`
}

type ClientDocker struct {
	Host string `json:"host,omitempty"`
}

type ClientMailjet struct {
	APIKey    string `json:"apiKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

type ClientSMTP struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ClientOperator struct {
	BaseURL string `json:"baseUrl,omitempty"`
}

type Repository struct {
	// Type of db to use
	Store  string `json:"store,omitempty"`
	Secret string `json:"secret,omitempty"`
}

type OAuth struct {
	Google   OAuthClientCredentials `json:"google,omitempty"`
	Github   OAuthClientCredentials `json:"github,omitempty"`
	Facebook OAuthClientCredentials `json:"facebook,omitempty"`
}

type OAuthClientCredentials struct {
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

type Cluster struct {
	// Type of the cluster - either docker or k8s
	Type ClusterType `json:"type,omitempty"`

	DevRegistry DevRegistry `json:"devRegistry,omitempty"`
	Git         ClusterGit  `json:"git,omitempty"`
}

type ClusterGit struct {
	URL         string         `json:"url,omitempty"`
	Branch      string         `json:"branch,omitempty"`
	PathPrefix  string         `json:"pathPrefix,omitempty"`
	Credentials GitCredentials `json:"credentials,omitempty"`
	Author      GitAuthor      `json:"author,omitempty"`
	Templates   GitTemplates   `json:"templates,omitempty"`
}

type GitCredentials struct {
	HTTPS HTTPSCredential `json:"https,omitempty"`
	SSH   SSHCredential   `json:"ssh,omitempty"`
}

type HTTPSCredential struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type SSHCredential struct {
	PrivateKey         string `json:"privateKey,omitempty"`
	PrivateKeyPassword string `json:"password,omitempty"`
}

type GitAuthor struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type GitTemplates struct {
	Rollout string `json:"rollout,omitempty"`
	Delete  string `json:"delete,omitempty"`
}

type DevRegistry struct {
	Host        string `json:"host,omitempty"`
	ClusterHost string `json:"clusterHost,omitempty"`
}

type ClusterType string

const (
	ClusterTypeDocker     ClusterType = "docker"
	ClusterTypeKubernetes ClusterType = "k8s"
)

type Email struct {
	From string `json:"from,omitempty"`
	Type string `json:"type,omitempty"`
}

type EmailType string

const (
	EmailTypeNoEmail = ""
	EmailTypeMailjet = "mailjet"
	EmailTypeSMTP    = "smtp"
)

func NewDefaultPlatform() *PlatformConfig {
	cfg := &PlatformConfig{
		Port:      4747,
		PublicURL: "",
		Logging: Logging{
			DevMode: false,
			Level:   zapcore.InfoLevel,
		},
		TelemetryEnabled: true,
		Auth: Auth{
			Secret:             "",
			CertificateFile:    "",
			CertificateKeyFile: "",
		},
		Client: Client{
			Postgres: ClientPostgres{
				User:     "",
				Password: "",
				Host:     "",
				Port:     5432,
				Database: "rig",
				Insecure: false,
			},
			Mongo: ClientMongo{
				Host: "",
			},
			Docker: ClientDocker{
				Host: "",
			},
			Mailjet: ClientMailjet{
				APIKey:    "",
				SecretKey: "",
			},
			Operator: ClientOperator{
				BaseURL: "",
			},
		},
		Repository: Repository{
			Store:  "postgres",
			Secret: "",
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
			Git: ClusterGit{
				Branch:     "main",
				PathPrefix: `apps/{{ .Project.Name }}/{{ .Capsule.Name }}/`,
				Templates: GitTemplates{
					Rollout: `Rig Platform rollout #{{ .Rollout.ID }} of {{ .Capsule.Name }}

Rollout initiated by {{ .Initiator.Name }} at {{ .Rollout.CreatedAt }}.
`,
					Delete: `Rig Platform delete of {{ .Capsule.Name }}

Capsule deleted by {{ .Initiator.Name }}.
`,
				},
				Author: GitAuthor{
					Name:  "rig-platform-change-roller",
					Email: "roll@rig.dev",
				},
			},
		},
		Email: Email{
			From: "",
			Type: EmailTypeNoEmail,
		},
	}

	cfg.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "config.rig.dev",
		Version: "v1alpha1",
		Kind:    "PlatformConfig",
	})

	return cfg
}

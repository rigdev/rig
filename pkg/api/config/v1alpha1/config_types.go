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

	// PrometheusServiceMonitor defines if Rig should spawn a Prometheus ServiceMonitor per capsule
	// for use with a Prometheus Operator stack.
	PrometheusServiceMonitor *PrometheusServiceMonitor `json:"prometheusServiceMonitor,omitempty"`
}

// See https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.ServiceMonitor
// They are used as part of the Prometheus operator stack, see https://github.com/prometheus-operator/prometheus-operator

type PrometheusServiceMonitor struct {
	// The path which Prometheus should query on ports. Defaults to /metrics if not set.
	Path string `json:"path,omitempty"`
	// The port which Prometheus will query metrics on
	Port int `json:"port"`
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

	// TelemetryEnabled specifies wether or not we are allowed to collect usage
	// data. Defaults to true.
	TelemetryEnabled bool `json:"telemetryEnabled,omitempty"`

	// Auth holds authentication configuration.
	Auth Auth `json:"auth,omitempty"`

	// Client holds configuration for clients used in the platform.
	Client Client `json:"client,omitempty"`

	// Repository specifies the type of db to use along with secret key
	Repository Repository `json:"repository,omitempty"`

	// OAuth holds configuration for oauth2 clients, namely google, github and facebook.
	OAuth OAuth `json:"oauth,omitempty"`

	// Cluster holds cluster specific configuration
	Cluster Cluster `json:"cluster,omitempty"`

	// Email holds configuration for sending emails. Either using mailjet or using SMTP
	Email Email `json:"email,omitempty"`

	// Loggin holds information about the granularity of logging
	Logging Logging `json:"logging,omitempty"`
}

// Auth specifies authentication configuration.
type Auth struct {
	// Secret specifies a secret which will be used for jwt signatures.
	Secret string `json:"secret,omitempty"`

	// CertificateFile specifies a path to a PEM encoded certificate file which
	// will be used for validating jwt signatures.
	CertificateFile string `json:"certificateFile,omitempty"`

	// CertificateKeyFile specifies a path to a PEM encoded certificate key
	// which will be used for jwt signatures.
	CertificateKeyFile string `json:"certificateKeyFile,omitempty"`
}

// Client holds various client configuration
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

// Logging specifies logging configuration.
type Logging struct {
	// DevModeEnabled enables verbose logs and changes the logging format to be
	// more human readable.
	DevMode bool `json:"devMode,omitempty"`

	// Level sets the granularity of logging.
	Level zapcore.Level `json:"level,omitempty"`
}

// ClientPostgres specifies the configuration for the postgres client.
type ClientPostgres struct {
	// User is the database user used when connecting to the postgres database.
	User string `json:"user,omitempty"`

	// Password is the password used when connecting to the postgres database.
	Password string `json:"password,omitempty"`

	// Host is the host where the postgres database can be reached.
	Host string `json:"host,omitempty"`

	// Port is the port of the postgres database server.
	Port int `json:"port,omitempty"`

	// Database in the postgres server to use
	Database string `json:"database,omitempty"`

	// Insecure is wether to use SSL when connecting to the postgres server
	Insecure bool `json:"insecure,omitempty"`
}

// ClientMongo specifies the configuration for the mongo client.
type ClientMongo struct {
	// User is the database user used when connecting to the mongodb server.
	User string `json:"user,omitempty"`

	// Password is used when connecting to the mongodb server.
	Password string `json:"password,omitempty"`

	// Host of the mongo server. This is both the host and port.
	Host string `json:"host,omitempty"`
}

// ClientDocker specifies the configuration for the docker client.
type ClientDocker struct {
	// Host where the docker daemon can be reached.
	Host string `json:"host,omitempty"`
}

// ClientMailjet specifes the configuration for the mailjet client.
type ClientMailjet struct {
	// APIKey is the mailjet API key
	APIKey string `json:"apiKey,omitempty"`

	// SecretKey is the mailjet secret key
	SecretKey string `json:"secretKey,omitempty"`
}

// ClientSMTP specifies the configuration for the SMTP client.
type ClientSMTP struct {
	// Host is the SMTP server host.
	Host string `json:"host,omitempty"`

	// Port is the SMTP server port to use.
	Port int `json:"port,omitempty"`

	// Username used when connecting to the SMTP server.
	Username string `json:"username,omitempty"`

	// Password used when connecting to the SMTP server.
	Password string `json:"password,omitempty"`
}

// ClientOperator specifies the configuration for the operator client.
type ClientOperator struct {
	// BaseURL is the URL used to connect to the operator API
	BaseURL string `json:"baseUrl,omitempty"`
}

// Repository specifies repository configuration
type Repository struct {
	// Store is what database will be used can be either postgres or mongodb.
	Store string `json:"store,omitempty"`

	// Secret is a secret key used for encrypting sensitive data before saving
	// it in the database.
	Secret string `json:"secret,omitempty"`
}

// OAuth specifies configuration for different OAuth providers.
type OAuth struct {
	// Google specifies OAuth client configuration for google.
	Google OAuthClientCredentials `json:"google,omitempty"`

	// Github specifies OAuth client configuration for github.
	Github OAuthClientCredentials `json:"github,omitempty"`

	// Facebook specifies OAuth client configuration for facebook.
	Facebook OAuthClientCredentials `json:"facebook,omitempty"`
}

// OAuthClientCredentials specifies a set of OAuth client credentials.
type OAuthClientCredentials struct {
	// ClientID is the OAuth client ID.
	ClientID string `json:"clientId,omitempty"`

	// ClientSecret is the OAuth client secret.
	ClientSecret string `json:"clientSecret,omitempty"`
}

// Cluster specifies cluster configuration
type Cluster struct {
	// Type of the cluster - either `docker` or `k8s`.
	Type ClusterType `json:"type,omitempty"`

	// DevRegistry configuration
	DevRegistry DevRegistry `json:"devRegistry,omitempty"`

	// Git sets up gitops write back for this cluster.
	Git ClusterGit `json:"git,omitempty"`
}

// ClusterGit specifies configuration for git integration. This can be used to
// tie rig into a gitops setup.
type ClusterGit struct {
	// URL is the git repository URL.
	URL string `json:"url,omitempty"`

	// Branch to commit changes to.
	Branch string `json:"branch,omitempty"`

	// PathPrefix path to commit to in git repository.
	PathPrefix string `json:"pathPrefix,omitempty"`

	// Credentials to use when connecting to git.
	Credentials GitCredentials `json:"credentials,omitempty"`

	// Author used when creating commits.
	Author GitAuthor `json:"author,omitempty"`

	// Templates used for commit messages.
	Templates GitTemplates `json:"templates,omitempty"`
}

// GitCredentials specifies how to authenticate against git.
type GitCredentials struct {
	// HTTPS specifies basic auth credentials.
	HTTPS HTTPSCredential `json:"https,omitempty"`

	// SSH specifies SSH credentials.
	SSH SSHCredential `json:"ssh,omitempty"`
}

// HTTPSCredential specifies basic auth credentials
type HTTPSCredential struct {
	// Username is the basic auth user name
	Username string `json:"username,omitempty"`

	// Password is the basic auth password
	Password string `json:"password,omitempty"`
}

// SSHCredential specifies SSH credentials
type SSHCredential struct {
	// PrivateKey is a PEM encoded SSH private key.
	PrivateKey string `json:"privateKey,omitempty"`

	// PrivateKeyPassword is an optional password for the SSH private key.
	PrivateKeyPassword string `json:"password,omitempty"`
}

// GitAuthor specifies a git commit author
type GitAuthor struct {
	// Name of author
	Name string `json:"name,omitempty"`

	// Email of author
	Email string `json:"email,omitempty"`
}

// GitTemplates specifies the templates used for creating commits.
type GitTemplates struct {
	// Rollout specifies the template used for rollout commits.
	Rollout string `json:"rollout,omitempty"`

	// Delete specifies the template used for delete commits.
	Delete string `json:"delete,omitempty"`
}

// DevRegistry specifies configuration for the dev registry support.
type DevRegistry struct {
	// Host is the host used in image names when pushing to the registry from
	// outside of the cluster.
	Host string `json:"host,omitempty"`

	// ClusterHost is the host where the registry can be reached from within
	// the cluster. Any image which is named after `Host` will be rename to use
	// `ClusterHost` instead. This ensures that the image can be pulled from
	// within the cluster.
	ClusterHost string `json:"clusterHost,omitempty"`
}

// ClusterType is a cluster type.
type ClusterType string

const (
	// ClusterTypeDocker is the docker cluster type.
	ClusterTypeDocker ClusterType = "docker"
	// ClusterTypeKubernetes is the kubernetes cluster type.
	ClusterTypeKubernetes ClusterType = "k8s"
)

// Email holds configuration for sending emails. Either using mailjet or using SMTP
type Email struct {
	// From is who is set as the sender of rig emails.
	From string `json:"from,omitempty"`

	// Type is what client rig should use to send emails.
	Type EmailType `json:"type,omitempty"`
}

// EmailType represents a type of mailing provider
type EmailType string

const (
	// EmailTypeNoEmail disables mail sending.
	EmailTypeNoEmail = ""
	// EmailTypeMailjet uses the mailjet API for sending emails.
	EmailTypeMailjet = "mailjet"
	// EmailTypeSMTP uses regular SMTP for sending emails.
	EmailTypeSMTP = "smtp"
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

package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/oauth2"
	"github.com/rigdev/rig/internal/repository"
	auth_service "github.com/rigdev/rig/internal/service/auth"
	proj_serv "github.com/rigdev/rig/internal/service/project"
	user_serv "github.com/rigdev/rig/internal/service/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/service"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"k8s.io/utils/strings/slices"
)

type Handler struct {
	cfg    config.Config
	oauth2 *oauth2.Providers
	ps     proj_serv.Service
	us     user_serv.Service
	auth   *auth_service.Service
	rs     repository.Secret
	logger *zap.Logger
}

func New(
	cfg config.Config,
	oauth2 *oauth2.Providers,
	us user_serv.Service,
	ps proj_serv.Service,
	auth *auth_service.Service,
	logger *zap.Logger,
	rs repository.Secret,
) *Handler {
	return &Handler{
		cfg:    cfg,
		oauth2: oauth2,
		ps:     ps,
		us:     us,
		auth:   auth,
		logger: logger,
		rs:     rs,
	}
}

func (h *Handler) ServiceName() string {
	return ""
}

func (h *Handler) Build() (string, string, service.HandlerFunc) {
	return http.MethodGet, "/oauth/callback", func(w http.ResponseWriter, r *http.Request) error {
		state, code, err := initOauth2Callback(r)
		if err != nil {
			return err
		}

		ctx := auth.WithProjectID(r.Context(), state.ProjectId)

		us, err := h.us.GetSettings(ctx)
		if err != nil {
			return err
		}

		var oidcprovider oauth2.Provider
		var provider *settings.OauthProviderSettings
		h.logger.Info("state.ProviderType", zap.String("state.ProviderType", state.ProviderType.String()))
		switch state.ProviderType {
		case model.OauthProvider_OAUTH_PROVIDER_GOOGLE:
			if !us.GetOauthSettings().GetGoogle().GetAllowLogin() {
				return errors.FailedPreconditionErrorf("google login not allowed")
			}
			oidcprovider = h.oauth2.Google
			provider = us.GetOauthSettings().GetGoogle()
		case model.OauthProvider_OAUTH_PROVIDER_GITHUB:
			if !us.GetOauthSettings().GetGithub().GetAllowLogin() {
				return errors.FailedPreconditionErrorf("github login not allowed")
			}
			oidcprovider = h.oauth2.Github
			provider = us.GetOauthSettings().GetGithub()
		case model.OauthProvider_OAUTH_PROVIDER_FACEBOOK:
			if !us.GetOauthSettings().GetFacebook().GetAllowLogin() {
				return errors.FailedPreconditionErrorf("facebook login not allowed")
			}
			oidcprovider = h.oauth2.Facebook
			provider = us.GetOauthSettings().GetFacebook()
		default:
			return errors.InvalidArgumentErrorf("invalid oauth provider type")
		}

		// check if state.AppRedirect is in user.UserSettings.OauthProviders.CallbackUrls
		if !slices.Contains(us.GetOauthSettings().GetCallbackUrls(), state.AppRedirect) {
			return errors.NotFoundErrorf("Invalid callback url")
		}

		h.logger.Debug("Getting Credentials")
		sid, err := uuid.Parse(provider.GetSecretId())
		if err != nil {
			return err
		}

		creds := model.ProviderCredentials{}
		bytes, err := h.rs.Get(ctx, sid)
		if err != nil {
			return err
		}

		if err := proto.Unmarshal(bytes, &creds); err != nil {
			return err
		}

		callBackURL, err := url.JoinPath(h.cfg.PublicURL, "oauth/callback")
		if err != nil {
			return fmt.Errorf("could not build callback URL: %w", err)
		}

		claims, err := oidcprovider.Validate(ctx, &creds, code, callBackURL)
		if err != nil {
			return err
		}

		token, err := h.auth.LoginOauth2(ctx, claims, provider)
		if err != nil {
			return err
		}

		values := url.Values{}
		values.Add("access_token", token.GetAccessToken())
		values.Add("refresh_token", token.GetRefreshToken())

		url, err := url.Parse(state.AppRedirect)
		if err != nil {
			return err
		}

		url.RawQuery = values.Encode()
		if url.Scheme == "" {
			url.Scheme = "https"
		}
		http.Redirect(w, r, url.String(), http.StatusFound)
		return nil
	}
}

func initOauth2Callback(r *http.Request) (*auth.State, string, error) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		return nil, "", errors.InvalidArgumentErrorf("invalid oauth2 state or code")
	}
	// unmarshall state into bytearray
	stateBytes, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		return nil, "", err
	}
	// unmarshall state into state struct
	var stateStruct auth.State
	if err := json.Unmarshal(stateBytes, &stateStruct); err != nil {
		return nil, "", err
	}
	return &stateStruct, code, nil
}

package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/crypto"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var googleScopes = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"openid",
	"https://www.googleapis.com/auth/userinfo.profile",
}

type Provider struct {
	Provider   *oidc.Provider
	Endpoint   oauth2.Endpoint
	RequireSSL bool
	Logger     *zap.Logger
}

type Claims struct {
	Id        string `json:"sub"`
	Issuer    string `json:"iss"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Image     string `json:"picture"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
}

func (g *Provider) Validate(ctx context.Context, creds *model.ProviderCredentials, code, redirectUrl string) (*auth.Oauth2Claims, error) {
	if creds.GetPublicKey() == "" {
		return nil, errors.New("missing required oidcProvider id")
	} else if redirectUrl == "" {
		return nil, errors.New("missing required redirect url")
	} else if code == "" {
		return nil, errors.New("missing required code")
	}
	oauthConfig := &oauth2.Config{
		ClientID:     creds.GetPublicKey(),
		RedirectURL:  redirectUrl,
		ClientSecret: creds.GetPrivateKey(),
		Scopes:       googleScopes,
		Endpoint:     g.Endpoint,
	}
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok || rawIdToken == "" {
		return nil, errors.New("missing required id token")
	}
	idToken, err := g.Provider.Verifier(&oidc.Config{ClientID: creds.GetPublicKey()}).Verify(ctx, rawIdToken)
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}
	if claims.Id == "" {
		return nil, errors.New("google-oauth2: missing required id")
	}
	if claims.Issuer == "" {
		claims.Issuer = "google"
	}
	oauthClaims := &auth.Oauth2Claims{
		Sub:   claims.Id,
		Iss:   claims.Issuer,
		Email: claims.Email,
		Image: claims.Image,
	}
	if claims.FirstName == "" || claims.LastName == "" {
		splitName := strings.Split(claims.Name, " ")
		if len(splitName) > 1 {
			oauthClaims.FirstName = splitName[0]
			oauthClaims.LastName = splitName[1]
		} else {
			oauthClaims.FirstName = splitName[0]
		}
	} else {
		oauthClaims.FirstName = claims.FirstName
		oauthClaims.LastName = claims.LastName
	}

	return oauthClaims, nil
}

func (g *Provider) Test(ctx context.Context, privateKey, publicKey, redirectUrl string) error {
	return errors.New("not implemented")
}

func (g *Provider) RedirectUrl(redirectUrl, appRedirect string, projectId uuid.UUID, creds *model.ProviderCredentials) (string, error) {
	if creds == nil {
		return "", errors.New("missing required provider credentials")
	} else if redirectUrl == "" {
		return "", errors.New("missing required redirect url")
	} else if creds.GetPublicKey() == "" {
		return "", errors.New("missing required public key")
	} else if creds.GetPrivateKey() == "" {
		return "", errors.New("missing required private key")
	}
	// values := url.Values{}
	// values.Add("grant_type", "authorization_code")
	oauthConfig := &oauth2.Config{
		RedirectURL:  redirectUrl,
		ClientID:     creds.GetPublicKey(),
		ClientSecret: creds.GetPrivateKey(),
		Scopes:       googleScopes,
		Endpoint:     g.Endpoint,
	}
	rand, err := crypto.GenerateSymmetricKey(15, crypto.AlphaNum)
	if err != nil {
		return "", err
	}
	state := &auth.State{
		Rand:         rand,
		ProjectId:    projectId,
		AppRedirect:  appRedirect,
		ProviderType: model.OauthProvider_OAUTH_PROVIDER_GOOGLE,
	}
	jsonState, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	base64State := base64.StdEncoding.EncodeToString(jsonState)
	if err != nil {
		return "", err
	}
	authUrl := oauthConfig.AuthCodeURL(base64State)
	return authUrl, nil
}

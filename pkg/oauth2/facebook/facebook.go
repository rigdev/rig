package facebook

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
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var facebookScopes = []string{
	"email",
	"public_profile",
	"openid",
}

type Provider struct {
	Provider *oidc.Provider
	Endpoint oauth2.Endpoint
	Logger   *zap.Logger
}

type Claims struct {
	Id        string `json:"sub"`
	Issuer    string `json:"iss"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Image     string `json:"picture"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Verified  bool   `json:"verified"`
}

func (f *Provider) Validate(ctx context.Context, creds *model.ProviderCredentials, code, redirectUrl string) (*auth.Oauth2Claims, error) {
	if creds.GetPublicKey() == "" {
		return nil, errors.New("missing required provider public key")
	} else if redirectUrl == "" {
		return nil, errors.New("missing required redirect url")
	} else if code == "" {
		return nil, errors.New("missing required code")
	}
	oauthConfig := &oauth2.Config{
		ClientID:     creds.GetPublicKey(),
		RedirectURL:  redirectUrl,
		ClientSecret: creds.GetPrivateKey(),
		Scopes:       facebookScopes,
		Endpoint:     f.Endpoint,
	}
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok || rawIdToken == "" {
		return nil, errors.New("missing required id token")
	}
	idToken, err := f.Provider.Verifier(&oidc.Config{ClientID: creds.GetPublicKey()}).Verify(ctx, rawIdToken)
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}
	if claims.Id == "" {
		return nil, errors.New("facebook-oauth2: missing required id")
	}

	oauthClaims := &auth.Oauth2Claims{
		Sub:   claims.Id,
		Iss:   claims.Issuer,
		Email: claims.Email,
		Image: claims.Image,
	}
	splitName := strings.Split(claims.Name, " ")
	if len(splitName) > 1 {
		oauthClaims.FirstName = splitName[0]
		oauthClaims.LastName = splitName[1]
	} else {
		oauthClaims.FirstName = splitName[0]
	}
	return oauthClaims, nil
}

func (f *Provider) Test(ctx context.Context, privateKey, publicKey, redirectUrl string) error {
	return errors.New("not implemented")
}

func (f *Provider) RedirectUrl(redirectUrl, appRedirect string, projectID string, creds *model.ProviderCredentials) (string, error) {
	if creds == nil {
		return "", errors.New("missing required credentials")
	}
	if redirectUrl == "" {
		return "", errors.New("missing required redirect url")
	}
	oauthConfig := &oauth2.Config{
		RedirectURL:  redirectUrl,
		ClientID:     creds.GetPublicKey(),
		ClientSecret: creds.GetPrivateKey(),
		Scopes:       facebookScopes,
		Endpoint:     f.Endpoint,
	}
	rand, err := crypto.GenerateSymmetricKey(15, crypto.AlphaNum)
	if err != nil {
		return "", err
	}
	state := &auth.State{
		Rand:         rand,
		ProjectId:    projectID,
		AppRedirect:  appRedirect,
		ProviderType: model.OauthProvider_OAUTH_PROVIDER_FACEBOOK,
	}
	jsonState, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	base64State := base64.StdEncoding.EncodeToString(jsonState)
	if err != nil {
		return "", err
	}
	return oauthConfig.AuthCodeURL(base64State), nil
}

package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/crypto"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var githubScopes = []string{"id-token, read:user"}

type Provider struct {
	Provider *oidc.Provider
	Endpoint oauth2.Endpoint
	Logger   *zap.Logger
}

type Claims struct {
	Id     int    `json:"id"`
	Issuer string `json:"iss"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Image  string `json:"avatar_url"`
}

func (g *Provider) Validate(ctx context.Context, creds *model.ProviderCredentials, code, redirectUrl string) (*auth.Oauth2Claims, error) {
	if creds.GetPublicKey() == "" {
		return nil, errors.New("missing required provider credentials id")
	} else if redirectUrl == "" {
		return nil, errors.New("missing required redirect url")
	} else if code == "" {
		return nil, errors.New("missing required code")
	}
	oauthConfig := &oauth2.Config{
		ClientID:     creds.GetPublicKey(),
		RedirectURL:  redirectUrl,
		ClientSecret: creds.GetPrivateKey(),
		Scopes:       githubScopes,
		Endpoint:     g.Endpoint,
	}
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	// GitHub does not support OpenID Connect -> we need to exchange the access token for a user
	rawAccessToken, ok := token.Extra("access_token").(string)
	if !ok || rawAccessToken == "" {
		return nil, errors.New("missing required access token")
	}
	bearer := "Bearer " + rawAccessToken
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var claims Claims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, err
	}

	if claims.Id == 0 {
		return nil, errors.New("missing required id")
	}
	if claims.Issuer == "" {
		claims.Issuer = "github"
	}

	oauthClaims := &auth.Oauth2Claims{
		Sub:   strconv.Itoa(claims.Id),
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

func (g *Provider) Test(ctx context.Context, privateKey, publicKey, redirectUrl string) error {
	return errors.New("not implemented")
}

func (g *Provider) RedirectUrl(redirectUrl, appRedirect string, projectID string, creds *model.ProviderCredentials) (string, error) {
	if creds == nil {
		return "", errors.New("missing required provider credentials")
	} else if redirectUrl == "" {
		return "", errors.New("missing required redirect url")
	}
	oauthConfig := &oauth2.Config{
		RedirectURL:  redirectUrl,
		ClientID:     creds.GetPublicKey(),
		ClientSecret: creds.GetPrivateKey(),
		Scopes:       githubScopes,
		Endpoint:     g.Endpoint,
	}

	rand, err := crypto.GenerateSymmetricKey(15, crypto.AlphaNum)
	if err != nil {
		return "", err
	}
	state := &auth.State{
		Rand:         rand,
		ProjectId:    projectID,
		AppRedirect:  appRedirect,
		ProviderType: model.OauthProvider_OAUTH_PROVIDER_GITHUB,
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

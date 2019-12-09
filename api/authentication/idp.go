package authentication

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	Provider *oidc.Provider
	Config   oauth2.Config
	Ctx      context.Context
}

func NewAuthenticator() (*Authenticator, error) {
	if os.Getenv("IDENTITY_PROVIDER_URL") == "" {
		return nil, errors.New("missing identity provider url configuration")
	}
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, os.Getenv("IDENTITY_PROVIDER_URL"))
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     os.Getenv("IDENTITY_PROVIDER_CLIENT_ID"),
		ClientSecret: os.Getenv("IDENTITY_PROVIDER_SECRET"),
		RedirectURL:  os.Getenv("IDENTITY_PROVIDER_CALLBACK"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
		Ctx:      ctx,
	}, nil
}

func VerifyIdentity(address string) (map[string]interface{}, error) {
	decodedAddress, err := url.QueryUnescape(address)
	if err != nil {
		return nil, err
	}
	urlAddress, err := url.Parse(strings.TrimSpace(decodedAddress))
	if err != nil {
		return nil, err
	}
	authenticator, err := NewAuthenticator()
	if err != nil {
		return nil, err
	}
	token, err := authenticator.Config.Exchange(context.TODO(), urlAddress.Query().Get("code"))
	if err != nil {
		return nil, err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}
	oidcConfig := &oidc.Config{
		ClientID: os.Getenv("IDENTITY_PROVIDER_CLIENT_ID"),
	}

	idToken, err := authenticator.Provider.Verifier(oidcConfig).Verify(context.TODO(), rawIDToken)
	if err != nil {
		return nil, errors.New("failed to verify ID Token")
	}
	// Getting now the userInfo
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		return nil, err
	}
	return profile, nil
}

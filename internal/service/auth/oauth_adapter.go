package auth

import (
	"errors"
	"fmt"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

type oauthIdentity struct {
	ExternalID string
	Issuer     string
	AuthType   string
}

type oauthProviderAdapter interface {
	ResolveIdentity(
		setting *settingModel.OAuth2Setting,
		code string,
		oauthState *authModel.OAuthState,
	) (*oauthIdentity, error)
}

type (
	githubOAuthAdapter struct{}
	googleOAuthAdapter struct{}
	qqOAuthAdapter     struct{}
	customOAuthAdapter struct{}
)

func getOAuthProviderAdapter(provider string) (oauthProviderAdapter, error) {
	switch provider {
	case string(commonModel.OAuth2GITHUB):
		return &githubOAuthAdapter{}, nil
	case string(commonModel.OAuth2GOOGLE):
		return &googleOAuthAdapter{}, nil
	case string(commonModel.OAuth2QQ):
		return &qqOAuthAdapter{}, nil
	case string(commonModel.OAuth2CUSTOM):
		return &customOAuthAdapter{}, nil
	default:
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
}

func (a *githubOAuthAdapter) ResolveIdentity(
	setting *settingModel.OAuth2Setting,
	code string,
	_ *authModel.OAuthState,
) (*oauthIdentity, error) {
	tokenResp, err := exchangeGithubCodeForToken(setting, code)
	if err != nil {
		return nil, err
	}
	userInfo, err := fetchGitHubUserInfo(setting, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}
	return &oauthIdentity{
		ExternalID: fmt.Sprint(userInfo.ID),
		AuthType:   string(authModel.AuthTypeOAuth2),
	}, nil
}

func (a *googleOAuthAdapter) ResolveIdentity(
	setting *settingModel.OAuth2Setting,
	code string,
	_ *authModel.OAuthState,
) (*oauthIdentity, error) {
	tokenResp, err := exchangeGoogleCodeForToken(setting, code)
	if err != nil {
		return nil, err
	}
	userInfo, err := fetchGoogleUserInfo(setting, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}
	return &oauthIdentity{
		ExternalID: userInfo.Sub,
		AuthType:   string(authModel.AuthTypeOAuth2),
	}, nil
}

func (a *qqOAuthAdapter) ResolveIdentity(
	setting *settingModel.OAuth2Setting,
	code string,
	_ *authModel.OAuthState,
) (*oauthIdentity, error) {
	tokenResp, err := exchangeQQCodeForToken(setting, code)
	if err != nil {
		return nil, err
	}
	openIDResp, err := fetchQQUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}
	return &oauthIdentity{
		ExternalID: openIDResp.OpenID,
		AuthType:   string(authModel.AuthTypeOAuth2),
	}, nil
}

func (a *customOAuthAdapter) ResolveIdentity(
	setting *settingModel.OAuth2Setting,
	code string,
	oauthState *authModel.OAuthState,
) (*oauthIdentity, error) {
	accessToken, idToken, err := exchangeCustomCodeForToken(setting, code)
	if err != nil {
		return nil, err
	}

	if setting.IsOIDC {
		oauthID, err := fetchCustomUserInfo(setting, accessToken, idToken, oauthState.Nonce)
		if err != nil {
			return nil, err
		}
		return &oauthIdentity{
			ExternalID: oauthID,
			Issuer:     setting.Issuer,
			AuthType:   string(authModel.AuthTypeOIDC),
		}, nil
	}

	oauthID, err := fetchCustomUserInfo(setting, accessToken, "", "")
	if err != nil {
		return nil, err
	}
	return &oauthIdentity{
		ExternalID: oauthID,
		AuthType:   string(authModel.AuthTypeOAuth2),
	}, nil
}

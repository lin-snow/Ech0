package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	model "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/transaction"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type AuthService struct {
	transactor     transaction.Transactor
	repository     Repository
	authRepo       AuthRepo
	settingService SettingService
}

func NewAuthService(
	tx transaction.Transactor,
	repository Repository,
	authRepo AuthRepo,
	settingService SettingService,
) *AuthService {
	return &AuthService{
		transactor:     tx,
		repository:     repository,
		authRepo:       authRepo,
		settingService: settingService,
	}
}

func (authService *AuthService) RevokeToken(jti string, remainTTL time.Duration) {
	authService.authRepo.RevokeToken(jti, remainTTL)
}

func (authService *AuthService) IsTokenRevoked(jti string) bool {
	return authService.authRepo.IsTokenRevoked(jti)
}

func (authService *AuthService) ExchangeOAuthCode(code string) (*authModel.TokenPair, error) {
	return authService.authRepo.GetAndDeleteOAuthCode(code)
}

func (authService *AuthService) Login(loginDto *authModel.LoginDto) (*authModel.TokenPair, error) {
	if loginDto.Username == "" || loginDto.Password == "" {
		return nil, errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
	}

	loginDto.Password = cryptoUtil.MD5Encrypt(loginDto.Password)
	user, err := authService.repository.GetUserByUsername(context.Background(), loginDto.Username)
	if err != nil {
		return nil, errors.New(commonModel.USER_NOTFOUND)
	}
	if user.Password != loginDto.Password {
		return nil, errors.New(commonModel.PASSWORD_INCORRECT)
	}
	return authService.issueUserToken(user)
}

func (authService *AuthService) issueUserToken(user model.User) (*authModel.TokenPair, error) {
	accessClaims := jwtUtil.CreateClaims(user)
	accessToken, err := jwtUtil.GenerateToken(accessClaims)
	if err != nil {
		return nil, err
	}

	refreshClaims := jwtUtil.CreateRefreshClaims(user)
	refreshToken, err := jwtUtil.GenerateToken(refreshClaims)
	if err != nil {
		return nil, err
	}

	return &authModel.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.Config().Auth.Jwt.Expires,
	}, nil
}

func (authService *AuthService) BindOAuth(
	ctx context.Context,
	provider string,
	redirectURI string,
) (string, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	user, err := authService.repository.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if !user.IsAdmin {
		return "", bindingPermissionError(provider)
	}

	setting, err := authService.getOAuthSetting(provider)
	if err != nil {
		return "", err
	}

	state, nonce, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionBind),
		userID,
		redirectURI,
		provider,
	)
	if err != nil {
		return "", err
	}

	authorizeURL := authService.buildOAuthAuthorizeURL(setting, provider, state, nonce)
	if authorizeURL == "" {
		return "", errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return authorizeURL, nil
}

func (authService *AuthService) GetOAuthLoginURL(provider string, redirectURI string) (string, error) {
	setting, err := authService.getOAuthSetting(provider)
	if err != nil {
		return "", err
	}

	state, nonce, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionLogin),
		"",
		redirectURI,
		provider,
	)
	if err != nil {
		return "", err
	}

	authorizeURL := authService.buildOAuthAuthorizeURL(setting, provider, state, nonce)
	if authorizeURL == "" {
		return "", errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return authorizeURL, nil
}

func (authService *AuthService) HandleOAuthCallback(
	provider string,
	code string,
	state string,
) (string, error) {
	setting, err := authService.getOAuthSetting(provider)
	if err != nil {
		return "", err
	}

	oauthState, err := jwtUtil.ParseOAuthState(state)
	if err != nil {
		return "", err
	}

	if oauthState.Provider != provider {
		return "", errors.New(commonModel.INVALID_PARAMS)
	}

	adapter, err := getOAuthProviderAdapter(provider)
	if err != nil {
		return "", err
	}
	identity, err := adapter.ResolveIdentity(setting, code, oauthState)
	if err != nil {
		logUtil.Error("resolve oauth identity failed", zap.String("provider", provider), zap.Error(err))
		return "", err
	}

	return authService.resolveOAuthCallback(
		oauthState,
		provider,
		identity.ExternalID,
		identity.Issuer,
		identity.AuthType,
	)
}

func (authService *AuthService) getOAuthSetting(provider string) (*settingModel.OAuth2Setting, error) {
	var setting settingModel.OAuth2Setting
	systemCtx := viewer.WithContext(context.Background(), viewer.NewSystemViewer())
	if err := authService.settingService.GetOAuth2Setting(systemCtx, &setting, true); err != nil {
		return nil, err
	}

	if setting.Provider != provider {
		return nil, errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	if !setting.Enable {
		return nil, errors.New(commonModel.OAUTH2_NOT_ENABLED)
	}

	if setting.ClientID == "" || setting.RedirectURI == "" || setting.AuthURL == "" || setting.TokenURL == "" ||
		setting.UserInfoURL == "" ||
		setting.ClientSecret == "" {
		return nil, errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return &setting, nil
}

func (authService *AuthService) buildOAuthAuthorizeURL(
	setting *settingModel.OAuth2Setting,
	provider, state, nonce string,
) string {
	scope := ""
	if len(setting.Scopes) > 0 {
		scope = strings.Join(setting.Scopes, " ")
	}
	if setting.IsOIDC {
		scope = "openid " + scope
	}

	switch provider {
	case string(commonModel.OAuth2GITHUB):
		config := oauth2.Config{
			ClientID:    setting.ClientID,
			RedirectURL: setting.RedirectURI,
			Scopes:      setting.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  setting.AuthURL,
				TokenURL: setting.TokenURL,
			},
		}
		return config.AuthCodeURL(state)
	case string(commonModel.OAuth2GOOGLE):
		config := oauth2.Config{
			ClientID:    setting.ClientID,
			RedirectURL: setting.RedirectURI,
			Scopes:      setting.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  setting.AuthURL,
				TokenURL: setting.TokenURL,
			},
		}
		return config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	case string(commonModel.OAuth2QQ):
		params := url.Values{}
		params.Set("response_type", "code")
		params.Set("client_id", setting.ClientID)
		params.Set("redirect_uri", setting.RedirectURI)
		params.Set("state", state)
		params.Set("display", "pc")
		if scope != "" {
			params.Set("scope", scope)
		}
		return fmt.Sprintf("%s?%s", setting.AuthURL, params.Encode())
	case string(commonModel.OAuth2CUSTOM):
		config := oauth2.Config{
			ClientID:    setting.ClientID,
			RedirectURL: setting.RedirectURI,
			Scopes:      setting.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  setting.AuthURL,
				TokenURL: setting.TokenURL,
			},
		}
		opts := []oauth2.AuthCodeOption{}
		if setting.IsOIDC && nonce != "" {
			opts = append(opts, oauth2.SetAuthURLParam("nonce", nonce))
		}
		return config.AuthCodeURL(state, opts...)
	default:
		return ""
	}
}

func bindingPermissionError(provider string) error {
	switch provider {
	case string(commonModel.OAuth2GITHUB):
		return errors.New(commonModel.NO_PERMISSION_BINDING_GITHUB)
	case string(commonModel.OAuth2GOOGLE):
		return errors.New(commonModel.NO_PERMISSION_BINDING_GOOGLE)
	case string(commonModel.OAuth2QQ):
		return errors.New(commonModel.NO_PERMISSION_BINDING_QQ)
	case string(commonModel.OAuth2CUSTOM):
		return errors.New(commonModel.NO_PERMISSION_BINDING_CUSTOM)
	default:
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
}

func (authService *AuthService) resolveOAuthCallback(
	oauthState *authModel.OAuthState,
	provider, externalID, issuer, authType string,
) (string, error) {
	switch oauthState.Action {
	case string(authModel.OAuth2ActionLogin):
		if oauthState.UserID != "" {
			logUtil.Warn(
				"auth audit",
				zap.String("provider", provider),
				zap.String("action", "oauth_login"),
				zap.String("user_id", ""),
				zap.String("result", "fail"),
				zap.String("reason", "unexpected_user_id_in_login_state"),
			)
			return "", errors.New(commonModel.INVALID_PARAMS)
		}

		var (
			user model.User
			err  error
		)

		if authType == string(authModel.AuthTypeOIDC) {
			user, err = authService.repository.GetUserByOIDC(
				context.Background(),
				provider,
				externalID,
				issuer,
			)
		} else {
			user, err = authService.repository.GetUserByOAuthID(
				context.Background(),
				provider,
				externalID,
			)
		}
		if err != nil {
			logUtil.Error("fetch user by oauth id failed", zap.String("provider", provider), zap.Error(err))
			logUtil.Warn(
				"auth audit",
				zap.String("provider", provider),
				zap.String("action", "oauth_login"),
				zap.String("user_id", ""),
				zap.String("result", "fail"),
				zap.String("reason", "identity_not_bound_or_lookup_failed"),
			)
			return "", err
		}

		tokenPair, err := authService.issueUserToken(user)
		if err != nil {
			logUtil.Error("generate oauth login token failed", zap.String("provider", provider), zap.Error(err))
			logUtil.Warn(
				"auth audit",
				zap.String("provider", provider),
				zap.String("action", "oauth_login"),
				zap.String("user_id", user.ID),
				zap.String("result", "fail"),
				zap.String("reason", "issue_token_failed"),
			)
			return "", err
		}

		redirectURL, err := authService.parseAndValidateClientRedirect(oauthState.Redirect)
		if err != nil {
			return "", err
		}

		code := cryptoUtil.GenerateRandomString(32)
		authService.authRepo.StoreOAuthCode(code, tokenPair, 60*time.Second)
		query := redirectURL.Query()
		query.Set("code", code)
		redirectURL.RawQuery = query.Encode()
		logUtil.Info(
			"auth audit",
			zap.String("provider", provider),
			zap.String("action", "oauth_login"),
			zap.String("user_id", user.ID),
			zap.String("result", "success"),
			zap.String("reason", ""),
		)

		return redirectURL.String(), nil

	case string(authModel.OAuth2ActionBind):
		if oauthState.UserID == "" {
			logUtil.Warn(
				"auth audit",
				zap.String("provider", provider),
				zap.String("action", "oauth_bind"),
				zap.String("user_id", ""),
				zap.String("result", "fail"),
				zap.String("reason", "missing_user_id"),
			)
			return "", errors.New(commonModel.INVALID_PARAMS)
		}

		if err := authService.transactor.Run(context.Background(), func(ctx context.Context) error {
			return authService.repository.BindOAuth(
				ctx,
				oauthState.UserID,
				provider,
				externalID,
				issuer,
				authType,
			)
		}); err != nil {
			logUtil.Warn(
				"auth audit",
				zap.String("provider", provider),
				zap.String("action", "oauth_bind"),
				zap.String("user_id", oauthState.UserID),
				zap.String("result", "fail"),
				zap.String("reason", "bind_persist_failed"),
			)
			return "", err
		}

		redirectURL, err := authService.parseAndValidateClientRedirect(oauthState.Redirect)
		if err != nil {
			return "", err
		}
		query := redirectURL.Query()
		query.Set("bind", "success")
		redirectURL.RawQuery = query.Encode()
		logUtil.Info(
			"auth audit",
			zap.String("provider", provider),
			zap.String("action", "oauth_bind"),
			zap.String("user_id", oauthState.UserID),
			zap.String("result", "success"),
			zap.String("reason", ""),
		)
		return redirectURL.String(), nil
	default:
		return "", errors.New(commonModel.INVALID_PARAMS)
	}
}

func (authService *AuthService) parseAndValidateClientRedirect(redirect string) (*url.URL, error) {
	redirectURL, err := url.Parse(redirect)
	if err != nil || redirectURL == nil {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	if !redirectURL.IsAbs() || redirectURL.Host == "" {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	if redirectURL.Scheme != "http" && redirectURL.Scheme != "https" {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}

	allowed := config.Config().Auth.Redirect.AllowedReturnURLs
	if authService.settingService != nil {
		systemCtx := viewer.WithContext(context.Background(), viewer.NewSystemViewer())
		var oauthSetting settingModel.OAuth2Setting
		if err := authService.settingService.GetOAuth2Setting(systemCtx, &oauthSetting, true); err == nil &&
			len(oauthSetting.AuthRedirectAllowedReturnURLs) > 0 {
			allowed = oauthSetting.AuthRedirectAllowedReturnURLs
		}
	}
	if len(allowed) == 0 {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	matched := false
	for _, item := range allowed {
		allowURL, parseErr := url.Parse(strings.TrimSpace(item))
		if parseErr != nil || allowURL == nil || allowURL.Host == "" {
			continue
		}
		if strings.EqualFold(redirectURL.Scheme, allowURL.Scheme) &&
			strings.EqualFold(redirectURL.Host, allowURL.Host) {
			matched = true
			break
		}
	}
	if !matched {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}

	return redirectURL, nil
}

const passkeySessionTTL = 5 * time.Minute

const (
	passkeyRegKey   = "passkey:reg"
	passkeyLoginKey = "passkey:login"
)

type passkeySessionCache struct {
	Session    webauthn.SessionData
	Origin     string
	DeviceName string
}

type webauthnUser struct {
	u           model.User
	userHandle  []byte
	credentials []webauthn.Credential
}

func (w *webauthnUser) WebAuthnID() []byte {
	return w.userHandle
}

func (w *webauthnUser) WebAuthnName() string {
	return w.u.Username
}

func (w *webauthnUser) WebAuthnDisplayName() string {
	return w.u.Username
}

func (w *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return w.credentials
}

func newNonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func getPasskeyRegisterSessionKey(nonce string) string {
	return fmt.Sprintf("%s:%s", passkeyRegKey, nonce)
}

func getPasskeyLoginSessionKey(nonce string) string {
	return fmt.Sprintf("%s:%s", passkeyLoginKey, nonce)
}

func makeUserHandle(userID string) []byte {
	return []byte(userID)
}

func userIDFromHandle(handle []byte) string {
	return string(handle)
}

func (authService *AuthService) newWebAuthn(rpID, origin string) (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPDisplayName: "Ech0",
		RPID:          rpID,
		RPOrigins:     []string{origin},
	})
}

func (authService *AuthService) getWebauthnUserByID(
	userID string,
) (*webauthnUser, model.User, error) {
	u, err := authService.repository.GetUserByID(context.Background(), userID)
	if err != nil {
		return nil, model.User{}, err
	}

	passkeys, err := authService.repository.ListPasskeysByUserID(userID)
	if err != nil {
		return nil, model.User{}, err
	}

	credentials := make([]webauthn.Credential, 0, len(passkeys))
	for _, pk := range passkeys {
		var cred webauthn.Credential
		if err := json.Unmarshal([]byte(pk.CredentialJSON), &cred); err != nil {
			continue
		}
		cred.Authenticator.SignCount = pk.SignCount
		credentials = append(credentials, cred)
	}

	return &webauthnUser{
		u:           u,
		userHandle:  makeUserHandle(userID),
		credentials: credentials,
	}, u, nil
}

func (authService *AuthService) PasskeyRegisterBegin(
	ctx context.Context,
	rpID, origin, deviceName string,
) (authModel.PasskeyRegisterBeginResp, error) {
	var resp authModel.PasskeyRegisterBeginResp
	userID := viewer.MustFromContext(ctx).UserID()

	wa, err := authService.newWebAuthn(rpID, origin)
	if err != nil {
		return resp, err
	}

	wUser, _, err := authService.getWebauthnUserByID(userID)
	if err != nil {
		return resp, err
	}

	if strings.TrimSpace(deviceName) == "" {
		deviceName = "Passkey"
	}

	creation, session, err := wa.BeginRegistration(
		wUser,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithAuthenticatorSelection(
			webauthn.SelectAuthenticator(
				"",
				func() *bool { b := true; return &b }(),
				string(protocol.VerificationPreferred),
			),
		),
	)
	if err != nil {
		return resp, err
	}

	nonce, err := newNonce()
	if err != nil {
		return resp, err
	}

	authService.repository.CacheSetPasskeySession(
		getPasskeyRegisterSessionKey(nonce),
		passkeySessionCache{
			Session:    *session,
			Origin:     origin,
			DeviceName: deviceName,
		},
		passkeySessionTTL,
	)

	resp.Nonce = nonce
	resp.PublicKey = &creation.Response
	return resp, nil
}

func (authService *AuthService) PasskeyRegisterFinish(
	ctx context.Context,
	rpID, origin, nonce string,
	credential json.RawMessage,
) error {
	userID := viewer.MustFromContext(ctx).UserID()
	cacheKey := getPasskeyRegisterSessionKey(nonce)
	cached, err := authService.repository.CacheGetPasskeySession(cacheKey)
	if err != nil {
		return errors.New(commonModel.INVALID_PARAMS)
	}
	authService.repository.CacheDeletePasskeySession(cacheKey)

	sess, ok := cached.(passkeySessionCache)
	if !ok {
		return errors.New(commonModel.INVALID_PARAMS)
	}
	if sess.Origin != origin {
		return errors.New(commonModel.INVALID_PARAMS)
	}

	wa, err := authService.newWebAuthn(rpID, origin)
	if err != nil {
		return err
	}

	wUser, _, err := authService.getWebauthnUserByID(userID)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest(
		"POST",
		"http://localhost/passkey/register/finish",
		bytes.NewReader(credential),
	)
	req.Header.Set("Content-Type", "application/json")

	cred, err := wa.FinishRegistration(wUser, sess.Session, req)
	if err != nil {
		return err
	}

	credID := base64.RawURLEncoding.EncodeToString(cred.ID)
	credJSON, _ := json.Marshal(cred)
	publicKey := base64.RawURLEncoding.EncodeToString(cred.PublicKey)
	aaguid := base64.RawURLEncoding.EncodeToString(cred.Authenticator.AAGUID)

	passkey := authModel.Passkey{
		UserID:         userID,
		CredentialID:   credID,
		CredentialJSON: string(credJSON),
		PublicKey:      publicKey,
		SignCount:      cred.Authenticator.SignCount,
		LastUsedAt:     time.Now().UTC().Unix(),
		DeviceName:     sess.DeviceName,
		AAGUID:         aaguid,
	}

	return authService.transactor.Run(context.Background(), func(ctx context.Context) error {
		return authService.repository.CreatePasskey(ctx, &passkey)
	})
}

func (authService *AuthService) PasskeyLoginBegin(
	rpID, origin string,
) (authModel.PasskeyLoginBeginResp, error) {
	var resp authModel.PasskeyLoginBeginResp

	wa, err := authService.newWebAuthn(rpID, origin)
	if err != nil {
		return resp, err
	}

	assertion, session, err := wa.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return resp, err
	}

	nonce, err := newNonce()
	if err != nil {
		return resp, err
	}

	authService.repository.CacheSetPasskeySession(
		getPasskeyLoginSessionKey(nonce),
		passkeySessionCache{
			Session: *session,
			Origin:  origin,
		},
		passkeySessionTTL,
	)

	resp.Nonce = nonce
	resp.PublicKey = &assertion.Response
	return resp, nil
}

func (authService *AuthService) PasskeyLoginFinish(
	rpID, origin, nonce string,
	credential json.RawMessage,
) (*authModel.TokenPair, error) {
	cacheKey := getPasskeyLoginSessionKey(nonce)
	cached, err := authService.repository.CacheGetPasskeySession(cacheKey)
	if err != nil {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	authService.repository.CacheDeletePasskeySession(cacheKey)

	sess, ok := cached.(passkeySessionCache)
	if !ok {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	if sess.Origin != origin {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}

	wa, err := authService.newWebAuthn(rpID, origin)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(
		"POST",
		"http://localhost/passkey/login/finish",
		bytes.NewReader(credential),
	)
	req.Header.Set("Content-Type", "application/json")

	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		credID := base64.RawURLEncoding.EncodeToString(rawID)
		pk, err := authService.repository.GetPasskeyByCredentialID(credID)
		if err != nil {
			return nil, err
		}

		expected := makeUserHandle(pk.UserID)
		if len(userHandle) > 0 && !bytes.Equal(userHandle, expected) {
			return nil, errors.New(commonModel.INVALID_PARAMS)
		}

		wUser, _, err := authService.getWebauthnUserByID(pk.UserID)
		if err != nil {
			return nil, err
		}
		return wUser, nil
	}

	user, credentialObj, err := wa.FinishPasskeyLogin(handler, sess.Session, req)
	if err != nil {
		return nil, err
	}

	uid := userIDFromHandle(user.WebAuthnID())
	if uid == "" {
		credID := base64.RawURLEncoding.EncodeToString(credentialObj.ID)
		pk, err2 := authService.repository.GetPasskeyByCredentialID(credID)
		if err2 != nil {
			return nil, err2
		}
		uid = pk.UserID
	}

	credID := base64.RawURLEncoding.EncodeToString(credentialObj.ID)
	pk, err := authService.repository.GetPasskeyByCredentialID(credID)
	if err == nil {
		_ = authService.transactor.Run(context.Background(), func(ctx context.Context) error {
			return authService.repository.UpdatePasskeyUsage(
				ctx,
				pk.ID,
				credentialObj.Authenticator.SignCount,
				time.Now().UTC().Unix(),
			)
		})
	}

	u, err := authService.repository.GetUserByID(context.Background(), uid)
	if err != nil {
		return nil, err
	}

	tokenPair, err := authService.issueUserToken(u)
	if err != nil {
		return nil, err
	}
	return tokenPair, nil
}

func (authService *AuthService) ListPasskeys(ctx context.Context) ([]authModel.PasskeyDeviceDto, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	passkeys, err := authService.repository.ListPasskeysByUserID(userID)
	if err != nil {
		return nil, err
	}

	devs := make([]authModel.PasskeyDeviceDto, 0, len(passkeys))
	for _, pk := range passkeys {
		devs = append(devs, authModel.PasskeyDeviceDto{
			ID:         pk.ID,
			DeviceName: pk.DeviceName,
			AAGUID:     pk.AAGUID,
			LastUsedAt: pk.LastUsedAt,
			CreatedAt:  pk.CreatedAt,
		})
	}
	return devs, nil
}

func (authService *AuthService) DeletePasskey(ctx context.Context, passkeyID string) error {
	userID := viewer.MustFromContext(ctx).UserID()
	return authService.transactor.Run(ctx, func(txCtx context.Context) error {
		return authService.repository.DeletePasskeyByID(txCtx, userID, passkeyID)
	})
}

func (authService *AuthService) UpdatePasskeyDeviceName(
	ctx context.Context,
	passkeyID string,
	deviceName string,
) error {
	userID := viewer.MustFromContext(ctx).UserID()
	if strings.TrimSpace(deviceName) == "" {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	return authService.transactor.Run(ctx, func(txCtx context.Context) error {
		return authService.repository.UpdatePasskeyDeviceName(
			txCtx,
			userID,
			passkeyID,
			deviceName,
		)
	})
}

func (authService *AuthService) GetOAuthInfo(
	ctx context.Context,
	provider string,
) (model.OAuthInfoDto, error) {
	var oauthInfo model.OAuthInfoDto
	userId := viewer.MustFromContext(ctx).UserID()

	user, err := authService.repository.GetUserByID(ctx, userId)
	if err != nil {
		return oauthInfo, err
	}

	if !user.IsAdmin {
		return oauthInfo, bindingPermissionError(provider)
	}

	var oauth2Setting settingModel.OAuth2Setting
	if err := authService.settingService.GetOAuth2Setting(viewer.WithContext(ctx, viewer.NewUserViewer(user.ID)), &oauth2Setting, true); err != nil {
		return oauthInfo, err
	}
	isOIDC := oauth2Setting.IsOIDC
	issuer := oauth2Setting.Issuer
	authType := string(authModel.AuthTypeOAuth2)
	if isOIDC {
		authType = string(authModel.AuthTypeOIDC)
	}

	var oauthInfoBinding model.UserExternalIdentity
	if isOIDC {
		oauthInfoBinding, err = authService.repository.GetOAuthOIDCInfo(
			user.ID,
			provider,
			issuer,
		)
		if err != nil {
			return oauthInfo, err
		}
	} else {
		oauthInfoBinding, err = authService.repository.GetOAuthInfo(user.ID, provider)
		if err != nil {
			return oauthInfo, err
		}
	}

	oauthInfo = model.OAuthInfoDto{
		Provider: oauthInfoBinding.Provider,
		UserID:   oauthInfoBinding.UserID,
		OAuthID:  oauthInfoBinding.Subject,
		Issuer:   oauthInfoBinding.Issuer,
		AuthType: authType,
	}

	return oauthInfo, nil
}

func exchangeGithubCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.GitHubTokenResponse, error) {
	token, err := exchangeOAuthCode(setting, code)
	if err != nil {
		return nil, err
	}
	return &authModel.GitHubTokenResponse{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Scope:       fmt.Sprint(token.Extra("scope")),
	}, nil
}

func fetchGitHubUserInfo(
	setting *settingModel.OAuth2Setting,
	accessToken string,
) (*authModel.GitHubUser, error) {
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("GitHub 用户信息请求失败: " + string(body))
	}

	var user authModel.GitHubUser
	_ = json.Unmarshal(body, &user)
	return &user, nil
}

func exchangeGoogleCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.GoogleTokenResponse, error) {
	token, err := exchangeOAuthCode(setting, code)
	if err != nil {
		return nil, err
	}
	expiresIn := int64(0)
	if !token.Expiry.IsZero() {
		expiresIn = int64(time.Until(token.Expiry).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}
	}
	return &authModel.GoogleTokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		ExpiresIn:    expiresIn,
		RefreshToken: token.RefreshToken,
		Scope:        fmt.Sprint(token.Extra("scope")),
		IDToken:      fmt.Sprint(token.Extra("id_token")),
	}, nil
}

func fetchGoogleUserInfo(
	setting *settingModel.OAuth2Setting,
	accessToken string,
) (*authModel.GoogleUser, error) {
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Google 用户信息请求失败: " + string(body))
	}

	var user authModel.GoogleUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func exchangeQQCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.QQTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", setting.ClientID)
	data.Set("client_secret", setting.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", setting.RedirectURI)
	data.Set("fmt", "json")
	data.Set("need_openid", "1")

	req, _ := http.NewRequest("GET", setting.TokenURL+"?"+data.Encode(), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("QQ token 响应错误: " + string(body))
	}

	raw := strings.TrimSpace(string(body))
	if strings.HasPrefix(raw, "callback(") && strings.HasSuffix(raw, ");") {
		raw = strings.TrimPrefix(raw, "callback(")
		raw = strings.TrimSuffix(raw, ");")
		raw = strings.TrimSpace(raw)
	}

	var tokenResp authModel.QQTokenResponse
	if err := json.Unmarshal([]byte(raw), &tokenResp); err == nil {
		if tokenResp.AccessToken != "" {
			return &tokenResp, nil
		}
	}

	vals, err := url.ParseQuery(raw)
	if err == nil && vals.Get("access_token") != "" {
		tokenResp.AccessToken = vals.Get("access_token")
		tokenResp.RefreshToken = vals.Get("refresh_token")
		tokenResp.ExpiresIn, _ = strconv.ParseInt(vals.Get("expires_in"), 10, 64)
		tokenResp.OpenID = vals.Get("openid")
		return &tokenResp, nil
	}

	return nil, errors.New("无法解析 QQ token 响应: " + string(body))
}

func fetchQQUserInfo(accessToken string) (*authModel.QQOpenIDResponse, error) {
	openIDURL := "https://graph.qq.com/oauth2.0/me" + "?access_token=" + url.QueryEscape(
		accessToken,
	) + "&fmt=json"
	req, _ := http.NewRequest("GET", openIDURL, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("QQ openid 请求失败: " + string(body))
	}

	var openIDResp authModel.QQOpenIDResponse
	if err := json.Unmarshal(body, &openIDResp); err != nil {
		return nil, err
	}

	return &openIDResp, nil
}

func exchangeCustomCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (accessToken string, idToken string, err error) {
	token, err := exchangeOAuthCode(setting, code)
	if err != nil {
		return "", "", err
	}

	accessToken = token.AccessToken
	if accessToken == "" {
		return "", "", errors.New("custom token 响应缺少 access_token")
	}

	if setting.IsOIDC {
		idToken = fmt.Sprint(token.Extra("id_token"))
		if idToken == "" {
			return "", "", errors.New("OIDC 响应缺少 id_token")
		}
	}

	return accessToken, idToken, nil
}

func exchangeOAuthCode(setting *settingModel.OAuth2Setting, code string) (*oauth2.Token, error) {
	config := oauth2.Config{
		ClientID:     setting.ClientID,
		ClientSecret: setting.ClientSecret,
		RedirectURL:  setting.RedirectURI,
		Scopes:       setting.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  setting.AuthURL,
			TokenURL: setting.TokenURL,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func fetchCustomUserInfo(
	setting *settingModel.OAuth2Setting,
	accessToken, idToken, expectedNonce string,
) (string, error) {
	if setting.IsOIDC {
		if idToken == "" {
			return "", errors.New("OIDC id_token is empty")
		}

		claims, err := jwtUtil.ParseAndVerifyIDToken(
			idToken,
			setting.Issuer,
			setting.JWKSURL,
			setting.ClientID,
			expectedNonce,
		)
		if err != nil {
			return "", err
		}

		return claims["sub"].(string), nil
	}

	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Custom 用户信息请求失败: " + string(body))
	}

	var userData map[string]any
	if err := json.Unmarshal(body, &userData); err != nil {
		return "", err
	}

	for _, key := range []string{"id", "sub", "user_id", "uid", "openid"} {
		if val, ok := userData[key]; ok {
			if id := fmt.Sprint(val); id != "" && id != "<nil>" {
				return id, nil
			}
		}
	}

	return "", errors.New("custom 用户信息缺少唯一标识字段 (id/sub/user_id/uid)")
}

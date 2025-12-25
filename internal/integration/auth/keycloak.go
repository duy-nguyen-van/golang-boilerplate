package auth

import (
	"context"
	"fmt"
	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/httpclient"
	"golang-boilerplate/internal/monitoring"

	"golang-boilerplate/internal/logger"

	"github.com/Nerzal/gocloak/v13"
	"github.com/getsentry/sentry-go"
)

// KeycloakAuth implements AuthService using Keycloak
type KeycloakAuth struct {
	client     *gocloak.GoCloak
	restClient httpclient.RestClient
	config     *config.Config
}

// NewKeycloakAuth creates a new Keycloak authentication service
func NewKeycloakAuth(cfg *config.Config, restClient httpclient.RestClient) (*KeycloakAuth, error) {
	return &KeycloakAuth{
		client:     gocloak.NewClient(cfg.KeycloakURL),
		restClient: restClient,
		config:     cfg,
	}, nil
}

// Login performs client login and returns an access token
func (a *KeycloakAuth) ClientLogin() (*TokenInfo, error) {
	token, err := a.client.LoginClient(context.Background(), a.config.KeycloakClientID, a.config.KeycloakSecret, a.config.KeycloakRealm)
	if err != nil {
		if hub := monitoring.GetSentryHub(context.Background()); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "login")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("client_id", a.config.KeycloakClientID)
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorf("Failed to login to keycloak: %v", err)
		return nil, errors.ExternalServiceError("Failed to login to keycloak", err).
			WithOperation("login").
			WithResource("keycloak")
	}
	return &TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		TokenType:    token.TokenType,
	}, nil
}

// GetUserInfo retrieves user information from Keycloak
func (a *KeycloakAuth) GetUserInfo(token string) (*User, error) {
	userInfo, err := a.client.GetUserInfo(context.Background(), token, a.config.KeycloakRealm)
	if err != nil {
		if hub := monitoring.GetSentryHub(context.Background()); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "get_user_info")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorf("Failed to get user info: %v", err)
		return nil, errors.ExternalServiceError("Failed to get user info", err).
			WithOperation("get_user_info").
			WithResource("keycloak")
	}

	return &User{
		Sub:               *userInfo.Sub,
		EmailVerified:     *userInfo.EmailVerified,
		Name:              *userInfo.Name,
		PreferredUsername: *userInfo.PreferredUsername,
		GivenName:         *userInfo.GivenName,
		FamilyName:        *userInfo.FamilyName,
		Email:             *userInfo.Email,
	}, nil
}

// ValidateToken validates a Keycloak token
func (a *KeycloakAuth) ValidateToken(token string) (*gocloak.IntroSpectTokenResult, error) {
	result, err := a.client.RetrospectToken(context.Background(), token, a.config.KeycloakClientID, a.config.KeycloakSecret, a.config.KeycloakRealm)

	if err != nil {
		if hub := monitoring.GetSentryHub(context.Background()); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "validate_token")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorf("Failed to validate token: %v", err)
		return nil, errors.ExternalServiceError("Failed to validate token", err).
			WithOperation("validate_token").
			WithResource("keycloak")
	}

	return result, nil
}

func (a *KeycloakAuth) DecodeAccessToken(ctx context.Context, token string, realm string, claims *TokenClaims) (*TokenClaims, error) {
	_, err := a.client.DecodeAccessTokenCustomClaims(ctx, token, realm, claims)
	if err != nil {
		// Capture invalid claims error in Sentry
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("auth_error", "invalid_claims")
				scope.SetTag("service", "fast-ai")
				scope.SetTag("environment", a.config.AppEnv)
				scope.SetExtra("error_details", err.Error())
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorf("Failed to decode access token custom claims: %v", err)
		return nil, errors.ExternalServiceError("Failed to decode access token custom claims", err).
			WithOperation("decode_access_token_custom_claims").
			WithResource("keycloak")
	}

	return claims, nil
}

func (a *KeycloakAuth) GetRealm() string {
	return a.config.KeycloakRealm
}

func (a *KeycloakAuth) GetClaimsKey() string {
	return a.config.KeycloakKeyClaim
}

// GetRequestingPartyToken exchanges an access token for an RPT with evaluated permissions
func (a *KeycloakAuth) GetRequestingPartyToken(ctx context.Context, accessToken string, opts RequestingPartyTokenOptions) (*JWT, error) {
	rpt, err := a.client.GetRequestingPartyToken(ctx, accessToken, a.config.KeycloakRealm, gocloak.RequestingPartyTokenOptions(opts))
	if err != nil {
		// Capture error in Sentry
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "get_rpt")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorf("Failed to get RPT: %v", err)
		return nil, errors.ExternalServiceError("Failed to get RPT", err).
			WithOperation("get_rpt").
			WithResource("keycloak")
	}
	return &JWT{
		AccessToken:      rpt.AccessToken,
		IDToken:          rpt.IDToken,
		ExpiresIn:        rpt.ExpiresIn,
		RefreshExpiresIn: rpt.RefreshExpiresIn,
		RefreshToken:     rpt.RefreshToken,
		TokenType:        rpt.TokenType,
		NotBeforePolicy:  rpt.NotBeforePolicy,
		SessionState:     rpt.SessionState,
		Scope:            rpt.Scope,
	}, nil
}

func (a *KeycloakAuth) CreateUser(ctx context.Context, adminToken string, userDto *dtos.CreateUserRequest) (*User, error) {
	emailVerified := true
	enabled := true
	userID, err := a.client.CreateUser(ctx, adminToken, a.config.KeycloakRealm, gocloak.User{
		Email:           &userDto.Email,
		Username:        &userDto.Email,
		EmailVerified:   &emailVerified,
		Enabled:         &enabled,
		RequiredActions: &[]string{"VERIFY_EMAIL", "UPDATE_PASSWORD"},
	})

	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "create_user")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		return nil, errors.ExternalServiceError("Failed to create user", err).
			WithOperation("create_user").
			WithResource("keycloak")
	}

	return &User{
		ID: userID,
	}, nil
}

func (a *KeycloakAuth) getClients(ctx context.Context, adminToken string) ([]*gocloak.Client, error) {
	kcClients, err := a.client.GetClients(ctx, adminToken, a.config.KeycloakRealm, gocloak.GetClientsParams{ClientID: gocloak.StringP(a.config.KeycloakClientID)})
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "get_clients")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorw("Failed to get clients",
			"realm", a.config.KeycloakRealm,
			"client_id", a.config.KeycloakClientID,
			"error", err,
		)
		return nil, errors.ExternalServiceError("Failed to get clients", err).
			WithOperation("get_clients").
			WithResource("keycloak")
	}
	return kcClients, nil
}

func (a *KeycloakAuth) getClientRole(ctx context.Context, adminToken string, clientID string, roleName string) (*gocloak.Role, error) {
	kcRole, err := a.client.GetClientRole(ctx, adminToken, a.config.KeycloakRealm, clientID, roleName)
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "get_client_role")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorw("Failed to get client role",
			"client_id", clientID,
			"role", roleName,
			"realm", a.config.KeycloakRealm,
			"error", err,
		)

		return nil, errors.ExternalServiceError("Failed to get client role", err).
			WithOperation("get_client_role").
			WithResource("keycloak")
	}

	return kcRole, nil
}

func (a *KeycloakAuth) AddClientRolesToUser(ctx context.Context, adminToken string, userID string, clientID string, role string) error {
	kcClients, err := a.getClients(ctx, adminToken)
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "get_clients")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				scope.SetExtra("client_id", clientID)
				scope.SetExtra("user_id", userID)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorw("Failed to get clients",
			"realm", a.config.KeycloakRealm,
			"client_id", clientID,
			"user_id", userID,
			"error", err,
		)

		return errors.ExternalServiceError("Failed to get client", err).
			WithOperation("get_client").
			WithResource("keycloak").
			WithContext("client_id", clientID)
	}

	if len(kcClients) == 0 {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "get_client")
				scope.SetExtra("error_details", "Client not found")
				scope.SetExtra("realm", a.config.KeycloakRealm)
				scope.SetExtra("client_id", clientID)
				scope.SetExtra("user_id", userID)
				hub.CaptureException(errors.NotFoundError("Client not found", nil))
			})
		}
		return errors.ExternalServiceError("Client not found", nil).
			WithOperation("get_client").
			WithResource("keycloak").
			WithContext("client_id", clientID)
	}
	clientID = *kcClients[0].ID

	kcRole, err := a.getClientRole(ctx, adminToken, clientID, role)
	if err != nil {
		return errors.ExternalServiceError("Failed to get client role", err).
			WithOperation("get_client_role").
			WithResource("keycloak").
			WithContext("client_id", clientID).
			WithContext("role", role)
	}

	err = a.client.AddClientRolesToUser(ctx, adminToken, a.config.KeycloakRealm, clientID, userID, []gocloak.Role{*kcRole})
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "add_client_roles_to_user")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				scope.SetExtra("user_id", userID)
				scope.SetExtra("client_id", clientID)
				scope.SetExtra("role", *kcRole.ID)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorw("Failed to add client roles to user",
			"user_id", userID,
			"client_id", clientID,
			"role", *kcRole.ID,
			"realm", a.config.KeycloakRealm,
			"error", err,
		)

		return errors.ExternalServiceError("Failed to add client roles to user", err).
			WithOperation("add_client_roles_to_user").
			WithResource("keycloak").
			WithContext("client_id", clientID).
			WithContext("role", role)
	}
	return nil
}

func (a *KeycloakAuth) SetPassword(ctx context.Context, adminToken string, userID string, password string, temporary bool) error {
	err := a.client.SetPassword(ctx, adminToken, userID, a.config.KeycloakRealm, password, temporary)
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "set_password")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		return errors.ExternalServiceError("Failed to set password", err).
			WithOperation("set_password").
			WithResource("keycloak")
	}
	return nil
}

func (a *KeycloakAuth) SendVerificationMail(ctx context.Context, adminToken string, userID string, params SendVerificationMailParams) error {
	err := a.client.SendVerifyEmail(ctx, adminToken, userID, a.config.KeycloakRealm, gocloak.SendVerificationMailParams(params))
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "send_verification_email")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorf("Failed to send verification email: %v", err)
		return errors.ExternalServiceError("Failed to send verification email", err).
			WithOperation("send_verification_email").
			WithResource("keycloak")
	}
	return nil
}

func (a *KeycloakAuth) GetClientID() string {
	return a.config.KeycloakClientID
}

func (a *KeycloakAuth) GetRedirectURI() string {
	return a.config.KeycloakRedirectURI
}

// getOrganizationName safely extracts the organization name from JWT token claims
func (a *KeycloakAuth) GetOrganization(userClaims *TokenClaims) (Organization, error) {
	// First try to get from Organization field
	if userClaims.Organization != nil {
		for orgKey, orgData := range userClaims.Organization {
			if orgData != nil {
				// The organization name is the key, and we can get the ID from the data
				var orgID string
				if id, exists := orgData["id"]; exists {
					if idStr, ok := id.(string); ok {
						orgID = idStr
					}
				}

				return Organization{
					Name: orgKey, // The key is the organization name
					ID:   orgID,
				}, nil
			}
		}
	}

	// If not found in Organization, try OrganizationWildcard field
	if userClaims.OrganizationWildcard != nil {
		for orgKey, orgData := range userClaims.OrganizationWildcard {
			if orgData != nil {
				// The organization name is the key, and we can get the ID from the data
				var orgID string
				if id, exists := orgData["id"]; exists {
					if idStr, ok := id.(string); ok {
						orgID = idStr
					}
				}

				return Organization{
					Name: orgKey, // The key is the organization name
					ID:   orgID,
				}, nil
			}
		}
	}

	return Organization{}, errors.NotFoundError("Organization name not found in token claims", nil).
		WithOperation("get_organization_name").
		WithResource("keycloak")
}

// getHeaders creates headers for Keycloak API calls
func (a *KeycloakAuth) getHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
}

func (a *KeycloakAuth) AddUserToOrganization(ctx context.Context, adminToken string, userID string, organizationID string) error {
	// Use the correct Keycloak Organizations API endpoint
	url := fmt.Sprintf("%s/admin/realms/%s/organizations/%s/members",
		a.config.KeycloakURL, a.config.KeycloakRealm, organizationID)

	// Request body for inviting existing user to organization
	// The body should be just the user ID as a plain string
	requestBody := userID

	var response map[string]interface{}
	var errorResponse map[string]interface{}

	_, err := a.restClient.Post(url, requestBody, &response, &errorResponse, a.getHeaders(adminToken))
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "keycloak_adapter")
				scope.SetTag("operation", "add_user_to_organization")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("user_id", userID)
				scope.SetExtra("organization_id", organizationID)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorw("Failed to add user to organization via Keycloak API",
			"error", err.Error(),
			"user_id", userID,
			"organization_id", organizationID,
			"url", url,
		)

		return errors.ExternalServiceError("Failed to add user to organization", err).
			WithOperation("add_user_to_organization").
			WithResource("keycloak").
			WithContext("user_id", userID).
			WithContext("organization_id", organizationID)
	}

	logger.Sugar.Infow("Successfully added user to organization via Keycloak API",
		"user_id", userID,
		"organization_id", organizationID,
	)

	return nil
}

func (a *KeycloakAuth) UpdateUser(ctx context.Context, adminToken string, userID string, userDto *dtos.UpdateUserRequest) error {
	enabled := true
	if userDto.Status == constants.UserStatusInactive {
		enabled = false
	}

	err := a.client.UpdateUser(ctx, adminToken, a.config.KeycloakRealm, gocloak.User{
		ID:      &userID,
		Enabled: &enabled,
	})

	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("adapter", "keycloak")
				scope.SetTag("operation", "update_user")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("realm", a.config.KeycloakRealm)
				scope.SetExtra("user_id", userID)
				scope.SetExtra("body_request", userDto)
				hub.CaptureException(err)
			})
		}
		logger.Sugar.Errorw("Failed to update user",
			"user_id", userID,
			"body_request", userDto,
			"realm", a.config.KeycloakRealm,
			"error", err,
		)

		return errors.ExternalServiceError("Failed to update user", err).
			WithOperation("update_user").
			WithResource("keycloak").
			WithContext("user_id", userID)
	}
	return nil
}

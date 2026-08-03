package auth

import (
	"context"
	"fmt"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/httpclient"

	"github.com/Nerzal/gocloak/v13"
	jwt "github.com/golang-jwt/jwt/v5"
)

// User represents an authenticated user
type User struct {
	ID                string    `json:"id"`
	Sub               string    `json:"sub"`
	EmailVerified     bool      `json:"email_verified"`
	Name              string    `json:"name"`
	PreferredUsername string    `json:"preferred_username"`
	GivenName         string    `json:"given_name"`
	FamilyName        string    `json:"family_name"`
	Email             string    `json:"email"`
	Roles             []string  `json:"roles"`
	ExpiresAt         time.Time `json:"expires_at"`
	Enabled           *bool     `json:"enabled,omitempty"`
}

// TokenInfo represents token information
type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// AuthorizationPermissions represents Keycloak UMA authorization data within a token (RPT)
type AuthorizationPermissions struct {
	Permissions []struct {
		ResourceName string   `json:"rsname"`
		Scopes       []string `json:"scopes"`
	} `json:"permissions"`
}

// TokenClaims represents the complete structure of our JWT claims
type TokenClaims struct {
	jwt.MapClaims
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	Organization         map[string]map[string]interface{} `json:"organization"`
	OrganizationWildcard map[string]map[string]interface{} `json:"organization:*"`
	Email                string                            `json:"email"`
	EmailVerified        bool                              `json:"email_verified"`
	Sub                  string                            `json:"sub"`
	Name                 string                            `json:"name"`
	PreferredUsername    string                            `json:"preferred_username"`
	GivenName            string                            `json:"given_name"`
	FamilyName           string                            `json:"family_name"`
	Scope                string                            `json:"scope"`
	AllowedOrigins       []string                          `json:"allowed-origins"`

	// Present in RPT tokens when Authorization Services are enabled
	Authorization AuthorizationPermissions `json:"authorization"`
}

type JWT struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

// RequestingPartyTokenOptions represents the options to obtain a requesting party token
type RequestingPartyTokenOptions struct {
	GrantType                     *string   `json:"grant_type,omitempty"`
	Ticket                        *string   `json:"ticket,omitempty"`
	ClaimToken                    *string   `json:"claim_token,omitempty"`
	ClaimTokenFormat              *string   `json:"claim_token_format,omitempty"`
	RPT                           *string   `json:"rpt,omitempty"`
	Permissions                   *[]string `json:"-"`
	PermissionResourceFormat      *string   `json:"permission_resource_format,omitempty"`
	PermissionResourceMatchingURI *bool     `json:"permission_resource_matching_uri,string,omitempty"`
	Audience                      *string   `json:"audience,omitempty"`
	ResponseIncludeResourceName   *bool     `json:"response_include_resource_name,string,omitempty"`
	ResponsePermissionsLimit      *uint32   `json:"response_permissions_limit,omitempty"`
	SubmitRequest                 *bool     `json:"submit_request,string,omitempty"`
	ResponseMode                  *string   `json:"response_mode,omitempty"`
	SubjectToken                  *string   `json:"subject_token,omitempty"`
}

type SendVerificationMailParams struct {
	ClientID    *string
	RedirectURI *string
}

type Organization struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	GetRealm() string
	DecodeAccessToken(ctx context.Context, token string, realm string, claims *TokenClaims) (*TokenClaims, error)
	ValidateToken(token string) (*gocloak.IntroSpectTokenResult, error)
	ClientLogin() (*TokenInfo, error)
	GetUserInfo(token string) (*User, error)
	GetClaimsKey() string

	// Obtain an RPT (Requesting Party Token) using UMA to evaluate permissions
	GetRequestingPartyToken(ctx context.Context, accessToken string, opts RequestingPartyTokenOptions) (*JWT, error)
	CreateUser(ctx context.Context, adminToken string, userDto *dtos.CreateUserRequest) (*User, error)
	SetPassword(ctx context.Context, adminToken string, userID string, password string, temporary bool) error
	SendVerificationMail(ctx context.Context, adminToken string, userID string, params SendVerificationMailParams) error
	GetClientID() string
	GetRedirectURI() string
	GetOrganization(userClaims *TokenClaims) (Organization, error)
	AddUserToOrganization(ctx context.Context, adminToken string, userID string, organizationID string) error
	AddClientRolesToUser(ctx context.Context, adminToken string, userID string, clientID string, role string) error
	UpdateUser(ctx context.Context, adminToken string, userID string, userDto *dtos.UpdateUserRequest) error
}

func ProvideAuth(
	cfg *config.Config,
	restClient httpclient.RestClient,
) (AuthService, error) {
	switch cfg.AuthProvider {
	case constants.AuthProviderKeycloak:
		keycloakAuth, err := NewKeycloakAuth(cfg, restClient)
		if err != nil {
			return nil, errors.ExternalServiceError("Failed to initialize Keycloak auth", err).
				WithOperation("initialize_auth_provider").
				WithResource("auth")
		}
		return keycloakAuth, nil
	default:
		return nil, errors.InternalError("Invalid auth provider", fmt.Errorf("invalid auth provider: %s", cfg.AuthProvider)).
			WithOperation("initialize_auth_provider").
			WithResource("auth").
			WithContext("auth_provider", cfg.AuthProvider)
	}
}

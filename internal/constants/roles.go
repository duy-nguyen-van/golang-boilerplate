package constants

// Role constants for RBAC
// These roles should match the roles defined in Keycloak

// Realm-level roles (applied to all clients)
const (
	RoleAdmin     = "admin"     // Full system administrator
	RoleUser      = "user"      // Regular user
	RoleModerator = "moderator" // Content moderator
)

// Client-level roles (specific to this application)
const (
	// User management roles
	RoleUserManager = "user-manager" // Can manage users
	RoleUserViewer  = "user-viewer"  // Can view users only
	RoleUserCreator = "user-creator" // Can create users
	RoleUserEditor  = "user-editor"  // Can edit users
	RoleUserDeleter = "user-deleter" // Can delete users

	// Company management roles
	RoleCompanyManager = "company-manager" // Can manage companies
	RoleCompanyViewer  = "company-viewer"  // Can view companies only
	RoleCompanyCreator = "company-creator" // Can create companies
	RoleCompanyEditor  = "company-editor"  // Can edit companies
	RoleCompanyDeleter = "company-deleter" // Can delete companies
)

// RoleGroups groups related roles for easier middleware usage
var (
	// AdminRoles includes all administrative roles
	AdminRoles = []string{RoleAdmin}

	// UserManagementRoles includes all roles that can manage users
	UserManagementRoles = []string{RoleAdmin, RoleUserManager, RoleUserCreator, RoleUserEditor, RoleUserDeleter}

	// UserViewRoles includes roles that can view users
	UserViewRoles = []string{RoleAdmin, RoleUserManager, RoleUserViewer, RoleUserCreator, RoleUserEditor, RoleUserDeleter}

	// CompanyManagementRoles includes all roles that can manage companies
	CompanyManagementRoles = []string{RoleAdmin, RoleCompanyManager, RoleCompanyCreator, RoleCompanyEditor, RoleCompanyDeleter}

	// CompanyViewRoles includes roles that can view companies
	CompanyViewRoles = []string{RoleAdmin, RoleCompanyManager, RoleCompanyViewer, RoleCompanyCreator, RoleCompanyEditor, RoleCompanyDeleter}
)

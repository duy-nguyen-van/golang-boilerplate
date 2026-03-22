# RBAC (Role-Based Access Control) Guide

This guide explains how to implement and use RBAC in your application.

## Overview

Your codebase supports **two authorization models**:

1. **RBAC (Role-Based Access Control)** - Uses roles assigned to users
2. **UMA/ABAC (Permission-Based)** - Uses resource#scope permissions (currently active)

Both models work with Keycloak and can be used together or separately.

## RBAC Architecture

### Components

1. **Role Constants** (`internal/constants/roles.go`)

   - Defines all available roles
   - Groups related roles for easier management

2. **RequireRole Middleware** (`internal/middlewares/auth.go`)

   - Checks if user has required roles
   - Extracts roles from JWT token claims
   - Supports both realm-level and client-level roles

3. **TokenClaims** (`internal/integration/auth/auth.go`)
   - Contains `RealmAccess.Roles` (realm-level roles)
   - Contains `ResourceAccess[clientID].Roles` (client-level roles)

## How RBAC Works

### 1. Role Extraction

Roles are extracted from JWT tokens in two places:

- **Realm Roles**: `token.realm_access.roles`

  - Applied to all clients in the realm
  - Examples: `admin`, `user`, `moderator`

- **Client Roles**: `token.resource_access[client-id].roles`
  - Specific to your application
  - Examples: `user-manager`, `company-creator`

### 2. Role Checking

The `RequireRole` middleware:

1. Extracts `TokenClaims` from request context
2. Collects all roles (realm + client)
3. Checks if user has any of the required roles (OR logic)
4. Returns 403 Forbidden if no match

## Usage Examples

### Basic Usage

```go
import (
    "golang-boilerplate/internal/constants"
    middlewares "golang-boilerplate/internal/middlewares"
)

// Require a single role
userGroup.POST("", handler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin),
)

// Require any of multiple roles (OR logic)
userGroup.PUT("/:id", handler.UpdateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserManager),
)

// Use role groups (spread slice)
userGroup.GET("", handler.GetUsers,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.UserViewRoles...),
)
```

### Group-Level RBAC

Apply RBAC to an entire route group:

```go
adminGroup := v1.Group("/admin")
adminGroup.Use(middlewares.AuthMiddleware(authService))
adminGroup.Use(middlewares.RequireRole(cfg, constants.RoleAdmin))

// All routes in this group require admin role
adminGroup.GET("/stats", adminHandler.GetStats)
adminGroup.POST("/config", adminHandler.UpdateConfig)
```

## Available Roles

### Realm-Level Roles

- `admin` - Full system administrator
- `user` - Regular user
- `moderator` - Content moderator

### Client-Level Roles

**User Management:**

- `user-manager` - Can manage users
- `user-viewer` - Can view users only
- `user-creator` - Can create users
- `user-editor` - Can edit users
- `user-deleter` - Can delete users

**Company Management:**

- `company-manager` - Can manage companies
- `company-viewer` - Can view companies only
- `company-creator` - Can create companies
- `company-editor` - Can edit companies
- `company-deleter` - Can delete companies

### Role Groups

Pre-defined groups for common use cases:

```go
constants.AdminRoles              // [admin]
constants.UserManagementRoles    // [admin, user-manager, user-creator, ...]
constants.UserViewRoles          // [admin, user-manager, user-viewer, ...]
constants.CompanyManagementRoles // [admin, company-manager, ...]
constants.CompanyViewRoles      // [admin, company-manager, company-viewer, ...]
```

## Keycloak Setup

### 1. Create Realm Roles (Optional)

1. Go to Keycloak Admin Console
2. Navigate to **Realm Settings > Roles**
3. Create roles: `admin`, `user`, `moderator`
4. Assign to users at realm level

### 2. Create Client Roles (Recommended)

1. Go to your **Client** in Keycloak
2. Navigate to **Roles** tab
3. Create roles:
   - `user-manager`, `user-viewer`, `user-creator`, `user-editor`, `user-deleter`
   - `company-manager`, `company-viewer`, `company-creator`, `company-editor`, `company-deleter`
4. Enable **"Add to ID token"** and **"Add to access token"** in Client Settings

### 3. Assign Roles to Users

1. Go to **Users > Select User > Role Mappings**
2. Assign **Realm Roles** or **Client Roles** as needed
3. Roles will appear in JWT token:

   ```json
   {
     "realm_access": {
       "roles": ["admin", "user"]
     },
     "resource_access": {
       "your-client-id": {
         "roles": ["user-manager", "company-viewer"]
       }
     }
   }
   ```

### 4. Token Configuration

Ensure your Keycloak client has:

- ✅ **Add to ID token** enabled
- ✅ **Add to access token** enabled
- ✅ **Client authentication** enabled (for service accounts)

## Switching from Permission-Based to RBAC

### Current (Permission-Based)

```go
userGroup.POST("", handler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequirePermission(cfg, authService, "user", "create"),
)
```

### With RBAC

```go
import "golang-boilerplate/internal/constants"

userGroup.POST("", handler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserCreator),
)
```

## Comparison: RBAC vs Permission-Based

| Feature            | RBAC                  | Permission-Based (UMA)            |
| ------------------ | --------------------- | --------------------------------- |
| **Granularity**    | Role-based            | Resource#Scope                    |
| **Keycloak Setup** | Roles                 | Authorization Services + Policies |
| **Token Type**     | Access Token          | RPT (Requesting Party Token)      |
| **Performance**    | Faster (no RPT call)  | Slower (RPT exchange)             |
| **Flexibility**    | Less flexible         | More flexible                     |
| **Use Case**       | Simple role hierarchy | Fine-grained permissions          |

## Best Practices

1. **Use Client Roles** for application-specific permissions
2. **Use Realm Roles** for cross-application roles (admin, user)
3. **Group Related Roles** using role groups in constants
4. **Always Authenticate First** - Use `AuthMiddleware` before `RequireRole`
5. **Document Role Requirements** - Add comments explaining why roles are required

## Example: Complete Route with RBAC

```go
// User routes with RBAC
userGroup := v1.Group("/users")

// Create - Only admins and user managers
userGroup.POST("", userHandler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserManager, constants.RoleUserCreator),
)

// Read - Admins, managers, and viewers
userGroup.GET("/:id", userHandler.GetOneByID,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.UserViewRoles...),
)

// Update - Admins and managers
userGroup.PUT("/:id", userHandler.UpdateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserManager, constants.RoleUserEditor),
)

// Delete - Only admins
userGroup.DELETE("/:id", userHandler.DeleteUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin),
)

// List - All viewers
userGroup.GET("", userHandler.GetUsers,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.UserViewRoles...),
)
```

## Troubleshooting

### Issue: "User not authenticated" error

**Cause**: `TokenClaims` not found in context  
**Solution**: Ensure `AuthMiddleware` is called before `RequireRole`

### Issue: "Insufficient permissions" even with correct role

**Cause**: Role not in token or wrong client ID  
**Solution**:

1. Check JWT token contains the role
2. Verify client ID matches Keycloak configuration
3. Ensure role is assigned to user in Keycloak

### Issue: Client roles not appearing in token

**Cause**: Token configuration not enabled  
**Solution**: Enable "Add to access token" in Keycloak client settings

## See Also

- `examples/rbac_example.go` - Complete working examples
- `internal/middlewares/auth.go` - Middleware implementation
- `internal/constants/roles.go` - Role definitions

# RBAC Quick Start Guide

## What Changed

1. ✅ **Fixed `RequireRole` middleware** - Now correctly extracts roles from `TokenClaims`
2. ✅ **Created role constants** - `internal/constants/roles.go`
3. ✅ **Added examples** - `examples/rbac_example.go`
4. ✅ **Updated router** - Added RBAC examples in comments

## Quick Usage

### Step 1: Import constants

```go
import "golang-boilerplate/internal/constants"
```

### Step 2: Apply RBAC to routes

```go
// Single role
userGroup.POST("", handler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin),
)

// Multiple roles (OR logic - user needs ANY of these)
userGroup.PUT("/:id", handler.UpdateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserManager),
)

// Role groups
userGroup.GET("", handler.GetUsers,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.UserViewRoles...),
)
```

### Step 3: Configure Keycloak

1. Create client roles in Keycloak (e.g., `user-manager`, `company-viewer`)
2. Assign roles to users
3. Enable "Add to access token" in client settings

## Available Roles

See `internal/constants/roles.go` for all available roles and role groups.

## Full Documentation

See `docs/RBAC_GUIDE.md` for complete documentation.

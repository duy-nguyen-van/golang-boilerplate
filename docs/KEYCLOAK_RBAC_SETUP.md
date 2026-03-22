# Keycloak RBAC Setup Guide (Client-Specific)

This guide walks you through setting up Role-Based Access Control (RBAC) in Keycloak for a specific client application. This approach uses **client-level roles**, which are scoped to your application and provide better isolation and security.

## Prerequisites

- Keycloak is running (via `docker-compose up` or standalone)
- Access to Keycloak Admin Console (default: <http://localhost:8080>)
- Admin credentials (default: `admin` / `admin`)

## Step 1: Access Keycloak Admin Console

1. Open your browser and navigate to: `http://localhost:8080`
2. Click **Administration Console**
3. Log in with admin credentials:
   - Username: `admin` (or from `KEYCLOAK_ADMIN` env var)
   - Password: `admin` (or from `KEYCLOAK_ADMIN_PASSWORD` env var)

## Step 2: Select Your Realm

1. In the top-left dropdown, select your realm (default: `master` or your custom realm)
2. If you need to create a new realm:
   - Click the realm dropdown → **Create Realm**
   - Enter realm name (e.g., `golang-boilerplate`)
   - Click **Create**

## Step 3: Configure Your Client

Before creating client roles, ensure your client is properly configured. Client roles are specific to your application and provide better security isolation.

### 3.1 Navigate to Clients

1. In the left sidebar, click **Clients**
2. Find your client (the one matching `KEYCLOAK_CLIENT_ID` from your `.env` file)
3. If it doesn't exist, click **Create client**:
   - Client type: **OpenID Connect**
   - Client ID: (your `KEYCLOAK_CLIENT_ID` value)
   - Click **Next**
   - Capability config: Enable **Client authentication** (for service accounts)
   - Click **Next**
   - Login settings: Add your redirect URIs
   - Click **Save**

### 3.2 Configure Client Settings

1. Click on your client to open its settings
2. Go to the **Settings** tab
3. Ensure these settings are configured:

   **Access settings:**

   - ✅ **Client authentication**: ON (if using service accounts)
   - ✅ **Authorization**: ON (if using UMA/permissions)
   - **Access token lifespan**: 5 minutes (default)

   **Login settings:**

   - **Valid redirect URIs**: Add your application URLs
   - **Web origins**: Add your application origins

   **Token settings:**

   - ✅ **Add to ID token**: ON (important for roles)
   - ✅ **Add to access token**: ON (important for roles)
   - ✅ **Add to userinfo**: ON (optional)

4. Click **Save**

## Step 4: Create Client-Level Roles

Client roles are specific to your application and provide better security isolation. All roles will be created under your client.

### 4.1 Navigate to Client Roles

1. In your client settings, click the **Roles** tab
2. You'll see the **Client roles** section

### 4.2 Create User Management Roles

Click **Create role** and create each role:

**Role: `user-manager`**

- Role name: `user-manager`
- Description: `Can manage users`
- Click **Save**

**Role: `user-viewer`**

- Role name: `user-viewer`
- Description: `Can view users only`
- Click **Save**

**Role: `user-creator`**

- Role name: `user-creator`
- Description: `Can create users`
- Click **Save**

**Role: `user-editor`**

- Role name: `user-editor`
- Description: `Can edit users`
- Click **Save**

**Role: `user-deleter`**

- Role name: `user-deleter`
- Description: `Can delete users`
- Click **Save**

### 4.3 Create Company Management Roles

Continue creating roles:

**Role: `company-manager`**

- Role name: `company-manager`
- Description: `Can manage companies`
- Click **Save**

**Role: `company-viewer`**

- Role name: `company-viewer`
- Description: `Can view companies only`
- Click **Save**

**Role: `company-creator`**

- Role name: `company-creator`
- Description: `Can create companies`
- Click **Save**

**Role: `company-editor`**

- Role name: `company-editor`
- Description: `Can edit companies`
- Click **Save**

**Role: `company-deleter`**

- Role name: `company-deleter`
- Description: `Can delete companies`
- Click **Save**

## Step 5: Assign Client Roles to Users

Now assign client roles to your users so they can access protected resources.

### 5.1 Navigate to Users

1. In the left sidebar, click **Users**
2. Find the user you want to assign roles to (or create a new user)
3. Click on the username to open user details

### 5.2 Assign Client Roles

1. Click the **Role mapping** tab
2. Click **Assign role**
3. **Important**: Check **Filter by clients** checkbox
4. Select your client from the dropdown (the one matching `KEYCLOAK_CLIENT_ID`)
5. Select the client roles you want to assign (e.g., `user-manager`, `company-viewer`)
6. Click **Assign**

### 5.3 Verify Role Assignment

After assigning roles, you should see them listed in the **Role mapping** tab under **Client roles** for your specific client:

- **Client roles**: Shows assigned client-level roles for your client
- Verify the roles appear under the correct client ID

## Step 6: Verify Token Configuration

Ensure client roles appear in JWT tokens.

### 6.1 Check Client Mapper Settings

1. Go to your **Client** → **Client scopes** tab
2. Click on **dedicated** (or your client's scope)
3. Go to the **Mappers** tab
4. Ensure the **Client roles mapper** exists:

   **Client roles mapper (Required):**

   - Name: `client roles` (or similar)
   - Mapper type: `User Client Role`
   - Client ID: (your client ID - must match `KEYCLOAK_CLIENT_ID`)
   - Token Claim Name: `resource_access.${client_id}.roles`
   - Add to ID token: ✅ ON
   - Add to access token: ✅ ON (critical for RBAC)
   - Add to userinfo: ✅ ON (optional)

5. If the mapper doesn't exist, click **Create mapper** → **By configuration** → Select **User Client Role** and configure it as above

### 6.2 Test Token

1. Get an access token for a user with assigned client roles
2. Decode the JWT token at <https://jwt.io>
3. Verify the token contains client roles in `resource_access`:

```json
{
  "resource_access": {
    "your-client-id": {
      "roles": ["user-manager", "company-viewer"]
    }
  }
}
```

**Note**: For client-specific RBAC, your application will read roles from `resource_access.{client-id}.roles`, not from `realm_access.roles`.

## Step 7: Update Your Application Code

Once client roles are configured in Keycloak, update your routes to use client-specific RBAC.

### 7.1 Switch from Permission-Based to RBAC

In `cmd/server/routes/router.go`, replace `RequirePermission` with `RequireRole`:

```go
import "golang-boilerplate/internal/constants"

// Before (Permission-Based)
userGroup.POST("", userHandler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequirePermission(cfg, authService, "user", "create"),
)

// After (Client-Specific RBAC)
userGroup.POST("", userHandler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleUserManager, constants.RoleUserCreator),
)
```

**Note**: The `RequireRole` middleware automatically reads roles from `resource_access.{client-id}.roles` based on your `KEYCLOAK_CLIENT_ID` configuration.

### 7.2 Example Route Configuration

```go
// User routes with client-specific RBAC
userGroup := v1.Group("/users")

// Create - User managers and creators
userGroup.POST("", userHandler.CreateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleUserManager, constants.RoleUserCreator),
)

// Read - Managers and viewers
userGroup.GET("/:id", userHandler.GetOneByID,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.UserViewRoles...),
)

// Update - Managers and editors
userGroup.PUT("/:id", userHandler.UpdateUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleUserManager, constants.RoleUserEditor),
)

// Delete - Managers and deleters
userGroup.DELETE("/:id", userHandler.DeleteUser,
    middlewares.AuthMiddleware(authService),
    middlewares.RequireRole(cfg, constants.RoleUserManager, constants.RoleUserDeleter),
)
```

## Step 8: Test RBAC

### 8.1 Test with Different Client Roles

1. **Create test users** and assign different client roles:

   - User A: `user-manager` + `company-manager` (client roles)
   - User B: `user-viewer` only (client role)
   - User C: `company-manager` only (client role)

2. **Get access tokens** for each user

3. **Test API endpoints**:
   - User A should access user and company management endpoints
   - User B should only access user GET/view endpoints
   - User C should only access company management endpoints

### 8.2 Verify Authorization

- ✅ **200 OK**: User has required client role
- ❌ **403 Forbidden**: User lacks required client role
- ❌ **401 Unauthorized**: Invalid or missing token

### 8.3 Verify Token Claims

Decode each user's JWT token and verify:

- Roles appear in `resource_access.{your-client-id}.roles`
- No roles in `realm_access.roles` (unless you're also using realm roles)
- Client ID matches your `KEYCLOAK_CLIENT_ID` configuration

## Troubleshooting

### Issue: Client roles not appearing in token

**Solution:**

1. Check client settings: **Add to access token** must be ✅ ON (Step 3.2)
2. Verify client roles mapper exists and is configured correctly (Step 6.1)
3. Ensure roles are assigned to the user under the correct client (Step 5.2)
4. Verify the mapper's Client ID matches your `KEYCLOAK_CLIENT_ID` exactly
5. Request a new token after role assignment (tokens are cached)

### Issue: "Insufficient permissions" error

**Solution:**

1. Verify the role name matches exactly (case-sensitive) - check `internal/constants/roles.go`
2. Check if role is in `resource_access.{client-id}.roles` (not `realm_access.roles`)
3. Ensure `KEYCLOAK_CLIENT_ID` in your `.env` matches the Keycloak client ID exactly
4. Decode the JWT token at <https://jwt.io> and verify:
   - Roles are present in `resource_access.{your-client-id}.roles`
   - Client ID in the token matches your configuration
5. Verify the user has the role assigned in Keycloak (Step 5.2)

### Issue: Client roles not found in application

**Solution:**

1. Ensure roles are created under the correct client (not realm roles)
2. Verify client ID in your application (`KEYCLOAK_CLIENT_ID`) matches Keycloak client ID
3. Check token mapper includes client roles and Client ID is set correctly
4. Verify the `RequireRole` middleware is reading from `resource_access.{client-id}.roles`

### Issue: Wrong client ID in token

**Solution:**

1. Verify `KEYCLOAK_CLIENT_ID` in your `.env` file matches the client ID in Keycloak
2. Check the client roles mapper has the correct Client ID configured
3. Ensure you're using the correct client when requesting tokens

## Quick Reference: Client Role Names

All roles are client-specific and should be created under your client:

### User Management Roles

- `user-manager` - Can manage users
- `user-viewer` - Can view users only
- `user-creator` - Can create users
- `user-editor` - Can edit users
- `user-deleter` - Can delete users

### Company Management Roles

- `company-manager` - Can manage companies
- `company-viewer` - Can view companies only
- `company-creator` - Can create companies
- `company-editor` - Can edit companies
- `company-deleter` - Can delete companies

**Important**: These roles must be created as **client roles** (not realm roles) under your specific client.

## Next Steps

1. ✅ Client configured in Keycloak
2. ✅ Client roles created under your client
3. ✅ Client roles assigned to users
4. ✅ Client roles mapper configured and verified
5. ✅ Token contains roles in `resource_access.{client-id}.roles`
6. ✅ Application code updated to use client-specific RBAC
7. ✅ Tested with different user roles

## Additional Resources

- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [RBAC Guide](./RBAC_GUIDE.md) - Application-side RBAC documentation
- [RBAC Quick Start](./RBAC_QUICK_START.md) - Quick reference
- [RBAC Examples](../examples/rbac_example.go) - Code examples

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Verify your Keycloak client configuration matches this guide
3. Ensure all roles are created as **client roles** (not realm roles)
4. Verify `KEYCLOAK_CLIENT_ID` matches your Keycloak client ID exactly
5. Review the application logs for detailed error messages
6. Decode JWT tokens at <https://jwt.io> to verify client role claims in `resource_access.{client-id}.roles`

## Why Client-Specific RBAC?

Using client-specific roles provides:

- **Better isolation**: Roles are scoped to your application only
- **Security**: No cross-application role leakage
- **Flexibility**: Each client can have its own role structure
- **Clarity**: Clear separation between different applications in the same realm

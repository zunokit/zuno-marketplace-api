# Authentication Implementation Gaps

**Analysis Date:** 2025-11-16
**Compared:** Current implementation vs `authen-flow.md` specification

---

## ✅ Implemented Features

### 1. Sign-in (SIWE) with Set-Cookie ✅
**Location:** `services/graphql-gateway/graph/schema.resolvers.go:19-44`

```graphql
mutation VerifySiwe($accountId: String!, $message: String!, $signature: String!) {
  verifySiwe(accountId: $accountId, message: $message, signature: $signature) {
    accessToken
    expiresAt
    userId
    address
    chainId
  }
}
```

**Implementation:**
- ✅ Calls auth-service `VerifySiwe` gRPC
- ✅ Sets `refresh_token` HttpOnly cookie
- ✅ Returns `accessToken` in response body
- ✅ Does NOT return `refreshToken` in body (security best practice)

**Status:** Complete ✅

---

### 2. Refresh Session ✅
**Location:** `services/graphql-gateway/graph/schema.resolvers.go:46-88`

```graphql
mutation RefreshSession($refreshToken: String!, $userAgent: String, $ipAddress: String) {
  refreshSession(refreshToken: $refreshToken, userAgent: $userAgent, ipAddress: $ipAddress) {
    accessToken
    expiresAt
    userId
  }
}
```

**Implementation:**
- ✅ Accepts `refreshToken` as parameter
- ✅ Falls back to cookie if parameter not provided
- ✅ Updates `refresh_token` cookie with new token
- ✅ Does NOT return `refreshToken` in body

**Status:** Complete ✅

---

### 3. Logout (Revoke Session) ✅
**Location:** `services/graphql-gateway/graph/schema.resolvers.go:90-107`

```graphql
mutation RevokeSession($sessionId: ID!) {
  revokeSession(sessionId: $sessionId)
}
```

**Implementation:**
- ✅ Calls auth-service `RevokeSession` gRPC
- ✅ Clears `refresh_token` cookie (Max-Age=-1)

**Status:** Complete ✅

---

### 4. Auth Middleware ✅
**Location:** `services/graphql-gateway/internal/middleware/auth.go`

**Implementation:**
- ✅ Validates JWT from `Authorization: Bearer <token>`
- ✅ Adds user claims to context
- ✅ Non-blocking (individual resolvers check auth requirements)
- ✅ `RequireAuth()` helper for protected resolvers

**Status:** Complete ✅

---

## ❌ Missing Features

### 1. Auto-retry on 401 ❌
**Spec:** `authen-flow.md` - Flow #3

```mermaid
FE → GQL: request with expired access token
GQL → FE: 401 UNAUTHENTICATED
FE → GQL: refreshSession() using cookie
GQL → FE: new accessToken
FE → GQL: replay original request with new token
```

**Current Status:**
- ❌ **NOT implemented** - This is a client-side/frontend feature
- Backend provides necessary endpoints (`refreshSession`)
- Frontend needs to implement retry logic

**Action Required:**
- Frontend: Implement GraphQL error interceptor
- Detect 401 errors
- Call `refreshSession` mutation
- Retry original request with new `accessToken`

**Priority:** 🔴 High (UX critical)

---

### 2. Silent Session Restore ❌
**Spec:** `authen-flow.md` - Flow #4

```mermaid
FE → GQL: me() on app load (no Authorization header)
GQL: Check for refresh_token cookie
GQL → AUTH: RefreshSession if cookie exists
GQL → FE: {user, accessToken} or {user: null}
```

**Current Status:**
- ❌ `me` query requires authentication (line 183-206)
- ❌ Cannot call `me()` without valid access token
- ❌ No automatic cookie-based restore

**Current Implementation:**
```go
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    // Extract user claims from context
    claims, err := middleware.RequireAuth(ctx)  // ❌ This fails if no token
    if err != nil {
        return nil, err
    }
    // ...
}
```

**Required Changes:**
```go
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    // 1. Try to get claims from existing access token
    claims, ok := middleware.GetUserClaims(ctx)

    // 2. If no valid access token, try refresh from cookie
    if !ok {
        if req, ok := appcontext.GetHTTPRequest(ctx); ok {
            if cookieToken, err := cookie.GetRefreshTokenFromCookie(req); err == nil {
                // Call RefreshSession internally
                refreshResp, err := r.AuthClient.RefreshSession(ctx, &pb.RefreshSessionRequest{
                    RefreshToken: cookieToken,
                })
                if err == nil {
                    // Set new cookie & return user
                    claims = &middleware.UserClaims{UserID: refreshResp.UserId}
                }
            }
        }
    }

    // 3. If still no auth, return nil user (not error)
    if claims == nil {
        return nil, nil  // Frontend shows "Connect Wallet" button
    }

    // 4. Fetch user data
    // ...
}
```

**Action Required:**
- Modify `me` query to support optional authentication
- Auto-refresh from cookie if no access token provided
- Return `null` user instead of error when not authenticated

**Priority:** 🔴 High (UX critical for SSR/page refresh)

---

### 3. WebSocket Support for Subscriptions ❌
**Spec:** `authen-flow.md` - Flow #6

```mermaid
FE → WS: connection_init {accessToken}
WS → FE: connection_ack
... subscription data streaming ...
FE → WS: connection_update {newAccessToken} when refreshed
```

**Current Status:**
- ❌ No WebSocket/Subscription implementation
- ❌ GraphQL schema has no `subscription` type
- ❌ No WS authentication handler

**Action Required:**
1. Add `subscription` type to GraphQL schema
2. Implement WebSocket transport (e.g., `graphql-ws` protocol)
3. Add WS authentication:
   - Accept `accessToken` in `connection_init` payload
   - Validate JWT
   - Support token refresh via `connection_update` or reconnect
4. Example subscriptions:
   ```graphql
   type Subscription {
     onCollectionCreated(userId: ID!): Collection!
     onNftMinted(collectionAddress: String!): NFT!
     onOfferReceived(userId: ID!): Offer!
   }
   ```

**Priority:** 🟡 Medium (needed for real-time features)

---

### 4. RevokeSessionByRefreshToken not exposed ❌
**Spec:** Auth proto has `RevokeSessionByRefreshToken` but not in GraphQL

**Current Status:**
- ✅ Protobuf: `RevokeSessionByRefreshToken` exists in `auth.proto`
- ❌ GraphQL: Only `revokeSession(sessionId)` exposed
- ❌ Cannot logout using refresh token from cookie

**Current Schema:**
```graphql
type Mutation {
  revokeSession(sessionId: ID!): Boolean!  # Requires session ID
}
```

**Required Schema:**
```graphql
type Mutation {
  revokeSession(sessionId: ID): Boolean!  # sessionId optional
  # OR add separate mutation:
  logout: Boolean!  # Uses refresh_token from cookie
}
```

**Required Implementation:**
```go
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
    // Get refresh token from cookie
    var token string
    if req, ok := appcontext.GetHTTPRequest(ctx); ok {
        if cookieToken, err := cookie.GetRefreshTokenFromCookie(req); err == nil {
            token = cookieToken
        }
    }

    if token == "" {
        return false, fmt.Errorf("no active session")
    }

    // Revoke by refresh token
    res, err := r.AuthClient.RevokeSessionByRefreshToken(ctx, &pb.RevokeSessionByRefreshTokenRequest{
        RefreshToken: token,
    })
    if err != nil {
        return false, err
    }

    // Clear cookie
    if w, ok := appcontext.GetHTTPResponse(ctx); ok {
        cookie.ClearRefreshTokenCookie(w)
    }

    return res.Success, nil
}
```

**Action Required:**
- Add `logout` mutation to schema
- Implement resolver using `RevokeSessionByRefreshToken`

**Priority:** 🟢 Low (current `revokeSession` works, but less convenient)

---

## ⚠️ Security Issues

### 1. Cookie Security Settings ⚠️
**Location:** `services/graphql-gateway/internal/cookie/cookie.go:14-23`

**Current:**
```go
http.SetCookie(w, &http.Cookie{
    Name:     RefreshTokenCookie,
    Value:    token,
    Path:     "/",
    MaxAge:   maxAge,
    HttpOnly: true,       // ✅ Good
    Secure:   false,      // ⚠️ INSECURE for production
    SameSite: http.SameSiteLaxMode,  // ⚠️ Spec requires Strict
})
```

**Spec Requirements:**
```
HttpOnly: true
Secure: true
SameSite: Strict
Path: /
Max-Age: 30 days
```

**Issues:**
1. ❌ `Secure: false` - Allows cookie over HTTP (vulnerable to MITM)
2. ⚠️ `SameSite: Lax` - Less strict than spec (allows some cross-site requests)

**Recommended Fix:**
```go
func SetRefreshTokenCookie(w http.ResponseWriter, token string, maxAge int) {
    // Detect environment
    secure := os.Getenv("ENVIRONMENT") != "development"

    http.SetCookie(w, &http.Cookie{
        Name:     RefreshTokenCookie,
        Value:    token,
        Path:     "/",
        MaxAge:   maxAge,
        HttpOnly: true,
        Secure:   secure,  // true in production
        SameSite: http.SameSiteStrictMode,  // Strict for better security
    })
}
```

**Action Required:**
- Make `Secure` environment-aware
- Change `SameSite` to `Strict`
- Add environment detection

**Priority:** 🔴 High (security vulnerability)

---

### 2. GraphQL Schema Type Mismatch ⚠️
**Location:** `services/graphql-gateway/schema.graphqls:118-122`

**Current Schema:**
```graphql
refreshSession(
  refreshToken: String!   # ⚠️ Marked as required
  userAgent: String
  ipAddress: String
): RefreshResponse!
```

**Actual Behavior:**
- Code accepts empty `refreshToken` and falls back to cookie
- Schema says it's required (`!`)
- **Mismatch** between schema and implementation

**Recommended Fix:**
```graphql
refreshSession(
  refreshToken: String    # Make optional (no !)
  userAgent: String
  ipAddress: String
): RefreshResponse!
```

**Or add comment:**
```graphql
"""
Refresh the current session.
- refreshToken: Optional. If not provided, will use refresh_token from HttpOnly cookie.
- userAgent: Optional. User agent string for session tracking.
- ipAddress: Optional. IP address for session tracking.
"""
refreshSession(
  refreshToken: String
  userAgent: String
  ipAddress: String
): RefreshResponse!
```

**Action Required:**
- Make `refreshToken` optional in schema
- Add documentation

**Priority:** 🟡 Medium (works but misleading)

---

### 3. AuthResponse Returns Empty refreshToken ⚠️
**Location:** `services/graphql-gateway/graph/schema.resolvers.go:38`

**Current:**
```go
return &model.AuthResponse{
    AccessToken:  res.AccessToken,
    RefreshToken: "",  // ⚠️ Always empty
    ExpiresAt:    res.ExpiresAt,
    UserID:       res.UserId,
    Address:      res.Address,
    ChainID:      res.ChainId,
}, nil
```

**Schema:**
```graphql
type AuthResponse {
  accessToken: String!
  refreshToken: String!   # ⚠️ Non-null but always empty
  expiresAt: String!
  userId: String!
  address: String!
  chainId: String!
}
```

**Issue:**
- Schema declares `refreshToken: String!` (non-null)
- Implementation always returns empty string
- Violates GraphQL contract

**Recommended Fix:**

**Option 1: Make nullable in schema**
```graphql
type AuthResponse {
  accessToken: String!
  refreshToken: String    # Nullable (no !)
  expiresAt: String!
  userId: String!
  address: String!
  chainId: String!
}
```

**Option 2: Remove from schema entirely**
```graphql
type AuthResponse {
  accessToken: String!
  # refreshToken removed - use cookie only
  expiresAt: String!
  userId: String!
  address: String!
  chainId: String!
}
```

**Action Required:**
- Choose Option 2 (remove field) for best security
- Update schema

**Priority:** 🟡 Medium (technical debt)

---

## 📋 Summary

### Completed Features: 4/7
| Feature | Status | Priority |
|---------|--------|----------|
| SIWE Sign-in + Cookie | ✅ Complete | - |
| Refresh Session | ✅ Complete | - |
| Logout | ✅ Complete | - |
| Auth Middleware | ✅ Complete | - |
| Auto-retry 401 | ❌ Missing | 🔴 High |
| Silent Restore | ❌ Missing | 🔴 High |
| WebSocket Auth | ❌ Missing | 🟡 Medium |

### Security Issues: 3
| Issue | Severity | Priority |
|-------|----------|----------|
| Cookie Secure: false | 🔴 Critical | 🔴 High |
| SameSite: Lax vs Strict | 🟡 Medium | 🟡 Medium |
| Schema type mismatches | 🟡 Medium | 🟡 Medium |

---

## 🎯 Action Plan

### Phase 1: Security Fixes (Immediate)
1. ✅ Fix cookie security settings
   - Environment-aware `Secure` flag
   - Change to `SameSite: Strict`
2. ✅ Fix schema type mismatches
   - Make `refreshToken` optional in mutations
   - Remove `refreshToken` from `AuthResponse`

### Phase 2: Core UX Features (Sprint 1)
3. ✅ Implement silent session restore
   - Modify `me` query for optional auth
   - Auto-refresh from cookie
   - Return `null` instead of error when unauthenticated
4. ✅ Add `logout` mutation
   - Use refresh token from cookie
   - Cleaner UX than `revokeSession(sessionId)`

### Phase 3: Advanced Features (Sprint 2)
5. ✅ WebSocket/Subscription support
   - Add subscription transport
   - Implement WS authentication
   - Add real-time subscriptions

### Phase 4: Frontend (Parallel)
6. ✅ Implement auto-retry interceptor
   - GraphQL error handling
   - Automatic refresh on 401
   - Replay failed requests

---

## 📝 Code Examples

### Example: Silent Session Restore Implementation

**GraphQL Schema:**
```graphql
type Query {
  """
  Get current authenticated user.
  If no access token provided, attempts to restore session from cookie.
  Returns null if not authenticated (not an error).
  """
  me: User  # Nullable
}
```

**Resolver:**
```go
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    var userID string

    // Try existing access token first
    if claims, ok := middleware.GetUserClaims(ctx); ok {
        userID = claims.UserID
    } else {
        // Try silent refresh from cookie
        if req, ok := appcontext.GetHTTPRequest(ctx); ok {
            if token, err := cookie.GetRefreshTokenFromCookie(req); err == nil {
                resp, err := r.AuthClient.RefreshSession(ctx, &pb.RefreshSessionRequest{
                    RefreshToken: token,
                })
                if err == nil {
                    userID = resp.UserId
                    // Set new cookie
                    if w, ok := appcontext.GetHTTPResponse(ctx); ok {
                        cookie.SetRefreshTokenCookie(w, resp.RefreshToken, cookie.MaxAge)
                    }
                }
            }
        }
    }

    // Not authenticated - return nil, not error
    if userID == "" {
        return nil, nil
    }

    // Fetch user
    res, err := r.UserClient.GetUser(ctx, &pb.GetUserRequest{UserId: userID})
    if err != nil {
        return nil, err
    }

    return &model.User{
        ID:        res.User.Id,
        Status:    res.User.Status,
        CreatedAt: res.User.CreatedAt,
        Profile:   profileFromProto(res.Profile),
    }, nil
}
```

**Frontend Usage:**
```typescript
// App.tsx - on mount
const { data } = useQuery(ME_QUERY);

if (data?.me) {
  // User logged in - silently restored session
  console.log('Logged in as:', data.me);
} else {
  // Not logged in - show Connect button
  console.log('Not authenticated');
}
```

---

## 📚 References

- **Spec:** `authen-flow.md`
- **Implementation:** `services/graphql-gateway/`
- **Auth Service:** `services/auth-service/`
- **Cookie Utils:** `services/graphql-gateway/internal/cookie/`

---

**Next Review:** After Phase 1-2 implementation

# Load Balancer with JWT Authentication

## Overview
This Go-based load balancer routes incoming requests to backend servers based on a round-robin strategy, with special routing rules for admin users. It also implements JWT authentication to extract user roles from tokens.

## Functions Documentation

### `GetNextServer(role string) string`
**Description:**
Determines the next backend server to route a request to.

**Parameters:**
- `role` (string): The role extracted from the JWT token.

**Returns:**
- `string`: The URL of the backend server to which the request should be forwarded.

**Behavior:**
- If the role is "admin", the request is always sent to `backend1.local`.
- Otherwise, the request is distributed in a round-robin manner across all backends.

---

### `HandleRequest(w http.ResponseWriter, r *http.Request)`
**Description:**
Handles incoming HTTP requests and forwards them to the selected backend server.

**Parameters:**
- `w` (http.ResponseWriter): The response writer to send data back to the client.
- `r` (http.Request): The incoming HTTP request.

**Behavior:**
- Extracts the user role from the request header.
- Determines the appropriate backend server using `GetNextServer`.
- Creates a reverse proxy to forward the request to the selected backend.
- Logs the request routing details.

---

### `parseJWT(tokenString string) (string, error)`
**Description:**
Parses a JWT token to extract the user role.

**Parameters:**
- `tokenString` (string): The JWT token extracted from the request header.

**Returns:**
- `string`: The role of the user.
- `error`: An error if the token is invalid or does not contain the role claim.

**Behavior:**
- Decodes and validates the JWT using a predefined secret key.
- Extracts the `role` claim from the token payload.
- Returns an error if the token is invalid or missing the role claim.

---

### `JwtMiddleware(next http.Handler) http.Handler`
**Description:**
Middleware to enforce JWT authentication on incoming requests.

**Parameters:**
- `next` (http.Handler): The next HTTP handler to call if authentication is successful.

**Returns:**
- `http.Handler`: A wrapped handler that checks JWT authentication before passing requests forward.

**Behavior:**
- Checks for an `Authorization` header in the request.
- Extracts and verifies the JWT token.
- If valid, sets the `Role` header in the request and forwards it to the next handler.
- Returns a `401 Unauthorized` response if the token is missing or invalid.

---

### `main()`
**Description:**
Initializes the HTTP server and registers the request handler with JWT authentication middleware.

**Behavior:**
- Wraps `HandleRequest` with `JwtMiddleware`.
- Starts an HTTP server listening on port 8080.
- Logs server startup and handles errors.
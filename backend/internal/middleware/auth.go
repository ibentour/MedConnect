// Package middleware provides Gin middleware for authentication (JWT),
// role-based access control (RBAC), and audit logging for MedConnect Oriental.
package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"medconnect-oriental/backend/internal/models"
)

// ──────────────────────────────────────────────────────────────────────
// JWT Claims
// ──────────────────────────────────────────────────────────────────────

// Claims represents the JWT payload for authenticated users.
type Claims struct {
	UserID       uuid.UUID   `json:"user_id"`
	Username     string      `json:"username"`
	Role         models.Role `json:"role"`
	DeptID       *uuid.UUID  `json:"dept_id,omitempty"`
	FacilityName string      `json:"facility_name"`
	jwt.RegisteredClaims
}

// ──────────────────────────────────────────────────────────────────────
// Token Generation
// ──────────────────────────────────────────────────────────────────────

// GenerateToken creates a signed JWT for an authenticated user.
// Token expires after 24 hours.
func GenerateToken(user models.User, jwtSecret string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		DeptID:       user.DeptID,
		FacilityName: user.FacilityName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "medconnect-oriental",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ──────────────────────────────────────────────────────────────────────
// JWT Authentication Middleware
// ──────────────────────────────────────────────────────────────────────

// JWTAuthMiddleware validates the Bearer token from the Authorization header
// or from the "token" query parameter, parses the claims, and injects them into the Gin context.
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""

		// First try to get token from Authorization header
		if authHeader != "" {
			// Expect "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Authorization header must be in format: Bearer <token>",
				})
				return
			}
			tokenString = parts[1]
		} else {
			// Try to get token from query parameter (for direct browser viewing of attachments)
			tokenString = c.Query("token")
			if tokenString == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Authorization header is required",
				})
				return
			}
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token is not valid",
			})
			return
		}

		// Inject claims into context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("dept_id", claims.DeptID)
		c.Set("facility_name", claims.FacilityName)
		c.Set("claims", claims)

		c.Next()
	}
}

// ──────────────────────────────────────────────────────────────────────
// RBAC Middleware
// ──────────────────────────────────────────────────────────────────────

// RBACMiddleware restricts access to routes based on user roles.
// Only users with one of the specified roles can proceed.
func RBACMiddleware(allowedRoles ...models.Role) gin.HandlerFunc {
	roleSet := make(map[models.Role]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		userRole, ok := roleVal.(models.Role)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid role in token",
			})
			return
		}

		if !roleSet[userRole] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Access denied: role '%s' is not authorized for this resource", userRole),
			})
			return
		}

		c.Next()
	}
}

// ──────────────────────────────────────────────────────────────────────
// Context Helpers
// ──────────────────────────────────────────────────────────────────────

// GetUserIDFromContext extracts the authenticated user's UUID from the Gin context.
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetUsernameFromContext extracts the authenticated user's username from the Gin context.
func GetUsernameFromContext(c *gin.Context) string {
	val, _ := c.Get("username")
	username, _ := val.(string)
	return username
}

// GetUserRoleFromContext extracts the authenticated user's role from the Gin context.
func GetUserRoleFromContext(c *gin.Context) models.Role {
	val, _ := c.Get("user_role")
	role, _ := val.(models.Role)
	return role
}

// GetDeptIDFromContext extracts the authenticated user's department ID from the Gin context.
func GetDeptIDFromContext(c *gin.Context) *uuid.UUID {
	val, exists := c.Get("dept_id")
	if !exists {
		return nil
	}
	id, ok := val.(*uuid.UUID)
	if !ok {
		return nil
	}
	return id
}

// ──────────────────────────────────────────────────────────────────────
// Token Validation Utility
// ──────────────────────────────────────────────────────────────────────

// ValidateTokenString validates a JWT token string and returns the claims.
// This is useful for endpoints that need to accept tokens via query parameters.
func ValidateTokenString(tokenString string, jwtSecret string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is required")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}

package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security-related HTTP headers to all responses.
// These headers help protect against common web vulnerabilities:
// - XSS attacks
// - Clickjacking
// - MIME type sniffing
// - Information leakage via referrer
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing - forces browser to respect declared Content-Type
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking by disallowing the page to be embedded in iframes
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS filter in browsers (legacy but still useful for older browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control referrer information sent with requests
		// strict-origin-when-cross-origin: Send full URL for same-origin requests,
		// only origin for cross-origin requests, and nothing for downgrades (HTTPS->HTTP)
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Prevent DNS prefetching to reduce information leakage
		c.Header("X-DNS-Prefetch-Control", "off")

		// Disable browser features that could be exploited
		// This is a restrictive policy - adjust based on your application's needs
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy - restrictive default
		// Note: For a React SPA, you may need to adjust this to allow:
		// - 'unsafe-inline' for scripts (or use nonces/hashes)
		// - 'unsafe-eval' if using eval (not recommended)
		// - connect-src for API calls
		// - img-src for images
		// This is a baseline policy - customize based on your specific needs
		csp := "default-src 'self'; " +
			"script-src 'self'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		// Note: HSTS (Strict-Transport-Security) should be added when running behind HTTPS
		// Uncomment and adjust max-age when deploying with TLS
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}

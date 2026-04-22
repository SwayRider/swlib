package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	jwt5 "github.com/golang-jwt/jwt/v5"
	"github.com/swayrider/swlib/jwt"
	log "github.com/swayrider/swlib/logger"
)

// Test RSA key pair for JWT tests
var (
	testPrivateKeyPEM string
	testPublicKeyPEM  string
	testLogger        *log.Logger
)

func init() {
	// Generate test RSA key pair
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	testPrivateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	testPublicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	// Configure JWT
	jwt.Configure("TestIssuer", "TestAudience")

	// Create test logger
	testLogger = log.New(log.WithComponent("security-test"))
}

// Helper to create test tokens
func createTestUserToken(userID string, isAdmin bool, accountLevel string, emailVerified bool) string {
	email := "test@example.com"
	openID := &jwt.OpenIDClaims{
		Email:         &email,
		EmailVerified: &emailVerified,
	}
	swClaims := jwt.NewSwayRiderUserClaims(isAdmin, accountLevel)
	_, token, _, _ := jwt.GenerateToken(userID, openID, swClaims, testPrivateKeyPEM, 15*time.Minute)
	return string(token)
}

func createTestServiceToken(serviceID string, scopes []string) string {
	swClaims := jwt.NewSwayRiderServiceClaims(scopes)
	_, token, _, _ := jwt.GenerateToken(serviceID, nil, swClaims, testPrivateKeyPEM, 15*time.Minute)
	return string(token)
}

func getTestPublicKeys() ([]string, error) {
	return []string{testPublicKeyPEM}, nil
}

func getTestPublicKeysError() ([]string, error) {
	return nil, errors.New("failed to get keys")
}

// Helper to reset endpoint profiles between tests
func resetEndpointProfiles() {
	endpointProfiles = make(map[string]EndpointProfile)
}

// =============================================================================
// Context Tests
// =============================================================================

func TestGetClaims(t *testing.T) {
	claims := &jwt.Claims{
		RegisteredClaims: jwt5.RegisteredClaims{
			Subject: "user-123",
		},
	}

	ctx := context.WithValue(context.Background(), ClaimsKey, claims)

	result, ok := GetClaims(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if result.Subject != "user-123" {
		t.Errorf("expected subject 'user-123', got '%s'", result.Subject)
	}
}

func TestGetClaims_NotPresent(t *testing.T) {
	ctx := context.Background()

	result, ok := GetClaims(ctx)
	if ok {
		t.Error("expected ok to be false")
	}
	if result != nil {
		t.Error("expected result to be nil")
	}
}

func TestGetClaims_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), ClaimsKey, "not a claims object")

	result, ok := GetClaims(ctx)
	if ok {
		t.Error("expected ok to be false for wrong type")
	}
	if result != nil {
		t.Error("expected result to be nil for wrong type")
	}
}

func TestGetJwt(t *testing.T) {
	ctx := context.WithValue(context.Background(), JwtKey, "test-jwt-token")

	result, ok := GetJwt(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if result != "test-jwt-token" {
		t.Errorf("expected 'test-jwt-token', got '%s'", result)
	}
}

func TestGetJwt_NotPresent(t *testing.T) {
	ctx := context.Background()

	result, ok := GetJwt(ctx)
	if ok {
		t.Error("expected ok to be false")
	}
	if result != "" {
		t.Error("expected empty string")
	}
}

func TestGetRefreshToken(t *testing.T) {
	ctx := context.WithValue(context.Background(), RefreshKey, "refresh-token-value")

	result, ok := GetRefreshToken(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if result != "refresh-token-value" {
		t.Errorf("expected 'refresh-token-value', got '%s'", result)
	}
}

func TestGetRefreshToken_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := GetRefreshToken(ctx)
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestGetOrigIp(t *testing.T) {
	ctx := context.WithValue(context.Background(), OrigIpKey, "192.168.1.1")

	result, ok := GetOrigIp(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if result != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", result)
	}
}

func TestGetOrigIp_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := GetOrigIp(ctx)
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestGetUserAgent(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserAgentKey, "Mozilla/5.0")

	result, ok := GetUserAgent(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if result != "Mozilla/5.0" {
		t.Errorf("expected 'Mozilla/5.0', got '%s'", result)
	}
}

func TestGetUserAgent_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := GetUserAgent(ctx)
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestGetHost(t *testing.T) {
	ctx := context.WithValue(context.Background(), HostKey, "example.com")

	result, ok := GetHost(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if result != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", result)
	}
}

func TestGetHost_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := GetHost(ctx)
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestGetSecure(t *testing.T) {
	ctx := context.WithValue(context.Background(), SecureKey, true)

	result, ok := GetSecure(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if !result {
		t.Error("expected true")
	}
}

func TestGetSecure_False(t *testing.T) {
	ctx := context.WithValue(context.Background(), SecureKey, false)

	result, ok := GetSecure(ctx)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if result {
		t.Error("expected false")
	}
}

func TestGetSecure_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := GetSecure(ctx)
	if ok {
		t.Error("expected ok to be false")
	}
}

// =============================================================================
// Endpoint Profile Registration Tests
// =============================================================================

func TestPublicEndpoint(t *testing.T) {
	resetEndpointProfiles()

	PublicEndpoint("/api/health")

	profile := GetEndpointProfile("/api/health")
	if !profile.AllowPublic {
		t.Error("expected AllowPublic to be true")
	}
}

func TestPublicEndpoint_WithMethod(t *testing.T) {
	resetEndpointProfiles()

	PublicEndpoint("/api/data", "GET")

	profile := GetEndpointProfileForMethod("/api/data", "GET")
	if !profile.AllowPublic {
		t.Error("expected AllowPublic to be true for GET")
	}

	// POST should not be public
	profile = GetEndpointProfileForMethod("/api/data", "POST")
	if profile.AllowPublic {
		t.Error("expected AllowPublic to be false for POST")
	}
}

func TestPublicEndpoint_UpdatesExisting(t *testing.T) {
	resetEndpointProfiles()

	// Set an admin endpoint first
	AdminEndpoint("/api/mixed")
	// Then make it also public
	PublicEndpoint("/api/mixed")

	profile := GetEndpointProfile("/api/mixed")
	if !profile.AllowPublic {
		t.Error("expected AllowPublic to be true")
	}
	if !profile.RequiresAdmin {
		t.Error("expected RequiresAdmin to still be true")
	}
}

func TestAdminEndpoint(t *testing.T) {
	resetEndpointProfiles()

	AdminEndpoint("/api/admin/users")

	profile := GetEndpointProfile("/api/admin/users")
	if !profile.RequiresAdmin {
		t.Error("expected RequiresAdmin to be true")
	}
}

func TestAdminEndpoint_WithMethod(t *testing.T) {
	resetEndpointProfiles()

	AdminEndpoint("/api/admin", "DELETE")

	profile := GetEndpointProfileForMethod("/api/admin", "DELETE")
	if !profile.RequiresAdmin {
		t.Error("expected RequiresAdmin to be true for DELETE")
	}

	profile = GetEndpointProfileForMethod("/api/admin", "GET")
	if profile.RequiresAdmin {
		t.Error("expected RequiresAdmin to be false for GET")
	}
}

func TestUnverifiedEndpoint(t *testing.T) {
	resetEndpointProfiles()

	UnverifiedEndpoint("/api/profile")

	profile := GetEndpointProfile("/api/profile")
	if !profile.AllowUnverified {
		t.Error("expected AllowUnverified to be true")
	}
}

func TestDenyVerifiedEndpoint(t *testing.T) {
	resetEndpointProfiles()

	DenyVerifiedEndpoint("/api/verify")

	profile := GetEndpointProfile("/api/verify")
	if !profile.DenyVerified {
		t.Error("expected DenyVerified to be true")
	}
}

func TestAllowExpiredJwtEndpoint(t *testing.T) {
	resetEndpointProfiles()

	AllowExpiredJwtEndpoint("/api/refresh")

	profile := GetEndpointProfile("/api/refresh")
	if !profile.AllowExpiredJwt {
		t.Error("expected AllowExpiredJwt to be true")
	}
}

func TestServiceClientEndpoint(t *testing.T) {
	resetEndpointProfiles()

	ServiceClientEndpoint("/api/internal/mail", []string{"mail:send", "mail:read"})

	profile := GetEndpointProfile("/api/internal/mail")
	if !profile.AllowService {
		t.Error("expected AllowService to be true")
	}
	if len(profile.AllowedScopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(profile.AllowedScopes))
	}
}

func TestServiceClientEndpoint_EmptyScopes(t *testing.T) {
	resetEndpointProfiles()

	ServiceClientEndpoint("/api/internal/any", nil)

	profile := GetEndpointProfile("/api/internal/any")
	if !profile.AllowService {
		t.Error("expected AllowService to be true")
	}
	if len(profile.AllowedScopes) != 1 || profile.AllowedScopes[0] != "*" {
		t.Error("expected wildcard scope when no scopes provided")
	}
}

func TestSetEndpointProfile(t *testing.T) {
	resetEndpointProfiles()

	profile := EndpointProfile{
		AllowPublic:         true,
		RequiresAdmin:       true,
		AllowedAccountTypes: []string{"premium"},
	}
	SetEndpointProfile("/api/custom", profile)

	result := GetEndpointProfile("/api/custom")
	if !result.AllowPublic {
		t.Error("expected AllowPublic to be true")
	}
	if !result.RequiresAdmin {
		t.Error("expected RequiresAdmin to be true")
	}
	if len(result.AllowedAccountTypes) != 1 {
		t.Error("expected 1 allowed account type")
	}
}

func TestGetEndpointProfileForMethod_Fallback(t *testing.T) {
	resetEndpointProfiles()

	// Set profile without method
	PublicEndpoint("/api/fallback")

	// Should fall back to the general profile
	profile := GetEndpointProfileForMethod("/api/fallback", "GET")
	if !profile.AllowPublic {
		t.Error("expected to fall back to general profile")
	}
}

func TestGetEndpointProfileForMethod_MethodSpecific(t *testing.T) {
	resetEndpointProfiles()

	// Set general profile
	PublicEndpoint("/api/specific")
	// Override for POST
	AdminEndpoint("/api/specific", "POST")

	getProfile := GetEndpointProfileForMethod("/api/specific", "GET")
	if !getProfile.AllowPublic {
		t.Error("GET should use general profile")
	}

	postProfile := GetEndpointProfileForMethod("/api/specific", "POST")
	if !postProfile.RequiresAdmin {
		t.Error("POST should use method-specific profile")
	}
}

func TestGetEndpointProfileForMethod_CaseInsensitive(t *testing.T) {
	resetEndpointProfiles()

	AdminEndpoint("/api/case", "POST")

	// Should work with lowercase
	profile := GetEndpointProfileForMethod("/api/case", "post")
	if !profile.RequiresAdmin {
		t.Error("expected case-insensitive method matching")
	}
}

func TestEndpointProfiles(t *testing.T) {
	resetEndpointProfiles()

	PublicEndpoint("/api/one")
	AdminEndpoint("/api/two")

	profiles := EndpointProfiles()
	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestEndpointKeys_EmptyMethod(t *testing.T) {
	resetEndpointProfiles()

	// Empty string method should register without method prefix
	PublicEndpoint("/api/empty", "")

	profile := GetEndpointProfile("/api/empty")
	if !profile.AllowPublic {
		t.Error("expected profile to be registered")
	}
}

// =============================================================================
// Endpoint Profile Evaluate Tests
// =============================================================================

func TestEvaluate_PublicEndpoint_NoToken(t *testing.T) {
	resetEndpointProfiles()
	PublicEndpoint("/api/public")
	profile := GetEndpointProfile("/api/public")

	claims, err := profile.Evaluate(nil, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error for public endpoint without token, got: %v", err)
	}
	if claims != nil {
		t.Error("expected nil claims for unauthenticated request")
	}
}

func TestEvaluate_PublicEndpoint_WithValidToken(t *testing.T) {
	resetEndpointProfiles()
	PublicEndpoint("/api/public")
	profile := GetEndpointProfile("/api/public")

	token := createTestUserToken("user-123", false, "standard", true)
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if claims == nil {
		t.Fatal("expected claims to be present")
	}
	if claims.Subject != "user-123" {
		t.Errorf("expected subject 'user-123', got '%s'", claims.Subject)
	}
}

func TestEvaluate_PublicEndpoint_WithInvalidToken(t *testing.T) {
	resetEndpointProfiles()
	PublicEndpoint("/api/public")
	profile := GetEndpointProfile("/api/public")

	token := "invalid-token"
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	// Public endpoint should succeed even with invalid token
	if err != nil {
		t.Fatalf("expected no error for public endpoint with invalid token, got: %v", err)
	}
	if claims != nil {
		t.Error("expected nil claims for invalid token")
	}
}

func TestEvaluate_ProtectedEndpoint_NoToken(t *testing.T) {
	resetEndpointProfiles()
	// Default profile requires authentication
	profile := EndpointProfile{}

	_, err := profile.Evaluate(nil, getTestPublicKeys, testLogger)
	if err != ErrNoAuthToken {
		t.Errorf("expected ErrNoAuthToken, got: %v", err)
	}
}

func TestEvaluate_ProtectedEndpoint_ValidToken(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{}

	token := createTestUserToken("user-456", false, "standard", true)
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if claims.Subject != "user-456" {
		t.Errorf("expected subject 'user-456', got '%s'", claims.Subject)
	}
}

func TestEvaluate_ProtectedEndpoint_InvalidToken(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{}

	token := "invalid-token"
	_, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != ErrInvalidJwt {
		t.Errorf("expected ErrInvalidJwt, got: %v", err)
	}
}

func TestEvaluate_AdminEndpoint_AdminUser(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{RequiresAdmin: true}

	token := createTestUserToken("admin-user", true, "admin", true)
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error for admin user, got: %v", err)
	}
	if claims.Subject != "admin-user" {
		t.Errorf("expected subject 'admin-user', got '%s'", claims.Subject)
	}
}

func TestEvaluate_AdminEndpoint_NonAdminUser(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{RequiresAdmin: true}

	token := createTestUserToken("regular-user", false, "standard", true)
	_, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != ErrUserNotAdmin {
		t.Errorf("expected ErrUserNotAdmin, got: %v", err)
	}
}

func TestEvaluate_UnverifiedUser_DefaultProfile(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{} // Default requires verified

	token := createTestUserToken("unverified", false, "standard", false)
	_, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != ErrUserNotVerified {
		t.Errorf("expected ErrUserNotVerified, got: %v", err)
	}
}

func TestEvaluate_UnverifiedUser_AllowUnverified(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{AllowUnverified: true}

	token := createTestUserToken("unverified", false, "standard", false)
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error with AllowUnverified, got: %v", err)
	}
	if claims.Subject != "unverified" {
		t.Errorf("expected subject 'unverified', got '%s'", claims.Subject)
	}
}

func TestEvaluate_VerifiedUser_DenyVerified(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{DenyVerified: true, AllowUnverified: true}

	token := createTestUserToken("verified", false, "standard", true)
	_, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != ErrUserAlreadyVerified {
		t.Errorf("expected ErrUserAlreadyVerified, got: %v", err)
	}
}

func TestEvaluate_UnverifiedUser_DenyVerified(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{DenyVerified: true, AllowUnverified: true}

	token := createTestUserToken("unverified", false, "standard", false)
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error for unverified user, got: %v", err)
	}
	if claims.Subject != "unverified" {
		t.Errorf("expected subject 'unverified', got '%s'", claims.Subject)
	}
}

func TestEvaluate_AllowedAccountTypes_Matching(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{AllowedAccountTypes: []string{"premium", "enterprise"}}

	token := createTestUserToken("premium-user", false, "premium", true)
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error for allowed account type, got: %v", err)
	}
	if claims.Subject != "premium-user" {
		t.Errorf("expected subject 'premium-user', got '%s'", claims.Subject)
	}
}

func TestEvaluate_AllowedAccountTypes_NotMatching(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{AllowedAccountTypes: []string{"premium", "enterprise"}}

	token := createTestUserToken("free-user", false, "free", true)
	_, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != ErrUserMissingRequiredAccount {
		t.Errorf("expected ErrUserMissingRequiredAccount, got: %v", err)
	}
}

func TestEvaluate_ServiceClaims_Allowed(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{AllowService: true, AllowedScopes: []string{"mail:send"}}

	token := createTestServiceToken("mail-service", []string{"mail:send", "mail:read"})
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error for service with matching scope, got: %v", err)
	}
	if claims.Subject != "mail-service" {
		t.Errorf("expected subject 'mail-service', got '%s'", claims.Subject)
	}
}

func TestEvaluate_ServiceClaims_WildcardScope(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{AllowService: true, AllowedScopes: []string{"*"}}

	token := createTestServiceToken("any-service", []string{"anything"})
	claims, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != nil {
		t.Fatalf("expected no error for wildcard scope, got: %v", err)
	}
	if claims.Subject != "any-service" {
		t.Errorf("expected subject 'any-service', got '%s'", claims.Subject)
	}
}

func TestEvaluate_ServiceClaims_NotAllowed(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{} // Service not allowed by default

	token := createTestServiceToken("service", []string{"some:scope"})
	_, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != ErrServiceClientNotAllowed {
		t.Errorf("expected ErrServiceClientNotAllowed, got: %v", err)
	}
}

func TestEvaluate_ServiceClaims_MissingScope(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{AllowService: true, AllowedScopes: []string{"required:scope"}}

	token := createTestServiceToken("service", []string{"different:scope"})
	_, err := profile.Evaluate(&token, getTestPublicKeys, testLogger)
	if err != ErrServiceClientNotAllowed {
		t.Errorf("expected ErrServiceClientNotAllowed for missing scope, got: %v", err)
	}
}

func TestEvaluate_PublicKeysFnError(t *testing.T) {
	resetEndpointProfiles()
	profile := EndpointProfile{}

	token := createTestUserToken("user", false, "standard", true)
	_, err := profile.Evaluate(&token, getTestPublicKeysError, testLogger)
	if err != ErrNoKeys {
		t.Errorf("expected ErrNoKeys, got: %v", err)
	}
}

// =============================================================================
// Error Variables Tests
// =============================================================================

func TestErrorVariables(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{ErrMetaDataNotFound, "metadata not found"},
		{ErrNoAuthToken, "no authorization token"},
		{ErrInvalidAuthHeader, "invalid authorization header"},
		{ErrInvalidJwt, "invalid jwt"},
		{ErrUserNotAdmin, "user is not admin"},
		{ErrUserMissingRequiredAccount, "user is missing required account"},
		{ErrUserNotVerified, "user is not verified"},
		{ErrUserAlreadyVerified, "user is already verified"},
		{ErrNoKeys, "no public keys"},
		{ErrServiceClientNotAllowed, "service client not allowed"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.expected {
			t.Errorf("expected error '%s', got '%s'", tt.expected, tt.err.Error())
		}
	}
}

// =============================================================================
// Context Keys Tests
// =============================================================================

func TestContextKeys(t *testing.T) {
	if ClaimsKey != "claims" {
		t.Errorf("expected ClaimsKey 'claims', got '%s'", ClaimsKey)
	}
	if JwtKey != "jwt" {
		t.Errorf("expected JwtKey 'jwt', got '%s'", JwtKey)
	}
	if RefreshKey != "refresh" {
		t.Errorf("expected RefreshKey 'refresh', got '%s'", RefreshKey)
	}
	if OrigIpKey != "originalIp" {
		t.Errorf("expected OrigIpKey 'originalIp', got '%s'", OrigIpKey)
	}
	if UserAgentKey != "userAgent" {
		t.Errorf("expected UserAgentKey 'userAgent', got '%s'", UserAgentKey)
	}
	if HostKey != "host" {
		t.Errorf("expected HostKey 'host', got '%s'", HostKey)
	}
	if SecureKey != "secure" {
		t.Errorf("expected SecureKey 'secure', got '%s'", SecureKey)
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkGetClaims(b *testing.B) {
	claims := &jwt.Claims{
		RegisteredClaims: jwt5.RegisteredClaims{Subject: "user-123"},
	}
	ctx := context.WithValue(context.Background(), ClaimsKey, claims)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetClaims(ctx)
	}
}

func BenchmarkEvaluate_ValidToken(b *testing.B) {
	resetEndpointProfiles()
	profile := EndpointProfile{}
	token := createTestUserToken("user-bench", false, "standard", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profile.Evaluate(&token, getTestPublicKeys, testLogger)
	}
}

func BenchmarkGetEndpointProfileForMethod(b *testing.B) {
	resetEndpointProfiles()
	PublicEndpoint("/api/bench")
	AdminEndpoint("/api/bench", "POST")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEndpointProfileForMethod("/api/bench", "GET")
	}
}

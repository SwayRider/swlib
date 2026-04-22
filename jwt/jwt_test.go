package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	jwt5 "github.com/golang-jwt/jwt/v5"
)

// Test RSA key pair generated once for all tests
var (
	testPrivateKey    *rsa.PrivateKey
	testPrivateKeyPEM string
	testPublicKeyPEM  string
)

func init() {
	// Generate a 2048-bit RSA key pair for testing
	var err error
	testPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("failed to generate test RSA key: " + err.Error())
	}

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(testPrivateKey)
	testPrivateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&testPrivateKey.PublicKey)
	if err != nil {
		panic("failed to marshal public key: " + err.Error())
	}
	testPublicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	// Configure JWT for tests
	Configure("TestIssuer", "TestAudience")
}

// Helper to create a pointer to a string
func strPtr(s string) *string {
	return &s
}

// Helper to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

func TestConfigure(t *testing.T) {
	// Save original values
	originalIssuer := jwtIssuer
	originalAudience := jwtAudience
	defer func() {
		jwtIssuer = originalIssuer
		jwtAudience = originalAudience
	}()

	Configure("NewIssuer", "NewAudience")

	if jwtIssuer != "NewIssuer" {
		t.Errorf("expected issuer 'NewIssuer', got '%s'", jwtIssuer)
	}
	if jwtAudience != "NewAudience" {
		t.Errorf("expected audience 'NewAudience', got '%s'", jwtAudience)
	}
}

func TestGenerateToken_UserClaims(t *testing.T) {
	userID := "user-123"
	email := "test@example.com"
	openIDClaims := &OpenIDClaims{
		Email:         &email,
		EmailVerified: boolPtr(true),
	}
	swClaims := NewSwayRiderUserClaims(true, "premium")

	jwtID, token, validUntil, err := GenerateToken(
		userID,
		openIDClaims,
		swClaims,
		testPrivateKeyPEM,
		DefaultTTL,
	)

	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if jwtID == "" {
		t.Error("expected non-empty jwtID")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if validUntil.Before(time.Now()) {
		t.Error("expected validUntil to be in the future")
	}
	if validUntil.After(time.Now().Add(DefaultTTL + time.Second)) {
		t.Error("expected validUntil to be within TTL range")
	}
}

func TestGenerateToken_ServiceClaims(t *testing.T) {
	serviceID := "service-mail"
	swClaims := NewSwayRiderServiceClaims(jwt5.ClaimStrings{"mail:send", "mail:read"})

	jwtID, token, validUntil, err := GenerateToken(
		serviceID,
		nil, // no OpenID claims for services
		swClaims,
		testPrivateKeyPEM,
		time.Hour,
	)

	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if jwtID == "" {
		t.Error("expected non-empty jwtID")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if validUntil.Before(time.Now().Add(59 * time.Minute)) {
		t.Error("expected validUntil to be about 1 hour from now")
	}
}

func TestGenerateToken_InvalidPrivateKey(t *testing.T) {
	_, _, _, err := GenerateToken(
		"user-123",
		nil,
		NewSwayRiderUserClaims(false, "standard"),
		"invalid-key",
		DefaultTTL,
	)

	if err == nil {
		t.Error("expected error for invalid private key")
	}
}

func TestVerifyToken_Valid(t *testing.T) {
	userID := "user-456"
	email := "verify@example.com"
	openIDClaims := &OpenIDClaims{
		Email:         &email,
		EmailVerified: boolPtr(true),
		Name:          strPtr("Test User"),
	}
	swClaims := NewSwayRiderUserClaims(false, "standard")

	_, token, _, err := GenerateToken(
		userID,
		openIDClaims,
		swClaims,
		testPrivateKeyPEM,
		DefaultTTL,
	)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := VerifyToken(string(token), testPublicKeyPEM, VerifyDefault)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("expected subject '%s', got '%s'", userID, claims.Subject)
	}
	if claims.Email == nil || *claims.Email != email {
		t.Errorf("expected email '%s', got '%v'", email, claims.Email)
	}
	if claims.Name == nil || *claims.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%v'", claims.Name)
	}

	userClaims, ok := claims.SwayRiderClaims.(*SwayRiderUserClaims)
	if !ok {
		t.Fatal("expected SwayRiderUserClaims")
	}
	if userClaims.IsAdmin {
		t.Error("expected IsAdmin to be false")
	}
	if userClaims.AccountLevel != "standard" {
		t.Errorf("expected AccountLevel 'standard', got '%s'", userClaims.AccountLevel)
	}
}

func TestVerifyToken_ServiceClaims(t *testing.T) {
	serviceID := "service-auth"
	scopes := jwt5.ClaimStrings{"user:read", "user:write"}
	swClaims := NewSwayRiderServiceClaims(scopes)

	_, token, _, err := GenerateToken(
		serviceID,
		nil,
		swClaims,
		testPrivateKeyPEM,
		DefaultTTL,
	)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := VerifyToken(string(token), testPublicKeyPEM, VerifyDefault)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}

	serviceClaims, ok := claims.SwayRiderClaims.(*SwayRiderServiceClaims)
	if !ok {
		t.Fatal("expected SwayRiderServiceClaims")
	}
	if len(serviceClaims.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(serviceClaims.Scopes))
	}
}

func TestVerifyToken_InvalidPublicKey(t *testing.T) {
	_, token, _, _ := GenerateToken(
		"user-123",
		nil,
		NewSwayRiderUserClaims(false, "standard"),
		testPrivateKeyPEM,
		DefaultTTL,
	)

	_, err := VerifyToken(string(token), "invalid-key", VerifyDefault)
	if err == nil {
		t.Error("expected error for invalid public key")
	}
}

func TestVerifyToken_WrongKey(t *testing.T) {
	// Generate a different key pair
	otherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	otherPublicKeyBytes, _ := x509.MarshalPKIXPublicKey(&otherKey.PublicKey)
	otherPublicKeyPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: otherPublicKeyBytes,
	}))

	_, token, _, _ := GenerateToken(
		"user-123",
		nil,
		NewSwayRiderUserClaims(false, "standard"),
		testPrivateKeyPEM,
		DefaultTTL,
	)

	_, err := VerifyToken(string(token), otherPublicKeyPEM, VerifyDefault)
	if err == nil {
		t.Error("expected error when verifying with wrong key")
	}
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	// Generate a token with very short TTL
	_, token, _, _ := GenerateToken(
		"user-123",
		nil,
		NewSwayRiderUserClaims(false, "standard"),
		testPrivateKeyPEM,
		1*time.Millisecond,
	)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err := VerifyToken(string(token), testPublicKeyPEM, VerifyDefault)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerifyToken_ExpiredTokenWithOmitValidation(t *testing.T) {
	// Generate a token with very short TTL
	_, token, _, _ := GenerateToken(
		"user-123",
		nil,
		NewSwayRiderUserClaims(false, "standard"),
		testPrivateKeyPEM,
		1*time.Millisecond,
	)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Should succeed with VerifyOmitClaimsValidation
	claims, err := VerifyToken(string(token), testPublicKeyPEM, VerifyOmitClaimsValidation)
	if err != nil {
		t.Fatalf("expected success with VerifyOmitClaimsValidation, got: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("expected subject 'user-123', got '%s'", claims.Subject)
	}
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	_, err := VerifyToken("not.a.valid.token", testPublicKeyPEM, VerifyDefault)
	if err == nil {
		t.Error("expected error for invalid token format")
	}
}

func TestVerifyToken_WrongAudience(t *testing.T) {
	// Save original audience
	originalAudience := jwtAudience
	defer func() { jwtAudience = originalAudience }()

	// Generate token with one audience
	Configure("TestIssuer", "Audience1")
	_, token, _, _ := GenerateToken(
		"user-123",
		nil,
		NewSwayRiderUserClaims(false, "standard"),
		testPrivateKeyPEM,
		DefaultTTL,
	)

	// Try to verify with different audience
	Configure("TestIssuer", "Audience2")
	_, err := VerifyToken(string(token), testPublicKeyPEM, VerifyDefault)
	if err == nil {
		t.Error("expected error for wrong audience")
	}

	// Restore for other tests
	Configure("TestIssuer", "TestAudience")
}

func TestOpenIDClaims_MapClaims(t *testing.T) {
	name := "John Doe"
	email := "john@example.com"
	locale := "en-US"

	claims := OpenIDClaims{
		Name:          &name,
		Email:         &email,
		EmailVerified: boolPtr(true),
		Locale:        &locale,
	}

	m := claims.MapClaims()

	if m["name"] != name {
		t.Errorf("expected name '%s', got '%v'", name, m["name"])
	}
	if m["email"] != email {
		t.Errorf("expected email '%s', got '%v'", email, m["email"])
	}
	if m["email_verified"] != true {
		t.Errorf("expected email_verified true, got '%v'", m["email_verified"])
	}
	if m["locale"] != locale {
		t.Errorf("expected locale '%s', got '%v'", locale, m["locale"])
	}

	// Should not include nil fields
	if _, exists := m["phone_number"]; exists {
		t.Error("expected phone_number to not be in map")
	}
}

func TestOpenIDClaims_FromMapClaims(t *testing.T) {
	m := map[string]any{
		"name":           "Jane Doe",
		"email":          "jane@example.com",
		"email_verified": true,
		"given_name":     "Jane",
		"family_name":    "Doe",
	}

	var claims OpenIDClaims
	err := claims.FromMapClaims(m)
	if err != nil {
		t.Fatalf("FromMapClaims failed: %v", err)
	}

	if claims.Name == nil || *claims.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got '%v'", claims.Name)
	}
	if claims.Email == nil || *claims.Email != "jane@example.com" {
		t.Errorf("expected email 'jane@example.com', got '%v'", claims.Email)
	}
	if claims.EmailVerified == nil || !*claims.EmailVerified {
		t.Error("expected email_verified true")
	}
	if claims.GivenName == nil || *claims.GivenName != "Jane" {
		t.Errorf("expected given_name 'Jane', got '%v'", claims.GivenName)
	}
	if claims.FamilyName == nil || *claims.FamilyName != "Doe" {
		t.Errorf("expected family_name 'Doe', got '%v'", claims.FamilyName)
	}
}

func TestSwayRiderUserClaims_MapClaims(t *testing.T) {
	claims := NewSwayRiderUserClaims(true, "premium")

	m := claims.MapClaims()

	if m["kind"] != "SwayRiderUserClaims" {
		t.Errorf("expected kind 'SwayRiderUserClaims', got '%v'", m["kind"])
	}
	if m["is_admin"] != true {
		t.Errorf("expected is_admin true, got '%v'", m["is_admin"])
	}
	if m["account_level"] != "premium" {
		t.Errorf("expected account_level 'premium', got '%v'", m["account_level"])
	}
}

func TestSwayRiderUserClaims_FromMapClaims(t *testing.T) {
	m := map[string]any{
		"kind":          "SwayRiderUserClaims",
		"is_admin":      false,
		"account_level": "free",
	}

	var claims SwayRiderUserClaims
	err := claims.FromMapClaims(m)
	if err != nil {
		t.Fatalf("FromMapClaims failed: %v", err)
	}

	if claims.IsAdmin {
		t.Error("expected IsAdmin false")
	}
	if claims.AccountLevel != "free" {
		t.Errorf("expected AccountLevel 'free', got '%s'", claims.AccountLevel)
	}
}

func TestSwayRiderUserClaims_FromMapClaims_WrongKind(t *testing.T) {
	m := map[string]any{
		"kind":          "SwayRiderServiceClaims",
		"is_admin":      false,
		"account_level": "free",
	}

	var claims SwayRiderUserClaims
	err := claims.FromMapClaims(m)
	if err == nil {
		t.Error("expected error for wrong kind")
	}
}

func TestSwayRiderServiceClaims_MapClaims(t *testing.T) {
	scopes := jwt5.ClaimStrings{"read", "write", "delete"}
	claims := NewSwayRiderServiceClaims(scopes)

	m := claims.MapClaims()

	if m["kind"] != "SwayRiderServiceClaims" {
		t.Errorf("expected kind 'SwayRiderServiceClaims', got '%v'", m["kind"])
	}

	scopesFromMap, ok := m["scopes"].(jwt5.ClaimStrings)
	if !ok {
		t.Fatal("expected scopes to be ClaimStrings")
	}
	if len(scopesFromMap) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(scopesFromMap))
	}
}

func TestSwayRiderServiceClaims_FromMapClaims(t *testing.T) {
	m := map[string]any{
		"kind":   "SwayRiderServiceClaims",
		"scopes": []any{"scope1", "scope2"},
	}

	var claims SwayRiderServiceClaims
	err := claims.FromMapClaims(m)
	if err != nil {
		t.Fatalf("FromMapClaims failed: %v", err)
	}

	if len(claims.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(claims.Scopes))
	}
	if claims.Scopes[0] != "scope1" || claims.Scopes[1] != "scope2" {
		t.Errorf("unexpected scopes: %v", claims.Scopes)
	}
}

func TestSwayRiderClaimsFromMapClaims_UserClaims(t *testing.T) {
	m := map[string]any{
		"swayrider": map[string]any{
			"kind":          "SwayRiderUserClaims",
			"is_admin":      true,
			"account_level": "admin",
		},
	}

	claims, err := SwayRiderClaimsFromMapClaims(m)
	if err != nil {
		t.Fatalf("SwayRiderClaimsFromMapClaims failed: %v", err)
	}

	userClaims, ok := claims.(*SwayRiderUserClaims)
	if !ok {
		t.Fatal("expected *SwayRiderUserClaims")
	}
	if !userClaims.IsAdmin {
		t.Error("expected IsAdmin true")
	}
}

func TestSwayRiderClaimsFromMapClaims_ServiceClaims(t *testing.T) {
	m := map[string]any{
		"swayrider": map[string]any{
			"kind":   "SwayRiderServiceClaims",
			"scopes": []any{"mail:send"},
		},
	}

	claims, err := SwayRiderClaimsFromMapClaims(m)
	if err != nil {
		t.Fatalf("SwayRiderClaimsFromMapClaims failed: %v", err)
	}

	serviceClaims, ok := claims.(*SwayRiderServiceClaims)
	if !ok {
		t.Fatal("expected *SwayRiderServiceClaims")
	}
	if len(serviceClaims.Scopes) != 1 || serviceClaims.Scopes[0] != "mail:send" {
		t.Errorf("unexpected scopes: %v", serviceClaims.Scopes)
	}
}

func TestSwayRiderClaimsFromMapClaims_MissingSwayrider(t *testing.T) {
	m := map[string]any{}

	_, err := SwayRiderClaimsFromMapClaims(m)
	if err == nil {
		t.Error("expected error for missing swayrider claims")
	}
}

func TestSwayRiderClaimsFromMapClaims_InvalidKind(t *testing.T) {
	m := map[string]any{
		"swayrider": map[string]any{
			"kind": "InvalidKind",
		},
	}

	_, err := SwayRiderClaimsFromMapClaims(m)
	if err == nil {
		t.Error("expected error for invalid kind")
	}
}

func TestSwayRiderClaimsFromMapClaims_InvalidType_NoPanic(t *testing.T) {
	// Security test: ensure invalid swayrider claim types don't cause a panic
	// This tests the fix for unsafe type assertion that could crash the server

	testCases := []struct {
		name  string
		value any
	}{
		{"string value", "not a map"},
		{"int value", 12345},
		{"bool value", true},
		{"float value", 3.14},
		{"slice value", []string{"a", "b"}},
		{"nil interface", (*map[string]any)(nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := map[string]any{
				"swayrider": tc.value,
			}

			// This should NOT panic - it should return an error
			_, err := SwayRiderClaimsFromMapClaims(m)
			if err == nil {
				t.Error("expected error for invalid swayrider type")
			}
		})
	}
}

func TestDefaultTTL(t *testing.T) {
	if DefaultTTL != 15*time.Minute {
		t.Errorf("expected DefaultTTL to be 15 minutes, got %v", DefaultTTL)
	}
}

func TestAccessToken_String(t *testing.T) {
	token := AccessToken("test-token-value")
	if string(token) != "test-token-value" {
		t.Error("AccessToken should be convertible to string")
	}
}

func TestOpenIDClaims_SetUpdatedTime(t *testing.T) {
	var claims OpenIDClaims
	now := time.Now()

	claims.SetUpdatedTime(now)

	if claims.UpdatedTime == nil {
		t.Fatal("expected UpdatedTime to be set")
	}
	if claims.UpdatedTime.Time.Unix() != now.Unix() {
		t.Error("UpdatedTime should match the set time")
	}
}

func TestOpenIDClaims_SetAuthTime(t *testing.T) {
	var claims OpenIDClaims
	now := time.Now()

	claims.SetAuthTime(now)

	if claims.AuthTime == nil {
		t.Fatal("expected AuthTime to be set")
	}
	if claims.AuthTime.Time.Unix() != now.Unix() {
		t.Error("AuthTime should match the set time")
	}
}

// Benchmark token generation
func BenchmarkGenerateToken(b *testing.B) {
	swClaims := NewSwayRiderUserClaims(false, "standard")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = GenerateToken(
			"user-benchmark",
			nil,
			swClaims,
			testPrivateKeyPEM,
			DefaultTTL,
		)
	}
}

// Benchmark token verification
func BenchmarkVerifyToken(b *testing.B) {
	swClaims := NewSwayRiderUserClaims(false, "standard")
	_, token, _, _ := GenerateToken(
		"user-benchmark",
		nil,
		swClaims,
		testPrivateKeyPEM,
		DefaultTTL,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyToken(string(token), testPublicKeyPEM, VerifyDefault)
	}
}

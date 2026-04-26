package auth

import (
	"testing"
)

const testSecret = "test-secret-key-for-unit-test"

func TestGenerateAndParseToken(t *testing.T) {
	token, err := GenerateToken(42, false, testSecret, 1)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", claims.UserID)
	}
	if claims.IsAdmin {
		t.Error("IsAdmin = true, want false")
	}
}

func TestAdminToken(t *testing.T) {
	token, err := GenerateToken(1, true, testSecret, 1)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if !claims.IsAdmin {
		t.Error("IsAdmin = false, want true")
	}
	if claims.UserID != 1 {
		t.Errorf("UserID = %d, want 1", claims.UserID)
	}
}

func TestParseInvalidToken(t *testing.T) {
	_, err := ParseToken("invalid.token.string", testSecret)
	if err == nil {
		t.Error("ParseToken should fail with invalid token")
	}
}

func TestParseWrongSecret(t *testing.T) {
	token, _ := GenerateToken(1, false, testSecret, 1)
	_, err := ParseToken(token, "wrong-secret")
	if err == nil {
		t.Error("ParseToken should fail with wrong secret")
	}
}

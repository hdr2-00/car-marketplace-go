package security

import (
	"errors"
	"strings"
	"testing"
	"time"

	"car-marketplace/internal/domain"
)

func TestValidateAccessTokenReportsExpiry(t *testing.T) {
	svc := NewJWTService(strings.Repeat("k", 40))
	token, err := svc.GenerateAccessToken(&domain.AuthClaims{UserID: 1}, -time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.ValidateAccessToken(token); !errors.Is(err, domain.ErrTokenExpired) {
		t.Fatalf("expired token: got %v, want ErrTokenExpired", err)
	}
}

func TestValidateAccessTokenRejectsOtherKey(t *testing.T) {
	token, _ := NewJWTService(strings.Repeat("a", 40)).GenerateAccessToken(&domain.AuthClaims{UserID: 1}, time.Minute)

	if _, err := NewJWTService(strings.Repeat("b", 40)).ValidateAccessToken(token); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("foreign token: got %v, want ErrInvalidToken", err)
	}
}

func TestValidateAccessTokenRoundTrip(t *testing.T) {
	svc := NewJWTService(strings.Repeat("k", 40))
	token, _ := svc.GenerateAccessToken(&domain.AuthClaims{UserID: 42, Email: "a@b.c"}, time.Minute)

	claims, err := svc.ValidateAccessToken(token)
	if err != nil || claims.UserID != 42 {
		t.Fatalf("round trip: claims=%v err=%v", claims, err)
	}
}

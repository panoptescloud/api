package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/domain/users"
	"github.com/panoptescloud/api/pkg/dto"
)

type hasher interface {
	Hash(input string) (dto.HashedValue, error)
}

type tokenRepo interface {
	Save(dto.RefreshToken) error
	ByToken(dto.HashedValue) (dto.RefreshToken, error)
	Delete(id uuid.UUID) error
}

type TokenManager struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	tokenRepo  tokenRepo
	hasher hasher
}

// GenerateRefreshToken returns a securely generated random string of n bytes,
// base64 URL encoded for safe transport/storage.
func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random bytes: %w", err)
	}

	// Base64 URL encoding avoids '+' and '/' characters
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (tm *TokenManager) GenerateJWT(userID users.UserID, name string) (string, error) {
	id, err := uuid.NewUUID()

	if err != nil {
		return "", err
	}

	tokenValue, err := generateRandomToken()
	if err != nil {
		return "", err
	}

	hashedToken, err := tm.hasher.Hash(tokenValue)

	if err != nil {
		return "", err
	}

	rt := dto.RefreshToken{
		ID:        id,
		Token:     hashedToken,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // 30 days
		UserID:    userID.WrappedUuid(),
	}

	if err := tm.tokenRepo.Save(rt); err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub":           userID.String(),
		"name":          name,
		"iat":           time.Now().Unix(),
		"exp":           time.Now().Add(time.Hour * 24).Unix(),
		"refresh_token": hashedToken.Original,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(tm.privateKey)
}

// TODO: see about tidying this up a bit
func (tm *TokenManager) RefreshJWT(tokenString string) (string, error) {
	
	token, err := tm.VerifyJWT(tokenString)

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims type")
	}

	refreshValue, ok := claims["refresh_token"].(string)
	if !ok || refreshValue == "" {
		return "", fmt.Errorf("missing refresh token in claims")
	}

	hashedToken, err := tm.hasher.Hash(refreshValue)
	if err != nil {
		return "", err
	}

	oldRT, err := tm.tokenRepo.ByToken(hashedToken)
	if err != nil {
		return "", fmt.Errorf("failed to find old refresh token: %w", err)
	}

	userID, err := users.NewUserID(oldRT.UserID.String())
	if err != nil {
		return "", err
	}

	name, ok := claims["name"].(string)

	if !ok || name == "" {
		return "", errors.New("failed to get name from old token")
	}

	newToken, err := tm.GenerateJWT(userID, name)

	if err != nil {
		return "", err
	}

	if err := tm.tokenRepo.Delete(oldRT.ID); err != nil {
		// TODO: Log warning, but don’t block the refresh
	}

	return newToken, nil
}

func (tm *TokenManager) VerifyJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func NewTokenManager(tokenRepo tokenRepo, hasher hasher, privateKeyPath string, publicKeyPath string) (*TokenManager, error) {
	privateKeyData, err := os.ReadFile(privateKeyPath)

	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)

	if err != nil {
		return nil, err
	}

	publicKeyData, err := os.ReadFile(publicKeyPath)

	if err != nil {
		return nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)

	if err != nil {
		return nil, err
	}

	return &TokenManager{
		publicKey:  publicKey,
		privateKey: privateKey,
		tokenRepo:  tokenRepo,
		hasher: hasher,
	}, nil
}

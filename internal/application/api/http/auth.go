package http

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func (js *JWTService) Generate(subject string, name string) (string, error) {
	claims := jwt.MapClaims{
		"sub":    subject,
		"name":   name,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
		"scopes": []string{"authenticated"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(js.privateKey)
}

func (js *JWTService) Verify(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return js.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func NewJWTService(privateKeyPath string, publicKeyPath string) (*JWTService, error) {

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

	return &JWTService{
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common/dto"
	usersdomain "github.com/panoptescloud/api/internal/users/domain"
	"github.com/panoptescloud/api/pkg/util/random"
)

type hasher interface {
	Hash(input string) (dto.HashedValue, error)
}

type tokenRepo interface {
	Save(dto.RefreshToken) error
	ByToken(dto.HashedValue) (dto.RefreshToken, error)
	Delete(id uuid.UUID) error
	DeleteByToken(token dto.HashedValue) error
}

type SessionManager struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	tokenRepo  tokenRepo
	hasher     hasher
}

func (tm *SessionManager) Create(userID usersdomain.UserID) (dto.Session, error) {
	id, err := uuid.NewUUID()

	if err != nil {
		return dto.Session{}, err
	}

	tokenValue, err := random.GenerateString()
	if err != nil {
		return dto.Session{}, err
	}

	hashedToken, err := tm.hasher.Hash(tokenValue)

	if err != nil {
		return dto.Session{}, err
	}

	issuedAt := time.Now()

	rt := dto.RefreshToken{
		ID:        id,
		Token:     hashedToken,
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt.Add(time.Hour * 24 * 14), // 14 days
		UserID:    userID.WrappedUuid(),
	}

	if err := tm.tokenRepo.Save(rt); err != nil {
		return dto.Session{}, err
	}

	jwtExpiresAt := issuedAt.Add(time.Minute * 5)
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": issuedAt.Unix(),
		"exp": jwtExpiresAt.Unix(),
	}

	csrf, err := random.GenerateString()

	if err != nil {
		return dto.Session{}, nil
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(tm.privateKey)

	if err != nil {
		return dto.Session{}, err
	}

	return dto.Session{
		JWT:          signedToken,
		RefreshToken: rt,
		CSRFToken:    csrf,
		UserID:       userID.String(),
		JWTExpiresAt: jwtExpiresAt,
		IssuedAt:     issuedAt,
	}, nil
}

// TODO: see about tidying this up a bit
func (tm *SessionManager) Refresh(hashedToken dto.HashedValue) (dto.Session, error) {
	oldRT, err := tm.tokenRepo.ByToken(hashedToken)
	if err != nil {
		return dto.Session{}, fmt.Errorf("failed to find old refresh token: %w", err)
	}

	userID, err := usersdomain.NewUserID(oldRT.UserID.String())
	if err != nil {
		return dto.Session{}, err
	}

	newSession, err := tm.Create(userID)

	if err != nil {
		return dto.Session{}, err
	}

	if err := tm.tokenRepo.Delete(oldRT.ID); err != nil {
		// TODO: Log warning, but don’t block the refresh
	}

	return newSession, nil
}

func (tm *SessionManager) DeleteRefreshToken(hashedToken dto.HashedValue) error {
	return tm.tokenRepo.DeleteByToken(hashedToken)
}

func (tm *SessionManager) VerifyJWT(tokenString string) (*jwt.Token, error) {
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

func NewSessionManager(tokenRepo tokenRepo, hasher hasher, privateKeyPath string, publicKeyPath string) (*SessionManager, error) {
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

	return &SessionManager{
		publicKey:  publicKey,
		privateKey: privateKey,
		tokenRepo:  tokenRepo,
		hasher:     hasher,
	}, nil
}

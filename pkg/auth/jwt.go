package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JWT struct {
	SecretKey string
}

type ContextKey string

const UserIDKey ContextKey = "user_id"

func NewJWT(secretKey string) *JWT {
	return &JWT{
		SecretKey: secretKey,
	}
}

func (j *JWT) CreateToken(userID string) (string, error) {
	claim := j.generateUserClaims(userID)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	return token.SignedString([]byte(j.SecretKey))
}

func (j *JWT) generateUserClaims(userID string) jwt.MapClaims {
	now := time.Now()
	claim := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24).Unix(),
	}

	return claim
}

func (j *JWT) ValidateToken(token string) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "access denied: invalid token")
		}

		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "access denied: invalid token")
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "access denied: invalid token")
	}

	return mapClaims, nil
}

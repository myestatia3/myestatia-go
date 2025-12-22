package security

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getJwtKey() []byte {
	key := os.Getenv("JWT_SECRET_KEY")
	if key == "" {
		key = "default_secret_key_for_dev_only" // Fallback or force error
	}
	return []byte(key)
}

type Claims struct {
	AgentID   string `json:"agent_id"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(agentID, companyID, role string) (string, error) {
	key := getJwtKey()
	if len(key) == 0 {
		return "", errors.New("JWT_SECRET_KEY is not set")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		AgentID:   agentID,
		CompanyID: companyID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func ValidateToken(tokenString string) (*Claims, error) {
	key := getJwtKey()
	if len(key) == 0 {
		return nil, errors.New("JWT_SECRET_KEY is not set")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

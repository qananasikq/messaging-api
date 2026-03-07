package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type JWT struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func New(secret, issuer string, accessTTL time.Duration) *JWT {
	return &JWT{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"usr"`
	jwt.RegisteredClaims
}

func (j *JWT) IssueAccessToken(userID uuid.UUID, username string) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(j.accessTTL)

	claims := Claims{
		UserID:   userID.String(),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return s, exp, nil
}

func (j *JWT) Parse(tokenStr string) (Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	}, jwt.WithIssuer(j.issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	if claims.UserID == "" {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

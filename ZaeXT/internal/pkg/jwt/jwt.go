package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	UserTier string `json:"user_tier"`
	jwt.RegisteredClaims
}

type JWT struct {
	signingKey []byte
}

var (
	ErrTokenExpired     = errors.New("token已过期，请重新登录")
	ErrTokenNotValidYet = errors.New("token未激活，请稍后再试")
	ErrTokenMalformed   = errors.New("token格式错误")
	ErrTokenInvalid     = errors.New("无法解析token")
)

func NewJWT(secretKey string) *JWT {
	return &JWT{
		signingKey: []byte(secretKey),
	}
}

func (j *JWT) GenerateToken(userID uint, username, userTier string, expiration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		UserTier: userTier,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			Issuer:    "ai-qa-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.signingKey)
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, ErrTokenInvalid
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrTokenInvalid
}

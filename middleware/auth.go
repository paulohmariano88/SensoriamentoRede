package middleware

import (
	"net/http"
	"strings"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secretKey")

// Gerar um Token
func GenerateJWT(username string) (string, error) {

	expirationTime := time.Now().Add(24 * time.Hour)  //Token valido por 1 dia

	claims := &jwt.RegisteredClaims{
		Subject: username,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}


func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Token ausente", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token Inválido", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
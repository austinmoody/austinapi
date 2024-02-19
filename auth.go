package main

import (
	"github.com/cristalhq/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
)

func VerifyToken(tokenString string) (*jwt.RegisteredClaims, error) {
	keyString := GetString("JWT_SECRET_KEY")
	key := []byte(keyString)

	verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
	if err != nil {
		log.Printf("error creating JWT verifier: %v", err)
		return nil, err
	}

	rawToken := []byte(tokenString)
	token, err := jwt.Parse(rawToken, verifier)
	if err != nil {
		log.Printf("error parsing JWT token: %v", err)
		return nil, jwt.ErrInvalidFormat
	}

	err = verifier.Verify(token)
	if err != nil {
		log.Printf("unable to verify JWT token: %v", err)
		return nil, jwt.ErrInvalidKey
	}

	var claims jwt.RegisteredClaims
	errParseClaims := jwt.ParseClaims(rawToken, verifier, &claims)
	if errParseClaims != nil {
		log.Printf("error parsing JWT claims: %v", errParseClaims)
		return nil, jwt.ErrInvalidKey
	}

	validAudience := claims.IsForAudience(GetString("JWT_AUDIENCE"))
	validId := claims.IsID(GetString("JWT_UNIQUE_ID"))
	validIssuer := claims.IsIssuer(GetString("JWT_ISSUER"))
	validNow := claims.IsValidAt(time.Now())

	if !validAudience || !validId || !validIssuer || !validNow {
		log.Printf("JWT token is invalid: invalid claims")
		return nil, jwt.ErrInvalidKey
	}

	return &claims, nil

}

func authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		_, err := VerifyToken(tokenString)

		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"api/auth"
	"api/http_server/authenticator"
	"api/http_server/http_util"
	"context"
	"fmt"
	"log"
	"net/http"
)

type ContextKey string

var (
	UserAuthDtoKey ContextKey = "UserAuthDtoKey"
)

type Middleware func(next http.HandlerFunc) http.HandlerFunc

func Authorize(
	next http.HandlerFunc, validator authenticator.Authenticator, group auth.Role,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := http_util.GetTokenFromHeader(authHeader)
		ctx := r.Context()

		isValid, username, err := validator.IsTokenValid(ctx, token, group)
		if err != nil || !isValid {
			if err != nil {
				log.Println(fmt.Errorf("failed token validation %w", err))
			}
			http_util.WriteJson(w, http.StatusUnauthorized, http_util.NewFailureResponse("Unauthorized"))
			return
		}

		updatedReq := r.WithContext(
			context.WithValue(ctx, UserAuthDtoKey, auth.AuthorizationDto{
				Header:   authHeader,
				Username: username,
				Role:     group,
			}),
		)
		next(w, updatedReq)
	}
}

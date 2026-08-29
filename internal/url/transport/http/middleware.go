package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/wxvn/grpc-mesh/internal/url/client"
)

type Middleware struct {
	auth *client.AuthClient
}

func NewMiddleware(auth *client.AuthClient) *Middleware {
	return &Middleware{auth: auth}
}

type contextKey string

var userIDKey contextKey = "user_id"

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		var prefix string = "Bearer "

		if !strings.HasPrefix(header, prefix) {
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if token == "" {
			http.Error(w, "access token is required", http.StatusUnauthorized)
			return
		}

		userID, err := m.auth.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		id, err := uuid.Parse(userID)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

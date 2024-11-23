package middleware

import (
	"github.com/zeromicro/go-zero/rest/handler"
	"net/http"
)

// CommonJwtAuthMiddleware : with jwt on the verification, no jwt on the verification
type CommonJwtAuthMiddleware struct {
	secret string
}

func NewCommonJwtAuthMiddleware(secret string) *CommonJwtAuthMiddleware {
	return &CommonJwtAuthMiddleware{
		secret: secret,
	}
}

func (m *CommonJwtAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header.Get("Authorization")) > 0 {
			//has jwt Authorization
			authHandler := handler.Authorize(m.secret)
			authHandler(next).ServeHTTP(w, r)
			return
		} else {
			//no jwt Authorization
			next(w, r)
		}
	}
}

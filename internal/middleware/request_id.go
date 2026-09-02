package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// HeaderXRequestID e o nome do header HTTP usado para propagar o request id.
	HeaderXRequestID = "X-Request-Id"
	requestIDKey     = "request_id"
)

// RequestID gera (ou propaga) um identificador unico por requisicao,
// expondo-o no header de resposta e no contexto da requisicao.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderXRequestID)
		if id == "" {
			id = uuid.NewString()
		}

		c.Set(requestIDKey, id)
		c.Header(HeaderXRequestID, id)
		c.Next()
	}
}

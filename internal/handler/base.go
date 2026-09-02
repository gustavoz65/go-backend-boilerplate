package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/middleware"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/server"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/validation"
)

type Handler struct {
	server *server.Server
}

func NewHandler(s *server.Server) Handler {
	return Handler{server: s}
}

// HandlerFunc representa um tipo de funcao que processa uma requisicao e retorna uma resposta ou um erro.
type HandlerFunc[Req any, Res any] func(c *gin.Context, req Req) (Res, error)

// HandlerFuncNoContent representa um tipo de funcao que processa uma requisicao sem retornar conteudo na resposta.
type HandlerFuncNoContent[Req any] func(c *gin.Context, req Req) error

// ResponseHandler e responsavel por manipular a resposta de uma operacao.
type ResponseHandler interface {
	Handle(c *gin.Context, result interface{})
	GetOperation() string
	AddAttributes(txn *newrelic.Transaction, result interface{})
}

// JSONResponseHandler manipula respostas no formato JSON.
type JSONResponseHandler struct {
	status int
}

func (h *JSONResponseHandler) Handle(c *gin.Context, result interface{}) {
	c.JSON(h.status, result)
}

func (h *JSONResponseHandler) GetOperation() string {
	return "handler"
}

func (h *JSONResponseHandler) AddAttributes(txn *newrelic.Transaction, result interface{}) {}

// NoContentResponseHandler manipula respostas sem conteudo (204 No Content).
type NoContentResponseHandler struct {
	status int
}

func (h *NoContentResponseHandler) Handle(c *gin.Context, result interface{}) {
	c.Status(h.status)
}

func (h *NoContentResponseHandler) GetOperation() string {
	return "handler_no_content"
}

func (h *NoContentResponseHandler) AddAttributes(txn *newrelic.Transaction, result interface{}) {}

// FileResponseHandler manipula respostas de download de arquivos.
type FileResponseHandler struct {
	status      int
	filename    string
	contentType string
}

func (h *FileResponseHandler) Handle(c *gin.Context, result interface{}) {
	data := result.([]byte)
	c.Header("Content-Disposition", "attachment; filename="+h.filename)
	c.Data(h.status, h.contentType, data)
}

func (h *FileResponseHandler) GetOperation() string {
	return "handler_file"
}

func (h *FileResponseHandler) AddAttributes(txn *newrelic.Transaction, result interface{}) {
	if txn != nil {
		txn.AddAttribute("file.name", h.filename)
		txn.AddAttribute("file.content_type", h.contentType)
		if data, ok := result.([]byte); ok {
			txn.AddAttribute("file.size_bytes", len(data))
		}
	}
}

// handleRequest e uma funcao generica que lida com a logica comum de manipulacao de requisicoes.
func handleRequest[Req any](
	c *gin.Context,
	req Req,
	handler func(c *gin.Context, req Req) (interface{}, error),
	responseHandler ResponseHandler,
) {
	start := time.Now()
	method := c.Request.Method
	route := c.FullPath()

	// Get New Relic transaction
	txn := newrelic.FromContext(c.Request.Context())
	if txn != nil {
		txn.AddAttribute("handler.name", route)
		responseHandler.AddAttributes(txn, nil)
	}

	// Get context-specific logger
	logger := middleware.GetLogger(c).With().
		Str("operation", responseHandler.GetOperation()).
		Str("method", method).
		Str("path", route).
		Str("route", route).
		Logger()

	logger.Info().Msg("handling request")

	// Validation com observability
	validationStart := time.Now()
	if err := validation.BindAndValidate(c, req); err != nil {
		validationDuration := time.Since(validationStart)

		logger.Error().
			Err(err).
			Dur("validation_duration", validationDuration).
			Msg("request validation failed")

		if txn != nil {
			txn.NoticeError(err)
			txn.AddAttribute("validation.status", "failed")
			txn.AddAttribute("validation.duration_ms", validationDuration.Milliseconds())
		}
		_ = c.Error(err)
		return
	}

	validationDuration := time.Since(validationStart)
	if txn != nil {
		txn.AddAttribute("validation.status", "success")
		txn.AddAttribute("validation.duration_ms", validationDuration.Milliseconds())
	}

	logger.Debug().
		Dur("validation_duration", validationDuration).
		Msg("request validation successful")

	// Execute handler with observability
	handlerStart := time.Now()
	result, err := handler(c, req)
	handlerDuration := time.Since(handlerStart)

	if err != nil {
		totalDuration := time.Since(start)

		logger.Error().
			Err(err).
			Dur("handler_duration", handlerDuration).
			Dur("total_duration", totalDuration).
			Msg("handler execution failed")

		if txn != nil {
			txn.NoticeError(err)
			txn.AddAttribute("handler.status", "error")
			txn.AddAttribute("handler.duration_ms", handlerDuration.Milliseconds())
			txn.AddAttribute("total.duration_ms", totalDuration.Milliseconds())
		}
		_ = c.Error(err)
		return
	}

	totalDuration := time.Since(start)

	if txn != nil {
		txn.AddAttribute("handler.status", "success")
		txn.AddAttribute("handler.duration_ms", handlerDuration.Milliseconds())
		txn.AddAttribute("total.duration_ms", totalDuration.Milliseconds())
		responseHandler.AddAttributes(txn, result)
	}

	logger.Info().
		Dur("handler_duration", handlerDuration).
		Dur("validation_duration", validationDuration).
		Dur("total_duration", totalDuration).
		Msg("request completed successfully")

	responseHandler.Handle(c, result)
}

// Handle serve para lidar com requisicoes que retornam uma resposta JSON.
func Handle[Req any, Res any](
	h Handler,
	handler HandlerFunc[Req, Res],
	status int,
	req Req,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleRequest(c, req, func(c *gin.Context, req Req) (interface{}, error) {
			return handler(c, req)
		}, &JSONResponseHandler{status: status})
	}
}

// HandleFile serve para lidar com requisicoes que retornam um arquivo para download.
func HandleFile[Req any](
	h Handler,
	handler HandlerFunc[Req, []byte],
	status int,
	req Req,
	filename string,
	contentType string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleRequest(c, req, func(c *gin.Context, req Req) (interface{}, error) {
			return handler(c, req)
		}, &FileResponseHandler{
			status:      status,
			filename:    filename,
			contentType: contentType,
		})
	}
}

// HandleNoContent serve para lidar com requisicoes que nao retornam conteudo na resposta.
func HandleNoContent[Req any](
	h Handler,
	handler HandlerFuncNoContent[Req],
	status int,
	req Req,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleRequest(c, req, func(c *gin.Context, req Req) (interface{}, error) {
			err := handler(c, req)
			return nil, err
		}, &NoContentResponseHandler{status: status})
	}
}

// SimpleHandler e um helper para handlers simples que nao usam BindAndValidate do base.
// Util para handlers que fazem seu proprio parsing (query params, path params, etc.)
func SimpleHandler(fn func(c *gin.Context)) gin.HandlerFunc {
	return fn
}

// JSONResponse helper para retornar JSON com status code
func JSONResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// SuccessResponse retorna uma resposta de sucesso padrao
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// CreatedResponse retorna uma resposta de criacao
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// NoContentResponse retorna uma resposta sem conteudo
func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

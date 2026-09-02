package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
)

const (
	MaxPageSize     = 100
	DefaultPageSize = 20
	DefaultPage     = 1
)

type QueryValidator struct {
	ctx *gin.Context
}

func NewQueryValidator(c *gin.Context) *QueryValidator {
	return &QueryValidator{ctx: c}
}

// GetUUID valida e retorna um UUID de query parameter
func (qv *QueryValidator) GetUUID(key string, required bool) (*uuid.UUID, error) {
	val := qv.ctx.Query(key)
	if val == "" {
		if required {
			return nil, errs.NewBadRequestError(fmt.Sprintf("%s é obrigatório", key), false, nil, nil, nil)
		}
		return nil, nil
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return nil, errs.NewBadRequestError(fmt.Sprintf("%s inválido", key), false, nil, nil, nil)
	}
	return &id, nil
}

// GetInt valida e retorna um int de query parameter com range
func (qv *QueryValidator) GetInt(key string, required bool, min, max int) (*int, error) {
	val := qv.ctx.Query(key)
	if val == "" {
		if required {
			return nil, errs.NewBadRequestError(fmt.Sprintf("%s é obrigatório", key), false, nil, nil, nil)
		}
		return nil, nil
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		return nil, errs.NewBadRequestError(fmt.Sprintf("%s deve ser um número inteiro", key), false, nil, nil, nil)
	}

	if num < min || num > max {
		return nil, errs.NewBadRequestError(
			fmt.Sprintf("%s deve estar entre %d e %d", key, min, max), false, nil, nil, nil)
	}

	return &num, nil
}

// GetDecimal valida e retorna um decimal de query parameter
func (qv *QueryValidator) GetDecimal(key string, required bool) (*decimal.Decimal, error) {
	val := qv.ctx.Query(key)
	if val == "" {
		if required {
			return nil, errs.NewBadRequestError(fmt.Sprintf("%s é obrigatório", key), false, nil, nil, nil)
		}
		return nil, nil
	}

	dec, err := decimal.NewFromString(val)
	if err != nil {
		return nil, errs.NewBadRequestError(fmt.Sprintf("%s inválido", key), false, nil, nil, nil)
	}

	return &dec, nil
}

// GetDate valida e retorna uma data de query parameter (formato: YYYY-MM-DD)
func (qv *QueryValidator) GetDate(key string, required bool) (*time.Time, error) {
	val := qv.ctx.Query(key)
	if val == "" {
		if required {
			return nil, errs.NewBadRequestError(fmt.Sprintf("%s é obrigatório", key), false, nil, nil, nil)
		}
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		return nil, errs.NewBadRequestError(
			fmt.Sprintf("%s deve estar no formato YYYY-MM-DD", key), false, nil, nil, nil)
	}

	return &t, nil
}

// GetBool valida e retorna um bool de query parameter
func (qv *QueryValidator) GetBool(key string) *bool {
	val := qv.ctx.Query(key)
	if val == "" {
		return nil
	}
	result := val == "true"
	return &result
}

// GetPagination valida e retorna página e tamanho de página com limites
func (qv *QueryValidator) GetPagination() (page, pageSize int, err error) {
	pagePtr, err := qv.GetInt("page", false, 1, 10000)
	if err != nil {
		return 0, 0, err
	}
	if pagePtr == nil {
		page = DefaultPage
	} else {
		page = *pagePtr
	}

	pageSizePtr, err := qv.GetInt("page_size", false, 1, MaxPageSize)
	if err != nil {
		return 0, 0, err
	}
	if pageSizePtr == nil {
		pageSize = DefaultPageSize
	} else {
		pageSize = *pageSizePtr
	}

	return page, pageSize, nil
}

// GetEnum valida e retorna um valor enum de query parameter
func (qv *QueryValidator) GetEnum(key string, required bool, allowedValues []string) (*string, error) {
	val := qv.ctx.Query(key)
	if val == "" {
		if required {
			return nil, errs.NewBadRequestError(fmt.Sprintf("%s é obrigatório", key), false, nil, nil, nil)
		}
		return nil, nil
	}

	for _, allowed := range allowedValues {
		if val == allowed {
			return &val, nil
		}
	}

	return nil, errs.NewBadRequestError(
		fmt.Sprintf("%s deve ser um dos valores: %v", key, allowedValues), false, nil, nil, nil)
}

// GetString retorna um string de query parameter com validação opcional de tamanho
func (qv *QueryValidator) GetString(key string, required bool, maxLength int) (*string, error) {
	val := qv.ctx.Query(key)
	if val == "" {
		if required {
			return nil, errs.NewBadRequestError(fmt.Sprintf("%s é obrigatório", key), false, nil, nil, nil)
		}
		return nil, nil
	}

	if maxLength > 0 && len(val) > maxLength {
		return nil, errs.NewBadRequestError(
			fmt.Sprintf("%s deve ter no máximo %d caracteres", key, maxLength), false, nil, nil, nil)
	}

	return &val, nil
}

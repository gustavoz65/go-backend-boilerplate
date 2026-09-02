package validation

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type Validatable interface {
	any
}

func BindAndValidate(c *gin.Context, req interface{}) error {
	if err := c.ShouldBind(req); err != nil {
		return errs.NewBadRequestError("Formato de requisicao invalido", false, nil, nil, nil)
	}

	if err := validate.Struct(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fieldErrors := make([]errs.FieldError, 0, len(validationErrors))
			for _, ve := range validationErrors {
				fieldErrors = append(fieldErrors, errs.FieldError{
					Field: toSnakeCase(ve.Field()),
					Error: formatValidationError(ve),
				})
			}
			return errs.NewBadRequestError("Erro de validacao", false, nil, fieldErrors, nil)
		}
		return errs.NewBadRequestError("Erro de validacao: "+err.Error(), false, nil, nil, nil)
	}

	return nil
}

func GetValidator() *validator.Validate {
	return validate
}

func formatValidationError(fe validator.FieldError) string {
	field := toSnakeCase(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("O campo '%s' e obrigatorio", field)
	case "email":
		return fmt.Sprintf("O campo '%s' deve ser um email valido", field)
	case "min":
		return fmt.Sprintf("O campo '%s' deve ter no minimo %s caracteres", field, fe.Param())
	case "max":
		return fmt.Sprintf("O campo '%s' deve ter no maximo %s caracteres", field, fe.Param())
	case "gt":
		return fmt.Sprintf("O campo '%s' deve ser maior que %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("O campo '%s' deve ser maior ou igual a %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("O campo '%s' deve ser menor ou igual a %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("O campo '%s' deve ser um dos valores: %s", field, fe.Param())
	case "len":
		return fmt.Sprintf("O campo '%s' deve ter exatamente %s caracteres", field, fe.Param())
	case "hexcolor":
		return fmt.Sprintf("O campo '%s' deve ser uma cor hexadecimal valida", field)
	case "gtfield":
		return fmt.Sprintf("O campo '%s' deve ser maior que '%s'", field, toSnakeCase(fe.Param()))
	case "nefield":
		return fmt.Sprintf("O campo '%s' deve ser diferente de '%s'", field, toSnakeCase(fe.Param()))
	default:
		return fmt.Sprintf("O campo '%s' e invalido", field)
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

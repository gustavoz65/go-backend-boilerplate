package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog representa um registro de auditoria no sistema
type AuditLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Action     string     `json:"action" db:"action"`
	EntityType string     `json:"entity_type" db:"entity_type"`
	EntityID   *uuid.UUID `json:"entity_id,omitempty" db:"entity_id"`
	OldValues  JSONMap    `json:"old_values,omitempty" db:"old_values"`
	NewValues  JSONMap    `json:"new_values,omitempty" db:"new_values"`
	IPAddress  *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  *string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// JSONMap é um tipo helper para campos JSON no banco
type JSONMap map[string]interface{}

// Value implementa driver.Valuer para salvar no banco
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implementa sql.Scanner para ler do banco
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// CreateAuditLogInput representa os dados para criar um log de auditoria
type CreateAuditLogInput struct {
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Action     string     `json:"action" validate:"required"`
	EntityType string     `json:"entity_type" validate:"required"`
	EntityID   *uuid.UUID `json:"entity_id,omitempty"`
	OldValues  JSONMap    `json:"old_values,omitempty"`
	NewValues  JSONMap    `json:"new_values,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
}

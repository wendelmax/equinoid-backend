package errors

import (
	"fmt"
)

var (
	ErrUserNotFound       = &NotFoundError{Resource: "user", Message: "usuário não encontrado"}
	ErrUserInactive       = &ValidationError{Field: "is_active", Message: "usuário inativo"}
	ErrUserEmailExists    = &ConflictError{Resource: "email", Message: "email já está em uso"}
	ErrInvalidCredentials = &AuthenticationError{Message: "credenciais inválidas"}

	ErrEquinoNotFound  = &NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	ErrEquinoidExists  = &ConflictError{Resource: "equinoid", Message: "EquinoId já existe"}
	ErrMicrochipExists = &ConflictError{Resource: "microchip_id", Message: "MicrochipID já existe"}

	ErrInvalidToken = &AuthenticationError{Message: "token inválido ou expirado"}
	ErrTokenExpired = &AuthenticationError{Message: "token expirado"}

	ErrValidationFailed = &ValidationError{Field: "validation", Message: "validação falhou"}
	ErrUnauthorized     = &AuthorizationError{Message: "não autorizado"}
	ErrForbidden        = &AuthorizationError{Message: "acesso negado"}
)

type NotFoundError struct {
	Resource string
	Message  string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	if e.ID != nil {
		return fmt.Sprintf("%s: %s (id: %v)", e.Resource, e.Message, e.ID)
	}
	return fmt.Sprintf("%s: %s", e.Resource, e.Message)
}

func (e *NotFoundError) WithID(id interface{}) *NotFoundError {
	e.ID = id
	return e
}

type ConflictError struct {
	Resource string
	Message  string
	Value    interface{}
}

func (e *ConflictError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("%s: %s (value: %v)", e.Resource, e.Message, e.Value)
	}
	return fmt.Sprintf("%s: %s", e.Resource, e.Message)
}

func (e *ConflictError) WithValue(value interface{}) *ConflictError {
	e.Value = value
	return e
}

type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validação falhou no campo '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validação falhou no campo '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) WithValue(value interface{}) *ValidationError {
	e.Value = value
	return e
}

type AuthenticationError struct {
	Message string
	Reason  string
}

func (e *AuthenticationError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Reason)
	}
	return e.Message
}

func (e *AuthenticationError) WithReason(reason string) *AuthenticationError {
	e.Reason = reason
	return e
}

type AuthorizationError struct {
	Message  string
	Action   string
	Resource string
}

func (e *AuthorizationError) Error() string {
	if e.Action != "" && e.Resource != "" {
		return fmt.Sprintf("%s: ação '%s' não permitida em '%s'", e.Message, e.Action, e.Resource)
	}
	return e.Message
}

func (e *AuthorizationError) WithAction(action, resource string) *AuthorizationError {
	e.Action = action
	e.Resource = resource
	return e
}

type DatabaseError struct {
	Operation string
	Message   string
	Err       error
}

func (e *DatabaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("erro de banco de dados na operação '%s': %s: %v", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("erro de banco de dados na operação '%s': %s", e.Operation, e.Message)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

func NewDatabaseError(operation, message string, err error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

type BusinessError struct {
	Code    string
	Message string
	Context map[string]interface{}
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewBusinessError(code, message string, context map[string]interface{}) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Context: context,
	}
}

func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

func IsConflict(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}

func IsValidation(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

func IsAuthentication(err error) bool {
	_, ok := err.(*AuthenticationError)
	return ok
}

func IsAuthorization(err error) bool {
	_, ok := err.(*AuthorizationError)
	return ok
}

func IsDatabase(err error) bool {
	_, ok := err.(*DatabaseError)
	return ok
}

func IsBusiness(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

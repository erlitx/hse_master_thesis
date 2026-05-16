package render

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Ошибка для JSON-ответа
type Err struct {
	Error string `json:"error"`
}

// Отправляет JSON с ошибкой
func Error(w http.ResponseWriter, err error, status int, message string) {
	log.Error().Err(err).Msg(message)

	err = unpack(err)
	err = fmt.Errorf("%s: %w", message, err)

	JSON(w, Err{Error: err.Error()}, status)
}

// Раскрывает обёрнутую ошибку
func unpack(err error) error {
	for {
		e := errors.Unwrap(err)
		if e == nil {
			break
		}

		err = e
	}

	return err
}

// Отправляет ошибку по карте доменных кодов
func ErrorSpecific(w http.ResponseWriter, err error, domainErrors map[error]ErrorMapping) {
	status, message := mapError(err, domainErrors)

	log.Error().Err(err).Msg(message)

	err = unpack(err)
	err = fmt.Errorf("%s: %w", message, err)

	JSON(w, Err{Error: err.Error()}, status)
}

// Сопоставление доменной ошибки и HTTP
type ErrorMapping struct {
	Status  int
	Message string
}

// Сопоставляет ошибку со статусом HTTP
func mapError(err error, domainErrors map[error]ErrorMapping) (int, string) {
	for domainErr, mapping := range domainErrors {
		if errors.Is(err, domainErr) {
			return mapping.Status, mapping.Message
		}
	}

	return http.StatusInternalServerError, "internal server error"
}

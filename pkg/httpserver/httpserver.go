package httpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Корневая конфигурация приложения
type Config struct {
	ServerPort string `envconfig:"HTTP_SERVER_PORT" default:"8081"`
	ClientPort string `envconfig:"HTTP_CLIENT_PORT" default:"8082"`
}

// HTTP-сервер
type Server struct {
	server *http.Server
	notify chan error
}

// Создаёт новый экземпляр
func New(handler http.Handler, port string) *Server {
	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 60 * time.Minute, // important for long responses
		IdleTimeout:  5 * time.Minute,
		Addr:         net.JoinHostPort("", port),
	}

	s := &Server{
		server: httpServer,
		notify: make(chan error, 1),
	}

	go s.start()

	log.Info().Msg("HTTP Server started on port: " + port)

	return s
}

// Запускает HTTP-сервер в горутине
func (s *Server) start() {
	s.notify <- s.server.ListenAndServe()
	close(s.notify)
}

// Канал ошибок завершения сервера
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Закрывает соединение
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("server - Close - s.server.Shutdown")
	}

	log.Info().Msg("HTTP Server closed")
}

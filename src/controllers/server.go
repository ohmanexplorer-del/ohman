package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Server struct {
	http   *http.Server
	engine http.Handler
}

func NewServer(db *gorm.DB, bot BotController, port int) *Server {
	engine := createRouter(db, bot)
	addr := fmt.Sprintf(":%d", port)
	return &Server{
		engine: engine,
		http:   &http.Server{Addr: addr, Handler: engine},
	}
}

func (s *Server) Start() {
	go func() {
		log.Printf("REST API listening on %s", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("REST API server error: %v", err)
		}
	}()
}

func (s *Server) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		log.Printf("REST API shutdown error: %v", err)
		return
	}
	log.Println("REST API server stopped")
}

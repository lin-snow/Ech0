// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type Server struct {
	GinEngine  *gin.Engine
	httpServer *http.Server
	listener   net.Listener
}

func (s *Server) Name() string {
	return "server"
}

func New(engine *gin.Engine) *Server {
	return &Server{
		GinEngine: engine,
	}
}

// listenAddr honors ECH0_SERVER_HOST. Empty, 0.0.0.0 and :: all mean "every
// interface" and keep the historical dual-stack bind (":port").
func listenAddr(host, port string) string {
	switch strings.TrimSpace(host) {
	case "", "0.0.0.0", "::", "[::]":
		host = ""
	}
	return net.JoinHostPort(host, port)
}

func (s *Server) Start(context.Context) error {
	if s.GinEngine == nil {
		return errors.New("gin engine is nil")
	}
	if s.listener != nil {
		return errors.New("http server already started")
	}

	port := config.Config().Server.Port
	PrintGreetings(port)

	s.httpServer = &http.Server{
		Addr:    listenAddr(config.Config().Server.Host, port),
		Handler: s.GinEngine,
	}

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.listener = listener

	go func() {
		if err := s.httpServer.Serve(listener); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logUtil.Panic(commonModel.GIN_RUN_FAILED, logUtil.Err(err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	shutdownCtx := ctx
	var cancel context.CancelFunc

	if ctx == nil {
		shutdownCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	if s.httpServer == nil {
		return nil
	} else {
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}

	s.httpServer = nil
	s.listener = nil
	return nil
}

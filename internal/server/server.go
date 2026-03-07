// Package server
//
//	@title			Ech0 API 文档
//	@version		1.0
//	@description	开源、自托管轻量级发布平台 Ech0 的 API 文档
//	@host			localhost:6277
//	@BasePath		/api
package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
)

// Server 是纯 HTTP runtime，只负责 gin/http 生命周期。
type Server struct {
	GinEngine  *gin.Engine
	httpServer *http.Server // 用于优雅停止服务器
	listener   net.Listener
}

func (s *Server) Name() string {
	return "server"
}

// New 创建一个新的 HTTP server 实例。
func New(engine *gin.Engine) *Server {
	return &Server{
		GinEngine: engine,
	}
}

// Start 启动服务器，并在返回前确认监听端口已成功绑定。
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
		Addr:    ":" + port,
		Handler: s.GinEngine,
	}

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.listener = listener

	// 监听成功后再异步进入 Serve。
	go func() {
		if err := s.httpServer.Serve(listener); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			errUtil.HandlePanicError(&commonModel.ServerError{
				Msg: commonModel.GIN_RUN_FAILED,
				Err: err,
			})
		}
	}()

	return nil
}

// Stop 优雅停止服务器
func (s *Server) Stop(ctx context.Context) error {
	// 使用传入的 context，如果没有则创建默认的 5 秒超时
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

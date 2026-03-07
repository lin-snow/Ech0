package server

import (
	"context"
	"net"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
)

func TestServerStopWithoutHTTPServer(t *testing.T) {
	s := &Server{}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

func TestServerStartFailsWhenPortAlreadyBound(t *testing.T) {
	config.Config().Server.Port = "6288"

	ln, err := net.Listen("tcp", ":6288")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() { _ = ln.Close() }()

	s := New(gin.New())
	if err := s.Start(context.Background()); err == nil {
		t.Fatalf("expected start failure when port is already bound")
	}
}

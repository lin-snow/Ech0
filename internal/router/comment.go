package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/captcha"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

func setupCommentRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	captchaHandler, err := captcha.NewHTTPHandler("/api")
	if err != nil {
		panic(err)
	}

	// Public
	appRouterGroup.PublicRouterGroup.Any("/cap/*any", gin.WrapH(captchaHandler))
	appRouterGroup.PublicRouterGroup.GET("/comments/form", middleware.NoCache(), h.CommentHandler.GetFormMeta())
	appRouterGroup.PublicRouterGroup.GET("/comments", middleware.NoCache(), h.CommentHandler.ListCommentsByEchoID())
	appRouterGroup.PublicRouterGroup.GET(
		"/comments/public",
		middleware.NoCache(),
		h.CommentHandler.ListPublicComments(),
	)
	appRouterGroup.PublicRouterGroup.POST("/comments", h.CommentHandler.CreateComment())

	// Admin Panel
	appRouterGroup.FullAuthRouterGroup.GET("/panel/comments", h.CommentHandler.ListPanelComments())
	appRouterGroup.FullAuthRouterGroup.GET("/panel/comments/:id", h.CommentHandler.GetCommentByID())
	appRouterGroup.FullAuthRouterGroup.PATCH("/panel/comments/:id/status", h.CommentHandler.UpdateCommentStatus())
	appRouterGroup.FullAuthRouterGroup.PATCH("/panel/comments/:id/hot", h.CommentHandler.UpdateCommentHot())
	appRouterGroup.FullAuthRouterGroup.DELETE("/panel/comments/:id", h.CommentHandler.DeleteComment())
	appRouterGroup.FullAuthRouterGroup.POST("/panel/comments/batch", h.CommentHandler.BatchAction())
	appRouterGroup.FullAuthRouterGroup.GET("/panel/comments/settings", h.CommentHandler.GetCommentSetting())
	appRouterGroup.FullAuthRouterGroup.PUT("/panel/comments/settings", h.CommentHandler.UpdateCommentSetting())
	appRouterGroup.FullAuthRouterGroup.POST("/panel/comments/settings/test-email", h.CommentHandler.TestCommentEmail())
}

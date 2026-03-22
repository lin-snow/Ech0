package router

import "github.com/lin-snow/ech0/internal/handler"

// setupInboxRoutes 配置收件箱相关路由
func setupInboxRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.FullAuthRouterGroup.GET("/inbox", h.InboxHandler.GetInboxList())
	appRouterGroup.FullAuthRouterGroup.GET("/inbox/unread", h.InboxHandler.GetUnreadInbox())
	appRouterGroup.FullAuthRouterGroup.PUT("/inbox/:id/read", h.InboxHandler.MarkInboxAsRead())
	appRouterGroup.FullAuthRouterGroup.DELETE("/inbox/:id", h.InboxHandler.DeleteInbox())
	appRouterGroup.FullAuthRouterGroup.DELETE("/inbox", h.InboxHandler.ClearInbox())
}

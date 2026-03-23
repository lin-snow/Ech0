package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func setupFileRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Auth
	appRouterGroup.AuthRouterGroup.GET(
		"/files",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.ListFiles(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/file/tree",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.ListFileTree(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/file/stream",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.StreamFileByPath,
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/file/:id",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.GetFileByID(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/file/:id/stream",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.StreamFileByID,
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/files/upload",
		middleware.RequireScopes(authModel.ScopeFileWrite),
		h.FileHandler.UploadFile(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/file/:id/meta",
		middleware.RequireScopes(authModel.ScopeFileWrite),
		h.FileHandler.UpdateFileMeta(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/files/external",
		middleware.RequireScopes(authModel.ScopeFileWrite),
		h.FileHandler.CreateExternalFile(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/file/:id",
		middleware.RequireScopes(authModel.ScopeFileWrite),
		h.FileHandler.DeleteFile(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/files/presign",
		middleware.RequireScopes(authModel.ScopeFileWrite),
		h.FileHandler.GetFilePresignURL(),
	)
}

package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	model "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/comment"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type CommentHandler struct {
	commentService service.Service
}

func NewCommentHandler(commentService service.Service) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

func (h *CommentHandler) GetFormMeta() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		h.attachOptionalViewer(ctx)
		data, err := h.commentService.GetFormMeta(
			ctx.Request.Context(),
			ctx.ClientIP(),
			resolveRequestBaseURL(ctx.Request),
		)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data, Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) ListCommentsByEchoID() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		echoID := strings.TrimSpace(ctx.Query("echo_id"))
		if echoID == "" {
			return res.Response{Msg: commonModel.INVALID_QUERY_PARAMS}
		}
		comments, err := h.commentService.ListPublicByEchoID(ctx.Request.Context(), echoID)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: comments, Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) ListPublicComments() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		limit := parseQueryInt(ctx, "limit", 30)
		comments, err := h.commentService.ListPublicComments(ctx.Request.Context(), limit)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: comments, Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) CreateComment() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		h.attachOptionalViewer(ctx)
		var dto model.CreateCommentDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		result, err := h.commentService.CreateComment(
			ctx.Request.Context(),
			ctx.ClientIP(),
			ctx.Request.UserAgent(),
			&dto,
		)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: result, Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) ListPanelComments() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		query := model.ListCommentQuery{
			Page:     parseQueryInt(ctx, "page", 1),
			PageSize: parseQueryInt(ctx, "page_size", 20),
			Keyword:  ctx.Query("keyword"),
			Status:   ctx.Query("status"),
			EchoID:   ctx.Query("echo_id"),
			Hot:      parseQueryBool(ctx, "hot"),
		}
		data, err := h.commentService.ListPanelComments(ctx.Request.Context(), query)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data, Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) GetCommentByID() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		id := strings.TrimSpace(ctx.Param("id"))
		data, err := h.commentService.GetCommentByID(ctx.Request.Context(), id)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data, Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) UpdateCommentStatus() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		id := strings.TrimSpace(ctx.Param("id"))
		var dto model.UpdateCommentStatusDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		if err := h.commentService.UpdateCommentStatus(ctx.Request.Context(), id, dto.Status); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) DeleteComment() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		id := strings.TrimSpace(ctx.Param("id"))
		if err := h.commentService.DeleteComment(ctx.Request.Context(), id); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Msg: commonModel.DELETE_SUCCESS}
	})
}

func (h *CommentHandler) UpdateCommentHot() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		id := strings.TrimSpace(ctx.Param("id"))
		var dto model.UpdateCommentHotDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		if err := h.commentService.UpdateCommentHot(ctx.Request.Context(), id, dto.Hot); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) BatchAction() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto model.BatchCommentActionDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		if err := h.commentService.BatchAction(ctx.Request.Context(), dto.Action, dto.IDs); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) GetCommentSetting() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		data, err := h.commentService.GetSystemSetting(ctx.Request.Context())
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data, Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) UpdateCommentSetting() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto model.SystemSetting
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		if err := h.commentService.UpdateSystemSetting(ctx.Request.Context(), dto); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Msg: commonModel.UPDATE_SETTINGS_SUCCESS}
	})
}

func (h *CommentHandler) TestCommentEmail() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto model.TestEmailRequest
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		if err := h.commentService.SendTestEmail(ctx.Request.Context(), dto.Setting); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *CommentHandler) attachOptionalViewer(ctx *gin.Context) {
	auth := strings.TrimSpace(ctx.GetHeader("Authorization"))
	v := service.ParseOptionalViewerFromAuthHeader(auth)
	if v != nil {
		viewer.AttachToRequest(&ctx.Request, v)
		return
	}
	viewer.AttachToRequest(&ctx.Request, viewer.NewNoopViewer())
}

func parseQueryInt(ctx *gin.Context, key string, def int) int {
	raw := strings.TrimSpace(ctx.Query(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

func parseQueryBool(ctx *gin.Context, key string) *bool {
	raw := strings.TrimSpace(ctx.Query(key))
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &v
}

func resolveRequestBaseURL(r *http.Request) string {
	if r == nil {
		return ""
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + host
}

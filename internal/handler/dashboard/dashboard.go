package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/dashboard"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetMetrics 获取系统指标
//
//	@Summary		获取系统指标
//	@Description	获取当前系统的各项运行指标，如 CPU 使用率、内存使用情况等
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=object}	"获取系统指标成功"
//	@Failure		200	{object}	res.Response				"获取系统指标失败"
//	@Router			/metrics [get]
func (dashboardHandler *DashboardHandler) GetMetrics() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		metrics, err := dashboardHandler.dashboardService.GetMetrics()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: metrics,
			Msg:  commonModel.GET_METRICS_SUCCESS,
		}
	})
}

// WSSubsribeMetrics 通过 WebSocket 订阅系统指标
//
//	@Summary		通过 WebSocket 订阅系统指标
//	@Description	通过 WebSocket 实时订阅系统的各项运行指标
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Router			/ws/metrics [get]
func (dashboardHandler *DashboardHandler) WSSubsribeMetrics() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			return
		}

		token = strings.Trim(token, `"`) // 去掉可能的双引号

		// 使用 JWT Util进行处理
		if _, err := jwtUtil.ParseToken(token); err != nil {
			return
		}

		if err := dashboardHandler.dashboardService.WSSubsribeMetrics(ctx.Writer, ctx.Request); err != nil {
			logUtil.GetLogger().
				Error("WebSocket Subscribe Metrics Failed", zap.String("Err", err.Error()))
		}
	})
}

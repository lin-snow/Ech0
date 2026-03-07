package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	model "github.com/lin-snow/ech0/internal/model/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	fmtUtil "github.com/lin-snow/ech0/internal/util/format"
)

type DashboardService struct {
	monitor       *monitor.Monitor
	commonService *commonService.CommonService
}

func NewDashboardService(
	monitor *monitor.Monitor,
	commonService *commonService.CommonService,
) *DashboardService {
	return &DashboardService{
		monitor:       monitor,
		commonService: commonService,
	}
}

func (dashboardService *DashboardService) GetMetrics() (model.Metrics, error) {
	return dashboardService.monitor.GetMetrics(), nil
}

func (s *DashboardService) WSSubsribeMetrics(w http.ResponseWriter, r *http.Request) error {
	// WebSocket 升级
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// 禁用日志输出
		// log.Printf("websocket upgrade failed: %v", err)
		return err
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		defer func() { _ = conn.Close() }()

		for {
			rawMetrics := s.monitor.GetMetrics()
			formatted := fmtUtil.FormatMetrics(&rawMetrics)

			resp := struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
				Data any    `json:"data"`
			}{
				Code: 1,
				Msg:  "metrics update",
				Data: formatted,
			}

			data, _ := json.Marshal(resp)

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				// 禁用日志输出
				// log.Println("write ws error:", err)
				return
			}

			time.Sleep(5 * time.Second)
		}
	}()
	return nil
}

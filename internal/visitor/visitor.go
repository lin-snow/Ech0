package visitor

import (
	"encoding/hex"
	"hash/fnv"
	"net/http"
	"time"
)

const (
	dateLayout = "2006-01-02"
	keepDays   = 7
	pvWindow   = 5 * time.Minute
	recordBuf  = 4096
)

type DayStat struct {
	Date string `json:"date"`
	PV   int64  `json:"pv"`
	UV   int64  `json:"uv"`
}

type Tracker struct {
	recordCh chan recordEvent
	queryCh  chan chan []DayStat
	todayCh  chan chan DayStat
	loadCh   chan loadRequest
}

func NewTracker() *Tracker {
	s := &Tracker{
		recordCh: make(chan recordEvent, recordBuf),
		queryCh:  make(chan chan []DayStat),
		todayCh:  make(chan chan DayStat),
		loadCh:   make(chan loadRequest),
	}
	go s.run()
	return s
}

func (s *Tracker) Record(r *http.Request, ip string) {
	if r == nil || r.Method != http.MethodGet {
		return
	}
	event := recordEvent{
		ipHash: hashIP(ip),
		at:     time.Now().UTC(),
	}
	// 非阻塞写入，队列满时丢弃事件，避免影响请求主路径。
	select {
	case s.recordCh <- event:
	default:
	}
}

func (s *Tracker) Last7Days() []DayStat {
	resp := make(chan []DayStat, 1)
	s.queryCh <- resp
	return <-resp
}

func (s *Tracker) TodayStat() DayStat {
	resp := make(chan DayStat, 1)
	s.todayCh <- resp
	return <-resp
}

func (s *Tracker) LoadHistory(history []DayStat) {
	done := make(chan struct{}, 1)
	s.loadCh <- loadRequest{
		history: history,
		done:    done,
	}
	<-done
}

type recordEvent struct {
	ipHash string
	at     time.Time
}

type loadRequest struct {
	history []DayStat
	done    chan struct{}
}

type runtimeState struct {
	byDay    map[string]DayStat
	today    string
	todayUV  map[string]struct{}
	lastPVAt map[string]time.Time
}

func (s *Tracker) run() {
	state := runtimeState{
		byDay:    make(map[string]DayStat, keepDays),
		todayUV:  make(map[string]struct{}),
		lastPVAt: make(map[string]time.Time),
	}
	for {
		select {
		case ev := <-s.recordCh:
			recordIP(&state, ev)
		case resp := <-s.queryCh:
			now := time.Now().UTC()
			rotateDay(&state, now.Format(dateLayout))
			gc(&state, now)
			resp <- snapshotLast7Days(state.byDay, now)
		case resp := <-s.todayCh:
			now := time.Now().UTC()
			today := now.Format(dateLayout)
			rotateDay(&state, today)
			gc(&state, now)
			stat, ok := state.byDay[today]
			if !ok {
				stat = DayStat{Date: today}
			}
			resp <- stat
		case req := <-s.loadCh:
			now := time.Now().UTC()
			rotateDay(&state, now.Format(dateLayout))
			gc(&state, now)
			loadHistory(&state, req.history, now.Format(dateLayout))
			req.done <- struct{}{}
		}
	}
}

func loadHistory(state *runtimeState, history []DayStat, today string) {
	for _, stat := range history {
		if stat.Date == "" || stat.Date == today {
			continue
		}
		state.byDay[stat.Date] = DayStat{
			Date: stat.Date,
			PV:   stat.PV,
			UV:   stat.UV,
		}
	}
}

func recordIP(state *runtimeState, ev recordEvent) {
	atUTC := ev.at.UTC()
	today := atUTC.Format(dateLayout)
	rotateDay(state, today)
	gc(state, atUTC)

	ipHash := ev.ipHash
	stat := state.byDay[today]
	stat.Date = today
	if canCountPV(state, ipHash, atUTC) {
		stat.PV++
	}
	if _, ok := state.todayUV[ipHash]; !ok {
		state.todayUV[ipHash] = struct{}{}
		stat.UV++
	}
	state.byDay[today] = stat
}

func rotateDay(state *runtimeState, today string) {
	if state.today == today {
		return
	}
	state.today = today
	state.todayUV = make(map[string]struct{})
}

func gc(state *runtimeState, now time.Time) {
	nowUTC := now.UTC()
	today := nowUTC.Format(dateLayout)
	cutoff := parseDate(today).AddDate(0, 0, -(keepDays - 1)).Format(dateLayout)
	for day := range state.byDay {
		if day < cutoff {
			delete(state.byDay, day)
		}
	}

	pvCutoff := nowUTC.Add(-24 * time.Hour)
	for ipHash, ts := range state.lastPVAt {
		if ts.Before(pvCutoff) {
			delete(state.lastPVAt, ipHash)
		}
	}
}

func parseDate(day string) time.Time {
	t, err := time.ParseInLocation(dateLayout, day, time.UTC)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}

func hashIP(ip string) string {
	if ip == "" {
		ip = "unknown"
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(ip))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func canCountPV(state *runtimeState, ipHash string, now time.Time) bool {
	last, ok := state.lastPVAt[ipHash]
	if ok && now.Sub(last) < pvWindow {
		return false
	}
	state.lastPVAt[ipHash] = now
	return true
}

func snapshotLast7Days(byDay map[string]DayStat, now time.Time) []DayStat {
	nowUTC := now.UTC()
	points := make([]DayStat, 0, keepDays)
	for i := keepDays - 1; i >= 0; i-- {
		day := nowUTC.AddDate(0, 0, -i).Format(dateLayout)
		stat, ok := byDay[day]
		if !ok {
			stat = DayStat{Date: day}
		}
		points = append(points, stat)
	}
	return points
}

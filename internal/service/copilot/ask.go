// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// AskOption is one reply a question offers.
//
// Both fields are shown as text and nothing else. For ask_user they are the
// model's own words, so they are never parsed, never matched, and never allowed
// to decide anything — a label reading "已批准" is three characters in a label.
// The one place a label is compared is a write confirmation, and there the
// affirmative label was written by this process, not by the model.
type AskOption struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// AskQuestion is one thing put to a person.
//
// Empty Options means the answer is free text. Detail is a preformatted block
// rendered under the question — it is where a write confirmation puts the
// concrete Echo it is about, so consent is given to a specific change rather
// than to a sentence describing one.
type AskQuestion struct {
	ID          string      `json:"id"`
	Text        string      `json:"text"`
	Header      string      `json:"header,omitempty"`
	Detail      string      `json:"detail,omitempty"`
	Options     []AskOption `json:"options,omitempty"`
	Multi       bool        `json:"multi,omitempty"`
	Recommended *int        `json:"recommended,omitempty"`
}

// AskAnswer is one reply, addressed to its question by id.
//
// Selected carries the labels picked; Custom carries typed text. Both empty is
// a reply that said nothing, which is not consent.
type AskAnswer struct {
	QuestionID string   `json:"question_id"`
	Selected   []string `json:"selected,omitempty"`
	Custom     string   `json:"custom,omitempty"`
}

// Ask is one round of questions a run is blocked on. AskID is what an answer
// quotes back: a run that asks twice would otherwise have no way to tell the
// two replies apart.
type Ask struct {
	AskID     string        `json:"ask_id"`
	Questions []AskQuestion `json:"questions"`
}

// AskExchange is one finished round as it is stored on the turn: what was asked
// and what came back. Both halves, because reopening the conversation has to
// reproduce the decision, not just the prompt that led to it.
type AskExchange struct {
	Questions []AskQuestion `json:"questions"`
	Answers   []AskAnswer   `json:"answers"`
}

// What one round may carry. Overshooting these costs the surplus, never the
// turn: losing a whole answer because the model offered a seventh option is a
// worse outcome than the one it was reaching for.
const (
	maxAskQuestions   = 4
	maxAskOptions     = 6
	maxAskTextRunes   = 200
	maxAskLabelRunes  = 64
	maxAskDetailRunes = 600
)

// askEvent is what a parked run tells the client.
//
// It travels on a channel of its own rather than through the agent event
// stream, because the agent loop knows nothing about asking: the question is
// emitted from inside a tool call, which is exactly where the run is parked.
type askEvent struct {
	Open   *Ask
	Closed string
}

// askRegistry is where a parked run and the request carrying its answer meet.
//
// One process, one map. Ech0 is a single binary and a run dies with the request
// that started it, so the rendezvous never has to outlive either — there is no
// second replica to reach and no stream left to resume on. A durable slot would
// buy nothing and would have to be swept by something.
type askRegistry struct {
	mu      sync.Mutex
	pending map[string]*pendingAsk
}

type pendingAsk struct {
	userID  string
	answers chan []AskAnswer
}

func newAskRegistry() *askRegistry {
	return &askRegistry{pending: make(map[string]*pendingAsk)}
}

func (r *askRegistry) open(askID, userID string) *pendingAsk {
	p := &pendingAsk{userID: userID, answers: make(chan []AskAnswer, 1)}
	r.mu.Lock()
	r.pending[askID] = p
	r.mu.Unlock()
	return p
}

func (r *askRegistry) close(askID string) {
	r.mu.Lock()
	delete(r.pending, askID)
	r.mu.Unlock()
}

// answer delivers a person's reply to the run waiting for it.
//
// The entry is removed under the same lock that found it, and that removal is
// what makes a round answerable exactly once: two tabs racing the same click
// cannot both find it. An unknown id, an already-answered one and somebody
// else's are one refusal on purpose — telling them apart would let a caller
// probe for live rounds it does not own.
func (r *askRegistry) answer(askID, userID string, answers []AskAnswer) error {
	r.mu.Lock()
	p, ok := r.pending[askID]
	if !ok || p.userID != userID {
		r.mu.Unlock()
		return errors.New(commonModel.CHAT_ASK_NOT_PENDING)
	}
	delete(r.pending, askID)
	r.mu.Unlock()

	p.answers <- answers
	return nil
}

// asker is the run-local half of one conversation's questions: one user, one
// event stream, one turn's record.
//
// Bound at construction rather than looked up. A tool that could pick its own
// asker would be a tool that could answer someone else's question, or show its
// question on someone else's stream.
type asker struct {
	registry *askRegistry
	events   chan<- askEvent
	userID   string
	budget   time.Duration
	strs     askStrings

	mu  sync.Mutex
	log []AskExchange
}

// Ask registers the round, shows it, waits for the reply, and records the
// exchange — in that order.
//
// Registered before shown, and that order is not stylistic: reversed, a client
// fast enough to answer before the slot existed would have its answer refused
// and would never be asked again.
func (a *asker) Ask(ctx context.Context, questions []AskQuestion) ([]AskAnswer, error) {
	questions = clampAskQuestions(questions)
	if len(questions) == 0 {
		return nil, errors.New(a.strs.NoQuestions)
	}

	ask := Ask{AskID: uuidUtil.NewV7(), Questions: questions}
	pending := a.registry.open(ask.AskID, a.userID)
	defer a.registry.close(ask.AskID)

	if !a.emit(ctx, askEvent{Open: &ask}) {
		return nil, ctx.Err()
	}

	waitCtx, cancel := context.WithTimeout(ctx, a.budget)
	defer cancel()

	select {
	case answers := <-pending.answers:
		a.record(AskExchange{Questions: questions, Answers: answers})
		return answers, nil
	case <-waitCtx.Done():
		a.emit(ctx, askEvent{Closed: ask.AskID})
		return nil, errors.New(a.strs.Unanswered)
	}
}

func (a *asker) emit(ctx context.Context, ev askEvent) bool {
	select {
	case a.events <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}

func (a *asker) record(ex AskExchange) {
	a.mu.Lock()
	a.log = append(a.log, ex)
	a.mu.Unlock()
}

// exchanges is what the turn stores. Read once the run is over.
func (a *asker) exchanges() []AskExchange {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.log
}

func clampAskQuestions(qs []AskQuestion) []AskQuestion {
	kept := make([]AskQuestion, 0, min(len(qs), maxAskQuestions))
	for i := range qs {
		if len(kept) == maxAskQuestions {
			break
		}
		q := qs[i]
		q.ID = strings.TrimSpace(q.ID)
		q.Text = clampRunes(q.Text, maxAskTextRunes)
		q.Header = clampRunes(q.Header, maxAskLabelRunes)
		q.Detail = clampRunes(q.Detail, maxAskDetailRunes)
		if q.Text == "" {
			continue
		}
		if q.ID == "" {
			q.ID = "q" + strconv.Itoa(i+1)
		}
		q.Options = clampAskOptions(q.Options)
		q.Recommended = clampRecommended(q.Recommended, len(q.Options))
		kept = append(kept, q)
	}
	return kept
}

func clampAskOptions(opts []AskOption) []AskOption {
	if len(opts) == 0 {
		return nil
	}
	kept := make([]AskOption, 0, min(len(opts), maxAskOptions))
	for i := range opts {
		if len(kept) == maxAskOptions {
			break
		}
		label := clampRunes(opts[i].Label, maxAskLabelRunes)
		if label == "" || slices.ContainsFunc(kept, func(k AskOption) bool { return k.Label == label }) {
			continue
		}
		kept = append(kept, AskOption{Label: label, Description: clampRunes(opts[i].Description, maxAskTextRunes)})
	}
	return kept
}

// clampRecommended drops an index that points at no option. A mark on nothing
// would render as a mark on whatever happens to sit at that position.
func clampRecommended(idx *int, options int) *int {
	if idx == nil || *idx < 0 || *idx >= options {
		return nil
	}
	return idx
}

func clampRunes(s string, limit int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit])
}

// answerFor finds the reply to one question. A round can be answered in any
// order, so position is not a key.
func answerFor(answers []AskAnswer, questionID string) (AskAnswer, bool) {
	for i := range answers {
		if answers[i].QuestionID == questionID {
			return answers[i], true
		}
	}
	return AskAnswer{}, false
}

// AnswerAsk hands a person's reply to the run parked on it.
//
// The reply is not read here. Labels are text and typed input is the person's
// own words; both are stored and handed back to the tool that asked, and
// neither is matched against anything on the way through.
func (s *CopilotService) AnswerAsk(ctx context.Context, askID string, answers []AskAnswer) error {
	askID = strings.TrimSpace(askID)
	if askID == "" || len(answers) == 0 {
		return errors.New(commonModel.CHAT_ASK_NOT_PENDING)
	}
	userID := viewer.MustFromContext(ctx).UserID()
	return s.asks.answer(askID, userID, answers)
}

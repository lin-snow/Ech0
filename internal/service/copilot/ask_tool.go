// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/lin-snow/ech0/internal/agent"
)

// askUserTool lets the model put something to the person and wait for the reply.
//
// EffectRead, because asking changes nothing. That matters twice: a question
// must not trip the mutation gate, and it must not be mistakeable for the
// confirmation a write tool goes through — the model asking "may I delete it?"
// is not consent to a deletion, and the two mechanisms are kept apart so it
// cannot become one.
//
// Interactive, because the work is a person deciding.
func (s *CopilotService) askUserTool(a *asker, locale string) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "ask_user",
			Description: askUserDescriptionFor(locale),
			Parameters:  json.RawMessage(askUserSchema),
		},
		Effect:      agent.EffectRead,
		Interactive: true,
		Run: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var in askUserArgs
			if err := json.Unmarshal(args, &in); err != nil {
				return agent.ToolOutput{}, err
			}
			questions := in.questions()
			answers, err := a.Ask(ctx, questions)
			if err != nil {
				return agent.ToolOutput{}, err
			}
			return agent.ToolOutput{Content: answerText(a.strs, questions, answers)}, nil
		},
	}
}

// askUserArgs mirrors askUserSchema and only it. The model's words arrive as
// data and leave as data.
type askUserArgs struct {
	Questions []struct {
		ID       string `json:"id"`
		Question string `json:"question"`
		Header   string `json:"header"`
		Options  []struct {
			Label       string `json:"label"`
			Description string `json:"description"`
		} `json:"options"`
		Multi       bool `json:"multi"`
		Recommended *int `json:"recommended"`
	} `json:"questions"`
}

func (in askUserArgs) questions() []AskQuestion {
	qs := make([]AskQuestion, 0, len(in.Questions))
	for _, q := range in.Questions {
		options := make([]AskOption, 0, len(q.Options))
		for _, opt := range q.Options {
			options = append(options, AskOption{Label: opt.Label, Description: opt.Description})
		}
		qs = append(qs, AskQuestion{
			ID:          q.ID,
			Text:        q.Question,
			Header:      q.Header,
			Options:     options,
			Multi:       q.Multi,
			Recommended: q.Recommended,
		})
	}
	return qs
}

// askUserSchema is hand-written, like every other tool schema in this package.
// Only questions is required, and within a question only its id and its text:
// options the model cannot honestly enumerate must be omissible, or it will
// invent them to satisfy the schema — and an invented shortlist is a leading
// question.
const askUserSchema = `{"type":"object","properties":{"questions":{"type":"array","description":"本轮要问的问题；相关的问题放在同一次调用里","items":{"type":"object","properties":{"id":{"type":"string","description":"这个问题的短名（如 \"tag\"），回答会带着它回来"},"question":{"type":"string","description":"问题本身，用用户的语言，一次只问一件事"},"header":{"type":"string","description":"可选，问题上方两三个字的小标题（如 \"标签\"）"},"options":{"type":"array","description":"提供的选项；无法确定真实候选时整个省略，让用户自由输入","items":{"type":"object","properties":{"label":{"type":"string","description":"按钮上显示的文字，几个字，不要整句"},"description":{"type":"string","description":"标签下面的第二行：选它意味着什么，尤其是代价"}},"required":["label"]}},"multi":{"type":"boolean","description":"true 表示可以同时选中多个选项"},"recommended":{"type":"integer","description":"你倾向的那个选项在 options 里的下标；仅作提示，永远不会被自动选中"}},"required":["id","question"]}}},"required":["questions"]}`

// mutatingTool builds a write tool as an agent.Mutation, so the loop — not this
// package, and certainly not the model — is what enforces the ordering.
//
// plan works out what would change and returns the question describing it plus
// the change itself. It must not perform anything: it runs before anyone has
// agreed to anything. Confirm is handed to the loop rather than called here,
// which is the whole point of the arrangement: there is no code path from a
// tool call to a write that does not pass through it, so "the model forgot to
// confirm" is not a state this can be in.
//
// Options[0] of the question is the affirmative by convention, and consented
// compares the person's pick against it. That label is written by this package;
// a label the model authored would let the model define consent.
func mutatingTool(
	a *asker,
	def agent.ToolDef,
	plan func(context.Context, json.RawMessage) (AskQuestion, applyFunc, error),
) agent.Tool {
	return agent.Tool{
		Def:    def,
		Effect: agent.EffectMutate,
		Mutation: &agent.Mutation{
			Plan: func(ctx context.Context, args json.RawMessage) (agent.Plan, error) {
				question, apply, err := plan(ctx, args)
				if err != nil {
					return agent.Plan{}, err
				}
				return agent.Plan{Prompt: question, Apply: apply}, nil
			},
			Confirm: func(ctx context.Context, p agent.Plan) (agent.Decision, error) {
				question, ok := p.Prompt.(AskQuestion)
				if !ok {
					return agent.Decision{Refusal: a.strs.declined("")}, nil
				}
				answers, err := a.Ask(ctx, []AskQuestion{question})
				if err != nil {
					return agent.Decision{}, err
				}
				approved, note := consented(question, answers)
				return agent.Decision{Approved: approved, Refusal: a.strs.declined(note)}, nil
			},
		},
	}
}

// applyFunc is the change half of a write tool: everything the tool would do if
// the person says yes, and nothing it does before then.
type applyFunc = func(ctx context.Context) (agent.ToolOutput, error)

// consented reads a confirmation, and everything that is not an exact yes is a
// no.
//
// Typed text is never consent, not even text that reads like agreement: the
// person answered something other than the question, and the model is where
// that belongs — it comes back as the reason the write did not happen. The
// empty reply an expired picker produces lands here too, and lands on no.
func consented(q AskQuestion, answers []AskAnswer) (bool, string) {
	ans, ok := answerFor(answers, q.ID)
	if !ok {
		return false, ""
	}
	if note := strings.TrimSpace(ans.Custom); note != "" {
		return false, note
	}
	if len(q.Options) == 0 || len(ans.Selected) != 1 {
		return false, ""
	}
	return ans.Selected[0] == q.Options[0].Label, ""
}

// answerText is what the model reads. One question answers as itself; a round
// answers as a list, so the model cannot lose which reply went with which.
func answerText(strs askStrings, questions []AskQuestion, answers []AskAnswer) string {
	if len(questions) == 1 {
		return replyOf(strs, answers, questions[0].ID)
	}
	var b strings.Builder
	b.WriteString(strs.Answers)
	for i := range questions {
		b.WriteString("\n- " + questions[i].Text + "：" + replyOf(strs, answers, questions[i].ID))
	}
	return b.String()
}

// replyOf prefers typed text over a picked label: someone who typed instead of
// clicking meant the words, not the button they bypassed.
func replyOf(strs askStrings, answers []AskAnswer, questionID string) string {
	ans, ok := answerFor(answers, questionID)
	if !ok {
		return strs.NoAnswer
	}
	if custom := strings.TrimSpace(ans.Custom); custom != "" {
		return custom
	}
	if len(ans.Selected) > 0 {
		return strings.Join(ans.Selected, "、")
	}
	return strs.NoAnswer
}

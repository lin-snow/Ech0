// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lin-snow/ech0/internal/config"
)

const (
	// confirmationStateTTL bounds how long a client may take to answer a
	// destructive-action confirmation before the request must be restarted.
	confirmationStateTTL  = 5 * time.Minute
	confirmationKey       = "confirm"
	confirmationMethod    = "elicitation/create"
	confirmationTitle     = "Confirm destructive action"
	confirmationTargetMax = 120
)

// InputRequest is one server-to-client request carried by an
// InputRequiredResult. MCP 2026-07-28 replaces server-initiated requests with
// this multi round-trip requests (MRTR) pattern.
type InputRequest struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}

// InputRequiredResult is the MRTR interim result. It intentionally does not
// embed ResultEnvelope: resultType must stay "input_required" and cache hints
// must not be attached.
type InputRequiredResult struct {
	ResultType    string                  `json:"resultType"`
	InputRequests map[string]InputRequest `json:"inputRequests,omitempty"`
	RequestState  string                  `json:"requestState,omitempty"`
	Meta          map[string]any          `json:"_meta,omitempty"`
}

type elicitationFormParams struct {
	Mode            string         `json:"mode"`
	Message         string         `json:"message"`
	RequestedSchema map[string]any `json:"requestedSchema"`
}

// confirmationState is the integrity-protected context echoed back by the
// client. It binds the confirmation to one principal, one tool and one exact
// argument set so a retry cannot be replayed against a different target.
type confirmationState struct {
	UserID  string `json:"uid"`
	Tool    string `json:"tool"`
	ArgsSHA string `json:"args"`
	Expiry  int64  `json:"exp"`
}

type confirmationResponse struct {
	Action  string `json:"action"`
	Content struct {
		Confirm bool `json:"confirm"`
	} `json:"content"`
}

func clientSupportsElicitation(params *requestParams) bool {
	if params == nil {
		return false
	}
	capabilities, ok := params.Meta[metaKeyClientCapabilities].(map[string]any)
	if !ok {
		return false
	}
	_, ok = capabilities["elicitation"]
	return ok
}

func newInputRequired(tool string, args map[string]any, userID string) (*InputRequiredResult, error) {
	state, err := signConfirmationState(userID, tool, args)
	if err != nil {
		return nil, err
	}
	return &InputRequiredResult{
		ResultType: resultTypeInputRequired,
		InputRequests: map[string]InputRequest{
			confirmationKey: {
				Method: confirmationMethod,
				Params: elicitationFormParams{
					Mode:            "form",
					Message:         confirmationMessage(tool, args),
					RequestedSchema: confirmationSchema(),
				},
			},
		},
		RequestState: state,
		Meta: map[string]any{
			metaKeyServerInfo: serverInfo(),
		},
	}, nil
}

func confirmationSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			confirmationKey: map[string]any{
				"type":        "boolean",
				"title":       confirmationTitle,
				"description": "Run this operation. It cannot be undone.",
			},
		},
		"required": []string{confirmationKey},
	}
}

// confirmationMessage names the concrete target when the arguments expose one,
// so the user is not asked to approve an opaque tool call.
func confirmationMessage(tool string, args map[string]any) string {
	target := ""
	for _, key := range []string{"uuid", "post_id", "id", "name", "tag", "url"} {
		if value, ok := args[key].(string); ok && strings.TrimSpace(value) != "" {
			target = strings.TrimSpace(value)
			break
		}
	}
	if target == "" {
		return fmt.Sprintf("Run %s? This operation cannot be undone.", tool)
	}
	return fmt.Sprintf("Run %s on %q? This operation cannot be undone.", tool, truncateRunes(target, confirmationTargetMax))
}

func truncateRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit]) + "…"
}

func signConfirmationState(userID, tool string, args map[string]any) (string, error) {
	payload, err := json.Marshal(confirmationState{
		UserID:  userID,
		Tool:    tool,
		ArgsSHA: argsDigest(args),
		Expiry:  time.Now().UTC().Add(confirmationStateTTL).Unix(),
	})
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, config.Config().Security.JWTSecret)
	mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func verifyConfirmationState(state, userID, tool string, args map[string]any) error {
	payloadPart, signaturePart, ok := strings.Cut(state, ".")
	if !ok {
		return errors.New("malformed requestState")
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return errors.New("malformed requestState")
	}
	signature, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return errors.New("malformed requestState")
	}

	mac := hmac.New(sha256.New, config.Config().Security.JWTSecret)
	mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return errors.New("requestState signature mismatch")
	}

	var decoded confirmationState
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return errors.New("malformed requestState")
	}
	if decoded.UserID != userID {
		return errors.New("requestState belongs to a different principal")
	}
	if decoded.Tool != tool {
		return errors.New("requestState belongs to a different tool")
	}
	if decoded.ArgsSHA != argsDigest(args) {
		return errors.New("requestState does not match the tool arguments")
	}
	if time.Now().UTC().Unix() > decoded.Expiry {
		return errors.New("requestState expired")
	}
	return nil
}

func argsDigest(args map[string]any) string {
	if len(args) == 0 {
		return ""
	}
	payload, err := json.Marshal(args)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// readConfirmation interprets the client's inputResponses entry. responded is
// false when the client omitted the response entirely (it may be retrying the
// original request without fulfilling the elicitation).
func readConfirmation(params *requestParams) (responded bool, approved bool, err error) {
	if params == nil || len(params.InputResponses) == 0 {
		return false, false, nil
	}
	raw, ok := params.InputResponses[confirmationKey]
	if !ok {
		return false, false, nil
	}
	var response confirmationResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return true, false, err
	}
	return true, response.Action == "accept" && response.Content.Confirm, nil
}

func cancelledResult() *ToolCallResult {
	return textError("Operation cancelled by the user.")
}

// consumeConfirmation enforces one-shot use of an approval. A destructive
// requestState must not authorize a second execution, even if the client
// replays the exact same retry.
func (s *Server) consumeConfirmation(state string) bool {
	now := time.Now().UTC()
	s.confirmMu.Lock()
	defer s.confirmMu.Unlock()

	for key, expiry := range s.confirmed {
		if now.After(expiry) {
			delete(s.confirmed, key)
		}
	}
	if _, used := s.confirmed[state]; used {
		return false
	}
	s.confirmed[state] = now.Add(confirmationStateTTL)
	return true
}

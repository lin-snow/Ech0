// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/httperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type problemBody struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func runExecute(t *testing.T, res Response) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	Execute(func(*gin.Context) Response { return res })(c)
	return rec
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) problemBody {
	t.Helper()
	require.Equal(t, "application/problem+json; charset=utf-8", rec.Header().Get("Content-Type"))
	var body problemBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

func TestExecute_Payload(t *testing.T) {
	cases := []struct {
		name       string
		res        Response
		wantStatus int
	}{
		{name: "payload", res: Payload("payload"), wantStatus: http.StatusOK},
		{name: "created", res: Created(map[string]string{"id": "1"}), wantStatus: http.StatusCreated},
		{name: "no-content", res: NoContent(http.StatusNoContent), wantStatus: http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := runExecute(t, tc.res)
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus == http.StatusNoContent {
				assert.Empty(t, rec.Body.String())
			}
		})
	}
}

func TestExecute_PayloadIsTheResource(t *testing.T) {
	rec := runExecute(t, Payload(map[string]string{"name": "ech0"}))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"name":"ech0"}`, rec.Body.String())
}

func TestExecute_Failure(t *testing.T) {
	cases := []struct {
		name       string
		res        Response
		wantStatus int
		wantDetail string
	}{
		{
			name:       "status-error-keeps-its-message",
			res:        Failed(httperr.Edge(http.StatusBadRequest, "raw msg")),
			wantStatus: http.StatusBadRequest,
			wantDetail: "raw msg",
		},
		{
			name:       "plain-error-is-an-internal-error",
			res:        Failed(errors.New("boom")),
			wantStatus: http.StatusInternalServerError,
			wantDetail: "internal server error",
		},
		{
			name:       "explicit-status-wins",
			res:        FailedWithStatus(http.StatusTooManyRequests, httperr.Edge(http.StatusBadRequest, "slow down")),
			wantStatus: http.StatusTooManyRequests,
			wantDetail: "slow down",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := runExecute(t, tc.res)

			require.Equal(t, tc.wantStatus, rec.Code)
			body := decodeProblem(t, rec)
			assert.Equal(t, tc.wantStatus, body.Status)
			assert.Equal(t, http.StatusText(tc.wantStatus), body.Title)
			assert.Equal(t, tc.wantDetail, body.Detail)
		})
	}
}

func TestExecute_ServerErrorHidesItsCause(t *testing.T) {
	rec := runExecute(t, FailedWithStatus(http.StatusInternalServerError, errors.New("dial tcp 10.0.0.1:5432: i/o timeout")))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	body := decodeProblem(t, rec)
	assert.Equal(t, "internal server error", body.Detail)
	assert.NotContains(t, rec.Body.String(), "5432")
}

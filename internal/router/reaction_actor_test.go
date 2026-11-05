// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestSetupRouter_ReactionWritesUseAuthenticatedActor guards the read/write
// actor symmetry: hydration runs behind optional auth, so the write routes
// must resolve the same actor. When they did not, an authenticated visitor's
// reaction was stored under "v:<visitor>" while reads looked up "u:<user>",
// making every reaction look un-reacted after a reload.
func TestSetupRouter_ReactionWritesUseAuthenticatedActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := helpers.NewTestDB(t)

	repo := echoRepository.NewEchoRepository(func() *gorm.DB { return db }, helpers.NewTestCache())
	svc := echoService.NewEchoService(nil, nil, nil, repo, func() *busen.Bus { return busen.New() })

	bundle := buildTestHandlers()
	bundle.EchoHandler = echoHandler.NewEchoHandler(svc)

	engine := gin.New()
	SetupRouter(engine, bundle, buildTestMWDeps())

	const echoID = "0f0f0f0f-1111-4222-8333-444444444444"
	require.NoError(t, db.Create(&echoModel.Echo{ID: echoID, Content: "actor test", UserID: "u-owner"}).Error)

	user := userModel.User{ID: "u-reactor", Username: "reactor"}
	token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/echo/"+echoID+"/reactions/like", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Ech0-Visitor", "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var stored echoModel.Reaction
	require.NoError(t, db.First(&stored, "echo_id = ? AND emoji = ?", echoID, "like").Error)
	assert.Equal(t, "u:"+user.ID, stored.Actor)
}

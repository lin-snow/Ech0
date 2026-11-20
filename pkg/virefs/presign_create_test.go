// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateOnlyUploadSignsConstraints(t *testing.T) {
	fs, err := NewObjectFSFromConfig(t.Context(), &S3Config{Provider: ProviderR2, Endpoint: "https://objects.invalid", Bucket: "test", AccessKey: "example", SecretKey: "example"})
	require.NoError(t, err)
	req, err := fs.PresignCreate(t.Context(), "test.png", 1024, "image/png", 10*time.Minute)
	require.NoError(t, err)
	u, err := url.Parse(req.URL)
	require.NoError(t, err)
	signed := strings.Split(u.Query().Get("X-Amz-SignedHeaders"), ";")
	require.Contains(t, signed, "content-length")
	require.Contains(t, signed, "content-type")
	require.Contains(t, signed, "if-none-match")
	require.Equal(t, "600", u.Query().Get("X-Amz-Expires"))
}

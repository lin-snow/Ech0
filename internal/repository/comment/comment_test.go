// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository_test

import (
	"context"
	"testing"
	"time"

	model "github.com/lin-snow/ech0/internal/model/comment"
	commentRepository "github.com/lin-snow/ech0/internal/repository/comment"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*commentRepository.CommentRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return commentRepository.NewCommentRepository(func() *gorm.DB { return db }), db
}

func newComment(opts ...func(*model.Comment)) model.Comment {
	c := model.Comment{
		EchoID:   "echo-1",
		Nickname: "alice",
		Email:    "alice@example.com",
		Content:  "hello world",
		Status:   model.StatusPending,
		Source:   model.SourceGuest,
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func insert(t *testing.T, repo *commentRepository.CommentRepository, c model.Comment) string {
	t.Helper()
	require.NoError(t, repo.CreateComment(context.Background(), &c))
	require.NotEmpty(t, c.ID, "BeforeCreate 应生成 UUID")
	return c.ID
}

func countRows(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&model.Comment{}).Count(&n).Error)
	return n
}

func TestCountWithin_WindowAndFieldMatch(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Unix()

	insert(t, repo, newComment(func(c *model.Comment) { c.IPHash = "ip1"; c.CreatedAt = now - 5 }))
	insert(t, repo, newComment(func(c *model.Comment) { c.IPHash = "ip1"; c.CreatedAt = now - 10000 }))
	insert(t, repo, newComment(func(c *model.Comment) { c.IPHash = "ip2"; c.CreatedAt = now - 5 }))

	insert(t, repo, newComment(func(c *model.Comment) { c.Email = "bob@example.com"; c.CreatedAt = now - 5 }))
	insert(t, repo, newComment(func(c *model.Comment) { c.Email = "bob@example.com"; c.CreatedAt = now - 10000 }))
	insert(t, repo, newComment(func(c *model.Comment) { c.UserID = new("u-1"); c.CreatedAt = now - 5 }))
	insert(t, repo, newComment(func(c *model.Comment) { c.UserID = new("u-1"); c.CreatedAt = now - 10000 }))

	const window int64 = 3600

	t.Run("ip within window", func(t *testing.T) {
		got, err := repo.CountByIPWithin(ctx, "ip1", window)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got, "窗口外的 ip1 旧行不应计入")
	})
	t.Run("ip other value isolated", func(t *testing.T) {
		got, err := repo.CountByIPWithin(ctx, "ip2", window)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got)
	})
	t.Run("ip no match", func(t *testing.T) {
		got, err := repo.CountByIPWithin(ctx, "ip-nope", window)
		require.NoError(t, err)
		assert.Equal(t, int64(0), got)
	})
	t.Run("email within window", func(t *testing.T) {
		got, err := repo.CountByEmailWithin(ctx, "bob@example.com", window)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got)
	})
	t.Run("user within window", func(t *testing.T) {
		got, err := repo.CountByUserWithin(ctx, "u-1", window)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got)
	})
}

func TestCountWithin_TinyWindowExcludesEverything(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Unix()

	insert(t, repo, newComment(func(c *model.Comment) { c.IPHash = "ip-tiny"; c.CreatedAt = now - 30 }))

	got, err := repo.CountByIPWithin(ctx, "ip-tiny", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(0), got)
}

func TestCountWithin_EmptyValueShortCircuits(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Unix()
	insert(t, repo, newComment(func(c *model.Comment) { c.Email = ""; c.CreatedAt = now - 1 }))

	cases := []struct {
		name string
		call func() (int64, error)
	}{
		{"empty ip", func() (int64, error) { return repo.CountByIPWithin(ctx, "", 3600) }},
		{"blank ip", func() (int64, error) { return repo.CountByIPWithin(ctx, "   ", 3600) }},
		{"empty email", func() (int64, error) { return repo.CountByEmailWithin(ctx, "", 3600) }},
		{"blank email", func() (int64, error) { return repo.CountByEmailWithin(ctx, "  ", 3600) }},
		{"empty user", func() (int64, error) { return repo.CountByUserWithin(ctx, "", 3600) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.call()
			require.NoError(t, err)
			assert.Equal(t, int64(0), got, "空/空白值必须短路返回 0")
		})
	}
}

func TestExistsRecentDuplicate_IdentityPrecedence(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Unix()

	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "echo-dup"
		c.Content = "duplicate content"
		c.UserID = new("u-x")
		c.Email = "dup@example.com"
		c.IPHash = "ip-x"
		c.CreatedAt = now - 5
	}))

	const window int64 = 3600

	cases := []struct {
		name                           string
		echoID, content, email, ipHash string
		userID                         string
		want                           bool
	}{
		{
			name:   "user matches even with wrong email/ip",
			echoID: "echo-dup", content: "duplicate content",
			email: "other@example.com", ipHash: "ip-other", userID: "u-x", want: true,
		},
		{
			name:   "user mismatch wins over matching email/ip",
			echoID: "echo-dup", content: "duplicate content",
			email: "dup@example.com", ipHash: "ip-x", userID: "u-other", want: false,
		},
		{
			name:   "email used when user empty",
			echoID: "echo-dup", content: "duplicate content",
			email: "dup@example.com", ipHash: "ip-other", userID: "", want: true,
		},
		{
			name:   "email mismatch wins over matching ip when user empty",
			echoID: "echo-dup", content: "duplicate content",
			email: "no@example.com", ipHash: "ip-x", userID: "", want: false,
		},
		{
			name:   "ip used when user and email empty",
			echoID: "echo-dup", content: "duplicate content",
			email: "", ipHash: "ip-x", userID: "", want: true,
		},
		{
			name:   "no identity matches by echo+content only",
			echoID: "echo-dup", content: "duplicate content",
			email: "", ipHash: "", userID: "", want: true,
		},
		{
			name:   "different content not a duplicate",
			echoID: "echo-dup", content: "totally different",
			email: "", ipHash: "", userID: "u-x", want: false,
		},
		{
			name:   "different echo not a duplicate",
			echoID: "echo-other", content: "duplicate content",
			email: "", ipHash: "", userID: "u-x", want: false,
		},
		{
			name:   "trims echo and content before matching",
			echoID: "  echo-dup  ", content: "  duplicate content  ",
			email: "", ipHash: "", userID: "u-x", want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.ExistsRecentDuplicate(ctx, tc.echoID, tc.content, tc.email, tc.ipHash, tc.userID, window)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestExistsRecentDuplicate_WindowExcludesOld(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Unix()

	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "echo-old"
		c.Content = "old duplicate"
		c.UserID = new("u-old")
		c.CreatedAt = now - 10000
	}))

	t.Run("outside window -> false", func(t *testing.T) {
		got, err := repo.ExistsRecentDuplicate(ctx, "echo-old", "old duplicate", "", "", "u-old", 3600)
		require.NoError(t, err)
		assert.False(t, got)
	})
	t.Run("inside large window -> true", func(t *testing.T) {
		got, err := repo.ExistsRecentDuplicate(ctx, "echo-old", "old duplicate", "", "", "u-old", 100000)
		require.NoError(t, err)
		assert.True(t, got)
	})
}

func seedForList(t *testing.T, repo *commentRepository.CommentRepository) int64 {
	t.Helper()
	base := time.Now().UTC().Unix() - 1000
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "e1"
		c.Status = model.StatusApproved
		c.Hot = true
		c.Nickname = "alice"
		c.Content = "first golang post"
		c.CreatedAt = base + 1
	}))
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "e1"
		c.Status = model.StatusPending
		c.Nickname = "bob"
		c.Email = "needle@inbox.com"
		c.Content = "second"
		c.CreatedAt = base + 2
	}))
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "e2"
		c.Status = model.StatusApproved
		c.Nickname = "carol"
		c.Content = "third about golang"
		c.CreatedAt = base + 3
	}))
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "e2"
		c.Status = model.StatusRejected
		c.Nickname = "dave"
		c.Content = "fourth"
		c.CreatedAt = base + 4
	}))
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "e1"
		c.Status = model.StatusApproved
		c.Hot = true
		c.Nickname = "erin"
		c.Content = "fifth"
		c.CreatedAt = base + 5
	}))
	return base
}

func TestListComments_Filters(t *testing.T) {
	repo, _ := newRepo(t)
	seedForList(t, repo)
	ctx := context.Background()

	cases := []struct {
		name      string
		query     model.ListCommentQuery
		wantTotal int64
		wantItems int
	}{
		{
			name:      "no filter returns all",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100},
			wantTotal: 5, wantItems: 5,
		},
		{
			name:      "filter by echo id",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, EchoID: "e1"},
			wantTotal: 3, wantItems: 3,
		},
		{
			name:      "filter by status approved",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, Status: string(model.StatusApproved)},
			wantTotal: 3, wantItems: 3,
		},
		{
			name:      "filter by hot true",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, Hot: new(true)},
			wantTotal: 2, wantItems: 2,
		},
		{
			name:      "filter by hot false",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, Hot: new(false)},
			wantTotal: 3, wantItems: 3,
		},
		{
			name:      "keyword matches content",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, Keyword: "golang"},
			wantTotal: 2, wantItems: 2,
		},
		{
			name:      "keyword matches email",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, Keyword: "needle"},
			wantTotal: 1, wantItems: 1,
		},
		{
			name:      "keyword matches nickname",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, Keyword: "carol"},
			wantTotal: 1, wantItems: 1,
		},
		{
			name:      "keyword trimmed and no match",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, Keyword: "   "},
			wantTotal: 5, wantItems: 5,
		},
		{
			name:      "combined echo and status",
			query:     model.ListCommentQuery{Page: 1, PageSize: 100, EchoID: "e1", Status: string(model.StatusApproved)},
			wantTotal: 2, wantItems: 2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := repo.ListComments(ctx, tc.query)
			require.NoError(t, err)
			assert.Equal(t, tc.wantTotal, res.Total)
			assert.Len(t, res.Items, tc.wantItems)
		})
	}
}

func TestListComments_PaginationAndOrder(t *testing.T) {
	repo, _ := newRepo(t)
	seedForList(t, repo)
	ctx := context.Background()

	page1, err := repo.ListComments(ctx, model.ListCommentQuery{Page: 1, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), page1.Total, "Total 是全量而非本页数量")
	require.Len(t, page1.Items, 2)
	assert.Equal(t, "fifth", page1.Items[0].Content)
	assert.Equal(t, "fourth", page1.Items[1].Content)

	page2, err := repo.ListComments(ctx, model.ListCommentQuery{Page: 2, PageSize: 2})
	require.NoError(t, err)
	require.Len(t, page2.Items, 2)
	assert.Equal(t, "third about golang", page2.Items[0].Content)
	assert.Equal(t, "second", page2.Items[1].Content)

	page3, err := repo.ListComments(ctx, model.ListCommentQuery{Page: 3, PageSize: 2})
	require.NoError(t, err)
	require.Len(t, page3.Items, 1)
	assert.Equal(t, "first golang post", page3.Items[0].Content)
}

func TestBatchUpdateStatus(t *testing.T) {
	t.Run("empty ids is a no-op", func(t *testing.T) {
		repo, db := newRepo(t)
		id := insert(t, repo, newComment(func(c *model.Comment) { c.Status = model.StatusPending }))
		require.NoError(t, repo.BatchUpdateStatus(context.Background(), nil, model.StatusApproved))

		got, err := repo.GetCommentByID(context.Background(), id)
		require.NoError(t, err)
		assert.Equal(t, model.StatusPending, got.Status)
		assert.Equal(t, int64(1), countRows(t, db))
	})

	t.Run("updates only targeted ids", func(t *testing.T) {
		repo, _ := newRepo(t)
		ctx := context.Background()
		id1 := insert(t, repo, newComment(func(c *model.Comment) { c.Status = model.StatusPending }))
		id2 := insert(t, repo, newComment(func(c *model.Comment) { c.Status = model.StatusPending }))
		id3 := insert(t, repo, newComment(func(c *model.Comment) { c.Status = model.StatusPending }))

		require.NoError(t, repo.BatchUpdateStatus(ctx, []string{id1, id2}, model.StatusApproved))

		c1, _ := repo.GetCommentByID(ctx, id1)
		c2, _ := repo.GetCommentByID(ctx, id2)
		c3, _ := repo.GetCommentByID(ctx, id3)
		assert.Equal(t, model.StatusApproved, c1.Status)
		assert.Equal(t, model.StatusApproved, c2.Status)
		assert.Equal(t, model.StatusPending, c3.Status, "未列入的 id 不应被改动")
	})
}

func TestBatchDelete(t *testing.T) {
	t.Run("empty ids is a no-op", func(t *testing.T) {
		repo, db := newRepo(t)
		insert(t, repo, newComment())
		require.NoError(t, repo.BatchDelete(context.Background(), nil))
		assert.Equal(t, int64(1), countRows(t, db))
	})

	t.Run("deletes only targeted ids", func(t *testing.T) {
		repo, db := newRepo(t)
		ctx := context.Background()
		id1 := insert(t, repo, newComment())
		id2 := insert(t, repo, newComment())
		id3 := insert(t, repo, newComment())

		require.NoError(t, repo.BatchDelete(ctx, []string{id1, id2}))
		assert.Equal(t, int64(1), countRows(t, db))

		_, err := repo.GetCommentByID(ctx, id1)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		remaining, err := repo.GetCommentByID(ctx, id3)
		require.NoError(t, err)
		assert.Equal(t, id3, remaining.ID)
	})
}

func TestCreateGetUpdateDelete(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()

	id := insert(t, repo, newComment(func(c *model.Comment) { c.Status = model.StatusPending; c.Hot = false }))

	got, err := repo.GetCommentByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPending, got.Status)

	require.NoError(t, repo.UpdateCommentStatus(ctx, id, model.StatusApproved))
	require.NoError(t, repo.UpdateCommentHot(ctx, id, true))
	got, err = repo.GetCommentByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, model.StatusApproved, got.Status)
	assert.True(t, got.Hot)

	require.NoError(t, repo.DeleteComment(ctx, id))
	_, err = repo.GetCommentByID(ctx, id)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListPublicByEchoID(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	base := time.Now().UTC().Unix() - 100

	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "pub-e"
		c.Status = model.StatusApproved
		c.Content = "earlier"
		c.CreatedAt = base + 1
	}))
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "pub-e"
		c.Status = model.StatusApproved
		c.Content = "later"
		c.CreatedAt = base + 2
	}))
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "pub-e"
		c.Status = model.StatusPending
		c.Content = "hidden"
		c.CreatedAt = base + 3
	}))
	insert(t, repo, newComment(func(c *model.Comment) {
		c.EchoID = "other-e"
		c.Status = model.StatusApproved
		c.CreatedAt = base + 4
	}))

	out, err := repo.ListPublicByEchoID(ctx, "pub-e")
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "earlier", out[0].Content, "应按 created_at 升序")
	assert.Equal(t, "later", out[1].Content)
}

func TestListPublicComments(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	base := time.Now().UTC().Unix() - 100

	for i := range 3 {
		insert(t, repo, newComment(func(c *model.Comment) {
			c.Status = model.StatusApproved
			c.CreatedAt = base + int64(i)
		}))
	}
	insert(t, repo, newComment(func(c *model.Comment) { c.Status = model.StatusPending; c.CreatedAt = base + 10 }))

	t.Run("only approved, desc, limited", func(t *testing.T) {
		out, err := repo.ListPublicComments(ctx, 2)
		require.NoError(t, err)
		require.Len(t, out, 2)
		assert.Equal(t, model.StatusApproved, out[0].Status)
		assert.GreaterOrEqual(t, out[0].CreatedAt, out[1].CreatedAt)
	})

	t.Run("limit larger than rows", func(t *testing.T) {
		out, err := repo.ListPublicComments(ctx, 100)
		require.NoError(t, err)
		assert.Len(t, out, 3, "pending 不计入")
	})
}

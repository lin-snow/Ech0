// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

func (a *Adapter) registerEchoTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:        "search_posts",
		Title:       "Search Posts",
		Description: "Search posts by keyword and/or tag IDs. Returns paginated results: {items, total, page, page_size}. All parameters are optional; omitting everything returns the latest posts.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":      map[string]any{"type": "string", "description": "Full-text search keyword (matched against post content)"},
				"tag_ids":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Filter by one or more tag UUIDs (AND logic)"},
				"page":       map[string]any{"type": "integer", "description": "Page number, 1-based", "default": 1},
				"page_size":  map[string]any{"type": "integer", "description": "Results per page (1–100)", "default": 20},
				"sort_by":    map[string]any{"type": "string", "enum": []string{"created_at", "fav_count"}, "description": "Field to sort by", "default": "created_at"},
				"sort_order": map[string]any{"type": "string", "enum": []string{"desc", "asc"}, "description": "Sort direction", "default": "desc"},
			},
		},
		Annotations: readOnlyHints(),
	}, a.searchPosts, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_post",
		Title:       "Get Post",
		Description: "Retrieve a single post by UUID. Returns the full post object including content, tags, like count, and timestamps.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "Post UUID"},
			},
		},
		Annotations: readOnlyHints(),
	}, a.getPost, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "list_tags",
		Title:       "List Tags",
		Description: "List every tag in the system. Each entry contains the tag id, name, and usage count. No parameters needed.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Annotations: readOnlyHints(),
	}, a.listTags, authModel.ScopeEchoRead)

	echoFileSchema := map[string]any{
		"type":     "object",
		"required": []string{"file_id"},
		"properties": map[string]any{
			"file_id":    map[string]any{"type": "string", "format": "uuid", "description": "File UUID (obtained via the file upload REST API)"},
			"sort_order": map[string]any{"type": "integer", "description": "Display order (0-based); defaults to array index if omitted"},
		},
	}
	extensionSchema := map[string]any{
		"type":     "object",
		"required": []string{"type", "payload"},
		"properties": map[string]any{
			"type": map[string]any{
				"type":        "string",
				"enum":        []string{"MUSIC", "VIDEO", "GITHUBPROJ", "WEBSITE"},
				"description": "Extension type",
			},
			"payload": map[string]any{
				"type":        "object",
				"description": "Type-specific data. MUSIC: {url}; VIDEO: {videoId}; GITHUBPROJ: {repoUrl}; WEBSITE: {title, site}",
			},
		},
	}
	layoutEnum := []string{"waterfall", "grid", "horizontal", "carousel", "stack"}

	reg.RegisterTool(ToolDefinition{
		Name:  "create_post",
		Title: "Create Post",
		Description: "Create a new post. At least one of content, echo_files, or extension must be provided. " +
			"Returns {id, message}. To attach files, upload them first via the REST API " +
			"(POST <base_url>/api/files/upload, multipart/form-data, field 'file') " +
			"and pass the returned id values in echo_files. " +
			"Read the resource ech0://guide/file-upload for full upload instructions.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"content":    map[string]any{"type": "string", "description": "Post body (Markdown supported)"},
				"tags":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tag names to attach; non-existent tags are created automatically"},
				"private":    map[string]any{"type": "boolean", "description": "Mark the post as private (only visible to the owner)", "default": false},
				"layout":     map[string]any{"type": "string", "enum": layoutEnum, "description": "Image layout style", "default": "waterfall"},
				"echo_files": map[string]any{"type": "array", "items": echoFileSchema, "description": "Attached files (images, etc.) referenced by file_id"},
				"extension":  extensionSchema,
			},
		},
		Annotations: createHints(),
	}, a.createPost, authModel.ScopeEchoWrite)

	reg.RegisterTool(ToolDefinition{
		Name:  "update_post",
		Title: "Update Post",
		Description: "Update an existing post. Only supplied fields are changed, except echo_files and extension which fully replace " +
			"the existing values when provided. Returns {id, message}.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id":         map[string]any{"type": "string", "format": "uuid", "description": "Post UUID"},
				"content":    map[string]any{"type": "string", "description": "New body (Markdown)"},
				"tags":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "New tag name list (replaces all existing tags)"},
				"private":    map[string]any{"type": "boolean", "description": "Set visibility"},
				"layout":     map[string]any{"type": "string", "enum": layoutEnum, "description": "Image layout style"},
				"echo_files": map[string]any{"type": "array", "items": echoFileSchema, "description": "New attached files (replaces all existing attachments)"},
				"extension":  extensionSchema,
			},
		},
		Annotations: destructiveHints(),
	}, a.updatePost, authModel.ScopeEchoWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "delete_post",
		Title:       "Delete Post",
		Description: "Permanently delete a post. Returns {id, message}. This action cannot be undone.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "Post UUID"},
			},
		},
		Annotations: destructiveHints(),
	}, a.deletePost, authModel.ScopeEchoWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_today_posts",
		Title:       "Get Today's Posts",
		Description: "Return all posts created today. The boundary of 'today' depends on the timezone parameter. Returns an array of post objects.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"timezone": map[string]any{"type": "string", "description": "IANA timezone name (e.g. Asia/Shanghai, America/New_York). Defaults to UTC if omitted."},
			},
		},
		Annotations: readOnlyHints(),
	}, a.getTodayPosts, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "like_post",
		Title:       "Like Post",
		Description: "Increment the like (favourite) count of a post by one. Returns {id, message}.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "Post UUID"},
			},
		},
		Annotations: createHints(),
	}, a.likePost, authModel.ScopeEchoWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "delete_tag",
		Title:       "Delete Tag",
		Description: "Delete a tag and remove its association from all posts. Returns {id, message}. This action cannot be undone.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "Tag UUID"},
			},
		},
		Annotations: destructiveHints(),
	}, a.deleteTag, authModel.ScopeEchoWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_hot_posts",
		Title:       "Get Hot Posts",
		Description: "Return the most popular posts, ranked by a weighted score of likes and comment count. Returns an array of post objects.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"limit": map[string]any{"type": "integer", "description": "Number of posts to return (1–100)", "default": 5},
			},
		},
		Annotations: readOnlyHints(),
	}, a.getHotPosts, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_random_post",
		Title:       "Get Random Post",
		Description: "Return a single randomly selected post. Returns the post object, or null if there are no posts. No parameters needed.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Annotations: readOnlyHints(),
	}, a.getRandomPost, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_on_this_day_posts",
		Title:       "Get On This Day Posts",
		Description: "Return posts from past years that share the same month and day as today ('on this day' / this day in history). The day boundary depends on the timezone parameter. Returns an array of post objects.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"timezone": map[string]any{"type": "string", "description": "IANA timezone name (e.g. Asia/Shanghai, America/New_York). Defaults to UTC if omitted."},
			},
		},
		Annotations: readOnlyHints(),
	}, a.getOnThisDayPosts, authModel.ScopeEchoRead)
}

func (a *Adapter) registerEchoResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://tags",
		Name:        "tags",
		Title:       "All Tags",
		Description: "JSON array of all tags. Each object has id, name, and the number of posts using this tag.",
		MimeType:    "application/json",
		Cache:       privateCache(liveTTLMs),
	}, a.resourceTags, authModel.ScopeEchoRead)

	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://posts/recent",
		Name:        "recent_posts",
		Title:       "Recent Posts",
		Description: "The 20 most recently created posts (newest first). Append ?limit=N (1–100) to the URI to change the count.",
		MimeType:    "application/json",
		Cache:       privateCache(liveTTLMs),
	}, a.resourceRecentPosts, authModel.ScopeEchoRead)

	reg.RegisterResourceTemplate(ResourceTemplateDefinition{
		URITemplate: "ech0://posts/{id}",
		Name:        "post",
		Title:       "Post by ID",
		Description: "Full post object (content, tags, timestamps, like count) for a given UUID. Replace {id} with the post UUID.",
		MimeType:    "application/json",
		Cache:       privateCache(liveTTLMs),
	}, a.resourcePostByID, authModel.ScopeEchoRead)
}

func (a *Adapter) searchPosts(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	query := stringArg(args, "query")
	page := intArg(args, "page", 1)
	pageSize := min(intArg(args, "page_size", 20), 100)
	sortBy := stringArg(args, "sort_by")
	sortOrder := stringArg(args, "sort_order")

	var tagIDs []string
	if raw, ok := args["tag_ids"]; ok {
		if arr, ok := raw.([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					tagIDs = append(tagIDs, s)
				}
			}
		}
	}

	result, err := a.echoSvc.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:      page,
		PageSize:  pageSize,
		Search:    query,
		TagIDs:    tagIDs,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, err
	}
	return jsonResult(result)
}

func (a *Adapter) getPost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	echo, err := a.echoSvc.GetEchoById(ctx, id)
	if err != nil {
		return nil, err
	}
	return jsonResult(echo)
}

func (a *Adapter) listTags(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
	tags, err := a.echoSvc.GetAllTags()
	if err != nil {
		return nil, err
	}
	return jsonResult(tags)
}

func (a *Adapter) createPost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	content := stringArg(args, "content")
	echoFiles := buildEchoFiles(args)
	extension := buildExtension(args)

	if strings.TrimSpace(content) == "" && len(echoFiles) == 0 && extension == nil {
		return textError("at least one of content, echo_files, or extension is required"), nil
	}

	echo := &echoModel.Echo{
		Content:   content,
		Private:   boolArg(args, "private"),
		Layout:    stringArg(args, "layout"),
		Tags:      buildTags(args),
		EchoFiles: echoFiles,
		Extension: extension,
	}
	if err := a.echoSvc.PostEcho(ctx, echo); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": echo.ID, "message": "post created successfully"})
}

func (a *Adapter) updatePost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}

	echo := &echoModel.Echo{ID: id}
	if content, ok := args["content"]; ok {
		if s, ok := content.(string); ok {
			echo.Content = s
		}
	}
	if p, ok := args["private"]; ok {
		if b, ok := p.(bool); ok {
			echo.Private = b
		}
	}
	echo.Layout = stringArg(args, "layout")
	echo.Tags = buildTags(args)
	echo.EchoFiles = buildEchoFiles(args)
	echo.Extension = buildExtension(args)

	if err := a.echoSvc.UpdateEcho(ctx, echo); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "post updated successfully"})
}

func (a *Adapter) deletePost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.echoSvc.DeleteEchoById(ctx, id); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "post deleted successfully"})
}

func (a *Adapter) getTodayPosts(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	timezone := stringArg(args, "timezone")
	posts, err := a.echoSvc.GetTodayEchos(ctx, timezone)
	if err != nil {
		return nil, err
	}
	return jsonResult(posts)
}

func (a *Adapter) likePost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.echoSvc.LikeEcho(ctx, id); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "post liked successfully"})
}

func (a *Adapter) deleteTag(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.echoSvc.DeleteTag(ctx, id); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "tag deleted successfully"})
}

func (a *Adapter) getHotPosts(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	limit := min(max(intArg(args, "limit", 5), 1), 100)
	posts, err := a.echoSvc.GetHotEchos(ctx, limit)
	if err != nil {
		return nil, err
	}
	return jsonResult(posts)
}

func (a *Adapter) getRandomPost(ctx context.Context, _ map[string]any) (*ToolCallResult, error) {
	post, err := a.echoSvc.GetRandomEcho(ctx)
	if err != nil {
		return nil, err
	}
	return jsonResult(post)
}

func (a *Adapter) getOnThisDayPosts(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	timezone := stringArg(args, "timezone")
	posts, err := a.echoSvc.GetOnThisDayEchos(ctx, timezone)
	if err != nil {
		return nil, err
	}
	return jsonResult(posts)
}

func (a *Adapter) resourceTags(_ context.Context, _ string) (*ResourceReadResult, error) {
	tags, err := a.echoSvc.GetAllTags()
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(tags)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://tags", MimeType: "application/json", Text: string(data)}},
	}, nil
}

func (a *Adapter) resourcePostByID(ctx context.Context, uri string) (*ResourceReadResult, error) {
	id := strings.TrimPrefix(uri, "ech0://posts/")
	if id == "" || id == uri {
		return nil, fmt.Errorf("invalid post URI: %s", uri)
	}
	echo, err := a.echoSvc.GetEchoById(ctx, id)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(echo)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: uri, MimeType: "application/json", Text: string(data)}},
	}, nil
}

func (a *Adapter) resourceRecentPosts(ctx context.Context, uri string) (*ResourceReadResult, error) {
	limit := 20
	if parts := strings.SplitN(uri, "?limit=", 2); len(parts) == 2 {
		if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	result, err := a.echoSvc.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     1,
		PageSize: limit,
	})
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(result)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://posts/recent", MimeType: "application/json", Text: string(data)}},
	}, nil
}

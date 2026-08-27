// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type EchoService struct {
	transactor     transaction.Transactor
	commonService  CommonService
	fileService    FileService
	echoRepository Repository
	bus            *busen.Bus
}

func NewEchoService(
	tx transaction.Transactor,
	commonService CommonService,
	fileService FileService,
	echoRepository Repository,
	busProvider func() *busen.Bus,
) *EchoService {
	return &EchoService{
		transactor:     tx,
		commonService:  commonService,
		fileService:    fileService,
		echoRepository: echoRepository,
		bus:            busProvider(),
	}
}

func (echoService *EchoService) PostEcho(ctx context.Context, newEcho *model.Echo) error {
	userid := viewer.MustFromContext(ctx).UserID()
	newEcho.UserID = userid

	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}

	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	layout := strings.TrimSpace(newEcho.Layout)
	if layout == "" || (layout != model.LayoutWaterfall &&
		layout != model.LayoutGrid &&
		layout != model.LayoutHorizontal &&
		layout != model.LayoutCarousel &&
		layout != model.LayoutStack &&
		layout != model.LayoutNone) {
		newEcho.Layout = model.LayoutWaterfall
	}

	normalizedExt, err := normalizeEchoExtension(newEcho.Extension)
	if err != nil {
		return err
	}
	newEcho.Extension = normalizedExt

	newEcho.Username = user.Username

	if isEchoEmpty(newEcho) {
		return errors.New(commonModel.ECHO_CAN_NOT_BE_EMPTY)
	}

	if err := echoService.validateSingleFileCategory(ctx, newEcho); err != nil {
		return err
	}

	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		if err := echoService.ProcessEchoTags(txCtx, newEcho); err != nil {
			return err
		}
		return echoService.echoRepository.CreateEcho(txCtx, newEcho)
	}); err != nil {
		return err
	}

	echoService.echoRepository.InvalidateEchoCaches()

	savedEcho, fetchErr := echoService.echoRepository.GetEchosById(ctx, newEcho.ID)
	if fetchErr != nil {
		return fetchErr
	}
	if savedEcho != nil {
		eventbus.Notify(context.Background(), echoService.bus, event.EchoCreated{Echo: *savedEcho, User: user})
	}
	if err := echoService.fileService.ConfirmTempFiles(ctx, collectEchoFileIDs(newEcho)); err != nil {
		logUtil.GetLogger().Warn("confirm temp files after post echo failed", logUtil.Err(err))
	}

	return nil
}

func (echoService *EchoService) GetEchosByPage(
	ctx context.Context,
	pageQueryDto commonModel.PageQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	return echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     pageQueryDto.Page,
		PageSize: pageQueryDto.PageSize,
		Search:   pageQueryDto.Search,
	})
}

func (echoService *EchoService) DeleteEchoById(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	type deletableFileRef struct {
		key         string
		storageType string
	}
	var deletableFiles []deletableFileRef
	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		echo, err := echoService.echoRepository.GetEchosById(txCtx, id)
		if err != nil {
			return err
		}
		if echo == nil {
			return errors.New(commonModel.ECHO_NOT_FOUND)
		}

		for _, ef := range echo.EchoFiles {
			if ef.File.Key != "" && storage.NormalizeStorageType(ef.File.StorageType) != storage.StorageTypeExternal {
				deletableFiles = append(deletableFiles, deletableFileRef{
					key:         ef.File.Key,
					storageType: ef.File.StorageType,
				})
			}
			if ef.File.ID != "" {
				if err := echoService.fileService.DeleteFileRecord(txCtx, ef.File.ID); err != nil {
					return err
				}
			}
		}

		return echoService.echoRepository.DeleteEchoById(txCtx, id)
	}); err != nil {
		return err
	}

	echoService.echoRepository.InvalidateEchoCaches(id)

	eventbus.Notify(context.Background(), echoService.bus, event.EchoDeleted{Echo: model.Echo{ID: id}, User: user})

	for _, file := range deletableFiles {
		if err := echoService.fileService.DeleteStoredFile(file.storageType, file.key); err != nil {
			logUtil.GetLogger().Warn(
				"delete stored file after echo delete failed",
				slog.String("echo_id", id),
				slog.String("file_key", file.key),
				slog.String("storage_type", file.storageType),
				logUtil.Err(err),
			)
		}
	}

	return nil
}

func (echoService *EchoService) GetTodayEchos(ctx context.Context, timezone string) ([]model.Echo, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userid != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return nil, err
		}
		showPrivate = user.IsAdmin
	}

	todayEchos := echoService.echoRepository.GetTodayEchos(showPrivate, timezone)
	return todayEchos, nil
}

func (echoService *EchoService) GetHotEchos(ctx context.Context, limit int) ([]model.Echo, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userid != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return nil, err
		}
		showPrivate = user.IsAdmin
	}
	return echoService.echoRepository.GetHotEchos(limit, showPrivate)
}

func (echoService *EchoService) GetRandomEcho(ctx context.Context) (*model.Echo, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userid != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return nil, err
		}
		showPrivate = user.IsAdmin
	}
	return echoService.echoRepository.GetRandomEcho(showPrivate)
}

func (echoService *EchoService) GetOnThisDayEchos(ctx context.Context, timezone string) ([]model.Echo, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userid != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return nil, err
		}
		showPrivate = user.IsAdmin
	}
	return echoService.echoRepository.GetOnThisDayEchos(showPrivate, timezone), nil
}

func (echoService *EchoService) UpdateEcho(ctx context.Context, echo *model.Echo) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	layout := strings.TrimSpace(echo.Layout)
	if layout == "" || (layout != model.LayoutWaterfall &&
		layout != model.LayoutGrid &&
		layout != model.LayoutHorizontal &&
		layout != model.LayoutCarousel &&
		layout != model.LayoutStack &&
		layout != model.LayoutNone) {
		echo.Layout = model.LayoutWaterfall
	}

	normalizedExt, err := normalizeEchoExtension(echo.Extension)
	if err != nil {
		return err
	}
	echo.Extension = normalizedExt

	for i := range echo.EchoFiles {
		echo.EchoFiles[i].EchoID = echo.ID
	}

	if isEchoEmpty(echo) {
		return errors.New(commonModel.ECHO_CAN_NOT_BE_EMPTY)
	}

	if err := echoService.validateSingleFileCategory(ctx, echo); err != nil {
		return err
	}

	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		if err := echoService.ProcessEchoTags(txCtx, echo); err != nil {
			return err
		}
		return echoService.echoRepository.UpdateEcho(txCtx, echo)
	}); err != nil {
		return err
	}

	echoService.echoRepository.InvalidateEchoCaches(echo.ID)

	eventbus.Notify(context.Background(), echoService.bus, event.EchoUpdated{Echo: *echo, User: user})
	if err := echoService.fileService.ConfirmTempFiles(ctx, collectEchoFileIDs(echo)); err != nil {
		logUtil.GetLogger().Warn("confirm temp files after update echo failed", logUtil.Err(err))
	}

	return nil
}

func (echoService *EchoService) LikeEcho(ctx context.Context, id string) error {
	echo, err := echoService.echoRepository.GetEchosById(ctx, id)
	if err != nil {
		return err
	}
	if echo == nil {
		return errors.New(commonModel.ECHO_NOT_FOUND)
	}
	if echo.Private {
		userID := viewer.MustFromContext(ctx).UserID()
		if userID == "" {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userID)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}
	}

	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		return echoService.echoRepository.LikeEcho(txCtx, id)
	}); err != nil {
		return err
	}
	echoService.echoRepository.InvalidateEchoCaches(id)
	return nil
}

func (echoService *EchoService) GetEchoById(ctx context.Context, id string) (*model.Echo, error) {
	userId := viewer.MustFromContext(ctx).UserID()
	echo, err := echoService.echoRepository.GetEchosById(ctx, id)
	if err != nil {
		return nil, err
	}
	if echo == nil {
		return nil, errors.New(commonModel.ECHO_NOT_FOUND)
	}

	if userId == "" {
		if echo.Private {
			return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
		}
	} else {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userId)
		if err != nil {
			return nil, err
		}
		if echo.Private && !user.IsAdmin {
			return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
		}
	}

	return echo, nil
}

func (echoService *EchoService) GetAllTags() ([]model.Tag, error) {
	return echoService.echoRepository.GetAllTags()
}

func (echoService *EchoService) CreateTag(ctx context.Context, name string) (*model.Tag, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	cleaned := strings.TrimSpace(strings.TrimPrefix(name, "#"))
	if cleaned == "" {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	if !isSafeTagName(cleaned) {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}

	existing, err := echoService.echoRepository.GetTagsByNames(ctx, []string{cleaned})
	if err != nil {
		return nil, err
	}
	for _, t := range existing {
		if t != nil && t.Name == cleaned {
			return t, nil
		}
	}

	newTag := model.Tag{Name: cleaned, UsageCount: 0}
	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		return echoService.echoRepository.CreateTag(txCtx, &newTag)
	}); err != nil {
		return nil, err
	}
	return &newTag, nil
}

func (echoService *EchoService) DeleteTag(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		return echoService.echoRepository.DeleteTagById(txCtx, id)
	})
}

func (echoService *EchoService) ProcessEchoTags(ctx context.Context, echo *model.Echo) error {
	var processedTags []model.Tag

	var names []string
	for _, tag := range echo.Tags {
		name := strings.TrimSpace(strings.TrimPrefix(tag.Name, "#"))
		if name == "" {
			continue
		}
		if !isSafeTagName(name) {
			return errors.New(commonModel.INVALID_PARAMS)
		}
		names = append(names, name)
	}

	existingTags, err := echoService.echoRepository.GetTagsByNames(ctx, names)
	if err != nil {
		return err
	}

	existingMap := make(map[string]*model.Tag)
	for _, t := range existingTags {
		existingMap[t.Name] = t
	}

	for _, name := range names {
		if existing, ok := existingMap[name]; ok {
			if err := echoService.echoRepository.IncrementTagUsageCount(ctx, existing.ID); err != nil {
				return err
			}
			processedTags = append(processedTags, *existing)
		} else {
			newTag := model.Tag{Name: name, UsageCount: 1}
			if err := echoService.echoRepository.CreateTag(ctx, &newTag); err != nil {
				return err
			}
			processedTags = append(processedTags, newTag)
		}
	}

	echo.Tags = processedTags
	return nil
}

func (echoService *EchoService) GetEchosByTagId(
	ctx context.Context,
	tagId string,
	pageQueryDto commonModel.PageQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	return echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     pageQueryDto.Page,
		PageSize: pageQueryDto.PageSize,
		Search:   pageQueryDto.Search,
		TagIDs:   []string{tagId},
	})
}

func (echoService *EchoService) QueryEchos(
	ctx context.Context,
	queryDto commonModel.EchoQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	if queryDto.Page < 1 {
		queryDto.Page = 1
	}
	if queryDto.PageSize < 1 {
		queryDto.PageSize = 10
	}
	if queryDto.PageSize > 100 {
		queryDto.PageSize = 100
	}
	queryDto.Search = strings.TrimSpace(queryDto.Search)

	if queryDto.SortBy == "" {
		queryDto.SortBy = "created_at"
	}
	if queryDto.SortOrder == "" {
		queryDto.SortOrder = "desc"
	}

	userId := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userId != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userId)
		if err != nil {
			return commonModel.PageQueryResult[[]model.Echo]{}, err
		}
		showPrivate = user.IsAdmin
	}

	echos, total, err := echoService.echoRepository.QueryEchos(queryDto, showPrivate)
	if err != nil {
		return commonModel.PageQueryResult[[]model.Echo]{}, err
	}

	return commonModel.PageQueryResult[[]model.Echo]{
		Items: echos,
		Total: total,
	}, nil
}

func isSafeTagName(name string) bool {
	return !strings.ContainsAny(name, "<>\"'&")
}

func normalizeEchoExtension(ext *model.EchoExtension) (*model.EchoExtension, error) {
	if ext == nil {
		return nil, nil
	}

	extType := strings.TrimSpace(ext.Type)
	if extType == "" {
		return nil, nil
	}
	ext.Type = extType
	if ext.Payload == nil {
		return nil, fmt.Errorf("extension payload is required")
	}

	switch ext.Type {
	case model.Extension_MUSIC:
		url := strings.TrimSpace(getPayloadString(ext.Payload, "url"))
		if url == "" {
			return nil, fmt.Errorf("extension payload.url is required for MUSIC")
		}
		ext.Payload = map[string]any{"url": urlUtil.TrimURL(url)}
	case model.Extension_VIDEO:
		videoID := strings.TrimSpace(getPayloadString(ext.Payload, "videoId"))
		if videoID == "" {
			return nil, fmt.Errorf("extension payload.videoId is required for VIDEO")
		}
		ext.Payload = map[string]any{"videoId": videoID}
	case model.Extension_GITHUBPROJ:
		repoURL := strings.TrimSpace(getPayloadString(ext.Payload, "repoUrl"))
		if repoURL == "" {
			return nil, fmt.Errorf("extension payload.repoUrl is required for GITHUBPROJ")
		}
		ext.Payload = map[string]any{"repoUrl": urlUtil.TrimURL(repoURL)}
	case model.Extension_WEBSITE:
		title := strings.TrimSpace(getPayloadString(ext.Payload, "title"))
		site := strings.TrimSpace(getPayloadString(ext.Payload, "site"))
		if title == "" || site == "" {
			return nil, fmt.Errorf("extension payload.title and payload.site are required for WEBSITE")
		}
		ext.Payload = map[string]any{
			"title": title,
			"site":  urlUtil.TrimURL(site),
		}
	case model.Extension_LOCATION:
		lat, okLat := getPayloadFloat(ext.Payload, "latitude")
		lng, okLng := getPayloadFloat(ext.Payload, "longitude")
		if !okLat || !okLng {
			return nil, fmt.Errorf("extension payload.latitude and payload.longitude are required for LOCATION")
		}
		if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			return nil, fmt.Errorf("extension payload.latitude/longitude out of range for LOCATION")
		}
		placeholder := strings.TrimSpace(getPayloadString(ext.Payload, "placeholder"))
		if placeholder == "" {
			return nil, fmt.Errorf("extension payload.placeholder is required for LOCATION")
		}
		ext.Payload = map[string]any{
			"latitude":    lat,
			"longitude":   lng,
			"placeholder": placeholder,
		}
	case model.Extension_TWEET:
		url := strings.TrimSpace(getPayloadString(ext.Payload, "url"))
		username := strings.TrimSpace(getPayloadString(ext.Payload, "username"))
		statusID := strings.TrimSpace(getPayloadString(ext.Payload, "statusId"))
		if url == "" || username == "" || statusID == "" {
			return nil, fmt.Errorf("extension payload.url, payload.username and payload.statusId are required for TWEET")
		}
		ext.Payload = map[string]any{
			"url":      urlUtil.TrimURL(url),
			"username": username,
			"statusId": statusID,
		}
	default:
		return nil, fmt.Errorf("unsupported extension type: %s", ext.Type)
	}

	return ext, nil
}

func getPayloadString(payload map[string]any, key string) string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return value
}

func getPayloadFloat(payload map[string]any, key string) (float64, bool) {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func isEchoEmpty(echo *model.Echo) bool {
	if echo == nil {
		return true
	}
	content := strings.TrimSpace(echo.Content)
	return content == "" && len(echo.EchoFiles) == 0 && echo.Extension == nil
}

func (echoService *EchoService) validateSingleFileCategory(ctx context.Context, echo *model.Echo) error {
	ids := collectEchoFileIDs(echo)
	if len(ids) == 0 {
		return nil
	}

	files, err := echoService.fileService.GetFilesByIDs(ctx, ids)
	if err != nil {
		return err
	}

	counts := make(map[storage.Category]int, len(files))
	for _, f := range files {
		counts[storage.NormalizeCategory(f.Category)]++
	}

	if len(counts) > 1 {
		return errors.New(commonModel.ECHO_MIXED_FILE_CATEGORIES)
	}
	for cat, count := range counts {
		if (cat == storage.CategoryAudio || cat == storage.CategoryVideo) && count > 1 {
			return errors.New(commonModel.ECHO_MIXED_FILE_CATEGORIES)
		}
	}
	return nil
}

func collectEchoFileIDs(echo *model.Echo) []string {
	if echo == nil || len(echo.EchoFiles) == 0 {
		return nil
	}
	ids := make([]string, 0, len(echo.EchoFiles))
	for _, ef := range echo.EchoFiles {
		if strings.TrimSpace(ef.FileID) != "" {
			ids = append(ids, ef.FileID)
			continue
		}
		if strings.TrimSpace(ef.File.ID) != "" {
			ids = append(ids, ef.File.ID)
		}
	}
	return ids
}

package service

import (
	"context"
	"errors"
	"strings"

	"github.com/lin-snow/ech0/internal/event"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	commonRepository "github.com/lin-snow/ech0/internal/repository/common"
	repository "github.com/lin-snow/ech0/internal/repository/echo"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

type EchoService struct {
	transactor       transaction.Transactor
	commonService    *commonService.CommonService
	echoRepository   repository.EchoRepositoryInterface
	commonRepository commonRepository.CommonRepositoryInterface
	kvRepository     keyvalueRepository.KeyValueRepositoryInterface
	eventBus         event.IEventBus
}

func NewEchoService(
	tx transaction.Transactor,
	commonService *commonService.CommonService,
	echoRepository repository.EchoRepositoryInterface,
	commonRepository commonRepository.CommonRepositoryInterface,
	kvRepository keyvalueRepository.KeyValueRepositoryInterface,
	eventBusProvider func() event.IEventBus,
) *EchoService {
	return &EchoService{
		transactor:       tx,
		commonService:    commonService,
		echoRepository:   echoRepository,
		commonRepository: commonRepository,
		kvRepository:     kvRepository,
		eventBus:         eventBusProvider(),
	}
}

func (echoService *EchoService) PostEcho(userid uint, newEcho *model.Echo) error {
	newEcho.UserID = userid

	user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userid)
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
		layout != model.LayoutCarousel) {
		newEcho.Layout = model.LayoutWaterfall
	}

	if newEcho.Extension != "" && newEcho.ExtensionType != "" {
		switch newEcho.ExtensionType {
		case model.Extension_GITHUBPROJ:
			newEcho.Extension = httpUtil.TrimURL(newEcho.Extension)
		}
	} else {
		newEcho.Extension = ""
		newEcho.ExtensionType = ""
	}

	newEcho.Username = user.Username

	if newEcho.Content == "" && len(newEcho.EchoFiles) == 0 &&
		(newEcho.Extension == "" || newEcho.ExtensionType == "") {
		return errors.New(commonModel.ECHO_CAN_NOT_BE_EMPTY)
	}

	if err := echoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		if err := echoService.ProcessEchoTags(ctx, newEcho); err != nil {
			return err
		}
		return echoService.echoRepository.CreateEcho(ctx, newEcho)
	}); err != nil {
		return err
	}

	savedEcho, fetchErr := echoService.echoRepository.GetEchosById(context.Background(), newEcho.ID)
	if fetchErr != nil {
		return fetchErr
	}
	if savedEcho != nil {
		if pubErr := echoService.eventBus.Publish(
			context.Background(),
			event.NewEvent(
				event.EventTypeEchoCreated,
				event.EventPayload{
					event.EventPayloadEcho: *savedEcho,
					event.EventPayloadUser: user,
				},
			),
		); pubErr != nil {
			logUtil.GetLogger().Error(pubErr.Error())
		}
	}

	return nil
}

func (echoService *EchoService) GetEchosByPage(
	userid uint,
	pageQueryDto commonModel.PageQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	if pageQueryDto.Page < 1 {
		pageQueryDto.Page = 1
	}
	if pageQueryDto.PageSize < 1 || pageQueryDto.PageSize > 100 {
		pageQueryDto.PageSize = 10
	}

	showPrivate := false
	if userid != authModel.NO_USER_LOGINED {
		user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userid)
		if err != nil {
			return commonModel.PageQueryResult[[]model.Echo]{}, err
		}
		showPrivate = user.IsAdmin
	}

	echosByPage, total := echoService.echoRepository.GetEchosByPage(
		pageQueryDto.Page,
		pageQueryDto.PageSize,
		pageQueryDto.Search,
		showPrivate,
	)
	return commonModel.PageQueryResult[[]model.Echo]{
		Items: echosByPage,
		Total: total,
	}, nil
}

func (echoService *EchoService) DeleteEchoById(userid, id uint) error {
	user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	var fileKeys []string
	if err := echoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		echo, err := echoService.echoRepository.GetEchosById(ctx, id)
		if err != nil {
			return err
		}
		if echo == nil {
			return errors.New(commonModel.ECHO_NOT_FOUND)
		}

		for _, ef := range echo.EchoFiles {
			if ef.File.Key != "" {
				fileKeys = append(fileKeys, ef.File.Key)
				if err := echoService.commonService.DeleteFileRecord(ctx, ef.File.Key); err != nil {
					return err
				}
			}
		}

		return echoService.echoRepository.DeleteEchoById(ctx, id)
	}); err != nil {
		return err
	}

	if pubErr := echoService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeEchoDeleted,
			event.EventPayload{
				event.EventPayloadEcho: model.Echo{ID: id},
				event.EventPayloadUser: user,
			},
		),
	); pubErr != nil {
		logUtil.GetLogger().Error(pubErr.Error())
	}

	for _, fileKey := range fileKeys {
		_ = echoService.commonService.DeleteStoredFile(fileKey)
	}

	return nil
}

func (echoService *EchoService) GetTodayEchos(userid uint, timezone string) ([]model.Echo, error) {
	showPrivate := false
	if userid != authModel.NO_USER_LOGINED {
		user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userid)
		if err != nil {
			return nil, err
		}
		showPrivate = user.IsAdmin
	}

	todayEchos := echoService.echoRepository.GetTodayEchos(showPrivate, timezone)
	return todayEchos, nil
}

func (echoService *EchoService) UpdateEcho(userid uint, echo *model.Echo) error {
	user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userid)
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
		layout != model.LayoutCarousel) {
		echo.Layout = model.LayoutWaterfall
	}

	if echo.Extension != "" && echo.ExtensionType != "" {
		switch echo.ExtensionType {
		case model.Extension_GITHUBPROJ:
			echo.Extension = httpUtil.TrimURL(echo.Extension)
		}
	} else {
		echo.Extension = ""
		echo.ExtensionType = ""
	}

	for i := range echo.EchoFiles {
		echo.EchoFiles[i].EchoID = echo.ID
	}

	if echo.Content == "" && len(echo.EchoFiles) == 0 &&
		(echo.Extension == "" || echo.ExtensionType == "") {
		return errors.New(commonModel.ECHO_CAN_NOT_BE_EMPTY)
	}

	if err := echoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		if err := echoService.ProcessEchoTags(ctx, echo); err != nil {
			return err
		}
		return echoService.echoRepository.UpdateEcho(ctx, echo)
	}); err != nil {
		return err
	}

	if pubErr := echoService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeEchoUpdated,
			event.EventPayload{
				event.EventPayloadEcho: *echo,
				event.EventPayloadUser: user,
			},
		),
	); pubErr != nil {
		logUtil.GetLogger().Error(pubErr.Error())
	}

	return nil
}

func (echoService *EchoService) LikeEcho(id uint) error {
	return echoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		return echoService.echoRepository.LikeEcho(ctx, id)
	})
}

func (echoService *EchoService) GetEchoById(userId, id uint) (*model.Echo, error) {
	echo, err := echoService.echoRepository.GetEchosById(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if echo == nil {
		return nil, errors.New(commonModel.ECHO_NOT_FOUND)
	}

	if userId == authModel.NO_USER_LOGINED {
		if echo.Private {
			return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
		}
	} else {
		user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userId)
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

func (echoService *EchoService) DeleteTag(userid, id uint) error {
	user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return echoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		return echoService.echoRepository.DeleteTagById(ctx, id)
	})
}

func (echoService *EchoService) ProcessEchoTags(ctx context.Context, echo *model.Echo) error {
	var processedTags []model.Tag

	var names []string
	for _, tag := range echo.Tags {
		name := strings.TrimSpace(strings.TrimPrefix(tag.Name, "#"))
		if name != "" {
			names = append(names, name)
		}
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
	userId, tagId uint,
	pageQueryDto commonModel.PageQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	if pageQueryDto.Page < 1 {
		pageQueryDto.Page = 1
	}
	if pageQueryDto.PageSize < 1 || pageQueryDto.PageSize > 100 {
		pageQueryDto.PageSize = 10
	}
	pageQueryDto.Search = strings.TrimSpace(pageQueryDto.Search)

	showPrivate := false
	if userId != authModel.NO_USER_LOGINED {
		user, err := echoService.commonService.CommonGetUserByUserId(context.Background(), userId)
		if err != nil {
			return commonModel.PageQueryResult[[]model.Echo]{}, err
		}
		showPrivate = user.IsAdmin
	}

	echos, total, err := echoService.echoRepository.GetEchosByTagId(
		tagId,
		pageQueryDto.Page,
		pageQueryDto.PageSize,
		pageQueryDto.Search,
		showPrivate,
	)
	if err != nil {
		return commonModel.PageQueryResult[[]model.Echo]{}, err
	}

	return commonModel.PageQueryResult[[]model.Echo]{
		Items: echos,
		Total: total,
	}, nil
}

package service

import (
	"errors"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	repository "github.com/lin-snow/ech0/internal/repository/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
)

type EchoService struct {
	commonService  commonService.CommonServiceInterface
	echoRepository repository.EchoRepositoryInterface
}

func NewEchoService(commonService commonService.CommonServiceInterface, echoRepository repository.EchoRepositoryInterface) EchoServiceInterface {
	return &EchoService{
		commonService:  commonService,
		echoRepository: echoRepository,
	}
}

func (echoService *EchoService) PostEcho(userid uint, newEcho *model.Echo) error {
	newEcho.UserID = userid

	user, err := echoService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}

	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查Extension内容
	if newEcho.Extension != "" && newEcho.ExtensionType != "" {
		if newEcho.ExtensionType == model.Extension_MUSIC {

		} else if newEcho.ExtensionType == model.Extension_VIDEO {

		} else if newEcho.ExtensionType == model.Extension_GITHUBPROJ {
			// 处理GitHub项目的链接
			newEcho.Extension = httpUtil.TrimURL(newEcho.Extension)
		} else if newEcho.ExtensionType == model.Extension_WEBSITE {

		}
	} else {
		newEcho.Extension = ""
		newEcho.ExtensionType = ""
	}

	newEcho.Username = user.Username

	for i := range newEcho.Images {
		if newEcho.Images[i].ImageURL == "" {
			newEcho.Images[i].ImageSource = ""
		}
	}

	if newEcho.Content == "" && len(newEcho.Images) == 0 && (newEcho.Extension == "" || newEcho.ExtensionType == "") {
		return errors.New(commonModel.ECHO_CAN_NOT_BE_EMPTY)
	}

	return echoService.echoRepository.CreateEcho(newEcho)
}

func (echoService *EchoService) GetEchosByPage(userid uint, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error) {
	// 参数校验
	if pageQueryDto.Page < 1 {
		pageQueryDto.Page = 1
	}
	if pageQueryDto.PageSize < 1 || pageQueryDto.PageSize > 100 {
		pageQueryDto.PageSize = 10
	}

	//管理员登陆则支持查看隐私数据，否则不允许
	showPrivate := false
	if userid == authModel.NO_USER_LOGINED {
		showPrivate = false
	} else {
		user, err := echoService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return commonModel.PageQueryResult[[]model.Echo]{}, err
		}
		if !user.IsAdmin {
			showPrivate = false
		}
		showPrivate = true
	}

	echosByPage, total := echoService.echoRepository.GetEchosByPage(pageQueryDto.Page, pageQueryDto.PageSize, pageQueryDto.Search, showPrivate)
	result := commonModel.PageQueryResult[[]model.Echo]{
		Items: echosByPage,
		Total: total,
	}

	return result, nil
}

func (echoService *EchoService) DeleteEchoById(userid, id uint) error {
	user, err := echoService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查该留言是否存在图片
	echo, err := echoService.echoRepository.GetEchosById(id)
	if err != nil {
		return err
	}
	if echo == nil {
		return errors.New(commonModel.ECHO_NOT_FOUND)
	}

	// 删除Echo中的图片
	if len(echo.Images) > 0 {
		for _, img := range echo.Images {
			if err := echoService.commonService.DirectDeleteImage(img.ImageURL, img.ImageSource); err != nil {
				return err
			}
		}
	}

	return echoService.echoRepository.DeleteEchoById(id)
}

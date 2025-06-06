package service

import (
	"errors"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	repository "github.com/lin-snow/ech0/internal/repository/echo"
	userService "github.com/lin-snow/ech0/internal/service/user"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
)

type EchoService struct {
	echoRepository repository.EchoRepositoryInterface
	userService    userService.UserServiceInterface
}

func NewEchoService(echoRepository repository.EchoRepositoryInterface, userService userService.UserServiceInterface) EchoServiceInterface {
	return &EchoService{
		echoRepository: echoRepository,
		userService:    userService,
	}
}

func (echoService *EchoService) PostEcho(userid uint, newEcho *model.Echo) error {
	newEcho.UserID = userid

	user, err := echoService.userService.GetUserByID((int)(userid))
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

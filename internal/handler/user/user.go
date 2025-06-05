package handler

import (
	"github.com/gin-gonic/gin"
	model "github.com/lin-snow/ech0/internal/model/user"
	service "github.com/lin-snow/ech0/internal/service/user"
	"net/http"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Login 用户登陆
func (userHandler *UserHandler) Login(ctx *gin.Context) {
	// 从请求体获取用户名和密码
	var user model.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusOK, dto.Fail[string](models.InvalidRequestBodyMessage))
		return
	}

}

// Register 用户注册
func (userHandler *UserHandler) Register(ctx *gin.Context) {

}

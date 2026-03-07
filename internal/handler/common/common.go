package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/common"
	storageDomain "github.com/lin-snow/ech0/internal/storage"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
)

type CommonHandler struct {
	commonService *service.CommonService
}

func NewCommonHandler(commonService *service.CommonService) *CommonHandler {
	return &CommonHandler{
		commonService: commonService,
	}
}

func (commonHandler *CommonHandler) UploadFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		file, err := ctx.FormFile("file")
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		category := storageDomain.NormalizeCategory(ctx.PostForm("category"))
		userId := ctx.MustGet("userid").(uint)

		fileDto, err := commonHandler.commonService.UploadFile(userId, file, category)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: fileDto,
			Msg:  commonModel.UPLOAD_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) DeleteFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userId := ctx.MustGet("userid").(uint)

		var dto commonModel.FileDeleteDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := commonHandler.commonService.DeleteFile(userId, dto); err != nil {
			ctx.JSON(
				http.StatusOK,
				commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
					Msg: "",
					Err: err,
				})),
			)
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) GetStatus() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		_, err := commonHandler.commonService.GetSysAdmin()
		if err != nil {
			return res.Response{
				Code: commonModel.InitInstallCode,
				Msg:  commonModel.SIGNUP_FIRST,
			}
		}

		status, err := commonHandler.commonService.GetStatus()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: status,
			Msg:  commonModel.GET_STATUS_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) GetHeatMap() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		timezone := timezoneUtil.NormalizeTimezone(ctx.GetHeader(timezoneUtil.DefaultTimezoneHeader))
		heatMap, err := commonHandler.commonService.GetHeatMap(timezone)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: heatMap,
			Msg:  commonModel.GET_HEATMAP_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) GetRss(ctx *gin.Context) {
	atom, err := commonHandler.commonService.GenerateRSS(ctx)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
				Msg: "",
				Err: err,
			})),
		)
		return
	}

	ctx.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(atom))
}

func (commonHandler *CommonHandler) UploadAudio() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userId := ctx.MustGet("userid").(uint)

		file, err := ctx.FormFile("file")
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		audioUrl, err := commonHandler.commonService.UploadMusic(userId, file)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: audioUrl,
			Msg:  commonModel.UPLOAD_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) DeleteAudio() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userId := ctx.MustGet("userid").(uint)

		if err := commonHandler.commonService.DeleteMusic(userId); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) GetPlayMusic() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		musicUrl := commonHandler.commonService.GetPlayMusicUrl()

		return res.Response{
			Data: musicUrl,
			Msg:  commonModel.GET_MUSIC_URL_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) PlayMusic(ctx *gin.Context) {
	commonHandler.commonService.PlayMusic(ctx)
}

func (commonHandler *CommonHandler) HelloEch0() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		hello := struct {
			Hello   string `json:"hello"`
			Version string `json:"version"`
			Github  string `json:"github"`
		}{
			Hello:   "Hello, Ech0! 👋",
			Version: commonModel.Version,
			Github:  "https://github.com/lin-snow/Ech0",
		}

		return res.Response{
			Msg:  commonModel.GET_HELLO_SUCCESS,
			Data: hello,
		}
	})
}

func (commonHandler *CommonHandler) Healthz() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		return res.Response{
			Msg: commonModel.GET_HEALTHZ_SUCCESS,
			Data: struct {
				Status string `json:"status"`
			}{
				Status: "ok",
			},
		}
	})
}

func (commonHandler *CommonHandler) GetFilePresignURL() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userId := ctx.MustGet("userid").(uint)
		var s3Dto commonModel.GetPresignURLDto
		if err := ctx.ShouldBindJSON(&s3Dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		presignDto, err := commonHandler.commonService.GetFilePresignURL(userId, &s3Dto, "PUT")
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: presignDto,
			Msg:  commonModel.GET_S3_PRESIGN_URL_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) GetWebsiteTitle() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto commonModel.GetWebsiteTitleDto
		if err := ctx.ShouldBindQuery(&dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_QUERY_PARAMS,
				Err: err,
			}
		}
		title, err := commonHandler.commonService.GetWebsiteTitle(dto.WebSiteURL)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}
		return res.Response{
			Data: title,
			Msg:  commonModel.GET_WEBSITE_TITLE_SUCCESS,
		}
	})
}

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/event"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	repository "github.com/lin-snow/ech0/internal/repository/common"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	storageDomain "github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	imgUtil "github.com/lin-snow/ech0/internal/util/img"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	mdUtil "github.com/lin-snow/ech0/internal/util/md"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

const globalMusicFileIDKey = "global_music_file_id"

type CommonService struct {
	transactor         transaction.Transactor
	commonRepository   repository.CommonRepositoryInterface
	fs                 virefs.FS
	resolveURL         storageDomain.URLResolver
	echoRepository     echoRepository.EchoRepositoryInterface
	keyvalueRepository keyvalueRepository.KeyValueRepositoryInterface
	fileRepository     fileRepository.FileRepositoryInterface
	eventBus           event.IEventBus
	keyGen             storageDomain.KeyGenerator
}

func NewCommonService(
	tx transaction.Transactor,
	commonRepository repository.CommonRepositoryInterface,
	echoRepo echoRepository.EchoRepositoryInterface,
	kvRepo keyvalueRepository.KeyValueRepositoryInterface,
	fileRepo fileRepository.FileRepositoryInterface,
	fs virefs.FS,
	resolveURL storageDomain.URLResolver,
	eventBusProvider func() event.IEventBus,
) *CommonService {
	return &CommonService{
		transactor:         tx,
		commonRepository:   commonRepository,
		echoRepository:     echoRepo,
		keyvalueRepository: kvRepo,
		fileRepository:     fileRepo,
		fs:                 fs,
		resolveURL:         resolveURL,
		eventBus:           eventBusProvider(),
		keyGen:             storageDomain.NewRandomKeyGenerator(),
	}
}

func (s *CommonService) CommonGetUserByUserId(ctx context.Context, userId uint) (userModel.User, error) {
	return s.commonRepository.GetUserByUserId(ctx, userId)
}

func (s *CommonService) UploadFile(
	userId uint,
	file *multipart.FileHeader,
	category storageDomain.Category,
) (commonModel.FileDto, error) {
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userId)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	contentType := file.Header.Get("Content-Type")
	if !isAllowedType(contentType, config.Config().Upload.AllowedTypes) {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	limit := int64(config.Config().Upload.ImageMaxSize)
	if category == storageDomain.CategoryAudio {
		limit = int64(config.Config().Upload.AudioMaxSize)
	}
	if file.Size > limit {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_SIZE_EXCEED_LIMIT)
	}

	gen := s.keyGenForCategory(category, file.Filename)
	key, err := gen.GenerateKey(category, user.ID, file.Filename)
	if err != nil {
		return commonModel.FileDto{}, err
	}

	reader, err := file.Open()
	if err != nil {
		return commonModel.FileDto{}, err
	}
	defer func() { _ = reader.Close() }()

	var opts []virefs.PutOption
	if contentType != "" {
		opts = append(opts, virefs.WithContentType(contentType))
	}
	if err := s.fs.Put(context.Background(), key, reader, opts...); err != nil {
		return commonModel.FileDto{}, err
	}

	width, height := 0, 0
	if category.IsImageLike() {
		width, height, err = imgUtil.GetImageSizeFromFile(file)
		if err != nil {
			return commonModel.FileDto{}, err
		}
	}

	url := s.resolveURL(key)

	fileRecord := &fileModel.File{
		Key:         key,
		URL:         url,
		Name:        file.Filename,
		ContentType: contentType,
		Size:        file.Size,
		Category:    string(category),
		Width:       width,
		Height:      height,
		UserID:      user.ID,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		return commonModel.FileDto{}, err
	}

	uploadType := commonModel.ImageType
	if category == storageDomain.CategoryAudio {
		uploadType = commonModel.AudioType
	}

	user.Password = ""
	if err := s.eventBus.Publish(context.Background(), event.NewEvent(
		event.EventTypeResourceUploaded,
		event.EventPayload{
			event.EventPayloadUser: user,
			event.EventPayloadFile: file.Filename,
			event.EventPayloadURL:  url,
			event.EventPayloadSize: file.Size,
			event.EventPayloadType: uploadType,
		},
	)); err != nil {
		logUtil.GetLogger().Error("Failed to publish resource uploaded event", zap.String("error", err.Error()))
	}

	return commonModel.FileDto{
		ID:          fileRecord.ID,
		Key:         key,
		URL:         url,
		ContentType: contentType,
		Category:    string(category),
		Size:        file.Size,
		Width:       width,
		Height:      height,
	}, nil
}

func (s *CommonService) DeleteFile(userid uint, dto commonModel.FileDeleteDto) error {
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if dto.Key == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	if err := s.transactor.Run(context.Background(), func(ctx context.Context) error {
		return s.DeleteFileRecord(ctx, dto.Key)
	}); err != nil {
		return err
	}

	_ = s.DeleteStoredFile(dto.Key)
	return nil
}

func (s *CommonService) GetSysAdmin() (userModel.User, error) {
	return s.commonRepository.GetSysAdmin(context.Background())
}

func (s *CommonService) GetStatus() (commonModel.Status, error) {
	ctx := context.Background()

	sysuser, err := s.commonRepository.GetSysAdmin(ctx)
	if err != nil {
		return commonModel.Status{}, err
	}

	var users []commonModel.UserStatus
	allusers, err := s.commonRepository.GetAllUsers(ctx)
	if err != nil {
		return commonModel.Status{}, err
	}
	for _, user := range allusers {
		users = append(users, commonModel.UserStatus{
			UserID:   user.ID,
			UserName: user.Username,
			IsAdmin:  user.IsAdmin,
		})
	}

	echos, err := s.commonRepository.GetAllEchos(ctx, true)
	if err != nil {
		return commonModel.Status{}, err
	}

	return commonModel.Status{
		SysAdminID: sysuser.ID,
		Username:   sysuser.Username,
		Logo:       sysuser.Avatar,
		Users:      users,
		TotalEchos: len(echos),
	}, nil
}

func (s *CommonService) GetHeatMap(timezone string) ([]commonModel.Heatmap, error) {
	ctx := context.Background()
	loc := timezoneUtil.LoadLocationOrUTC(timezone)
	nowUser := time.Now().UTC().In(loc)
	startUser := time.Date(nowUser.Year(), nowUser.Month(), nowUser.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -29)
	endUserExclusive := startUser.AddDate(0, 0, 30)

	createdAtList, err := s.commonRepository.GetHeatMap(ctx, startUser.UTC(), endUserExclusive.UTC())
	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int)
	for _, createdAt := range createdAtList {
		day := createdAt.In(loc).Format("2006-01-02")
		countMap[day]++
	}

	var results [30]commonModel.Heatmap
	for i := 0; i < 30; i++ {
		date := startUser.AddDate(0, 0, i).Format("2006-01-02")
		results[i] = commonModel.Heatmap{
			Date:  date,
			Count: countMap[date],
		}
	}

	return results[:], nil
}

func (s *CommonService) GenerateRSS(ctx *gin.Context) (string, error) {
	echos, err := s.commonRepository.GetAllEchos(ctx.Request.Context(), false)
	if err != nil {
		return "", err
	}

	schema := "http"
	if ctx.Request.TLS != nil {
		schema = "https"
	}
	host := ctx.Request.Host
	feed := &feeds.Feed{
		Title:       "Ech0",
		Link:        &feeds.Link{Href: fmt.Sprintf("%s://%s/", schema, host)},
		Image:       &feeds.Image{Url: fmt.Sprintf("%s://%s/Ech0.svg", schema, host)},
		Description: "Ech0",
		Author:      &feeds.Author{Name: "Ech0"},
		Updated:     time.Now().UTC(),
	}

	for _, msg := range echos {
		renderedContent := mdUtil.MdToHTML([]byte(msg.Content))
		title := msg.Username + " - " + msg.CreatedAt.Format("2006-01-02")

		if len(msg.EchoFiles) > 0 {
			var imageContent []byte
			for _, ef := range msg.EchoFiles {
				imageContent = fmt.Appendf(
					imageContent,
					"<img src=\"%s\" alt=\"Image\" style=\"max-width:100%%;height:auto;\" />",
					ef.File.URL,
				)
			}
			renderedContent = append(imageContent, renderedContent...)
		}

		if len(msg.Tags) > 0 {
			for _, tag := range msg.Tags {
				renderedContent = fmt.Appendf(renderedContent, "<br /><span class=\"tag\">#%s</span>", tag.Name)
			}
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s://%s/echo/%d", schema, host, msg.ID)},
			Description: string(renderedContent),
			Author:      &feeds.Author{Name: msg.Username},
			Created:     msg.CreatedAt,
		})
	}

	return feed.ToAtom()
}

func (s *CommonService) UploadMusic(
	userId uint,
	file *multipart.FileHeader,
) (string, error) {
	fileDto, err := s.UploadFile(userId, file, storageDomain.CategoryAudio)
	if err != nil {
		return "", err
	}

	if err := s.transactor.Run(context.Background(), func(ctx context.Context) error {
		return s.keyvalueRepository.AddOrUpdateKeyValue(
			ctx, globalMusicFileIDKey,
			strconv.FormatUint(uint64(fileDto.ID), 10),
		)
	}); err != nil {
		_ = s.transactor.Run(context.Background(), func(ctx context.Context) error {
			return s.DeleteFileRecord(ctx, fileDto.Key)
		})
		_ = s.DeleteStoredFile(fileDto.Key)
		return "", err
	}

	return fileDto.URL, nil
}

func (s *CommonService) DeleteMusic(userid uint) error {
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	ctx := context.Background()
	val, err := s.keyvalueRepository.GetKeyValue(ctx, globalMusicFileIDKey)
	if err != nil || val == "" {
		return nil
	}

	fileID, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return nil
	}

	fileRecord, err := s.fileRepository.GetByID(ctx, uint(fileID))
	if err != nil {
		return nil
	}

	if err := s.transactor.Run(ctx, func(txCtx context.Context) error {
		if err := s.keyvalueRepository.DeleteKeyValue(txCtx, globalMusicFileIDKey); err != nil {
			return err
		}
		return s.fileRepository.Delete(txCtx, fileRecord.ID)
	}); err != nil {
		return err
	}

	_ = s.DeleteStoredFile(fileRecord.Key)
	return nil
}

func (s *CommonService) GetPlayMusicUrl() string {
	val, err := s.keyvalueRepository.GetKeyValue(context.Background(), globalMusicFileIDKey)
	if err != nil || val == "" {
		return ""
	}

	fileID, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return ""
	}

	fileRecord, err := s.fileRepository.GetByID(context.Background(), uint(fileID))
	if err != nil {
		return ""
	}

	return fileRecord.URL
}

func (s *CommonService) PlayMusic(ctx *gin.Context) {
	val, _ := s.keyvalueRepository.GetKeyValue(context.Background(), globalMusicFileIDKey)
	if val == "" {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}
	fileID, _ := strconv.ParseUint(val, 10, 64)
	fileRecord, err := s.fileRepository.GetByID(context.Background(), uint(fileID))
	if err != nil {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}

	contentType := fileRecord.ContentType
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")

	reader, err := s.fs.Get(context.Background(), fileRecord.Key)
	if err != nil {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}
	defer func() { _ = reader.Close() }()

	readSeeker, ok := reader.(io.ReadSeeker)
	if !ok {
		ctx.String(http.StatusInternalServerError, "音乐文件读取失败")
		return
	}

	http.ServeContent(ctx.Writer, ctx.Request, fileRecord.Name, fileRecord.CreatedAt, readSeeker)
}

func (s *CommonService) GetFilePresignURL(
	userid uint,
	s3Dto *commonModel.GetPresignURLDto,
	method string,
) (commonModel.PresignDto, error) {
	var result commonModel.PresignDto

	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return result, err
	}
	if !user.IsAdmin {
		return result, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if s3Dto.FileName == "" {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}

	contentType := s3Dto.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	category := storageDomain.CategoryImage
	if strings.HasPrefix(contentType, "audio") {
		category = storageDomain.CategoryAudio
	}

	if !isAllowedType(contentType, config.Config().Upload.AllowedTypes) {
		return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	key, err := s.keyGen.GenerateKey(category, userid, s3Dto.FileName)
	if err != nil {
		return result, err
	}

	p, ok := s.fs.(virefs.Presigner)
	if !ok {
		return result, fmt.Errorf("backend does not support presigned URLs")
	}
	req, err := p.PresignPut(context.Background(), key, 24*time.Hour)
	if err != nil {
		return result, err
	}

	url := s.resolveURL(key)

	fileRecord := &fileModel.File{
		Key:         key,
		URL:         url,
		Name:        s3Dto.FileName,
		ContentType: contentType,
		Category:    string(category),
		UserID:      userid,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		logUtil.GetLogger().Error("Failed to save presign file record", zap.String("error", err.Error()))
	}

	result.FileName = s3Dto.FileName
	result.ContentType = contentType
	result.Key = key
	result.PresignURL = req.URL
	result.FileURL = url

	return result, nil
}

func (s *CommonService) CleanupOrphanFiles() error {
	ctx := context.Background()
	threshold := time.Now().UTC().Add(-24 * time.Hour)

	files, err := s.fileRepository.GetOrphanFiles(ctx, threshold)
	if err != nil {
		return err
	}

	musicFileIDStr, _ := s.keyvalueRepository.GetKeyValue(ctx, globalMusicFileIDKey)
	musicFileID := uint(0)
	if musicFileIDStr != "" {
		if id, err := strconv.ParseUint(musicFileIDStr, 10, 64); err == nil {
			musicFileID = uint(id)
		}
	}

	for _, file := range files {
		if file.ID == musicFileID {
			continue
		}
		if file.Key != "" {
			_ = s.fs.Delete(ctx, file.Key)
		}
		_ = s.fileRepository.Delete(ctx, file.ID)
	}

	return nil
}

func (s *CommonService) GetWebsiteTitle(websiteURL string) (string, error) {
	websiteURL = httpUtil.TrimURL(websiteURL)

	body, err := httpUtil.SendRequest(websiteURL, "GET", httpUtil.Header{}, 10*time.Second)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("解析 HTML 失败: %w", err)
	}

	title := extractTitle(doc)
	if title == "" {
		return "", errors.New("未找到网站标题")
	}

	return title, nil
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
		return ""
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}
	return ""
}

func (s *CommonService) keyGenForCategory(category storageDomain.Category, fileName string) storageDomain.KeyGenerator {
	if category == storageDomain.CategoryAudio {
		ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
		if ext == "" {
			ext = ".bin"
		}
		return &storageDomain.StaticKeyGenerator{
			Name: "music" + ext,
		}
	}
	return s.keyGen
}

func isAllowedType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func (s *CommonService) DeleteFileRecord(ctx context.Context, key string) error {
	return s.fileRepository.DeleteByKey(ctx, key)
}

func (s *CommonService) DeleteStoredFile(key string) error {
	if key == "" {
		return nil
	}
	return s.fs.Delete(context.Background(), key)
}

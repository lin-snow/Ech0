package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"golang.org/x/sync/singleflight"

	"github.com/lin-snow/ech0/internal/agent"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
)

type AgentService struct {
	settingService settingService.SettingServiceInterface
	echoService    echoService.EchoServiceInterface
	todoService    todoService.TodoServiceInterface
	kvRepository   keyvalueRepository.KeyValueRepositoryInterface
	recentGenGroup singleflight.Group
}

func NewAgentService(
	settingService settingService.SettingServiceInterface,
	echoService echoService.EchoServiceInterface,
	todoService todoService.TodoServiceInterface,
	kvRepository keyvalueRepository.KeyValueRepositoryInterface,
) AgentServiceInterface {
	return &AgentService{
		settingService: settingService,
		echoService:    echoService,
		todoService:    todoService,
		kvRepository:   kvRepository,
	}
}

func (agentService *AgentService) GetRecent(ctx context.Context) (string, error) {
	const cacheKey = string(agent.GEN_RECENT)

	if value, ok := agentService.getRecentFromCache(cacheKey); ok {
		return value, nil
	}

	value, err, _ := agentService.recentGenGroup.Do(cacheKey, func() (any, error) {
		if cached, ok := agentService.getRecentFromCache(cacheKey); ok {
			return cached, nil
		}

		output, err := agentService.buildRecentSummary(ctx)
		if err != nil {
			return "", err
		}

		agentService.kvRepository.AddOrUpdateKeyValue(ctx, cacheKey, output)
		return output, nil
	})
	if err != nil {
		return "", err
	}

	recent, ok := value.(string)
	if !ok {
		return "", errors.New("recent summary type assertion failed")
	}

	return recent, nil
}

func (agentService *AgentService) getRecentFromCache(cacheKey string) (string, bool) {
	cachedValue, err := agentService.kvRepository.GetKeyValue(cacheKey)
	if err != nil {
		return "", false
	}

	value, ok := cachedValue.(string)
	return value, ok
}

func (agentService *AgentService) buildRecentSummary(ctx context.Context) (string, error) {
	echos, err := agentService.echoService.GetEchosByPage(authModel.NO_USER_LOGINED, commonModel.PageQueryDto{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		return "", err
	}

	var memos []*schema.Message
	for i, e := range echos.Items {
		content := fmt.Sprintf(
			"用户 %s 在 %s 发布了内容 %d ：%s 。 内容标签为：%v。",
			e.Username,
			e.CreatedAt.Format("2006-01-02 15:04"),
			i+1,
			e.Content,
			e.Tags,
		)

		memos = append(memos, &schema.Message{
			Role:    schema.User,
			Content: content,
		})
	}

	in := []*schema.Message{
		{
			Role: schema.System,
			Content: `
				你只能输出纯文本。
				不能输出代码块、格式化标记、Markdown 符号（如井号、星号、反引号、方括号、尖括号）。
				不能输出任何结构化格式（如列表、表格）。
				回复中只能出现正常文字、标点符号和 Emoji 和 换行。
				确保输出始终是自然语言连续文本。`,
		},
		{
			Role:    schema.User,
			Content: "请根据提供的近期互动内容（内容可能包括日常生活、句子诗词摘抄、吐槽等等），总结该用户最近的活动和状态，突出作者状态即可，不需要详细描述内容，如果没有任何内容，请回复作者最近很神秘~",
		},
	}

	in = append(in, memos...)

	var setting model.AgentSetting
	if err := agentService.settingService.GetAgentInfo(&setting); err != nil {
		return "", errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}

	output, err := agent.Generate(ctx, setting, in)
	if err != nil {
		return "", err
	}

	return output, nil
}

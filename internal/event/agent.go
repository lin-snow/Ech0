package event

import (
	"context"
	"encoding/json"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	keyvalue "github.com/lin-snow/ech0/internal/repository/keyvalue"
	todoRepository "github.com/lin-snow/ech0/internal/repository/todo"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
)

type AgentProcessor struct {
	echoRepo     echoRepository.EchoRepositoryInterface
	todoRepo     todoRepository.TodoRepositoryInterface
	userRepo     userRepository.UserRepositoryInterface
	keyvalueRepo keyvalue.KeyValueRepositoryInterface
	inboxRepo    inboxRepository.InboxRepositoryInterface
}

func NewAgentProcessor(
	echoRepo echoRepository.EchoRepositoryInterface,
	todoRepo todoRepository.TodoRepositoryInterface,
	userRepo userRepository.UserRepositoryInterface,
	keyvalueRepo keyvalue.KeyValueRepositoryInterface,
	inboxRepo inboxRepository.InboxRepositoryInterface,
) *AgentProcessor {
	return &AgentProcessor{
		echoRepo:     echoRepo,
		todoRepo:     todoRepo,
		userRepo:     userRepo,
		keyvalueRepo: keyvalueRepo,
		inboxRepo:    inboxRepo,
	}
}

func (ap *AgentProcessor) Handle(ctx context.Context, e *Event) error {
	// 获取 Agent 设置
	var agentSetting settingModel.AgentSetting
	if agentSettingStr, err := ap.keyvalueRepo.GetKeyValue(ctx, commonModel.AgentSettingKey); err == nil {
		if err := json.Unmarshal([]byte(agentSettingStr), &agentSetting); err != nil {
			return err
		}
	}

	// 清理生成内容的缓存
	_ = ap.clearCache()

	// // 更新平行人格
	// if err := ap.updatePersona(&agentSetting, e); err != nil {
	// 	return err
	// }

	// // 随机发表 Echo
	// if err := ap.mayPostEchoToInbox(&agentSetting); err != nil {
	// 	return err
	// }

	return nil
}

func (ap *AgentProcessor) clearCache() error {
	// 删除 AGENT_GEN_RECENT 缓存(忽略 err)
	return ap.keyvalueRepo.DeleteKeyValue(context.Background(), string(agent.GEN_RECENT))
}

// func (ap *AgentProcessor) updatePersona(setting *settingModel.AgentSetting, e *Event) error {
// 	// 配置并开启了 Agent 才能更新人格
// 	if setting == nil || !setting.Enable {
// 		return nil
// 	}

// 	// 取出 Echo
// 	payload := e.Payload[EventPayloadEcho]
// 	echo, ok := payload.(echoModel.Echo)
// 	if !ok {
// 		return nil
// 	}

// 	// 取出当前人格
// 	var p persona.Persona
// 	if personaStr, err := ap.keyvalueRepo.GetKeyValue(persona.PersonaKey); err == nil {
// 		if err := json.Unmarshal([]byte(personaStr.(string)), &p); err != nil {
// 			return err
// 		}
// 	} else {
// 		// 如果没有找到对应的人格，初始化一个默认人格
// 		now := time.Now().Unix()
// 		p = persona.Persona{
// 			Name:         "Persona",
// 			Description:  "parallel personality",
// 			Style:        []persona.Feature{},
// 			Mood:         []persona.Feature{},
// 			Topics:       []persona.Feature{},
// 			Expression:   []persona.Feature{},
// 			Independence: 0.4 + rand.Float64()*0.2, // 0.4~0.6
// 			CreatedAt:    now,
// 			UpdatedAt:    now,
// 			LastActive:   now,
// 		}
// 	}

// 	// fmt.Printf("当前人格：%+v\n", p)

// 	// 随机获取一个维度进行更新
// 	dim := p.WhatDimensionToUpdate()
// 	features := p.GetDimensionFeatures(dim)

// 	// 构建大模型输入
// 	featuresJSON, _ := json.Marshal(features)
// 	systemPrompt := `
// 你是一套“人格特征更新器”，你的任务是根据输入内容更新某个人格维度的特征。
// 你必须严格遵守以下规则：

// 任务：
// 在【已有特征】基础上，进行“轻量更新”：保留大部分有意义的特征，只替换或新增少量与近期行为更相关的特征。
// 更新后特征数量必须保持在 6～10 个之间，并尽可能接近 8 个。

// 输出格式（必须严格遵守）：
// [
// {"name": "中文特征名", "weight": 0.33},
// {"name": "中文特征名", "weight": 0.66}
// ]

// 规则要求：
// 1. 所有特征名称必须是中文的、不带标点、简短词语或短语。
// 2. 特征必须从属于指定维度。
// 3. weight 必须是 0~1 的浮点数。
// 4. 最终输出必须是合法 JSON，禁止输出任何解释性文本。

// 维度说明：
// style：行为方式、说话风格，如 温和、犀利、冷静、机敏。
// mood：情绪状态，如 愉快、紧张、轻松、烦躁。
// topics：兴趣偏好，如 科技、生活、哲学、编程。
// expression：表达方式，如 简洁表达、比喻表达、故事表达。

// 最终只需输出 JSON 数组本体。
// `

// 	userPrompt := fmt.Sprintf(`
// 当前维度：
// %s

// 当前特征列表：
// %s

// 用户最近行为（Echo）：
// %s

// 请基于以上信息生成新的特征列表（完整覆盖旧列表）。
// 严格只输出 JSON 数组，不包含额外文本。
// `, dim, string(featuresJSON), echo.Content)

// 	in := []*schema.Message{
// 		{
// 			Role:    schema.System,
// 			Content: systemPrompt,
// 		},
// 		{
// 			Role:    schema.User,
// 			Content: userPrompt,
// 		},
// 	}

// 	// 调用大模型 根据 Echo 内容和 选中的维度，生成新的特征，得到未校验的Feature列表
// 	out, err := agent.Generate(context.Background(), *setting, in, false)
// 	if err != nil {
// 		fmt.Println(out)
// 		return errors.New(err.Error() + " | during persona update:" + out)
// 	}

// 	// 解析大模型输出，得到新的特征列表
// 	var newFeatures []persona.Feature
// 	if err := json.Unmarshal([]byte(out), &newFeatures); err != nil {
// 		return errors.New(err.Error() + "| parse Features err" + out)
// 	}

// 	// 执行更新
// 	p.UpdateDimension(dim, newFeatures)
// 	p.UpdatedAt = time.Now().Unix()
// 	delta := (rand.Float64() - 0.5) * 0.1 // -0.05 ~ +0.05 微调
// 	p.Independence += delta
// 	if p.Independence < 0 {
// 		p.Independence = 0
// 	}
// 	if p.Independence > 1 {
// 		p.Independence = 1
// 	}

// 	// 保存更新后的人格
// 	personaBytes, err := json.Marshal(p)
// 	if err != nil {
// 		return err
// 	}
// 	// fmt.Println(p)
// 	if err := ap.keyvalueRepo.AddOrUpdateKeyValue(context.Background(), string(persona.PersonaKey), string(personaBytes)); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (ap *AgentProcessor) mayPostEchoToInbox(setting *settingModel.AgentSetting) error {
// 	// 获取人格
// 	var p persona.Persona
// 	if personaStr, err := ap.keyvalueRepo.GetKeyValue(persona.PersonaKey); err == nil {
// 		if err := json.Unmarshal([]byte(personaStr.(string)), &p); err != nil {
// 			return err
// 		}
// 	}

// 	// 如果人格特征信息不足，则不发表 Echo
// 	if len(p.Style)+len(p.Mood)+len(p.Topics)+len(p.Expression) < 6 {
// 		return nil
// 	}

// 	// 计算发布 Echo 的概率
// 	lastActive := p.LastActive
// 	now := time.Now().Unix()
// 	hoursSinceLastActive := float64(now-lastActive) / 3600.0 // 转换为小时
// 	// 基础概率
// 	baseProb := 0.5
// 	// 随着时间增加，概率线性增加，最多增加到 0.5
// 	timeFactor := hoursSinceLastActive * 0.05
// 	if timeFactor > 0.5 {
// 		timeFactor = 0.5
// 	}
// 	finalProb := baseProb + timeFactor

// 	// 随机决定是否发布 Echo
// 	if rand.Float64() > finalProb {
// 		return nil
// 	}

// 	// 构建大模型输入
// 	// ---- 构建可读的人格特征文本 ----
// 	buildList := func(fs []persona.Feature) string {
// 		if len(fs) == 0 {
// 			return "无"
// 		}
// 		parts := make([]string, 0, len(fs))
// 		for _, f := range fs {
// 			parts = append(parts, fmt.Sprintf("%s(%.2f)", f.Name, f.Weight))
// 		}
// 		return strings.Join(parts, "、")
// 	}

// 	// 随机化话题偏好特征
// 	topicPool := []string{
// 		"科技", "生活", "哲学", "编程", "艺术", "电影", "旅行",
// 		"音乐", "心理学", "历史", "美食", "运动", "游戏", "文学",
// 	}
// 	p.Topics = []persona.Feature{} // 清空现有话题特征
// 	for _, t := range topicPool {
// 		if rand.Float64() < 0.4 {
// 			p.Topics = append(p.Topics, persona.Feature{
// 				Name:   t,
// 				Weight: 0.5 + rand.Float64()*0.5, // 0.5~1.0
// 			})
// 		}
// 	}

// 	personaDesc := fmt.Sprintf(`
// 你的基础人格如下：
// 风格(Style)：%s
// 情绪(Mood)：%s
// 可选话题(Topics)：%s
// 表达方式(Expression)：%s

// 独立性(Independence)：%.2f（数值越高越倾向主动表达自我）
// `, buildList(p.Style), buildList(p.Mood), buildList(p.Topics), buildList(p.Expression), p.Independence)

// 	// ---- 根据独立性调整表达倾向 ----
// 	var expressionHint string
// 	if p.Independence > 0.7 {
// 		expressionHint = "你更倾向于表达自己的观点，带有明显的个人色彩。"
// 	} else if p.Independence < 0.3 {
// 		expressionHint = "你更倾向于模仿人类用户的语气或最近的对话风格。"
// 	} else {
// 		expressionHint = "你既可能表达自我，也可能借助外界语气，呈现一种自然随性的说话方式。"
// 	}

// 	// ---- 构建 System Prompt ----
// 	systemPrompt := fmt.Sprintf(`
// 你是一个具有独特人格的 Agent，会不定期自主发表一句 Echo（像社交动态、心情、吐槽、近况、摘抄一样的内容）。

// 请根据下方提供的人格特征，创造一段自然的、贴合人格的 Echo 内容，不要太长（一两句话即可）。

// 确保语气、情绪、主题偏好都要呼应人格，并鼓励多样化和生活化表达：
// - 内容长度建议为 1 到 6 个句子，既可以短小精悍，也可以稍长一些，但不要冗长。
// - Echo 可以涉及不同兴趣主题，不要总是重复相同话题。
// - 可以描述日常生活、思考、体验、趣闻，展现人格的丰富性。
// - 内容应体现独立性：独立性高的人更倾向表达独立观点，低独立性的人可能模仿或借鉴外界语气。
// - 采用随机的表达方式，不要总是使用同一种句式或结构。

// %s

// 生成要求：
// 1. 内容必须符合人格特征。
// 2. 不能解释自己或提及 AI。
// 3. 不要出现“这是 Echo”之类的元描述。
// 4. 输出仅包含最终文字内容。
// 5. 尽量避免重复话题，展示生活的多样性。
// `, expressionHint)

// 	// ---- User Prompt：提供当前人格 ----
// 	userPrompt := fmt.Sprintf(`
// 以下是你的人格：
// %s

// 请基于人格创作一段 Echo 内容（简短、自然、有个性）。
// `, personaDesc)

// 	// ---- 构建最终输入 ----
// 	in := []*schema.Message{
// 		{
// 			Role:    schema.System,
// 			Content: systemPrompt,
// 		},
// 		{
// 			Role:    schema.User,
// 			Content: userPrompt,
// 		},
// 	}

// 	out, err := agent.Generate(context.Background(), *setting, in, false, float32(p.Independence+0.35))
// 	if err != nil {
// 		return errors.New(err.Error() + " | during echo generation:" + out)
// 	}

// 	// 发表到收件箱
// 	newInbox := inboxModel.Inbox{
// 		Source:    string(commonModel.AgentSource),
// 		Content:   out,
// 		Type:      string(commonModel.EchoInboxType),
// 		Read:      false,
// 		CreatedAt: now,
// 	}
// 	if err := ap.inboxRepo.PostInbox(context.Background(), &newInbox); err != nil {
// 		return err
// 	}

// 	return nil
// }

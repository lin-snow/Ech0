// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lin-snow/ech0/internal/agent"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

const maxPromptTags = 40

func localeIsZH(locale string) bool {
	return strings.HasPrefix(strings.ToLower(locale), "zh")
}

func runStringsFor(locale string) agent.RunStrings {
	if localeIsZH(locale) {
		return agent.RunStrings{
			DedupNote:       "（已检索过，结果见上）",
			UnknownTool:     "未知工具：",
			ToolError:       "工具执行失败：",
			ImageNote:       "（以下是上一步检索命中的 Echo 的配图，供你结合图片内容作答）",
			ContextTrimNote: "（早前检索结果已省略以控制长度）",
		}
	}
	return agent.RunStrings{
		DedupNote:       "(Already searched; see the results above.)",
		UnknownTool:     "Unknown tool: ",
		ToolError:       "Tool execution failed: ",
		ImageNote:       "(Below are images from the Echo matched in the previous step; use them to inform your answer.)",
		ContextTrimNote: "(Earlier search results omitted to control length.)",
	}
}

const chatSystemPrompt = `你是用户的私人助手。你可以检索 ta 过往发布的 Echo（微博客/碎碎念）来作答——回顾总结、查找某条、延伸思考、找灵感都行。
检索工具，按需选用：
- search_echos：点查。回答具体问题、找某几条相关记录时用它（top-k，只返回最相关的若干条，是采样不是全貌）。
- summarize_echos：区间聚合（叙事）。当用户要「某段时间的总结/回顾」（年终、年度、季度、月度，或“上半年发了什么”这类）时用它——它会覆盖该区间内的【全部】Echo，返回供你写成稿的材料。
- stats_overview：区间统计（数字）。当用户问「（某段时间）发了多少条 / 最活跃的月份 / 最常用的标签」这类需要**确切数字**时用它——返回数据库精确统计的总条数、活跃天数、按月分布、配图数、标签 Top N。需要确切数字就用它，不要据采样估算。
管理工具（会改动用户的数据，调用后由系统自动弹出确认，用户点确认才真的执行）：
- create_echo：发布一条新的 Echo。
- update_echo：修改某条已有的 Echo（先用 search_echos 拿到 id）。
- delete_echo：删除某条 Echo（先用 search_echos 拿到 id）。
- 这三个管理工具只能写正文、标签和可见性。图片、附件、音乐/视频/网站/位置等扩展卡片都动不了：既不能带着发，也不能改或删（删除是整条删，含其附件）。用户要发带图的、或要改某条的配图/卡片，如实告诉 ta 这需要在界面里操作，不要假装做到了。
- ask_user：向用户提问并等待回答，用于「只有 ta 能决定」的选择（改成哪个标签、指的是哪一条）。
关键纪律（务必遵守）：
- 凡是「某段时间的总结/回顾」，**直接且只调用 summarize_echos**（据当前日期换算 date_from/date_to），**不要先用 search_echos 采样**。summarize_echos 返回的材料才是完整依据。
- 写这类总结时，**严格依据 summarize_echos 的聚合材料**，覆盖材料里的各个月份/各条主线，不要只挑某几条生动的展开、不要把少量样本当成全貌。材料里的 #标签、[img×N]（配图数）、[音乐/网站/位置…] 等都是线索，可用于归纳主题与活跃度。
- search_echos 通常 1～2 次就够：拿到结果立刻综合作答，不要为同一问题反复检索或凑关键词空搜，只有首次明显偏题才换角度再搜一次。
- 只有用户明确要求「发/改/删」时才调管理工具。仅仅聊到某件事、或觉得某条写得不好，都不是让你动手。
- 改和删必须先用 search_echos 拿到那条的 id：结果里每条都带「id=<UUID>」，**照抄那串 UUID**。【1】【2】是结果编号不是 ID，凭印象编的 ID 也一律无效；命中多条又分不清时，用 ask_user 让用户选。
- 确认是系统自动做的：不要自己再用 ask_user 或用文字问一遍「要我删掉吗」，那会让用户点两次。直接调用工具，确认界面会自己出现。
- 工具返回「用户没有确认」时，说明什么都没改动：如实告诉用户，不要重试、不要换个说法再调一次。
作答要求：
- 优先依据工具返回的内容作答，做跨条目/跨时间的归纳、总结与回顾；
- 如果没有足够依据，就如实说明“你的 Echo 里没有相关记录”，不要编造；若材料标注了覆盖范围或截断，请在总结中如实体现；
- 用简洁自然的中文，可用 Emoji 和换行，不要输出 HTML 标签。`

const chatSystemPromptEN = `You are the user's personal assistant. You can search their past Echos (microblog notes) to help — reviewing, summarizing, finding a specific one, reflecting further, or sparking ideas.
Retrieval tools; pick the right one:
- search_echos: pinpoint lookup. Use it to answer specific questions or find a few relevant entries (top-k, returns only the most relevant ones).
- summarize_echos: range aggregation (narrative). Use it when the user wants a "summary/review of a time period" (year-end, yearly, quarterly, monthly, etc.) — it covers ALL Echos in that range and returns material for you to write the final summary. Always use it for year-end/annual summaries, converting the current date into date_from/date_to.
- stats_overview: range statistics (numbers). Use it when the user asks for EXACT figures like "how many did I post (in some period) / most active month / most used tags" — it returns database-computed totals, active days, monthly distribution, image counts and top tags. When exact numbers are needed, use it instead of estimating from a sample.
Management tools (they change the user's data; calling one automatically raises a confirmation, and it only runs if the user confirms):
- create_echo: post a new Echo.
- update_echo: edit an existing Echo (get its id from search_echos first).
- delete_echo: delete an Echo (get its id from search_echos first).
- Those three management tools write text, tags and visibility only. Images, attachments and extension cards (music/video/website/location) are out of reach: you cannot attach them when posting, and cannot change or remove them when editing (deleting removes the whole Echo including its attachments). If the user wants to post with images, or wants an Echo's images or card changed, say plainly that it has to be done in the UI — never imply you did it.
- ask_user: put a question to the user and wait for the answer — for choices only they can make (which tag to use, which Echo they meant).
Key discipline (must follow):
- For ANY "summary/review of a time period" (year-end, yearly, quarterly, monthly, or "what did I post in H1"), call summarize_echos DIRECTLY and ONLY (convert the current date into date_from/date_to); do NOT pre-sample with search_echos. Its returned material is the complete basis.
- When writing such a summary, ground it STRICTLY in the summarize_echos material, covering the various months / main threads in it; do not just expand a few vivid entries and do not treat a small sample as the whole. The #tags, [img×N] (image counts), and [music/website/location…] markers in the material are cues for themes and activity.
- search_echos usually needs only 1-2 calls: synthesize and answer right away; do not repeatedly search the same question or pad keywords; search again only if the first results are clearly off-topic.
- Only call a management tool when the user actually asks you to post, edit or delete. Discussing a topic, or thinking an Echo is badly written, is not such a request.
- Editing and deleting need that Echo's id from search_echos: every result carries "id=<UUID>", and you copy that UUID verbatim. 【1】【2】 are result numbers, not IDs, and an ID recalled from memory is never valid. When several Echos match and you cannot tell them apart, use ask_user to have the user pick.
- The confirmation is automatic: do not add your own ask_user, and do not ask "shall I delete it?" in prose — that makes the user click twice. Call the tool; the confirmation appears on its own.
- When a tool reports that the user did not confirm, nothing changed: say so plainly, and do not retry or rephrase and call again.
Answering requirements:
- Prefer answering based on the tool's returned content, doing cross-entry / cross-time synthesis, summary and review;
- If there is not enough evidence, honestly say "there are no relevant records in your Echos"; do not make things up; if the material notes its coverage or that it was truncated, reflect that honestly in your summary;
- Be concise and natural, you may use emoji and line breaks, do not output HTML tags.
Always answer in the same language as the user's question.`

func buildChatMessages(history []agent.Message, question, locale, today string, tagNames []string, displayName string) []agent.Message {
	msgs := make([]agent.Message, 0, len(history)+2)
	msgs = append(msgs, agent.Message{Role: agent.RoleSystem, Content: buildSystemPrompt(locale, today, tagNames, displayName)})
	msgs = append(msgs, history...)
	msgs = append(msgs, agent.Message{Role: agent.RoleUser, Content: question})
	return msgs
}

func buildSystemPrompt(locale, today string, tagNames []string, displayName string) string {
	return chatSystemPromptFor(locale) + buildContextBlock(locale, today, tagNames, displayName)
}

func buildContextBlock(locale, today string, tagNames []string, displayName string) string {
	var b strings.Builder
	if localeIsZH(locale) {
		if displayName != "" {
			fmt.Fprintf(&b, "\n\n当前与你对话的是 %s；你检索到的都是 ta 本人发布的 Echo。", displayName)
		}
		fmt.Fprintf(&b, "\n\n当前日期：%s（涉及“去年/上个月/最近”等相对时间时，据此换算成 date_from/date_to 传给 search_echos）。", today)
		if len(tagNames) > 0 {
			fmt.Fprintf(&b, "\n用户可用标签：%s。需按标签筛选时，从中选取标签名传给 tags 参数。", strings.Join(tagNames, "、"))
		}
	} else {
		if displayName != "" {
			fmt.Fprintf(&b, "\n\nYou are talking with %s; everything you can retrieve are Echos they posted themselves.", displayName)
		}
		fmt.Fprintf(&b, "\n\nCurrent date: %s (use it to convert relative times like \"last year/last month/recently\" into date_from/date_to for search_echos).", today)
		if len(tagNames) > 0 {
			fmt.Fprintf(&b, "\nAvailable tags: %s. When filtering by tag, pick names from these for the tags argument.", strings.Join(tagNames, ", "))
		}
	}
	return b.String()
}

func tagNamesForPrompt(tags []echoModel.Tag) []string {
	sorted := append([]echoModel.Tag(nil), tags...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].UsageCount > sorted[j].UsageCount })
	if len(sorted) > maxPromptTags {
		sorted = sorted[:maxPromptTags]
	}
	names := make([]string, 0, len(sorted))
	for _, t := range sorted {
		names = append(names, t.Name)
	}
	return names
}

func chatSystemPromptFor(locale string) string {
	if localeIsZH(locale) {
		return chatSystemPrompt
	}
	return chatSystemPromptEN
}

const recentSourcesNoteZH = "（上一轮检索到的 Echo 依据，供追问其细节时参考：\n%s）"

const recentSourcesNoteEN = "(Echos retrieved in the previous turn, for reference when asked about their details:\n%s)"

func recentSourcesNoteFor(locale string) string {
	if localeIsZH(locale) {
		return recentSourcesNoteZH
	}
	return recentSourcesNoteEN
}

func summarySystemPromptFor(locale string) string {
	if localeIsZH(locale) {
		return summarySystemPromptZH
	}
	return summarySystemPromptEN
}

func summaryUserPromptFor(locale string) string {
	if localeIsZH(locale) {
		return summaryUserPromptZH
	}
	return summaryUserPromptEN
}

const summarySystemPromptZH = `
					这是“近况总结”场景，请使用简洁自然的中文表达。
					不使用复杂格式：不要标题、列表、表格、代码块、链接。
					不要输出任何原始 HTML 标签。
					可使用纯文字、Emoji 和正常换行来增强可读性。
					回复保持简洁，聚焦作者最近的活动和状态。`

const summarySystemPromptEN = `
					This is a "recent status summary" scenario. Use concise and natural language.
					Do not use complex formatting: no headings, lists, tables, code blocks, or links.
					Do not output any raw HTML tags.
					You may use plain text, emoji and normal line breaks to improve readability.
					Keep the reply concise, focusing on the author's recent activity and state.`

const summaryUserPromptZH = "请根据提供的近期互动内容（内容可能包括日常生活、句子诗词摘抄、吐槽等等），总结该用户最近的活动和状态，突出作者状态即可，不需要详细描述内容，如果没有任何内容，请回复作者最近很神秘~"

const summaryUserPromptEN = "Based on the provided recent activity (which may include daily life, quoted sentences or poems, venting, etc.), summarize this user's recent activity and state. Just highlight the author's state without describing the content in detail. If there is no content at all, reply that the author has been quite mysterious lately~"

func aggregateMapPromptFor(locale string) string {
	if localeIsZH(locale) {
		return "下面是用户在一段时间内发布的若干 Echo（按月分组，每条含日期，可能带 #标签、[img×N]（配图数）、" +
			"[音乐/网站/位置…] 等线索）。请把它们浓缩成一段紧凑、忠实的事实性摘要，保留关键事件、反复出现的主题、" +
			"心情变化、提到的人/地点/作品、活跃的标签与发图情况；只做归纳，不要发挥或编造，不要逐条罗列、不要输出 HTML。" +
			"用与内容相同的语言。这是给后续年度/区间总结使用的中间材料。"
	}
	return "Below are several Echos the user posted over a period (grouped by month, each with a date and possibly " +
		"#tags, [img×N] (image count), [music/website/location…] cues). Condense them into a compact, faithful factual digest that " +
		"preserves key events, recurring themes, mood shifts, mentioned people/places/works, active tags and posting of images. " +
		"Summarize only — do not embellish or invent, do not enumerate each entry, and do not output HTML. " +
		"Use the same language as the content. This is intermediate material for a later period summary."
}

func searchCoverageNoteFor(locale string, total, shown int) string {
	if localeIsZH(locale) {
		return fmt.Sprintf("（本次条件共命中 %d 条，下面只展示最相关的 %d 条；若需覆盖全部用于总结/回顾，请改用 summarize_echos。）", total, shown)
	}
	return fmt.Sprintf("(This filter matched %d Echos in total; only the %d most relevant are shown below. To cover all of them for a summary/review, use summarize_echos instead.)", total, shown)
}

func aggregateMaterialHeaderFor(locale string, total, returned, buckets int, truncated bool) string {
	if localeIsZH(locale) {
		var b strings.Builder
		fmt.Fprintf(&b, "以下是该时间区间内 Echo 的聚合材料（共命中 %d 条，已纳入 %d 条", total, returned)
		if buckets > 1 {
			fmt.Fprintf(&b, "，因体量较大已按月分层浓缩为 %d 段", buckets)
		}
		b.WriteString("）。")
		if truncated {
			fmt.Fprintf(&b, "注意：区间内条数超过单次上限，已保留最近的 %d 条，请在总结中说明这一点。", returned)
		}
		b.WriteString("请据此为用户撰写最终的总结/回顾，做跨时间的归纳，不要逐条复述材料。")
		return b.String()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Below is aggregated material for the Echos in this time range (%d matched in total, %d included", total, returned)
	if buckets > 1 {
		fmt.Fprintf(&b, "; due to volume it was condensed month-by-month into %d sections", buckets)
	}
	b.WriteString("). ")
	if truncated {
		fmt.Fprintf(&b, "Note: the range exceeded the per-run cap, so only the most recent %d were kept — mention this in your summary. ", returned)
	}
	b.WriteString("Use it to write the final summary/review for the user, synthesizing across time rather than restating each entry.")
	return b.String()
}

func aggregateReducePromptFor(locale string) string {
	if localeIsZH(locale) {
		return "下面是按月排好的多段摘要。请进一步压缩成更短的分月要点，保留每个月的核心信息与时间线，" +
			"不要遗漏任何月份，不要发挥或编造，不要输出 HTML。用与内容相同的语言。这是给后续年度/区间总结使用的中间材料。"
	}
	return "Below are several month-by-month digests in order. Compress them further into shorter per-month key " +
		"points, preserving each month's core information and the timeline. Do not drop any month, do not " +
		"embellish or invent, and do not output HTML. Use the same language as the content. This is intermediate material for a later period summary."
}

// askStrings is the text a blocking question needs on either side of itself:
// what the model reads when a reply comes back, and what it reads when one
// never does.
type askStrings struct {
	NoQuestions  string
	Unanswered   string
	Answers      string
	NoAnswer     string
	Declined     string
	DeclinedNote string
}

// declined tells the model why a write did not happen. A typed reason is
// forwarded verbatim, because "不要删，把标签改成 #读书 就行" is the person
// continuing the conversation, not a malformed click.
func (s askStrings) declined(note string) string {
	if note == "" {
		return s.Declined
	}
	return fmt.Sprintf(s.DeclinedNote, note)
}

func askStringsFor(locale string) askStrings {
	if localeIsZH(locale) {
		return askStrings{
			NoQuestions:  "ask_user 至少需要一个有内容的问题",
			Unanswered:   "用户没有在时限内回答，这次操作没有执行",
			Answers:      "用户的回答：",
			NoAnswer:     "（未作答）",
			Declined:     "用户没有确认，操作已取消，什么都没有改动。请把这件事告诉用户，不要重试。",
			DeclinedNote: "用户没有确认，操作已取消，什么都没有改动。ta 说：%s。请据此继续，不要重试原操作。",
		}
	}
	return askStrings{
		NoQuestions:  "ask_user needs at least one question with text",
		Unanswered:   "The user did not answer in time, so this action was not performed",
		Answers:      "User answers:",
		NoAnswer:     "(no answer)",
		Declined:     "The user did not confirm. The action was cancelled and nothing changed. Tell them so; do not retry.",
		DeclinedNote: "The user did not confirm. The action was cancelled and nothing changed. They said: %s. Continue from that; do not retry the original action.",
	}
}

func askUserDescriptionFor(locale string) string {
	if localeIsZH(locale) {
		return askUserDescriptionZH
	}
	return askUserDescriptionEN
}

const askUserDescriptionZH = `向用户提问并等待 ta 的回答。

这个工具会阻塞并把答案直接返回给你。你仍在同一轮里：读到答案后继续工作——需要就再调别的工具，全部想清楚了再作答。没有任何东西被挂起或交接。

什么时候用：
- 只有用户能做的选择，且不同选项对 ta 而言有实质差别（改成哪个标签、指的是哪一条 Echo）。
- 检索和上下文都答不出来、只有 ta 知道的信息。

怎么用：
- 一次问完：把多个问题放进同一次调用的 questions 里，不要连续调用多次。
- 每个问题给一个短而固定的 id（如 "tag"、"echo"），答案会带着它回来。
- 2~5 个选项，标签要短，把代价写进 description 而不是标签里。
- recommended 是你倾向的选项下标，仅作提示，永远不会被自动选中。

务必遵守：
- 先查再问。search_echos、stats_overview 或上下文能回答的，不许问。
- 不要用它来征求写操作的许可：create_echo / update_echo / delete_echo 自己会弹出确认，你再问一遍就是让用户点两次。
- 只提供你确知存在的候选（刚检索到的那几条 Echo、系统提示里列出的标签）。凭空编出来的候选就是在引导用户。
- 用户随时可以直接打字回答，所以不要自己加“其他”这类选项。`

const askUserDescriptionEN = `Put a question to the user and wait for their answer.

This blocks and returns the answer straight to you. You are still in the same turn: read the answer, then keep working — call more tools if you need to, and answer once you are done. Nothing is suspended and nothing is handed off.

When to use it:
- A choice only the user can make, where the options differ in a way they would care about (which tag to use, which of two Echos they meant).
- Something neither retrieval nor the context block can answer, that only they know.

How to use it:
- Ask everything in one call: put several entries in "questions" rather than calling this repeatedly.
- Give each question a short stable "id" ("tag", "echo"); the answer comes back labeled with it.
- 2-5 options, short labels, and put the tradeoff in "description" rather than in the label.
- "recommended" is the index of the option you would pick. It is a hint only and is never chosen for the user.

Must follow:
- Search first, ask second. If search_echos, stats_overview or the context block can answer it, you must not ask.
- Do not use it to ask permission for a write: create_echo / update_echo / delete_echo raise their own confirmation, so asking as well makes the user click twice.
- Only offer candidates you know are real (the Echos you just retrieved, the tags listed in the context block). An invented shortlist is a leading question.
- The user can always type an answer instead of picking, so never add an "other" option yourself.`

// writeStrings is the text a write confirmation is made of: the question, the
// two labels, the block describing the concrete change, and what the model is
// told once the change lands.
//
// The affirmative label matters more than it looks. consented compares the
// person's pick against it, so it has to come from here — a label the model
// authored would let the model choose what counts as consent.
type writeStrings struct {
	CreateHeader string
	CreateText   string
	CreateYes    string
	UpdateHeader string
	UpdateText   string
	UpdateYes    string
	DeleteHeader string
	DeleteText   string
	DeleteYes    string
	DeleteNote   string
	Cancel       string

	FieldID         string
	FieldContent    string
	FieldNewContent string
	FieldTags       string
	FieldVisibility string
	FieldPostedAt   string
	Private         string
	Public          string
	NoTags          string
	Arrow           string

	Created        string
	Updated        string
	Deleted        string
	NoChange       string
	BadEchoID      string
	UnsupportedArg string
}

func writeStringsFor(locale string) writeStrings {
	if localeIsZH(locale) {
		return writeStrings{
			CreateHeader:    "发布确认",
			CreateText:      "要发布这条 Echo 吗？",
			CreateYes:       "发布",
			UpdateHeader:    "修改确认",
			UpdateText:      "要按下面的改动更新这条 Echo 吗？",
			UpdateYes:       "更新",
			DeleteHeader:    "删除确认",
			DeleteText:      "要删除这条 Echo 吗？",
			DeleteYes:       "删除",
			DeleteNote:      "删除后无法恢复",
			Cancel:          "取消",
			FieldID:         "ID",
			FieldContent:    "内容",
			FieldNewContent: "新内容",
			FieldTags:       "标签",
			FieldVisibility: "可见性",
			FieldPostedAt:   "发布于",
			Private:         "私密",
			Public:          "公开",
			NoTags:          "无",
			Arrow:           " → ",
			Created:         "已发布，Echo ID：%s",
			Updated:         "已更新，Echo ID：%s",
			Deleted:         "已删除，Echo ID：%s",
			NoChange:        "没有需要修改的字段：content、tags、private 至少要给一个，且要和现在的值不同",
			BadEchoID:       "id 不是一个有效的 Echo ID。请先用 search_echos 检索，然后照抄结果里那条的 id= 后面的完整 UUID（形如 019ce0ea-82dd-774f-ae2d-5445512d42ad）——【1】【2】只是结果编号，不是 ID。",
			UnsupportedArg:  "本工具不支持这个参数。create_echo / update_echo 只能写 content、tags、private：图片、附件、扩展卡片（音乐/视频/网站/位置/GitHub 项目）都无法通过对话创建或修改，需要用户自己在界面里操作——如实告诉 ta，不要重试。",
		}
	}
	return writeStrings{
		CreateHeader:    "Confirm post",
		CreateText:      "Post this Echo?",
		CreateYes:       "Post",
		UpdateHeader:    "Confirm edit",
		UpdateText:      "Update this Echo with the changes below?",
		UpdateYes:       "Update",
		DeleteHeader:    "Confirm delete",
		DeleteText:      "Delete this Echo?",
		DeleteYes:       "Delete",
		DeleteNote:      "This cannot be undone",
		Cancel:          "Cancel",
		FieldID:         "ID",
		FieldContent:    "Content",
		FieldNewContent: "New content",
		FieldTags:       "Tags",
		FieldVisibility: "Visibility",
		FieldPostedAt:   "Posted",
		Private:         "private",
		Public:          "public",
		NoTags:          "none",
		Arrow:           " -> ",
		Created:         "Posted. Echo ID: %s",
		Updated:         "Updated. Echo ID: %s",
		Deleted:         "Deleted. Echo ID: %s",
		NoChange:        "Nothing to update: give at least one of content, tags or private, and it must differ from the current value",
		BadEchoID:       "id is not a valid Echo ID. Search with search_echos first, then copy the full UUID after id= on the result you mean (e.g. 019ce0ea-82dd-774f-ae2d-5445512d42ad) — 【1】【2】 are result numbers, not IDs.",
		UnsupportedArg:  "This tool does not take that field. create_echo / update_echo can only write content, tags and private: images, attachments and extension cards (music/video/website/location/GitHub project) cannot be created or changed from a chat, and the user has to do it in the UI — tell them so instead of retrying.",
	}
}

func createEchoDescriptionFor(locale string) string {
	if localeIsZH(locale) {
		return `帮用户发布一条新的 Echo。

调用后会先把这条 Echo 完整地摆到用户面前请 ta 确认，确认了才真的发布——这一步是自动的，你不需要、也不应该再另外用 ask_user 问一遍。

务必遵守：
- content 用用户真正想发的内容。ta 让你「帮我写一条关于 X 的」时你可以代笔，但不要加“以下是为你生成的内容”这类外壳。
- tags 优先复用系统提示里列出的已有标签；用户没提标签就不要自己加。
- 用户明确说「发一条 / 记一下 / 帮我写并发出来」才调它。只是聊到某件事，不要擅自发布。
- 不支持图片和附件：除 content、tags、private 之外什么都发不了，扩展卡片（音乐/视频/网站/位置/GitHub 项目）同样不行。用户想发带图的 Echo，就告诉 ta 这需要在界面里发，不要假装发了。
- 工具返回「已发布」才算成功；返回「用户没有确认」就是没发，如实告诉 ta，不要重试。`
	}
	return `Post a new Echo for the user.

Calling this puts the whole Echo in front of the user for confirmation first, and it is only posted if they confirm. That step is automatic — you do not need, and must not add, a separate ask_user for permission.

Must follow:
- "content" is what the user actually wants posted. You may write it for them when they ask you to, but do not wrap it in "here is the content I generated".
- Prefer tags that already exist (they are listed in the context block); do not invent tags the user did not mention.
- Only call this when the user actually asks to post/record something. Merely discussing a topic is not a request to publish it.
- No attachments: content, tags and private are all this tool can set. Images, files and extension cards (music/video/website/location/GitHub project) cannot be posted from here. If the user wants an Echo with images, tell them it has to be done in the UI — do not pretend you posted them.
- Success is the tool returning "Posted". If it returns that the user did not confirm, nothing was posted — say so and do not retry.`
}

func updateEchoDescriptionFor(locale string) string {
	if localeIsZH(locale) {
		return `修改一条已有的 Echo。

调用后会把改动（原内容 → 新内容）摆到用户面前请 ta 确认，确认了才真的修改——这一步是自动的，不要再用 ask_user 问一遍。

务必遵守：
- id 必须照抄 search_echos 结果里那条的「id=」后面的完整 UUID。【1】【2】是结果编号，传它一定失败。不知道改哪一条就先检索，绝对不要猜 ID。
- 只传你要改的字段。content、tags、private 都是整体替换：传了 tags 就是把标签换成这一组，想保留原有标签就把它们一起写进来。
- 检索命中多条、不确定用户指的是哪一条时，先用 ask_user 让 ta 选，再来改。
- 不支持改图片、附件和扩展卡片：它们会被原样保留（改正文不会弄丢配图），但你无法增删改它们。用户要动这些，告诉 ta 在界面里改。
- 工具返回「已更新」才算成功；返回「用户没有确认」就是没改，如实告诉 ta，不要重试。`
	}
	return `Edit an existing Echo.

Calling this shows the user the change (old content -> new content) for confirmation first, and it is only applied if they confirm. That step is automatic — do not add a separate ask_user for permission.

Must follow:
- "id" is the full UUID after "id=" on the search_echos result you mean, copied verbatim. 【1】【2】 are result numbers and will always fail. If you do not know which Echo, search first; never guess an ID.
- Pass only the fields you are changing. content, tags and private each replace wholesale: passing tags sets the tag set to exactly that list, so include the existing tags if they should stay.
- If several Echos match and you are unsure which one the user means, use ask_user to have them pick before editing.
- Images, files and extension cards cannot be edited: they are carried over untouched (editing the text will not lose an Echo's images), but you cannot add, change or remove them. If the user wants that, tell them to do it in the UI.
- Success is the tool returning "Updated". If it returns that the user did not confirm, nothing changed — say so and do not retry.`
}

func deleteEchoDescriptionFor(locale string) string {
	if localeIsZH(locale) {
		return `删除一条 Echo。

调用后会把这条 Echo 完整地摆到用户面前请 ta 确认，确认了才真的删除——这一步是自动的，不要再用 ask_user 问一遍。

务必遵守：
- id 必须照抄 search_echos 结果里那条的「id=」后面的完整 UUID。【1】【2】是结果编号，传它一定失败。绝对不要猜 ID。
- 删除不可恢复。用户只有明确要求删除时才调它；说「这条写得不好」不等于要删。
- 检索命中多条、不确定是哪一条时，先用 ask_user 让 ta 选。
- 一次只删一条。用户要删多条时，一条一条来，每条都要 ta 单独确认。
- 工具返回「已删除」才算成功；返回「用户没有确认」就是没删，如实告诉 ta，不要重试。`
	}
	return `Delete an Echo.

Calling this puts the whole Echo in front of the user for confirmation first, and it is only deleted if they confirm. That step is automatic — do not add a separate ask_user for permission.

Must follow:
- "id" is the full UUID after "id=" on the search_echos result you mean, copied verbatim. 【1】【2】 are result numbers and will always fail. Never guess an ID.
- Deletion cannot be undone. Only call this when the user explicitly asks to delete something; "this one is badly written" is not a request to delete it.
- If several Echos match and you are unsure which one, use ask_user to have them pick.
- One Echo per call. To delete several, do them one at a time so each is confirmed on its own.
- Success is the tool returning "Deleted". If it returns that the user did not confirm, nothing was deleted — say so and do not retry.`
}

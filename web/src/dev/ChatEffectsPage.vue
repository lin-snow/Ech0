<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="effects">
    <header class="effects__bar">
      <div>
        <h1>Chat component effects</h1>
        <p>Dev-only preview of every animated piece of the Copilot surface.</p>
      </div>
      <div class="effects__themes">
        <button
          v-for="theme in THEMES"
          :key="theme"
          class="effects__theme"
          :class="{ 'effects__theme--on': currentTheme === theme }"
          @click="setTheme(theme)"
        >
          {{ theme }}
        </button>
      </div>
    </header>

    <section class="panel">
      <h2>Inline status row</h2>
      <ChatActivity label="Thinking" :active="true" clock="1.4s" />
      <ChatActivity
        label="Searching echos"
        :active="true"
        detail="“mood this month”"
        clock="2.1s"
      />
      <ChatActivity
        label="Thought"
        :active="false"
        clock="3.0s"
        detail="A short note about what was read"
      >
        <template #glyph>
          <ChatDoneGlyph />
        </template>
        <p class="body-copy">
          Collapsible body: open it once and it stays open until clicked shut.
        </p>
      </ChatActivity>
      <div class="controls">
        <button class="ctl" @click="swapLabel = !swapLabel">swap label</button>
        <span class="note">Watch the clock slide, not jump.</span>
      </div>
      <ChatActivity
        :label="swapLabel ? 'Searched 3 time(s)' : 'Searching echos'"
        :active="true"
        :clock="swapLabel ? '2.4s' : '0.8s'"
      />
    </section>

    <section class="panel">
      <h2>Waiting &amp; reasoning</h2>
      <ChatWaiting />
      <ChatReasoning
        :text="'Checking the last few weeks for a thread: the harbour walk, the bug notes, and the slow mornings.'"
        :active="true"
      />
      <ChatReasoning
        :text="'The thread was mostly about slowing down after the move.'"
        :active="false"
        :duration-ms="4200"
      />
    </section>

    <section class="panel">
      <h2>Retrieval pill</h2>
      <ChatRetrieval
        :searches="['mood this month', 'recent focus']"
        :coverage="{ total: 42, returned: 12, buckets: 4, truncated: false }"
        :sources="sources"
        :active="true"
      />
      <ChatRetrieval
        :searches="['mood this month', 'recent focus']"
        :coverage="{ total: 42, returned: 12, buckets: 4, truncated: false }"
        :sources="sources"
        :active="false"
      />
      <ChatRetrieval
        :searches="[]"
        :coverage="{ total: 400, returned: 120, buckets: 8, truncated: true }"
        :active="false"
      />
    </section>

    <section class="panel">
      <h2>Sources</h2>
      <ChatSources :sources="rowSources" @open="openEcho" />
      <div class="controls">
        <button class="ctl" @click="wideSources = !wideSources">toggle 2 / 8 sources</button>
      </div>
      <ChatSources :sources="wideSources ? manySources : rowSources" @open="openEcho" />
      <p v-if="lastOpened" class="note">opened: {{ lastOpened }}</p>
    </section>

    <section class="panel">
      <h2>Ask picker (approval card)</h2>
      <ChatAskPicker :ask="singleAsk" :index="0" @answer="onAnswer" @back="() => {}" />
      <ChatAskPicker :ask="multiAsk" :index="0" @answer="onAnswer" @back="() => {}" />
      <ChatAskPicker
        :ask="flowAsk"
        :index="flowIndex"
        @answer="onFlowAnswer"
        @back="flowIndex = Math.max(0, flowIndex - 1)"
      />
      <ChatAskExchange :exchange="settledExchange" />
      <p v-if="lastAnswer" class="note">
        Last answer: {{ lastAnswer.question_id }} → {{ (lastAnswer.selected ?? []).join(' · ') }}
      </p>
    </section>

    <section class="panel">
      <h2>Actions &amp; streaming answer</h2>
      <ChatActions text="A demo answer" :can-retry="true" @retry="replay" />
      <div class="answer">
        <AnimatedMarkdown
          :content="streamed"
          :streaming="streaming"
          animation="blurIn"
          @update:revealing="revealing = $event"
        />
      </div>
      <div class="controls">
        <button class="ctl" @click="replay">
          {{ streaming ? 'streaming…' : 'replay stream' }}
        </button>
        <span class="note">revealing: {{ revealing }}</span>
      </div>
    </section>

    <section class="panel">
      <h2>Combined turn</h2>
      <div class="turn">
        <p class="bubble">How has my mood changed over this period?</p>
        <div class="trace">
          <ChatRetrieval
            :searches="['mood this month', 'recent focus']"
            :coverage="{ total: 42, returned: 12, buckets: 4, truncated: false }"
            :sources="sources"
            :active="false"
          />
        </div>
        <div class="answer">
          <TheMdPreview
            :content="'Mostly **slower mornings** and longer walks. The move shows up as a quieter week, then a return to the old rhythm.'"
          />
        </div>
        <ChatSources :sources="rowSources" @open="openEcho" />
        <ChatActions text="Combined turn answer" :can-retry="true" @retry="replay" />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { TheMdPreview } from '@/components/advanced/md'
import ChatActivity from '@/views/chat/modules/ChatActivity.vue'
import ChatActions from '@/views/chat/modules/ChatActions.vue'
import ChatAskExchange from '@/views/chat/modules/ChatAskExchange.vue'
import ChatAskPicker from '@/views/chat/modules/ChatAskPicker.vue'
import ChatReasoning from '@/views/chat/modules/ChatReasoning.vue'
import ChatRetrieval from '@/views/chat/modules/ChatRetrieval.vue'
import ChatSources from '@/views/chat/modules/ChatSources.vue'
import ChatDoneGlyph from '@/views/chat/modules/ChatDoneGlyph.vue'
import ChatWaiting from '@/views/chat/modules/ChatWaiting.vue'
import AnimatedMarkdown from '@/views/chat/modules/AnimatedMarkdown.vue'

const THEMES = ['light', 'dark', 'sunny'] as const
const currentTheme = ref<(typeof THEMES)[number]>('light')

const setTheme = (theme: (typeof THEMES)[number]) => {
  currentTheme.value = theme
  document.documentElement.className = theme
}

const swapLabel = ref(false)
const wideSources = ref(false)

const source = (
  id: string,
  content: string,
  created: number,
  ext?: string,
): App.Api.Chat.ChatSource => ({
  echo_id: id,
  content,
  username: 'lin-snow',
  echo_created: created,
  distance: 0.42,
  files: [],
  ...(ext
    ? { extension: { type: ext, payload: {} } as unknown as App.Api.Ech0.EchoExtension }
    : {}),
})

const sources = [
  source('e1', 'Walked the old harbour road at dusk, thinking about the move.', 1755000000),
  source(
    'e2',
    'Notes on the stubborn bug: the cache invalidated before the write landed.',
    1755100000,
  ),
  source('e3', 'A quiet week, mostly reading and sketching the new layout.', 1755200000),
]

const rowSources = [
  ...sources,
  source('e4', 'A quiet week, mostly reading and sketching the new layout.', 1755300000, 'WEBSITE'),
  source('e5', '', 1755400000, 'LOCATION'),
]

const manySources = Array.from({ length: 8 }, (_, i) =>
  source(
    `m${i}`,
    `Sample echo ${i + 1}: a short note from the timeline used to test the card grid at length.`,
    1755000000 + i * 86400,
    i % 3 === 0 ? 'WEBSITE' : undefined,
  ),
)

const openEcho = (id: string) => {
  lastOpened.value = id
}
const lastOpened = ref<string>('')

const singleAsk: App.Api.Chat.ChatAsk = {
  ask_id: 'a1',
  questions: [
    {
      id: 'q1',
      header: 'Pick a range',
      text: 'Which period should the summary cover?',
      options: [
        { label: 'Last 7 days', description: 'Short and recent' },
        { label: 'Last 30 days', description: 'A month of context' },
        { label: 'Everything', description: 'Slow, but complete' },
      ],
      recommended: 1,
    },
  ],
}

const multiAsk: App.Api.Chat.ChatAsk = {
  ask_id: 'a2',
  questions: [
    {
      id: 'q2',
      header: 'What to include',
      text: 'Which threads matter to you?',
      multi: true,
      options: [{ label: 'Work' }, { label: 'Reading' }, { label: 'Travel' }],
    },
  ],
}

const flowAsk: App.Api.Chat.ChatAsk = {
  ask_id: 'a3',
  questions: [
    {
      id: 'q5',
      header: 'Range',
      text: 'Which period should the summary cover?',
      options: [
        { label: 'Last 7 days', description: 'Short and recent' },
        { label: 'Last 30 days', description: 'A month of context' },
      ],
      recommended: 1,
    },
    {
      id: 'q6',
      header: 'Voice',
      text: 'How should it read?',
      options: [
        { label: 'Warm', description: 'Like a letter to yourself' },
        { label: 'Plain', description: 'Just the facts' },
      ],
    },
    {
      id: 'q7',
      header: 'Focus',
      text: 'Anything to leave out?',
      options: [
        { label: 'Work', description: 'Keep it personal' },
        { label: 'Nothing', description: 'Include everything' },
      ],
    },
  ],
}

const settledExchange: App.Api.Chat.ChatAskExchange = {
  questions: [
    { id: 'q3', header: 'Range', text: 'Summary range', options: [{ label: 'Last 30 days' }] },
    { id: 'q4', header: 'Voice', text: 'Tone', options: [{ label: 'Warm' }] },
  ],
  answers: [
    { question_id: 'q3', selected: ['Last 30 days'] },
    { question_id: 'q4', selected: ['Warm'] },
  ],
}

const lastAnswer = ref<App.Api.Chat.ChatAskAnswer | null>(null)
const onAnswer = (answer: App.Api.Chat.ChatAskAnswer) => {
  lastAnswer.value = answer
}

const flowIndex = ref(0)
const onFlowAnswer = (answer: App.Api.Chat.ChatAskAnswer) => {
  lastAnswer.value = answer
  if (flowIndex.value < flowAsk.questions.length - 1) flowIndex.value += 1
  else flowIndex.value = 0
}

const SAMPLE_ANSWER = `## What changed

Mostly **slower mornings** and longer walks. The move shows up as a quiet week, then a return to the old rhythm.

- the harbour walk kept the week anchored
- the bug notes were the only sharp edge
- reading came back on the weekend`

const streamed = ref(SAMPLE_ANSWER)
const streaming = ref(false)
const revealing = ref(false)

let streamTimer = 0

const replay = () => {
  if (streamTimer) clearInterval(streamTimer)
  streamed.value = ''
  streaming.value = true
  let cursor = 0
  streamTimer = window.setInterval(() => {
    cursor = Math.min(SAMPLE_ANSWER.length, cursor + 4)
    streamed.value = SAMPLE_ANSWER.slice(0, cursor)
    if (cursor >= SAMPLE_ANSWER.length) {
      clearInterval(streamTimer)
      streamTimer = 0
      streaming.value = false
    }
  }, 45)
}

onBeforeUnmount(() => {
  if (streamTimer) clearInterval(streamTimer)
})

setTheme(currentTheme.value)
</script>

<style scoped>
.effects {
  width: min(44rem, calc(100% - 2.5rem));
  margin: 2rem auto 6rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  color: var(--color-text-primary);
  font-family: var(--font-family-sans);
}

.effects__bar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.effects__bar h1 {
  margin: 0 0 0.25rem;
  font-size: 1.15rem;
}

.effects__bar p {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.78rem;
}

.effects__themes {
  display: flex;
  gap: 0.3rem;
}

.effects__theme {
  border: 1px solid var(--color-border-strong);
  border-radius: 999px;
  background: transparent;
  color: var(--color-text-muted);
  padding: 0.25rem 0.65rem;
  font-size: 0.7rem;
  cursor: pointer;
}

.effects__theme--on {
  border-color: var(--color-accent);
  background: var(--color-accent-soft);
  color: var(--color-text-primary);
}

.panel {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 0.95rem 1rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.9rem;
  background: var(--color-bg-surface);
}

.panel h2 {
  margin: 0;
  font-size: 0.72rem;
  font-weight: 500;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.body-copy {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.82rem;
}

.controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.ctl {
  border: 1px solid var(--color-border-strong);
  border-radius: 999px;
  background: var(--color-bg-muted);
  color: var(--color-text-secondary);
  padding: 0.25rem 0.7rem;
  font-size: 0.7rem;
  cursor: pointer;
}

.note {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.7rem;
}

.turn {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.bubble {
  align-self: flex-end;
  max-width: 85%;
  margin: 0;
  padding: 0.55rem 0.9rem;
  border-radius: 1.1rem 1.1rem 0.25rem;
  background: var(--color-accent-soft);
  font-size: 0.9rem;
  line-height: 1.6;
}

.trace {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.05rem;
}

.answer {
  font-size: 1rem;
}

.answer :deep(.echo-markdown) {
  line-height: 1.8;
}
</style>

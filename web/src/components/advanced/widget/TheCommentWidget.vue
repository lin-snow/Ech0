<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <WidgetShell title="Comment" :fluid="fluid">
    <template #icon>
      <span class="comment-doodle" aria-hidden="true">
        <span class="doodle-smile">
          <span class="smile-eye smile-eye-left"></span>
          <span class="smile-eye smile-eye-right"></span>
          <span class="smile-mouth"></span>
        </span>
      </span>
    </template>
    <div class="comment-teaser-body">
      <div class="comment-teaser-card">
        <button
          v-if="canJumpToEchoDetail"
          type="button"
          class="comment-jump-btn"
          v-tooltip="t('connectWidget.jumpToEchoDetail')"
          :aria-label="t('connectWidget.jumpToEchoDetail')"
          @click="handleJumpToEchoDetail"
        >
          <LinkTo class="w-4 h-4" />
        </button>
        <p v-if="commentLoading" class="comment-teaser-content">
          {{ t('connectWidget.randomCommentLoading') }}
        </p>
        <p v-else class="comment-teaser-content">{{ randomCommentContent }}</p>
      </div>
    </div>
  </WidgetShell>
</template>

<script setup lang="ts">
import LinkTo from '@/components/icons/linkto.vue'
import WidgetShell from './WidgetShell.vue'
import { fetchGetPublicComments } from '@/service/api'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

defineProps<{ fluid?: boolean }>()

const router = useRouter()
const { t } = useI18n()
const randomComment = ref<App.Api.Comment.CommentItem | null>(null)
const commentLoading = ref(false)
const canJumpToEchoDetail = computed(
  () => Boolean(randomComment.value?.echo_id) && !commentLoading.value,
)

const randomCommentContent = computed(() => {
  const content = randomComment.value?.content?.trim()
  if (!content) return String(t('connectWidget.noFeaturedComment'))
  return content.replace(/\s+/g, ' ').slice(0, 120)
})

const pickRandomComment = (items: App.Api.Comment.CommentItem[]) => {
  const hotComments = items.filter(
    (item) => item.status === 'approved' && item.hot && item.content?.trim(),
  )
  if (hotComments.length > 0) {
    randomComment.value = hotComments[Math.floor(Math.random() * hotComments.length)]
    return
  }

  const approvedComments = items.filter(
    (item) => item.status === 'approved' && item.content?.trim(),
  )
  if (approvedComments.length > 0) {
    randomComment.value = approvedComments[Math.floor(Math.random() * approvedComments.length)]
  }
}

const loadRandomComment = async () => {
  commentLoading.value = true
  try {
    const res = await fetchGetPublicComments(10)
    if (res.code === 1 && res.data?.length) {
      pickRandomComment(res.data)
    }
  } catch {
    randomComment.value = null
  } finally {
    commentLoading.value = false
  }
}

const handleJumpToEchoDetail = () => {
  const echoId = randomComment.value?.echo_id?.trim()
  if (!echoId) return
  router.push({
    name: 'echo',
    params: { echoId },
  })
}

onMounted(() => {
  void loadRandomComment()
})
</script>

<style scoped>
.comment-doodle {
  position: relative;
  display: block;
  width: 1.125rem;
  height: 1.125rem;
}

.comment-doodle .doodle-smile {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%) rotate(-5deg) scale(0.55);
}

.doodle-smile {
  position: relative;
  width: 34px;
  height: 34px;
  border-radius: 9999px;
  background: #f2e68f;
  border: 1.5px solid #3f362d;
  transform: rotate(-5deg);
}

.smile-eye {
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 9999px;
  background: #3f362d;
  top: 13px;
}

.smile-eye-left {
  left: 10px;
}

.smile-eye-right {
  left: 19px;
}

.smile-mouth {
  position: absolute;
  left: 11px;
  top: 18px;
  width: 10px;
  height: 6px;
  border-bottom: 1.5px solid #3f362d;
  border-radius: 0 0 10px 10px;
}

.comment-teaser-body {
  width: 100%;
}

.comment-teaser-card {
  position: relative;
  width: 100%;
  border: 1px solid var(--color-border-subtle);
  background-color: var(--comment-widget-card-bg);
  box-shadow: 0 8px 18px rgb(20 20 20 / 4%);
  padding: 14px 12px 12px;
  transform: rotate(-1.1deg);
  transform-origin: top center;
  transition: transform 220ms ease;
}

.comment-teaser-card:hover {
  transform: rotate(-0.4deg);
}

.comment-jump-btn {
  position: absolute;
  right: 8px;
  top: 8px;
  width: 26px;
  height: 26px;
  border-radius: 9999px;
  border: 1px solid var(--comment-widget-jump-border);
  background: var(--comment-widget-jump-bg);
  color: var(--color-text-secondary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transform: translateY(-1px) scale(0.96);
  pointer-events: none;
  transition:
    opacity 180ms ease,
    transform 180ms ease,
    color 180ms ease,
    border-color 180ms ease;
}

.comment-teaser-card:hover .comment-jump-btn,
.comment-jump-btn:focus-visible {
  opacity: 1;
  transform: translateY(0) scale(1);
  pointer-events: auto;
}

.comment-jump-btn:hover {
  color: var(--color-text-primary);
  border-color: var(--color-accent);
}

.comment-teaser-card::before {
  content: '';
  position: absolute;
  left: 50%;
  top: -7px;
  transform: translateX(-50%) rotate(-1.5deg);
  width: 42px;
  height: 12px;
  border-radius: 2px;
  background: var(--comment-widget-card-before-bg);
  box-shadow:
    0 1px 0 rgb(255 255 255 / 30%) inset,
    0 1px 2px rgb(0 0 0 / 8%);
  opacity: 0.95;
}

.comment-teaser-content {
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.65;
  white-space: normal;
  overflow-wrap: anywhere;
}
</style>

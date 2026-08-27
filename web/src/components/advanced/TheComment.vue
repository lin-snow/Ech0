<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div id="comments" class="w-full max-w-sm h-auto px-0 py-2 mx-auto">
    <div
      v-if="commentsClosed"
      class="rounded-lg border border-[var(--color-border-subtle)] p-3 text-sm text-[var(--color-text-muted)]"
      :class="{ 'mb-3': readOnly }"
    >
      {{ t('commentSection.disabled') }}
    </div>

    <template v-if="!commentsClosed || readOnly">
      <div class="mb-4 comment-list-board">
        <div class="mb-3 flex items-center justify-between gap-2">
          <button
            v-if="!readOnly"
            type="button"
            class="comment-pill-btn shrink-0"
            :aria-expanded="commentFormExpanded"
            @click="commentFormExpanded = !commentFormExpanded"
          >
            <Comments class="comment-pill-btn__icon" aria-hidden="true" />
            <span>{{ t('commentSection.publishComment') }}</span>
          </button>
          <span class="comment-count-text">
            {{ t('commentSection.commentCount', { count: comments.length }) }}
          </span>
        </div>

        <form
          v-if="commentFormExpanded && !readOnly"
          class="comment-form-panel mb-3"
          @submit.prevent="submitComment"
        >
          <div class="comment-form-head">
            <div class="flex items-center gap-1.5">
              <h3 class="comment-form-title">{{ t('commentSection.publishComment') }}</h3>
              <span class="comment-md-hint" v-tooltip="t('commentSection.markdownSupported')">
                <MarkdownIcon class="h-3 w-3" />
              </span>
            </div>
            <span class="comment-ready-indicator">
              <i
                class="comment-ready-dot"
                :class="profileReady ? 'is-ready' : 'is-not-ready'"
                aria-hidden="true"
              ></i>
            </span>
          </div>

          <div v-if="replyTarget" class="comment-reply-banner">
            <span class="truncate">
              {{ t('commentSection.replyingTo', { nickname: replyTarget.nickname }) }}
            </span>
            <button
              type="button"
              class="comment-reply-cancel"
              :aria-label="t('commentSection.cancelReply')"
              @click="cancelReply"
            >
              ✕
            </button>
          </div>

          <div v-if="!isPrivilegedUser" class="comment-id-grid">
            <input
              v-model.trim="form.nickname"
              type="text"
              class="comment-input-field comment-input-sm"
              :placeholder="t('commentSection.nicknameRequired')"
            />
            <input
              v-model.trim="form.email"
              type="email"
              class="comment-input-field comment-input-sm"
              :placeholder="t('commentSection.emailRequired')"
            />
            <input
              v-model.trim="form.website"
              type="url"
              class="comment-input-field comment-input-sm comment-id-grid__full"
              :placeholder="t('commentSection.websiteOptional')"
            />
          </div>

          <textarea
            v-model.trim="form.content"
            class="comment-input-field comment-textarea"
            :placeholder="t('commentSection.commentPlaceholder')"
            maxlength="200"
          />

          <div v-if="form.content" class="comment-preview-strip">
            <div class="comment-preview-strip__head">
              <BaseAvatar
                :seed="previewAvatarSeed"
                :size="18"
                alt="avatar"
                class="h-[18px] w-[18px] shrink-0 rounded-full object-cover"
              />
              <span class="comment-preview-strip__name">{{ previewNickname }}</span>
              <span class="comment-preview-strip__label">{{
                t('commentSection.livePreview')
              }}</span>
            </div>
            <TheMdPreview class="comment-md-content" :content="form.content" />
          </div>

          <input
            v-model="form.hp_field"
            type="text"
            tabindex="-1"
            autocomplete="off"
            class="hidden"
          />

          <div v-if="needCaptcha" class="comment-captcha-wrap">
            <div ref="captchaMountRef" class="comment-captcha-mount"></div>
            <p v-if="captchaError" class="comment-captcha-error">{{ captchaError }}</p>
          </div>

          <div class="comment-form-foot">
            <span class="comment-char-count" :class="{ 'is-over': contentTooLong }">
              {{ contentLength }}/200
            </span>
            <button
              v-if="showSubmitButton"
              type="submit"
              class="comment-submit-btn"
              :disabled="submitting || !canSubmit"
            >
              {{ submitting ? t('commentSection.submitting') : t('commentSection.submitComment') }}
            </button>
          </div>

          <div v-if="submitNotice" class="comment-notice">
            <div class="comment-notice__title">
              {{
                submitNotice.status === 'approved'
                  ? t('commentSection.commentPublished')
                  : t('commentSection.commentSubmittedPending')
              }}
            </div>
            <div class="comment-notice__time">{{ submitNoticeText }}</div>
            <div v-if="submitNotice.contentPreview" class="comment-notice__preview">
              {{ submitNotice.contentPreview }}
            </div>
          </div>
        </form>

        <div v-if="loading" class="text-sm text-[var(--color-text-muted)]">
          {{ t('commentSection.loading') }}
        </div>
        <div
          v-else-if="comments.length === 0 && !commentFormExpanded"
          class="text-sm text-[var(--color-text-muted)]"
        >
          {{ t('commentSection.empty') }}
        </div>
        <div v-else class="comment-thread">
          <article
            v-for="item in topLevelComments"
            :id="commentAnchorId(item.id)"
            :key="item.id"
            class="comment-card"
            :class="{ 'comment-anchor-flash': highlightedId === item.id }"
          >
            <span v-if="item.hot" class="comment-hot-badge">Hot</span>
            <div class="comment-row">
              <BaseAvatar
                :seed="getCommentAvatarSeed(item)"
                :size="28"
                alt="avatar"
                class="h-7 w-7 shrink-0 rounded-full object-cover"
              />
              <div class="min-w-0 flex-1">
                <div class="comment-meta">
                  <span class="comment-floor-no">#{{ numberOf(item) }}</span>
                  <a
                    v-if="item.website"
                    :href="item.website"
                    target="_blank"
                    rel="noreferrer"
                    class="comment-author-link"
                  >
                    {{ item.nickname }}
                  </a>
                  <span v-else class="comment-author">{{ item.nickname }}</span>
                  <Verified
                    v-if="item.source === 'system'"
                    v-tooltip="t('commentSection.verifiedUser')"
                    class="verified-badge-icon h-3! w-3! shrink-0 text-sky-500"
                  />
                  <span class="comment-dot">·</span>
                  <span class="comment-time">{{ formatDate(item.created_at) }}</span>
                </div>
                <TheMdPreview class="comment-md-content" :content="item.content" />
                <button
                  v-if="!readOnly"
                  type="button"
                  class="comment-reply-btn"
                  @click="startReply(item)"
                >
                  {{ t('commentSection.reply') }}
                </button>
              </div>
            </div>

            <div v-if="repliesOf(item.id).length" class="comment-replies">
              <div
                v-for="reply in repliesOf(item.id)"
                :id="commentAnchorId(reply.id)"
                :key="reply.id"
                class="comment-row comment-reply-row"
                :class="{ 'comment-anchor-flash': highlightedId === reply.id }"
              >
                <BaseAvatar
                  :seed="getCommentAvatarSeed(reply)"
                  :size="22"
                  alt="avatar"
                  class="h-[22px] w-[22px] shrink-0 rounded-full object-cover"
                />
                <div class="min-w-0 flex-1">
                  <div class="comment-meta">
                    <span class="comment-floor-no">#{{ numberOf(reply) }}</span>
                    <a
                      v-if="reply.website"
                      :href="reply.website"
                      target="_blank"
                      rel="noreferrer"
                      class="comment-author-link"
                    >
                      {{ reply.nickname }}
                    </a>
                    <span v-else class="comment-author">{{ reply.nickname }}</span>
                    <Verified
                      v-if="reply.source === 'system'"
                      v-tooltip="t('commentSection.verifiedUser')"
                      class="verified-badge-icon h-3! w-3! shrink-0 text-sky-500"
                    />
                    <span class="comment-dot">·</span>
                    <span class="comment-time">{{ formatDate(reply.created_at) }}</span>
                    <span v-if="reply.hot" class="comment-hot-inline">Hot</span>
                    <template v-if="parentNumberOf(reply)">
                      <span class="comment-dot">·</span>
                      <button
                        type="button"
                        class="comment-reply-ref"
                        v-tooltip="
                          t('commentSection.inReplyTo', { nickname: parentNicknameOf(reply) })
                        "
                        @click="jumpToComment(reply.parent_id)"
                      >
                        {{ t('commentSection.inReplyToFloor', { floor: parentNumberOf(reply) }) }}
                      </button>
                    </template>
                  </div>
                  <TheMdPreview class="comment-md-content" :content="reply.content" />
                  <button
                    v-if="!readOnly"
                    type="button"
                    class="comment-reply-btn"
                    @click="startReply(reply)"
                  >
                    {{ t('commentSection.reply') }}
                  </button>
                </div>
              </div>
            </div>
          </article>
        </div>
      </div>
    </template>
  </div>
</template>

<script lang="ts">
let capRejectionGuardInstalled = false
const installCapRejectionGuard = () => {
  if (capRejectionGuardInstalled || typeof window === 'undefined') return
  capRejectionGuardInstalled = true
  window.addEventListener('unhandledrejection', (event) => {
    const reason = event.reason
    if (reason instanceof Error && reason.message.includes('_ensureSize')) {
      event.preventDefault()
    }
  })
}
</script>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchCreateComment, fetchGetCommentFormMeta, fetchGetComments } from '@/service/api'
import { useUserStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { formatDate } from '@/utils/other'
import BaseAvatar from '@/components/common/BaseAvatar.vue'
import { TheMdPreview } from '@/components/advanced/md'
import Verified from '../icons/verified.vue'
import MarkdownIcon from '../icons/markdown.vue'
import Comments from '../icons/comments.vue'
import { useI18n } from 'vue-i18n'
import { isStaticMode } from '@/service/request/shared'

type CapSolveDetail = {
  token?: string
}

type CapErrorDetail = {
  error?: string
  message?: string
}

type CapWidgetElement = HTMLElement & {
  solve?: () => Promise<CapSolveDetail>
}

type SubmitNotice = {
  status: App.Api.Comment.CommentStatus
  contentPreview: string
  submittedAt: number
}

const route = useRoute()
const userStore = useUserStore()
const { t, locale } = useI18n()
const loading = ref(false)
const submitting = ref(false)
const comments = ref<App.Api.Comment.CommentItem[]>([])
const formMeta = ref<App.Api.Comment.FormMeta | null>(null)
const captchaMountRef = ref<HTMLElement | null>(null)
const captchaWidget = ref<CapWidgetElement | null>(null)
const captchaError = ref('')
const solvingCaptcha = ref(false)
const commentFormExpanded = ref(false)
const submitNotice = ref<SubmitNotice | null>(null)
let capWidgetLoadPromise: Promise<unknown> | null = null

const readOnly = isStaticMode()
const commentsClosed = computed(() => !!formMeta.value && !formMeta.value.enable_comment)

const form = reactive<App.Api.Comment.CreateCommentDto>({
  echo_id: '',
  parent_id: '',
  nickname: '',
  email: '',
  website: '',
  content: '',
  hp_field: '',
  form_token: '',
  captcha_token: '',
})

const replyTarget = ref<App.Api.Comment.CommentItem | null>(null)

const commentMap = computed(() => {
  const map = new Map<string, App.Api.Comment.CommentItem>()
  for (const c of comments.value) map.set(c.id, c)
  return map
})

const rootIdOf = (item: App.Api.Comment.CommentItem) => {
  const map = commentMap.value
  let cur = item
  const seen = new Set<string>()
  while (cur.parent_id && map.has(cur.parent_id) && !seen.has(cur.id)) {
    seen.add(cur.id)
    cur = map.get(cur.parent_id)!
  }
  return cur.id
}

const topLevelComments = computed(() =>
  comments.value.filter((c) => !c.parent_id || !commentMap.value.has(c.parent_id)),
)

const repliesOf = (rootId: string) =>
  comments.value.filter(
    (c) => c.parent_id && commentMap.value.has(c.parent_id) && rootIdOf(c) === rootId,
  )

const commentNumbers = computed(() => {
  const ordered = [...comments.value].sort((a, b) => {
    if (a.created_at !== b.created_at) return a.created_at - b.created_at
    return a.id < b.id ? -1 : a.id > b.id ? 1 : 0
  })
  const map = new Map<string, number>()
  ordered.forEach((c, index) => map.set(c.id, index + 1))
  return map
})

const numberOf = (item: App.Api.Comment.CommentItem) => commentNumbers.value.get(item.id) ?? 0

const parentNumberOf = (reply: App.Api.Comment.CommentItem) =>
  reply.parent_id ? (commentNumbers.value.get(reply.parent_id) ?? 0) : 0

const parentNicknameOf = (reply: App.Api.Comment.CommentItem) =>
  reply.parent_id ? (commentMap.value.get(reply.parent_id)?.nickname ?? '') : ''

const commentAnchorId = (id: string) => `comment-anchor-${id}`
const highlightedId = ref<string | null>(null)
let highlightTimer: ReturnType<typeof setTimeout> | null = null

const jumpToComment = async (id?: string | null) => {
  if (!id) return
  const el = document.getElementById(commentAnchorId(id))
  if (!el) return
  el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  if (highlightTimer) clearTimeout(highlightTimer)
  highlightedId.value = null
  await nextTick()
  highlightedId.value = id
  highlightTimer = setTimeout(() => {
    highlightedId.value = null
    highlightTimer = null
  }, 1600)
}

const startReply = (item: App.Api.Comment.CommentItem) => {
  replyTarget.value = item
  form.parent_id = item.id
  commentFormExpanded.value = true
  nextTick(() => {
    document.getElementById('comments')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  })
}

const cancelReply = () => {
  replyTarget.value = null
  form.parent_id = ''
}

const isPrivilegedUser = computed(() => {
  const user = userStore.user
  return Boolean(user && (user.is_admin || user.is_owner))
})

const canSubmit = computed(() => {
  if (!form.echo_id || !form.content) return false
  if (contentTooLong.value) return false
  if (!form.form_token) return false
  if (!isPrivilegedUser.value && (!form.nickname || !form.email)) return false
  return true
})

const needCaptcha = computed(() =>
  Boolean(formMeta.value?.captcha_enabled && formMeta.value?.captcha_api_endpoint),
)
const showSubmitButton = computed(() => !needCaptcha.value || Boolean(form.captcha_token))

const contentLength = computed(() => form.content.length)
const contentTooLong = computed(() => contentLength.value > 200)

const profileReady = computed(() => {
  if (isPrivilegedUser.value) return true
  return Boolean(form.nickname && form.email)
})

const previewNickname = computed(() => {
  if (isPrivilegedUser.value) return userStore.user?.username || String(t('commentSection.you'))
  return form.nickname || String(t('commentSection.you'))
})

const previewAvatarSeed = computed(() => {
  if (isPrivilegedUser.value) {
    return `${previewNickname.value}-system-preview`
  }
  return `${form.nickname || 'guest'}-${form.email || 'preview'}`
})

const getCommentAvatarSeed = (item: App.Api.Comment.CommentItem) => {
  return `${item.id}-${item.nickname}-${item.source}`
}

const submitNoticeText = computed(() => {
  if (!submitNotice.value) return ''
  const timeLabel = formatDate(submitNotice.value.submittedAt)
  if (submitNotice.value.status === 'approved') {
    return String(t('commentSection.publishedAt', { time: timeLabel }))
  }
  return String(t('commentSection.submittedAtPending', { time: timeLabel }))
})

const clearCaptchaWidget = () => {
  if (captchaWidget.value) {
    captchaWidget.value.remove()
    captchaWidget.value = null
  }
}

const ensureCapWidgetScript = async () => {
  if (typeof window === 'undefined') return
  installCapRejectionGuard()
  if (customElements.get('cap-widget')) return
  capWidgetLoadPromise ??= import('@cap.js/widget')
  await capWidgetLoadPromise
}

const mountCaptchaWidget = async () => {
  clearCaptchaWidget()
  if (!needCaptcha.value || !captchaMountRef.value) return
  try {
    await ensureCapWidgetScript()
  } catch {
    captchaError.value = String(t('commentSection.captchaLoadFailed'))
    return
  }
  captchaError.value = ''
  const widget = document.createElement('cap-widget') as CapWidgetElement
  widget.setAttribute('data-cap-api-endpoint', formMeta.value?.captcha_api_endpoint || '')
  const i18nAttrs: Record<string, string> = {
    'data-cap-i18n-initial-state': String(t('commentSection.capInitialState')),
    'data-cap-i18n-verifying-label': String(t('commentSection.capVerifyingLabel')),
    'data-cap-i18n-solved-label': String(t('commentSection.capSolvedLabel')),
    'data-cap-i18n-error-label': String(t('commentSection.capErrorLabel')),
    'data-cap-i18n-troubleshooting-label': String(t('commentSection.capTroubleshootingLabel')),
    'data-cap-i18n-wasm-disabled': String(t('commentSection.capWasmDisabled')),
    'data-cap-i18n-verify-aria-label': String(t('commentSection.capVerifyAriaLabel')),
    'data-cap-i18n-verifying-aria-label': String(t('commentSection.capVerifyingAriaLabel')),
    'data-cap-i18n-verified-aria-label': String(t('commentSection.capVerifiedAriaLabel')),
    'data-cap-i18n-error-aria-label': String(t('commentSection.capErrorAriaLabel')),
  }
  Object.entries(i18nAttrs).forEach(([key, value]) => widget.setAttribute(key, value))
  widget.addEventListener('solve', (event) => {
    const token = (event as CustomEvent<CapSolveDetail>).detail?.token
    form.captcha_token = token || ''
    if (!token) {
      captchaError.value = String(t('commentSection.captchaVerifyFailed'))
      return
    }
    captchaError.value = ''
  })
  widget.addEventListener('error', (event: Event) => {
    const detail = 'detail' in event ? (event as CustomEvent<CapErrorDetail>).detail : undefined
    const message = detail?.message || detail?.error
    form.captcha_token = ''
    captchaError.value = message || String(t('commentSection.captchaWidgetError'))
  })
  captchaMountRef.value.appendChild(widget)
  captchaWidget.value = widget
}

const ensureCaptchaToken = async () => {
  if (!needCaptcha.value) return true
  if (form.captcha_token) return true
  if (!captchaWidget.value) {
    captchaError.value = String(t('commentSection.captchaLoading'))
    return false
  }
  if (!captchaWidget.value.solve) {
    captchaError.value = String(t('commentSection.completeCaptchaFirst'))
    return false
  }
  solvingCaptcha.value = true
  captchaError.value = ''
  try {
    const result = await captchaWidget.value.solve()
    const token = result?.token || ''
    form.captcha_token = token
    if (!token) {
      captchaError.value = String(t('commentSection.completeCaptchaFirst'))
      return false
    }
    return true
  } catch {
    captchaError.value = String(t('commentSection.completeCaptchaFirst'))
    return false
  } finally {
    solvingCaptcha.value = false
  }
}

const GUEST_PROFILE_KEY = 'ech0:comment-guest-profile'

const loadGuestProfile = () => {
  if (isPrivilegedUser.value) return
  try {
    const raw = localStorage.getItem(GUEST_PROFILE_KEY)
    if (!raw) return
    const saved = JSON.parse(raw) as Partial<
      Pick<App.Api.Comment.CreateCommentDto, 'nickname' | 'email' | 'website'>
    >
    if (saved.nickname) form.nickname = saved.nickname
    if (saved.email) form.email = saved.email
    if (saved.website) form.website = saved.website
  } catch {}
}

const saveGuestProfile = () => {
  if (isPrivilegedUser.value) return
  try {
    localStorage.setItem(
      GUEST_PROFILE_KEY,
      JSON.stringify({
        nickname: form.nickname,
        email: form.email,
        website: form.website,
      }),
    )
  } catch {}
}

const loadData = async () => {
  const echoId = String(route.params.echoId || '')
  if (!echoId) return
  form.echo_id = echoId
  loadGuestProfile()
  loading.value = true
  try {
    const [metaRes, commentsRes] = await Promise.all([
      fetchGetCommentFormMeta(),
      fetchGetComments(echoId),
    ])
    if (metaRes.code === 1) {
      formMeta.value = metaRes.data
      form.form_token = metaRes.data.form_token
    }
    if (commentsRes.code === 1) {
      comments.value = commentsRes.data || []
    }
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.content = ''
  form.hp_field = ''
  form.captcha_token = ''
  form.parent_id = ''
  replyTarget.value = null
  captchaError.value = ''
  void mountCaptchaWidget()
}

const submitComment = async () => {
  if (!canSubmit.value || submitting.value) return
  if (solvingCaptcha.value) return
  if (!(await ensureCaptchaToken())) {
    theToast.error(String(t('commentSection.completeCaptchaFirst')))
    return
  }
  submitting.value = true
  try {
    const submittedContent = form.content.trim()
    const res = await fetchCreateComment(form)
    if (res.code === 1) {
      const status = res.data?.status || 'pending'
      saveGuestProfile()
      submitNotice.value = {
        status,
        contentPreview: submittedContent.slice(0, 80),
        submittedAt: Date.now(),
      }
      if (status === 'approved') {
        theToast.success(String(t('commentSection.commentPublished')))
      } else {
        theToast.success(String(t('commentSection.commentSubmittedPending')))
      }
      resetForm()
      if (status === 'approved') {
        await loadData()
      }
    }
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})

watch(
  () =>
    [
      needCaptcha.value,
      formMeta.value?.captcha_api_endpoint,
      captchaMountRef.value,
      locale.value,
    ] as const,
  async () => {
    form.captcha_token = ''
    captchaError.value = ''
    await nextTick()
    void mountCaptchaWidget()
  },
  { immediate: true, flush: 'post' },
)

onBeforeUnmount(() => {
  clearCaptchaWidget()
  if (highlightTimer) clearTimeout(highlightTimer)
})
</script>

<style scoped>
:global(:root) {
  --comment-grid-bg-color: var(--comment-grid-bg);
  --comment-grid-line-color: rgb(120 120 120 / 7%);
  --comment-card-bg: #faf8f2;
  --comment-card-border: var(--comment-sticky-border-color);
}

:global(:root.dark) {
  --comment-grid-bg-color: var(--comment-grid-bg);
  --comment-grid-line-color: rgb(240 240 240 / 5%);
  --comment-card-bg: #33312c;
  --comment-card-border: var(--comment-sticky-border-color);
}

.comment-list-board {
  position: relative;
}

:global(html.echo-detail-grid-bg),
:global(html.echo-detail-grid-bg body),
:global(html.echo-detail-grid-bg body #app) {
  background-color: var(--comment-grid-bg-color);
  background-image:
    linear-gradient(to right, var(--comment-grid-line-color) 1px, transparent 1px),
    linear-gradient(to bottom, var(--comment-grid-line-color) 1px, transparent 1px);
  background-size: 32px 32px;
  background-position: 0 0;
}

:global(html.echo-detail-grid-bg),
:global(html.echo-detail-grid-bg body) {
  margin: 0;
  padding: 0;
  min-height: 100%;
}

:global(html.echo-detail-grid-bg body #app) {
  min-height: 100vh;
}

.comment-thread {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 0.25rem;
}

.comment-card {
  position: relative;
  border: 1px solid var(--comment-card-border);
  background: var(--comment-card-bg);
  border-radius: 7px;
  padding: 0.6rem 0.7rem;
}

.comment-row {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.comment-reply-row {
  align-items: flex-start;
}

.comment-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.15rem 0.3rem;
  min-width: 0;
  line-height: 1.3;
}

.comment-floor-no {
  flex-shrink: 0;
  font-size: 0.62rem;
  font-weight: 600;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.comment-reply-ref {
  flex-shrink: 0;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  cursor: pointer;
  transition: color 0.15s ease;
}

.comment-reply-ref:hover {
  color: #0ea5e9;
}

.comment-author,
.comment-author-link {
  max-width: 9rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.comment-author-link {
  transition: color 0.18s ease;
}

.comment-author-link:hover {
  color: #0ea5e9;
}

.comment-dot {
  color: var(--color-text-muted);
}

.comment-time {
  color: var(--color-text-muted);
  white-space: nowrap;
  font-size: 0.72rem;
}

.verified-badge-icon {
  transform: translateY(0.5px);
}

.comment-hot-badge {
  position: absolute;
  top: 0.45rem;
  right: 0.5rem;
  z-index: 1;
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #ef4444;
}

.comment-hot-inline {
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #ef4444;
}

.comment-anchor-flash {
  border-radius: 7px;
  animation: comment-anchor-flash 1.6s ease-out;
}

@keyframes comment-anchor-flash {
  0%,
  55% {
    box-shadow: 0 0 0 2px rgb(14 165 233 / 50%);
  }

  100% {
    box-shadow: 0 0 0 2px rgb(14 165 233 / 0%);
  }
}

.comment-replies {
  margin-top: 0.55rem;
  margin-left: 0.35rem;
  padding-left: 0.7rem;
  border-left: 1px solid var(--comment-card-border);
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.comment-reply-btn {
  margin-top: 0.2rem;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color 0.15s ease;
}

.comment-reply-btn:hover {
  color: #0ea5e9;
}

.comment-md-content {
  margin-top: 0.2rem;
  color: var(--color-text-primary);
  font-size: 0.84rem;
  line-height: 1.6;
}

.comment-md-content :deep(p) {
  margin: 0.1rem 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.comment-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.65rem 0.2rem 0.4rem;
  border-radius: 9999px;
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
  background: transparent;
  font-size: 0.8rem;
  font-weight: 600;
  line-height: 1.2;
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease,
    background-color 0.15s ease;
}

.comment-pill-btn:hover {
  color: var(--color-text-primary);
  border-color: var(--color-border-strong);
}

.comment-pill-btn:focus,
.comment-pill-btn:focus-visible {
  outline: none;
  border-color: var(--color-border-subtle);
  box-shadow: 0 0 0 3px var(--color-bg-muted);
}

.comment-reply-btn:focus,
.comment-reply-ref:focus,
.comment-reply-cancel:focus,
.comment-submit-btn:focus,
.comment-author-link:focus {
  outline: none;
}

.comment-reply-btn:focus-visible,
.comment-reply-ref:focus-visible,
.comment-reply-cancel:focus-visible,
.comment-submit-btn:focus-visible,
.comment-author-link:focus-visible {
  outline: 1px solid var(--color-text-muted);
  outline-offset: 2px;
  border-radius: 4px;
}

.comment-pill-btn__icon {
  width: 0.95rem;
  height: 0.95rem;
  flex-shrink: 0;
}

.comment-count-text {
  color: var(--color-text-muted);
  font-size: 0.76rem;
  white-space: nowrap;
}

.comment-form-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  border: 1px solid var(--comment-card-border);
  border-radius: 8px;
  background: var(--comment-card-bg);
  padding: 0.7rem;
}

.comment-form-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.comment-form-title {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.comment-md-hint {
  display: inline-flex;
  align-items: center;
  color: var(--color-text-muted);
}

.comment-ready-indicator {
  display: inline-flex;
  align-items: center;
}

.comment-ready-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 9999px;
}

.comment-ready-dot.is-ready {
  background: #10b981;
}

.comment-ready-dot.is-not-ready {
  background: #d1d5db;
}

.comment-reply-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.28rem 0.5rem;
  border-radius: 6px;
  background: var(--color-bg-muted);
  font-size: 0.74rem;
  color: var(--color-text-secondary);
}

.comment-reply-cancel {
  flex-shrink: 0;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  line-height: 1;
  font-size: 0.8rem;
}

.comment-reply-cancel:hover {
  color: var(--color-text-primary);
}

.comment-id-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.4rem;
}

.comment-id-grid__full {
  grid-column: 1 / -1;
}

.comment-input-field {
  width: 100%;
  border: 1px solid var(--comment-input-border);
  border-radius: 6px;
  background: var(--comment-input-bg);
  color: var(--color-text-primary);
  transition:
    border-color 0.18s ease,
    box-shadow 0.2s ease,
    background-color 0.18s ease;
}

.comment-input-field::placeholder {
  color: var(--comment-input-placeholder);
}

.comment-input-field:hover {
  border-color: var(--comment-input-hover-border);
}

.comment-input-field:focus,
.comment-input-field:focus-visible {
  outline: none;
  border-color: var(--comment-input-focus-border);
  background: var(--comment-input-focus-bg);
  box-shadow: 0 0 0 3px var(--comment-input-focus-ring-outer);
}

.comment-input-sm {
  padding: 0.32rem 0.55rem;
  font-size: 0.8rem;
}

.comment-textarea {
  min-height: 3.25rem;
  padding: 0.4rem 0.55rem;
  font-size: 0.84rem;
  line-height: 1.55;
  resize: vertical;
}

.comment-preview-strip {
  border: 1px dashed var(--comment-card-border);
  border-radius: 6px;
  padding: 0.4rem 0.5rem;
}

.comment-preview-strip__head {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-bottom: 0.2rem;
}

.comment-preview-strip__name {
  max-width: 7rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.74rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.comment-preview-strip__label {
  font-size: 0.66rem;
  color: var(--color-text-muted);
}

.comment-captcha-wrap {
  width: 100%;
  box-sizing: border-box;
}

.comment-captcha-mount {
  min-height: 40px;
}

.comment-captcha-mount :deep(cap-widget) {
  box-sizing: border-box;
  display: block;
  width: 100% !important;
  max-width: 100% !important;
  min-width: 0;

  --cap-widget-width: 100%;
  --cap-widget-height: 40px;
}

.comment-captcha-error {
  font-size: 0.72rem;
  color: #ef4444;
}

.comment-form-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.comment-char-count {
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.comment-char-count.is-over {
  color: #ef4444;
}

.comment-submit-btn {
  border: none;
  border-radius: 6px;
  background: var(--color-text-primary);
  color: var(--color-bg-canvas);
  padding: 0.34rem 0.9rem;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.15s ease;
}

.comment-submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.comment-notice {
  border: 1px solid var(--color-border-subtle);
  border-radius: 6px;
  background: var(--color-bg-canvas);
  padding: 0.5rem 0.6rem;
}

.comment-notice__title {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.comment-notice__time {
  margin-top: 0.15rem;
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.comment-notice__preview {
  margin-top: 0.35rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 5px;
  padding: 0.25rem 0.4rem;
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

@media (width <= 640px) {
  .comment-submit-btn {
    min-height: 34px;
  }
}
</style>

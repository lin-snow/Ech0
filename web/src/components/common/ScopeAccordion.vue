<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ACCESS_SCOPE_GROUPS } from '@/constants/accessScopes'

const props = defineProps<{ scopes: string[] }>()

const { t } = useI18n()

type ScopeGroupView = {
  key: string
  label: string
  items: { value: string; label: string }[]
}

const uniqueScopes = computed(() => [...new Set(props.scopes)])

const scopeGroups = computed<ScopeGroupView[]>(() => {
  const requested = new Set(uniqueScopes.value)
  const known = new Set<string>()
  const groups: ScopeGroupView[] = []

  for (const group of ACCESS_SCOPE_GROUPS) {
    const items = group.items
      .filter((item) => requested.has(item.value))
      .map((item) => {
        known.add(item.value)
        return { value: item.value, label: String(t(item.labelKey)) }
      })
    if (items.length > 0) {
      groups.push({ key: group.labelKey, label: String(t(group.labelKey)), items })
    }
  }

  const other = uniqueScopes.value
    .filter((scope) => !known.has(scope))
    .map((scope) => ({ value: scope, label: scope }))
  if (other.length > 0) {
    groups.push({ key: 'other', label: String(t('scopeAccordion.otherScopes')), items: other })
  }

  return groups
})
</script>

<template>
  <div>
    <div class="mb-2 flex items-baseline justify-between gap-2">
      <p class="text-sm text-[var(--color-text-muted)]">{{ t('scopeAccordion.scopes') }}</p>
      <p class="text-[11px] text-[var(--color-text-muted)]">
        {{
          t('scopeAccordion.summary', { count: uniqueScopes.length, groups: scopeGroups.length })
        }}
      </p>
    </div>
    <div class="space-y-1.5">
      <details
        v-for="(group, index) in scopeGroups"
        :key="group.key"
        class="scope-group rounded-[var(--radius-md)] border border-[var(--color-border-subtle)] overflow-hidden"
        :open="index === 0"
      >
        <summary
          class="flex cursor-pointer list-none items-center justify-between gap-3 px-3 py-2.5 text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)]/60"
        >
          <span class="flex min-w-0 items-center gap-2">
            <span class="truncate font-medium">{{ group.label }}</span>
            <span
              class="shrink-0 rounded-full bg-[var(--color-bg-muted)] px-1.5 py-0.5 text-[11px] text-[var(--color-text-muted)]"
            >
              {{ group.items.length }}
            </span>
          </span>
          <svg
            class="scope-group__chevron h-4 w-4 shrink-0 text-[var(--color-text-muted)]"
            viewBox="0 0 20 20"
            fill="none"
            aria-hidden="true"
          >
            <path
              d="M5 7.5 10 12.5 15 7.5"
              stroke="currentColor"
              stroke-width="1.75"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </summary>
        <ul class="flex flex-wrap gap-1.5 px-3 pb-3">
          <li
            v-for="item in group.items"
            :key="item.value"
            :title="item.value"
            class="text-xs px-2 py-0.5 rounded-full bg-[var(--color-accent-soft)] text-[var(--color-text-primary)]"
          >
            {{ item.label }}
          </li>
        </ul>
      </details>
    </div>
  </div>
</template>

<style scoped>
.scope-group__chevron {
  transition: transform 160ms ease;
}

.scope-group[open] .scope-group__chevron {
  transform: rotate(180deg);
}

@media (prefers-reduced-motion: reduce) {
  .scope-group__chevron {
    transition: none;
  }
}
</style>

<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script lang="ts">
import { defineComponent, h, ref, watch } from 'vue'

export default defineComponent({
  name: 'DiffText',
  props: {
    text: { type: String, default: '' },
    animationClass: { type: String, default: 'am-blur-in' },
    duration: { type: Number, default: 600 },
    animate: { type: Boolean, default: true },
  },
  setup(props) {
    const chunks = ref<{ id: number; text: string }[]>([])
    let collected = ''
    let nextId = 0

    const pushPieces = (delta: string) => {
      for (const piece of delta.split(/(\s+)/)) {
        if (piece.length) chunks.value.push({ id: nextId++, text: piece })
      }
    }

    const truncateTo = (len: number) => {
      const kept: { id: number; text: string }[] = []
      let acc = 0
      for (const c of chunks.value) {
        if (acc + c.text.length <= len) {
          kept.push(c)
          acc += c.text.length
        } else {
          if (len > acc) kept.push({ id: c.id, text: c.text.slice(0, len - acc) })
          break
        }
      }
      chunks.value = kept
    }

    const sync = (input: string) => {
      if (!props.animate) {
        chunks.value = input ? [{ id: 0, text: input }] : []
        collected = input
        return
      }
      if (input === collected) return
      if (input.startsWith(collected)) {
        pushPieces(input.slice(collected.length))
        collected = input
        return
      }
      if (collected.startsWith(input)) {
        truncateTo(input.length)
        collected = input
        return
      }
      chunks.value = []
      pushPieces(input)
      collected = input
    }

    watch(() => props.text, sync, { immediate: true })

    return () =>
      chunks.value.map((c) =>
        h(
          'span',
          {
            key: c.id,
            class: ['am-tok', props.animate ? props.animationClass : null],
            style: props.animate ? { animationDuration: `${props.duration}ms` } : undefined,
          },
          c.text,
        ),
      )
  },
})
</script>

<style>
.am-tok {
  display: inline-block;
  white-space: pre-wrap;
  animation-iteration-count: 1;
  animation-fill-mode: both;

  animation-timing-function: ease-in-out;
}

.am-tok.am-blur-in {
  animation-name: am-blur-in;
}

.am-tok.am-fade-in {
  animation-name: am-fade-in;
}

@keyframes am-blur-in {
  from {
    opacity: 0;
    filter: blur(5px);
    transform: translateX(-0.12em);
  }

  to {
    opacity: 1;
    filter: blur(0);
    transform: translateX(0);
  }
}

@keyframes am-fade-in {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}
</style>

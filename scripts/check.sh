#!/usr/bin/env bash

set -u

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  C_HDR=$'\033[1;34m'
  C_HDR_WEB=$'\033[1;35m'
  C_HDR_ALL=$'\033[1;36m'
  C_OK=$'\033[1;32m'
  C_FAIL=$'\033[1;31m'
  C_RESET=$'\033[0m'
else
  C_HDR=""; C_HDR_WEB=""; C_HDR_ALL=""; C_OK=""; C_FAIL=""; C_RESET=""
fi

STEPS=(
  "spdx|$C_HDR_ALL|全仓：SPDX/Copyright 头补齐 (scripts/add-spdx-headers.mjs)|node scripts/add-spdx-headers.mjs"
  "go-fmt|$C_HDR|后端：格式化 (golangci-lint fmt)|golangci-lint fmt"
  "go-lint|$C_HDR|后端：Lint (golangci-lint run)|golangci-lint run"
  "openapi|$C_HDR|后端：OpenAPI 规格生成 (Huma, just openapi)|go run ./cmd/openapi-gen"
  "web-format|$C_HDR_WEB|前端 web：格式化 (prettier --write src/)|(cd web && pnpm format)"
  "web-lint|$C_HDR_WEB|前端 web：Lint (eslint . --fix)|(cd web && pnpm lint)"
  "web-style|$C_HDR_WEB|前端 web：Stylelint (stylelint --fix)|(cd web && pnpm lint:style)"
  "web-i18n|$C_HDR_WEB|前端 web：i18n 校验 (key / unused / hardcoded / pseudo-smoke)|(cd web && pnpm run i18n:check)"
)

keys=()
statuses=()
durations=()
overall=0
total_start=$(date +%s)

fmt_dur() {
  local s=$1
  if [ "$s" -ge 60 ]; then
    printf '%dm%02ds' "$((s / 60))" "$((s % 60))"
  else
    printf '%ds' "$s"
  fi
}

for entry in "${STEPS[@]}"; do
  IFS='|' read -r key color banner cmd <<< "$entry"
  printf '%s=== %s ===%s\n' "$color" "$banner" "$C_RESET"
  start=$(date +%s)
  if eval "$cmd"; then
    statuses+=("PASS")
  else
    statuses+=("FAIL")
    overall=1
  fi
  end=$(date +%s)
  keys+=("$key")
  durations+=("$(fmt_dur $((end - start)))")
done

total_elapsed=$(fmt_dur $(($(date +%s) - total_start)))

key_w=4
for k in "${keys[@]}"; do
  if [ "${#k}" -gt "$key_w" ]; then key_w=${#k}; fi
done
status_w=6
dur_w=8
for d in "${durations[@]}"; do
  if [ "${#d}" -gt "$dur_w" ]; then dur_w=${#d}; fi
done

repeat_dash() {
  local n=$1
  local i=0
  while [ "$i" -lt "$n" ]; do
    printf -- '-'
    i=$((i + 1))
  done
}

border() {
  printf '+'
  repeat_dash $((key_w + 2));    printf '+'
  repeat_dash $((status_w + 2)); printf '+'
  repeat_dash $((dur_w + 2));    printf '+\n'
}

echo
printf '%s== 运行结果 ==%s\n' "$C_HDR_ALL" "$C_RESET"
border
printf '| %-*s | %-*s | %*s |\n' "$key_w" "Step" "$status_w" "Status" "$dur_w" "Duration"
border

pass_count=0
fail_count=0
for i in "${!keys[@]}"; do
  s="${statuses[$i]}"
  if [ "$s" = "PASS" ]; then
    color="$C_OK"
    pass_count=$((pass_count + 1))
  else
    color="$C_FAIL"
    fail_count=$((fail_count + 1))
  fi
  pad=$((status_w - ${#s}))
  printf '| %-*s | %s%s%s%*s | %*s |\n' \
    "$key_w" "${keys[$i]}" \
    "$color" "$s" "$C_RESET" "$pad" "" \
    "$dur_w" "${durations[$i]}"
done
border
echo

if [ "$overall" -eq 0 ]; then
  printf '%s✓ %d/%d steps passed — %s elapsed%s\n' \
    "$C_OK" "$pass_count" "${#keys[@]}" "$total_elapsed" "$C_RESET"
else
  printf '%s✘ %d/%d steps failed — %s elapsed%s\n' \
    "$C_FAIL" "$fail_count" "${#keys[@]}" "$total_elapsed" "$C_RESET"
  printf '%s  Failed:%s' "$C_FAIL" "$C_RESET"
  for i in "${!keys[@]}"; do
    if [ "${statuses[$i]}" = "FAIL" ]; then
      printf ' %s' "${keys[$i]}"
    fi
  done
  echo
fi

exit "$overall"

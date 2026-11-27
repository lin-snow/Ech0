// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { i18n } from '@/locales'

// Kept in step with internal/httperr.TypeAccountDisabled.
export const ACCOUNT_DISABLED_TYPE = 'https://ech0.app/problems/account-disabled'

// Status is the behavior. The server sends no user-facing text, so every
// message the user reads is chosen here.
const PROBLEM_MESSAGES: Record<number, string> = {
  400: 'common.invalidRequest',
  401: 'common.unauthorized',
  403: 'common.forbidden',
  404: 'common.notFound',
  409: 'common.conflict',
  412: 'common.conflict',
  422: 'common.invalidRequest',
  429: 'common.rateLimited',
}

export const problemMessage = (problem: App.Api.Problem): string => {
  const key =
    PROBLEM_MESSAGES[problem.status] ??
    (problem.status >= 500 ? 'common.serverError' : 'common.requestFailed')
  return String(i18n.global.t(key))
}

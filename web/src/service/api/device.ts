// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

/** What the CLI asked for, so the approval page can show it before approving. */
export function fetchDevicePending(userCode: string) {
  return request<App.Api.Device.PendingInfo>({
    url: `/auth/device/pending?user_code=${encodeURIComponent(userCode)}`,
    method: 'GET',
    silentError: true,
  })
}

/** Approves the pending request; the CLI then receives its access token. */
export function fetchDeviceApprove(userCode: string) {
  return request<null>({
    url: '/auth/device/approve',
    method: 'POST',
    data: { user_code: userCode },
  })
}

// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

/** Validated authorization request rendered by the consent screen. */
export interface OAuthConsentContext {
  client_name: string
  client_uri?: string
  redirect_uri: string
  scopes: string[]
  resource: string
  state?: string
}

export interface OAuthAuthorizeParams {
  client_id: string
  redirect_uri: string
  response_type: string
  scope: string
  state?: string
  resource: string
  code_challenge: string
  code_challenge_method: string
}

export interface OAuthAuthorizePayload extends OAuthAuthorizeParams {
  decision: 'approve' | 'deny'
}

/** Reads the validated request so the page can show the client and scopes. */
export function fetchOAuthConsent(params: OAuthAuthorizeParams) {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value) query.set(key, value)
  }
  return request<OAuthConsentContext>({
    url: `/authz/consent?${query.toString()}`,
    method: 'GET',
    silentError: true,
  })
}

/** Approves or denies the request; the server returns the client redirect. */
export function submitOAuthConsent(payload: OAuthAuthorizePayload) {
  return request<{ redirect_url: string }>({
    url: '/authz/authorize',
    method: 'POST',
    data: payload,
    silentError: true,
  })
}

// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchLogin(loginParams: App.Api.Auth.LoginParams) {
  return request<App.Api.Auth.LoginResponse>({
    url: '/login',
    method: 'POST',
    data: loginParams,
  })
}

export function fetchSignup(signupParams: App.Api.Auth.SignupParams) {
  return request({
    url: '/register',
    method: 'POST',
    data: signupParams,
  })
}

export function fetchLogout() {
  return request<null>({
    url: '/auth/logout',
    method: 'POST',
    silentError: true,
  })
}

export function fetchExchangeCode(code: string) {
  return request<App.Api.Auth.TokenPairResponse>({
    url: '/auth/exchange',
    method: 'POST',
    data: { code },
  })
}

export function fetchPasskeyRegisterBegin(deviceName: string) {
  return request<App.Api.Auth.PasskeyRegisterBeginResp>({
    url: '/passkey/register/begin',
    method: 'POST',
    data: { device_name: deviceName },
  })
}

export function fetchPasskeyRegisterFinish(nonce: string, credential: unknown) {
  return request({
    url: '/passkey/register/finish',
    method: 'POST',
    data: { nonce, credential },
  })
}

export function fetchPasskeyLoginBegin() {
  return request<App.Api.Auth.PasskeyLoginBeginResp>({
    url: '/passkey/login/begin',
    method: 'POST',
    data: {},
  })
}

export function fetchPasskeyLoginFinish(nonce: string, credential: unknown) {
  return request<App.Api.Auth.TokenPairResponse>({
    url: '/passkey/login/finish',
    method: 'POST',
    data: { nonce, credential },
  })
}

export function fetchPasskeyDevices() {
  return request<App.Api.Auth.PasskeyDevice[]>({
    url: '/passkeys',
    method: 'GET',
  })
}

export function fetchDeletePasskeyDevice(id: string) {
  return request({
    url: `/passkeys/${id}`,
    method: 'DELETE',
  })
}

export function fetchUpdatePasskeyDeviceName(id: string, deviceName: string) {
  return request({
    url: `/passkeys/${id}`,
    method: 'PUT',
    data: { device_name: deviceName },
  })
}

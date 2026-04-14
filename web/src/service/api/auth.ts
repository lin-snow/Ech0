import { request } from '../request'

// 登录
export function fetchLogin(loginParams: App.Api.Auth.LoginParams) {
  return request<App.Api.Auth.LoginResponse>({
    url: '/login',
    method: 'POST',
    data: loginParams,
  })
}

// 注册
export function fetchSignup(signupParams: App.Api.Auth.SignupParams) {
  return request({
    url: '/register',
    method: 'POST',
    data: signupParams,
  })
}

// 登出（best-effort，错误静默处理）
export function fetchLogout() {
  return request<null>({
    url: '/auth/logout',
    method: 'POST',
    silentError: true,
  })
}

// 一次性 code 交换 token（OAuth 回调专用）
export function fetchExchangeCode(code: string) {
  return request<App.Api.Auth.TokenPairResponse>({
    url: '/auth/exchange',
    method: 'POST',
    data: { code },
  })
}

// Passkey / WebAuthn
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

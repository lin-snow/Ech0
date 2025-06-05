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

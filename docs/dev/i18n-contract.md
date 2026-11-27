# Ech0 i18n Contract

## 1. Locale Strategy

- Supported locales: `zh-CN`, `en-US`.
- Default locale: `zh-CN`.
- Fallback locale: `en-US`.
- Negotiation order:
  1. Explicit request value (`lang` query or `X-Locale` header)
  2. User preference (`user.locale`, for authenticated requests)
  3. System default locale (`system_settings.default_locale`)
  4. `Accept-Language`
  5. Fallback locale (`zh-CN`)
- Frontend startup order:
  1. `localStorage.locale`
  2. user profile locale (after login)
  3. `system_settings.default_locale`
  4. browser locale

## 2. API Contract

- Failures are RFC 9457 `application/problem+json` documents (see [api-contract-design.md](api-contract-design.md)); there is no success envelope and no server-sent `msg` for successful calls.
- The server produces **no user-facing text**. A problem document carries only:
  - `status`: the only signal a client branches on.
  - `type`: a URI, present only where two errors share a status (account disabled; the four device-flow states).
  - `detail`: English, for developers, logs and `curl`. Clients must never parse or display it as-is.
- Client rendering is chosen by status (`web/src/service/request/problem.ts`): 400/422 invalid request, 401 session expired, 403 forbidden, 404 not found, 409/412 conflict, 429 too many requests, 5xx server error.
- Diagnostics endpoints (connectivity tests for S3, agent, SMTP, webhook) opt into `onProblem` and show `detail` verbatim: there the English text *is* the answer.

## 3. Key Naming Rules

- Use semantic keys, never use source text as key.
- Recommended format: `domain.module.action`.
- Current examples:
  - `common.request_failed`
  - `auth.token_missing`
  - `dashboard.logs.tail_invalid`

## 4. Backend Rules

- A business error is an `internal/apperr` sentinel. `httperr.From` is the only place that turns it into a status and a developer-facing English `detail`. The status is the whole contract.
- Never write Chinese into an error message that can reach the wire. OpenAPI `Summary`/`Description`, LLM prompts and panic messages are not wire text and may stay in Chinese.
- The server has no translation catalog. `internal/i18n` only resolves which locale to use (default settings, user preference, `X-Locale`) and passes it to the copilot so the LLM answers in the user's language.
- Keep logs and telemetry machine-oriented; do not localize log field names.

## 5. Frontend Rules

- New UI text should be added to locale JSON and rendered via `t()`.
- API error display is derived from `status` (see §2); never from `detail`.
- `X-Locale` is still sent; the server uses it only to pick the copilot answer language and to validate `default_locale`.
- User settings should persist personal locale to backend (`user.locale`) and switch UI locale immediately after save.

## 6. Content/Template Rules

- User-generated content is not auto-translated by backend.
- System-generated content (built-in prompts) stays server-side and is not localized by key; it is not user-facing UI text.

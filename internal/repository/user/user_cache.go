// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import "fmt"

const (
	UsernameKeyPrefix = "username"
	IDKeyPrefix       = "id"
	PasskeyRegKey     = "passkey:reg"
	PasskeyLoginKey   = "passkey:login"
)

func GetUserIDKey(id string) string {
	return fmt.Sprintf("%s:%s", IDKeyPrefix, id)
}

func GetUsernameKey(username string) string {
	return fmt.Sprintf("%s:%s", UsernameKeyPrefix, username)
}

func GetPasskeyRegisterSessionKey(nonce string) string {
	return fmt.Sprintf("%s:%s", PasskeyRegKey, nonce)
}

func GetPasskeyLoginSessionKey(nonce string) string {
	return fmt.Sprintf("%s:%s", PasskeyLoginKey, nonce)
}

// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package uuid

import (
	"crypto/sha1"
	"uuid"
)

var NameSpaceURL = uuid.MustParse("6ba7b811-9dad-11d1-80b4-00c04fd430c8")

func NewV7() string {
	return uuid.NewV7().String()
}

func NewV5(namespace uuid.UUID, name []byte) uuid.UUID {
	buf := make([]byte, 0, len(namespace)+len(name))
	buf = append(buf, namespace[:]...)
	buf = append(buf, name...)
	sum := sha1.Sum(buf)

	id := uuid.UUID(sum[:16])
	id[6] = id[6]&0x0f | 0x50
	id[8] = id[8]&0x3f | 0x80
	return id
}

func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

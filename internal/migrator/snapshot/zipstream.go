// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"bytes"
	"encoding/binary"
	"math"
)

// One uncompressed entry lets workers append independently. The signed data
// descriptor carries the final CRC. ZIP64 keeps the original ech0.db entry
// readable by SQLite tooling and by the upstream snapshot importer above 4 GiB.
const zipEntry = "ech0.db"

type zipRecord struct{ bytes.Buffer }

func (r *zipRecord) u16(v uint16) { r.Write(binary.LittleEndian.AppendUint16(nil, v)) }
func (r *zipRecord) u32(v uint32) { r.Write(binary.LittleEndian.AppendUint32(nil, v)) }
func (r *zipRecord) u64(v uint64) { r.Write(binary.LittleEndian.AppendUint64(nil, v)) }
func zipSize(n uint64) uint32     { return uint32(min(n, uint64(math.MaxUint32))) }
func zipExtra(n uint64) []byte {
	if n < math.MaxUint32 {
		return nil
	}
	var b zipRecord
	b.u16(1)
	b.u16(16)
	b.u64(n)
	b.u64(n)
	return b.Bytes()
}
func zipVersion(n uint64) uint16 {
	if n >= math.MaxUint32 {
		return 45
	}
	return 20
}
func snapshotHeader(n uint64) []byte {
	var b zipRecord
	extra := zipExtra(n)
	b.u32(0x04034b50)
	b.u16(zipVersion(n))
	b.u16(8)
	b.u16(0)
	b.u16(0)
	b.u16(33)
	b.u32(0)
	b.u32(zipSize(n))
	b.u32(zipSize(n))
	b.u16(uint16(len(zipEntry)))
	b.u16(uint16(len(extra)))
	b.WriteString(zipEntry)
	b.Write(extra)
	return b.Bytes()
}
func snapshotFooter(n uint64, crc uint32) []byte {
	var b zipRecord
	extra := zipExtra(n)
	b.u32(0x08074b50)
	b.u32(crc)
	if n >= math.MaxUint32 {
		b.u64(n)
		b.u64(n)
	} else {
		b.u32(uint32(n))
		b.u32(uint32(n))
	}
	start := uint64(len(snapshotHeader(n))) + n + uint64(b.Len())
	desc := b.Len()
	b.u32(0x02014b50)
	b.u16(45)
	b.u16(zipVersion(n))
	b.u16(8)
	b.u16(0)
	b.u16(0)
	b.u16(33)
	b.u32(crc)
	b.u32(zipSize(n))
	b.u32(zipSize(n))
	b.u16(uint16(len(zipEntry)))
	b.u16(uint16(len(extra)))
	b.u16(0)
	b.u16(0)
	b.u16(0)
	b.u32(0)
	b.u32(0)
	b.WriteString(zipEntry)
	b.Write(extra)
	size := uint64(b.Len() - desc)
	if start >= math.MaxUint32 {
		b.u32(0x06064b50)
		b.u64(44)
		b.u16(45)
		b.u16(45)
		b.u32(0)
		b.u32(0)
		b.u64(1)
		b.u64(1)
		b.u64(size)
		b.u64(start)
		b.u32(0x07064b50)
		b.u32(0)
		b.u64(start + size)
		b.u32(1)
	}
	b.u32(0x06054b50)
	b.u16(0)
	b.u16(0)
	b.u16(1)
	b.u16(1)
	b.u32(zipSize(size))
	b.u32(zipSize(start))
	b.u16(0)
	return b.Bytes()
}

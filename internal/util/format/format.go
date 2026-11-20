// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

func ValidateCrontabExpression(expr string) error {
	_, err := ParseCrontabExpression(expr)
	return err
}

func ParseCrontabExpression(expr string) (cron.Schedule, error) {
	expr = strings.TrimSpace(expr)
	fields := strings.Fields(expr)

	switch len(fields) {
	case 5:
		s, err := cron.ParseStandard(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid 5-field cron expression: %w", err)
		}
		return s, nil
	case 6:
		parser := cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		)
		s, err := parser.Parse(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid 6-field cron expression: %w", err)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("cron expression must have 5 or 6 fields, got %d", len(fields))
	}

}

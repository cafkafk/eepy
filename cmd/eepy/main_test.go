/*
 * SPDX-FileCopyrightText: 2025 Christina Sørensen
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"testing"
	"time"
)

func TestGenerateSchedule(t *testing.T) {
	wakeTime, _ := time.Parse(timeFormat, "10:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("1h30m")
	schedule := generateSchedule(wakeTime, targetWakeTime, adjustment)

	if len(schedule) != 5 {
		t.Errorf("Expected schedule to have 5 entries, but got %d", len(schedule))
	}
}

func TestGenerateScheduleOnTarget(t *testing.T) {
	wakeTime, _ := time.Parse(timeFormat, "05:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("1h30m")
	schedule := generateSchedule(wakeTime, targetWakeTime, adjustment)

	if len(schedule) != 1 {
		t.Errorf("Expected schedule to have 1 entry, but got %d", len(schedule))
	}
}


func TestGenerateScheduleAheadOfTarget(t *testing.T) {
	wakeTime, _ := time.Parse(timeFormat, "04:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("1h30m")
	schedule := generateSchedule(wakeTime, targetWakeTime, adjustment)

	if len(schedule) != 1 {
		t.Errorf("Expected schedule to have 1 entry, but got %d", len(schedule))
	}
}

func TestGenerateScheduleWithDifferentAdjustment(t *testing.T) {
	wakeTime, _ := time.Parse(timeFormat, "10:00")
	targetWakeTime, _ := time.Parse(timeFormat, "08:00")
	adjustment, _ := time.ParseDuration("30m")
	schedule := generateSchedule(wakeTime, targetWakeTime, adjustment)

	if len(schedule) != 5 {
		t.Errorf("Expected schedule to have 5 entries, but got %d", len(schedule))
	}
}

func TestGenerateScheduleWithComplexAdjustment(t *testing.T) {
	wakeTime, _ := time.Parse(timeFormat, "10:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("3h45m")
	schedule := generateSchedule(wakeTime, targetWakeTime, adjustment)

	if len(schedule) != 3 {
		t.Errorf("Expected schedule to have 3 entries, but got %d", len(schedule))
	}
}

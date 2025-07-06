/*
 * SPDX-FileCopyrightText: 2025 Christina Sørensen
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGeneratePlan(t *testing.T) {
	// Redirect stdout to a buffer to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function
	wakeTime, _ := time.Parse(timeFormat, "10:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("1h30m")
	generatePlan(wakeTime, targetWakeTime, adjustment)

	// Restore stdout and read the buffer
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the output
	expected := "Your sleep calibration plan:"
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("Expected output to contain %q, but it did not", expected)
	}
}

func TestGeneratePlanOnTarget(t *testing.T) {
	// Redirect stdout to a buffer to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function
	wakeTime, _ := time.Parse(timeFormat, "05:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("1h30m")
	generatePlan(wakeTime, targetWakeTime, adjustment)

	// Restore stdout and read the buffer
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the output
	expected := "You are already at or ahead of your target wake time!"
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("Expected output to contain %q, but it did not", expected)
	}
}


func TestGeneratePlanAheadOfTarget(t *testing.T) {
	// Redirect stdout to a buffer to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function
	wakeTime, _ := time.Parse(timeFormat, "04:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("1h30m")
	generatePlan(wakeTime, targetWakeTime, adjustment)

	// Restore stdout and read the buffer
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the output
	expected := "You are already at or ahead of your target wake time!"
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("Expected output to contain %q, but it did not", expected)
	}

	expectedSurplus := "You got 1h0m0s extra sleep today. Nice!"
	if !strings.Contains(buf.String(), expectedSurplus) {
		t.Errorf("Expected output to contain %q, but it did not", expectedSurplus)
	}
}

func TestGeneratePlanWithDifferentAdjustment(t *testing.T) {
	// Redirect stdout to a buffer to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function
	wakeTime, _ := time.Parse(timeFormat, "10:00")
	targetWakeTime, _ := time.Parse(timeFormat, "08:00")
	adjustment, _ := time.ParseDuration("30m")
	generatePlan(wakeTime, targetWakeTime, adjustment)

	// Restore stdout and read the buffer
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the output
	output := buf.String()
	now := time.Now()
	expectedLines := []string{
		now.Format("Mon, Jan 2") + " (Day 1):",
		"  - Wake up at 10:00",
		now.AddDate(0, 0, 1).Format("Mon, Jan 2") + " (Day 2):",
		"  - Wake up at 09:30",
		now.AddDate(0, 0, 2).Format("Mon, Jan 2") + " (Day 3):",
		"  - Wake up at 09:00",
		now.AddDate(0, 0, 3).Format("Mon, Jan 2") + " (Day 4):",
		"  - Wake up at 08:30",
		now.AddDate(0, 0, 4).Format("Mon, Jan 2") + " (Day 5):",
		"  - Wake up at 08:00",
	}

	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("Expected output to contain %q, but it did not", line)
		}
	}
}

func TestGeneratePlanWithComplexAdjustment(t *testing.T) {
	// Redirect stdout to a buffer to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function
	wakeTime, _ := time.Parse(timeFormat, "10:00")
	targetWakeTime, _ := time.Parse(timeFormat, "05:00")
	adjustment, _ := time.ParseDuration("3h45m")
	generatePlan(wakeTime, targetWakeTime, adjustment)

	// Restore stdout and read the buffer
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the output
	output := buf.String()
	now := time.Now()
	expectedLines := []string{
		now.Format("Mon, Jan 2") + " (Day 1):",
		"  - Wake up at 10:00",
		now.AddDate(0, 0, 1).Format("Mon, Jan 2") + " (Day 2):",
		"  - Wake up at 06:15",
		now.AddDate(0, 0, 2).Format("Mon, Jan 2") + " (Day 3):",
		"  - Wake up at 05:00",
	}

	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("Expected output to contain %q, but it did not", line)
		}
	}
}

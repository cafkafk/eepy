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

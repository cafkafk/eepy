/*
 * SPDX-FileCopyrightText: 2025 Christina Sørensen
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/spf13/pflag"
)

const (
	idealSleepDuration = 9 * time.Hour
	minSleepDuration   = 7*time.Hour + 30*time.Minute
	timeFormat         = "15:04"
)

func main() {
	targetWakeTimeStr := pflag.String("target", "05:00", "Your target wake up time (HH:MM)")
	adjustmentStr := pflag.String("adjustment", "1h30m", "Adjustment per day")
	adb := pflag.Bool("adb", false, "Set alarm on Android device via ADB")
	noSkipToday := pflag.Bool("no-skip-today", false, "Do not skip setting an alarm for today")
	pflag.Parse()

	if len(pflag.Args()) != 1 {
		fmt.Println("Usage: eepy [wake-time] [flags]")
		pflag.PrintDefaults()
		os.Exit(1)
	}

	wakeTimeStr := pflag.Arg(0)
	wakeTime, err := time.Parse(timeFormat, wakeTimeStr)
	if err != nil {
		fmt.Printf("Error parsing wake-time: %v\n", err)
		os.Exit(1)
	}

	targetWakeTime, err := time.Parse(timeFormat, *targetWakeTimeStr)
	if err != nil {
		fmt.Printf("Error parsing target-wake-time: %v\n", err)
		os.Exit(1)
	}

	adjustment, err := time.ParseDuration(*adjustmentStr)
	if err != nil {
		fmt.Printf("Error parsing adjustment: %v\n", err)
		os.Exit(1)
	}

	plan := generatePlan(wakeTime, targetWakeTime, adjustment)

	if *adb {
		setAlarms(plan, *noSkipToday)
	}
}

func generatePlan(wakeTime, targetWakeTime time.Time, adjustment time.Duration) []time.Time {
	var plan []time.Time
	now := time.Now()
	if !wakeTime.After(targetWakeTime) {
		bedtime := wakeTime.Add(-idealSleepDuration)
		surplus := targetWakeTime.Sub(wakeTime)
		fmt.Println("You are already at or ahead of your target wake time!")
		fmt.Printf("Your bedtime for tonight is %s.\n", bedtime.Format(timeFormat))
		if surplus > 0 {
			fmt.Printf("You got %v extra sleep today. Nice!\n", surplus)
		}
		tomorrow := now.AddDate(0, 0, 1)
		wakeTimeWithDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), wakeTime.Hour(), wakeTime.Minute(), 0, 0, now.Location())
		plan = append(plan, wakeTimeWithDate)
		return plan
	}

	fmt.Println("Your sleep calibration plan:")
	fmt.Println("-----------------------------")
	fmt.Printf("Ideal sleep: %.1f hours. Minimum functional sleep: %.1f hours.\n", idealSleepDuration.Hours(), minSleepDuration.Hours())
	fmt.Println("-----------------------------")

	currentWakeTime := wakeTime
	day := 1

	for {
		dayOfPlan := now.AddDate(0, 0, day-1)
		bedtime := currentWakeTime.Add(-idealSleepDuration)
		fmt.Printf("%s (Day %d):\n", dayOfPlan.Format("Mon, Jan 2"), day)
		fmt.Printf("  - Wake up at %s\n", currentWakeTime.Format(timeFormat))
		fmt.Printf("  - Go to bed at %s\n", bedtime.Format(timeFormat))

		wakeTimeWithDate := time.Date(dayOfPlan.Year(), dayOfPlan.Month(), dayOfPlan.Day(), currentWakeTime.Hour(), currentWakeTime.Minute(), 0, 0, now.Location())
		plan = append(plan, wakeTimeWithDate)

		if !currentWakeTime.After(targetWakeTime) {
			break
		}

		currentWakeTime = currentWakeTime.Add(-adjustment)
		if currentWakeTime.Before(targetWakeTime) {
			currentWakeTime = targetWakeTime
		}
		day++
	}
	fmt.Println("-----------------------------")
	fmt.Println("You have reached your target sleep schedule!")
	return plan
}

func setAlarms(plan []time.Time, noSkipToday bool) {
	if len(plan) > 7 {
		fmt.Println("Error: Cannot schedule alarms for a plan longer than 7 days.")
		os.Exit(1)
	}

	if !noSkipToday && len(plan) > 0 {
		plan = plan[1:]
	}

	fmt.Println("Setting alarms via ADB...")

	for _, wakeTime := range plan {
		hour := wakeTime.Hour()
		minute := wakeTime.Minute()
		dayOfWeek := wakeTime.Weekday()

		// Map time.Weekday to Android Calendar constants
		// Sunday = 1, Monday = 2, ..., Saturday = 7
		androidDay := int(dayOfWeek) + 1

		fmt.Printf("Setting alarm for %s: %02d:%02d\n", wakeTime.Format("Mon, Jan 2"), hour, minute)

		args := []string{
			"shell", "am", "start",
			"-a", "android.intent.action.SET_ALARM",
			"--ei", "android.intent.extra.alarm.HOUR", strconv.Itoa(hour),
			"--ei", "android.intent.extra.alarm.MINUTES", strconv.Itoa(minute),
			"--eia", "android.intent.extra.alarm.DAYS", strconv.Itoa(androidDay),
			"--es", "android.intent.extra.alarm.MESSAGE", fmt.Sprintf("'Sleep Adjustment Wake Up: %s'", wakeTime.Format("Mon, Jan 2")),
		}

		cmd := exec.Command("adb", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing adb command for %s: %v\n", wakeTime.Format("Mon, Jan 2"), err)
			fmt.Printf("Output: %s\n", string(output))
		} else {
			fmt.Printf("Alarm for %s sent successfully.\n", wakeTime.Format("Mon, Jan 2"))
			if len(output) > 0 {
				fmt.Printf("Output: %s\n", string(output))
			}
		}
	}
}

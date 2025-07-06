/*
 * SPDX-FileCopyrightText: 2025 Christina Sørensen
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"fmt"
	"os"
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

	generatePlan(wakeTime, targetWakeTime, adjustment)
}


func generatePlan(wakeTime, targetWakeTime time.Time, adjustment time.Duration) {
	if !wakeTime.After(targetWakeTime) {
		bedtime := wakeTime.Add(-idealSleepDuration)
		surplus := targetWakeTime.Sub(wakeTime)
		fmt.Println("You are already at or ahead of your target wake time!")
		fmt.Printf("Your bedtime for tonight is %s.\n", bedtime.Format(timeFormat))
		if surplus > 0 {
			fmt.Printf("You got %v extra sleep today. Nice!\n", surplus)
		}
		return
	}

	fmt.Println("Your sleep calibration plan:")
	fmt.Println("-----------------------------")
	fmt.Printf("Ideal sleep: %.1f hours. Minimum functional sleep: %.1f hours.\n", idealSleepDuration.Hours(), minSleepDuration.Hours())
	fmt.Println("-----------------------------")

	currentWakeTime := wakeTime
	day := 1

	for {
		bedtime := currentWakeTime.Add(-idealSleepDuration)
		fmt.Printf("Day %d:\n", day)
		fmt.Printf("  - Wake up at %s\n", currentWakeTime.Format(timeFormat))
		fmt.Printf("  - Go to bed at %s\n", bedtime.Format(timeFormat))

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
}

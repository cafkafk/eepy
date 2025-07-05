package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	idealSleepDuration = 9 * time.Hour
	minSleepDuration   = 7*time.Hour + 30*time.Minute
	timeFormat         = "15:04"
)

func main() {
	wakeTimeStr := flag.String("wake-time", "", "Your wake up time for today (HH:MM)")
	targetWakeTimeStr := flag.String("target-wake-time", "05:00", "Your target wake up time (HH:MM)")
	adjustmentStr := flag.String("adjustment", "1h30m", "Adjustment per day")
	flag.Parse()

	if *wakeTimeStr == "" {
		fmt.Println("Error: wake-time is required.")
		flag.Usage()
		os.Exit(1)
	}

	wakeTime, err := time.Parse(timeFormat, *wakeTimeStr)
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
	fmt.Println("Your sleep calibration plan:")
	fmt.Println("-----------------------------")

	currentWakeTime := wakeTime
	day := 1

	// Loop until the current wake time is the same as or earlier than the target.
	for currentWakeTime.After(targetWakeTime) {
		bedtime := currentWakeTime.Add(-idealSleepDuration)
		if day == 1 {
			tiredBedtime := currentWakeTime.Add(-minSleepDuration)
			fmt.Printf("Day %d (Today):\n", day)
			fmt.Printf("  - Wake up at %s\n", currentWakeTime.Format(timeFormat))
			fmt.Printf("  - Go to bed at %s (for %.1f hours of sleep) or %s (for %.1f hours of sleep)\n",
				tiredBedtime.Format(timeFormat), minSleepDuration.Hours(), bedtime.Format(timeFormat), idealSleepDuration.Hours())
		} else {
			fmt.Printf("Day %d:\n", day)
			fmt.Printf("  - Wake up at %s\n", currentWakeTime.Format(timeFormat))
			fmt.Printf("  - Go to bed at %s (for %.1f hours of sleep)\n", bedtime.Format(timeFormat), idealSleepDuration.Hours())
		}

		currentWakeTime = currentWakeTime.Add(-adjustment)
		day++
	}

	// After the loop, print the final target schedule.
	bedtime := targetWakeTime.Add(-idealSleepDuration)
	fmt.Printf("Day %d:\n", day)
	fmt.Printf("  - Wake up at %s\n", targetWakeTime.Format(timeFormat))
	fmt.Printf("  - Go to bed at %s (for %.1f hours of sleep)\n", bedtime.Format(timeFormat), idealSleepDuration.Hours())


	fmt.Println("-----------------------------")
	fmt.Println("You have reached your target sleep schedule!")
}
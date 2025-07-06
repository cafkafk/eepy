/*
 * SPDX-FileCopyrightText: 2025 Christina Sørensen
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
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
	htmlOutput := pflag.Bool("html", false, "Generate an HTML visualization of the plan")
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

	if *htmlOutput {
		err := generateHTML(plan, adjustment, wakeTime, targetWakeTime)
		if err != nil {
			fmt.Printf("Error generating HTML: %v\n", err)
			os.Exit(1)
		}
	}

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

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Your Sleep Calibration Plan</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      Helvetica, Arial, sans-serif;
    background-color: #f0f2f5;
    color: #333;
    margin: 0;
    padding: 2rem;
  }
  .container {
    max-width: 800px;
    margin: 2rem auto;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }
  header {
    background-color: #4a5568;
    color: #fff;
    padding: 1.5rem 2rem;
    text-align: center;
  }
  h1 {
    margin: 0;
    font-size: 1.8rem;
    font-weight: 600;
  }
  .summary {
    padding: 1.5rem 2rem;
    background-color: #f7fafc;
    border-bottom: 1px solid #e2e8f0;
    display: flex;
    justify-content: space-around;
    align-items: center;
  }
  .summary-item {
    text-align: center;
  }
  .summary-item span {
    font-weight: 600;
    display: block;
  }
  .chart-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
    padding: 2rem;
  }
  .chart-container {
    padding: 1rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
  }
  table {
    width: 100%;
    border-collapse: collapse;
  }
  th,
  td {
    padding: 1rem 1.5rem;
    text-align: left;
    border-bottom: 1px solid #e2e8f0;
  }
  th {
    background-color: #f7fafc;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #718096;
  }
  tr:last-child td {
    border-bottom: none;
  }
  tr:hover {
    background-color: #f7fafc;
  }
  .emoji {
    font-size: 1.2rem;
    margin-right: 0.5rem;
  }
  .timeline {
    width: 100%;
    height: 20px;
    background-color: #e2e8f0;
    position: relative;
    border-radius: 4px;
    overflow: hidden;
  }
  .sleep-block {
    position: absolute;
    top: 0;
    height: 100%;
    background-color: #4a5568;
  }
</style>
</head>
<body>
  <div class="container">
    <header>
      <h1>Your Sleep Calibration Plan</h1>
    </header>
    <div class="summary">
      <div class="summary-item">
        <span>Days to Target</span>
        <span>{{.DaysToTarget}} days</span>
      </div>
      <div class="summary-item">
        <span>Adjustment per Day</span>
        <span>{{.Adjustment}}</span>
      </div>
      <div class="chart-container" style="width: 150px; height: 150px;">
        <canvas id="progressDonutChart"></canvas>
      </div>
    </div>
    <div class="chart-grid">
      <div class="chart-container">
        <canvas id="wakeUpAndBedtimeChart"></canvas>
      </div>
      <div class="chart-container">
        <canvas id="sleepDurationChart"></canvas>
      </div>
    </div>
    <table>
      <thead>
        <tr>
          <th><span class="emoji">📅</span>Date</th>
          <th><span class="emoji">⏰</span>Wake Up</th>
          <th><span class="emoji">😴</span>Bedtime</th>
          <th><span class="emoji">⏳</span>Duration</th>
          <th><span class="emoji">📊</span>Sleep Period</th>
        </tr>
      </thead>
      <tbody>
        {{range .Schedule}}
        <tr>
          <td>{{.Date}}</td>
          <td>{{.WakeTime}}</td>
          <td>{{.Bedtime}}</td>
          <td>{{.Duration}}</td>
          <td>
            <div class="timeline">
              {{range .SleepBlocks}}
              <div class="sleep-block" style="left: {{.Start}}%; width: {{.Width}}%;"></div>
              {{end}}
            </div>
          </td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
<script>
  const timeToFloat = (timeStr) => {
    const [hours, minutes] = timeStr.split(':').map(Number);
    return hours + minutes / 60;
  };

  const formatTime = (value) => {
    const hours = Math.floor(value);
    const minutes = Math.round((value - hours) * 60);
    return ('0' + hours).slice(-2) + ':' + ('0' + minutes).slice(-2);
  };

  // Progress Donut Chart
  const progressDonutCtx = document.getElementById('progressDonutChart').getContext('2d');
  new Chart(progressDonutCtx, {
    type: 'doughnut',
    data: {
      labels: ['Completed', 'Remaining'],
      datasets: [{
        data: [{{.Progress}}, 100 - {{.Progress}}],
        backgroundColor: ['#4a5568', '#e2e8f0'],
        borderWidth: 0
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      cutout: '70%',
      plugins: {
        legend: { display: false },
        tooltip: { enabled: false }
      }
    }
  });

  // Wake Up and Bedtime Chart
  const wakeUpAndBedtimeCtx = document.getElementById('wakeUpAndBedtimeChart').getContext('2d');
  new Chart(wakeUpAndBedtimeCtx, {
    type: 'line',
    data: {
      labels: {{.ChartLabels}},
      datasets: [
        {
          label: 'Wake Up Time',
          data: {{.WakeUpData}},
          borderColor: '#4a5568',
          backgroundColor: 'rgba(74, 85, 104, 0.2)',
          fill: true,
          tension: 0.1
        },
        {
          label: 'Bedtime',
          data: {{.BedtimeData}},
          borderColor: '#a0aec0',
          backgroundColor: 'rgba(160, 174, 192, 0.2)',
          fill: true,
          tension: 0.1
        }
      ]
    },
    options: {
      scales: {
        y: {
          ticks: { callback: (value) => formatTime(value) }
        }
      }
    }
  });

  // Sleep Duration Chart
  const sleepDurationCtx = document.getElementById('sleepDurationChart').getContext('2d');
  new Chart(sleepDurationCtx, {
    type: 'bar',
    data: {
      labels: {{.ChartLabels}},
      datasets: [{
        label: 'Sleep Duration (hours)',
        data: {{.DurationData}},
        backgroundColor: '#4a5568'
      }]
    },
    options: {
      scales: {
        y: {
          beginAtZero: true
        }
      }
    }
  });
</script>
</body>
</html>
`

type SleepBlock struct {
	Start float64 // percentage
	Width float64 // percentage
}

type ScheduleEntry struct {
	Date        string
	WakeTime    string
	Bedtime     string
	Duration    string
	SleepBlocks []SleepBlock
}

type TemplateData struct {
	DaysToTarget int
	Adjustment   string
	Schedule     []ScheduleEntry
	ChartLabels  []string
	WakeUpData   []float64
	BedtimeData  []float64
	DurationData []float64
	Progress     float64
}

func generateHTML(plan []time.Time, adjustment time.Duration, initialWakeTime, targetWakeTime time.Time) error {
	var schedule []ScheduleEntry
	var chartLabels []string
	var wakeUpData, bedtimeData, durationData []float64

	for _, wakeTime := range plan {
		bedtime := wakeTime.Add(-idealSleepDuration)
		duration := wakeTime.Sub(bedtime)

		var blocks []SleepBlock
		if bedtime.Day() != wakeTime.Day() && bedtime.Location() == wakeTime.Location() {
			start1 := float64(bedtime.Hour()*60+bedtime.Minute()) / 1440 * 100
			width1 := (1440.0 - float64(bedtime.Hour()*60+bedtime.Minute())) / 1440 * 100
			blocks = append(blocks, SleepBlock{Start: start1, Width: width1})
			start2 := 0.0
			width2 := float64(wakeTime.Hour()*60+wakeTime.Minute()) / 1440 * 100
			blocks = append(blocks, SleepBlock{Start: start2, Width: width2})
		} else {
			start := float64(bedtime.Hour()*60+bedtime.Minute()) / 1440 * 100
			width := float64(duration.Minutes()) / 1440 * 100
			blocks = append(blocks, SleepBlock{Start: start, Width: width})
		}

		schedule = append(schedule, ScheduleEntry{
			Date:        wakeTime.Format("Mon, Jan 2"),
			WakeTime:    wakeTime.Format(timeFormat),
			Bedtime:     bedtime.Format(timeFormat),
			Duration:    fmt.Sprintf("%.1f hours", duration.Hours()),
			SleepBlocks: blocks,
		})

		chartLabels = append(chartLabels, "`"+wakeTime.Format("Jan 2")+"`")
		wakeUpData = append(wakeUpData, float64(wakeTime.Hour())+float64(wakeTime.Minute())/60.0)
		bedtimeData = append(bedtimeData, float64(bedtime.Hour())+float64(bedtime.Minute())/60.0)
		durationData = append(durationData, duration.Hours())
	}

	totalAdjustment := initialWakeTime.Sub(targetWakeTime)
	currentAdjustment := initialWakeTime.Sub(plan[len(plan)-1])
	progress := 0.0
	if totalAdjustment > 0 {
		progress = (float64(currentAdjustment) / float64(totalAdjustment)) * 100
	}

	data := TemplateData{
		DaysToTarget: len(plan),
		Adjustment:   adjustment.String(),
		Schedule:     schedule,
		ChartLabels:  chartLabels,
		WakeUpData:   wakeUpData,
		BedtimeData:  bedtimeData,
		DurationData: durationData,
		Progress:     progress,
	}

	tmpl, err := template.New("schedule").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	tmpfile, err := ioutil.TempFile("", "sleep-schedule-*.html")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer tmpfile.Close()

	err = tmpl.Execute(tmpfile, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	fmt.Printf("Generated HTML report: %s\n", tmpfile.Name())

	cmd := exec.Command("xdg-open", tmpfile.Name())
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error opening file with xdg-open: %w", err)
	}

	return nil
}

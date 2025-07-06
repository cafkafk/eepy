/*
 * SPDX-FileCopyrightText: 2025 Christina Sørensen
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

const (
	idealSleepDuration = 9 * time.Hour
	minSleepDuration   = 7*time.Hour + 30*time.Minute
	timeFormat         = "15:04"
)

type Plan struct {
	InitialWakeTime time.Time
	TargetWakeTime  time.Time
	Adjustment      time.Duration
	Schedule        []time.Time
	StartDate       time.Time
}

var (
	configPath string
	historyPath string
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	configDir := filepath.Join(home, ".config", "eepy")
	configPath = filepath.Join(configDir, "plan.json")
	historyPath = filepath.Join(configDir, "history")

	if err := os.MkdirAll(historyPath, 0755); err != nil {
		fmt.Printf("Error creating history directory: %v\n", err)
		os.Exit(1)
	}

	targetWakeTimeStr := pflag.String("target", "05:00", "Your target wake up time (HH:MM)")
	adjustmentStr := pflag.String("adjustment", "1h30m", "Adjustment per day")
	adb := pflag.Bool("adb", false, "Set alarm on Android device via ADB")
	noSkipToday := pflag.Bool("no-skip-today", false, "Do not skip setting an alarm for today")
	htmlOutput := pflag.Bool("html", false, "Generate an HTML visualization of the plan")
	startDateStr := pflag.String("start-date", time.Now().Format("2006-01-02"), "The start date of the plan (YYYY-MM-DD)")
	pflag.Parse()

	startDate, err := time.Parse("2006-01-02", *startDateStr)
	if err != nil {
		fmt.Printf("Error parsing start-date: %v\n", err)
		os.Exit(1)
	}

	existingPlan, err := loadPlan()

	if len(pflag.Args()) == 0 {
		if err != nil {
			fmt.Println("No active sleep plan found. Create one by providing a wake-up time.")
			fmt.Println("Usage: eepy [wake-time] [flags]")
			pflag.PrintDefaults()
			os.Exit(1)
		}
		displayPlan(existingPlan)
		if *htmlOutput {
			if err := generateHTML(existingPlan); err != nil {
				fmt.Printf("Error generating HTML: %v\n", err)
			}
		}
		os.Exit(0)
	}

	wakeTimeStr := pflag.Arg(0)
	wakeTime, err := time.Parse(timeFormat, wakeTimeStr)
	if err != nil {
		fmt.Printf("Error parsing wake-time: %v\n", err)
		os.Exit(1)
	}

	if existingPlan != nil {
		fmt.Print("An active sleep plan already exists. Do you want to override it? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(input)) != "y" {
			fmt.Println("Operation cancelled.")
			os.Exit(0)
		}
		if err := archivePlan(existingPlan); err != nil {
			fmt.Printf("Error archiving existing plan: %v\n", err)
			os.Exit(1)
		}
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

	schedule := generateSchedule(wakeTime, targetWakeTime, adjustment, startDate)
	newPlan := &Plan{
		InitialWakeTime: wakeTime,
		TargetWakeTime:  targetWakeTime,
		Adjustment:      adjustment,
		Schedule:        schedule,
		StartDate:       startDate,
	}

	if err := savePlan(newPlan); err != nil {
		fmt.Printf("Error saving new plan: %v\n", err)
		os.Exit(1)
	}

	displayPlan(newPlan)

	if *htmlOutput {
		if err := generateHTML(newPlan); err != nil {
			fmt.Printf("Error generating HTML: %v\n", err)
		}
	}

	if *adb {
		setAlarms(newPlan.Schedule, *noSkipToday)
	}
}

func generateSchedule(wakeTime, targetWakeTime time.Time, adjustment time.Duration, startDate time.Time) []time.Time {
	var schedule []time.Time
	if !wakeTime.After(targetWakeTime) {
		wakeTimeWithDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), wakeTime.Hour(), wakeTime.Minute(), 0, 0, startDate.Location())
		schedule = append(schedule, wakeTimeWithDate)
		return schedule
	}

	currentWakeTime := wakeTime
	day := 0

	for {
		dayOfPlan := startDate.AddDate(0, 0, day)
		wakeTimeWithDate := time.Date(dayOfPlan.Year(), dayOfPlan.Month(), dayOfPlan.Day(), currentWakeTime.Hour(), currentWakeTime.Minute(), 0, 0, startDate.Location())
		schedule = append(schedule, wakeTimeWithDate)

		if !currentWakeTime.After(targetWakeTime) {
			break
		}

		currentWakeTime = currentWakeTime.Add(-adjustment)
		if currentWakeTime.Before(targetWakeTime) {
			currentWakeTime = targetWakeTime
		}
		day++
	}
	return schedule
}

func displayPlan(p *Plan) {
	fmt.Println("Your sleep calibration plan:")
	fmt.Println("-----------------------------")
	fmt.Printf("Ideal sleep: %.1f hours. Minimum functional sleep: %.1f hours.\n", idealSleepDuration.Hours(), minSleepDuration.Hours())
	fmt.Println("-----------------------------")
	for i, wakeTime := range p.Schedule {
		dayOfPlan := p.StartDate.AddDate(0, 0, i)
		bedtime := wakeTime.Add(-idealSleepDuration)
		fmt.Printf("%s (Day %d):\n", dayOfPlan.Format("Mon, Jan 2"), i+1)
		fmt.Printf("  - Wake up at %s\n", wakeTime.Format(timeFormat))
		fmt.Printf("  - Go to bed at %s\n", bedtime.Format(timeFormat))
	}
	fmt.Println("-----------------------------")
	if !p.Schedule[len(p.Schedule)-1].After(p.TargetWakeTime) {
		fmt.Println("You have reached your target sleep schedule!")
	}
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

func savePlan(p *Plan) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configPath, data, 0644)
}

func loadPlan() (*Plan, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var p Plan
	err = json.Unmarshal(data, &p)
	return &p, err
}

func archivePlan(p *Plan) error {
	files, err := ioutil.ReadDir(historyPath)
	if err != nil {
		return err
	}
	newFileName := fmt.Sprintf("plan-%d.json", len(files)+1)
	newPath := filepath.Join(historyPath, newFileName)
	return os.Rename(configPath, newPath)
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
  .donut-chart-container {
    position: relative;
    width: 150px;
    height: 150px;
    text-align: center;
  }
  .donut-chart-label {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 24px;
    font-weight: bold;
  }
  .donut-chart-sub-label {
    font-size: 12px;
    color: #718096;
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
      <div class="donut-chart-container">
        <canvas id="progressDonutChart"></canvas>
        <div class="donut-chart-label">
          {{printf "%.0f%%" .Progress}}
          <div class="donut-chart-sub-label">Progress</div>
        </div>
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

func generateHTML(p *Plan) error {
	var schedule []ScheduleEntry
	var chartLabels []string
	var wakeUpData, bedtimeData, durationData []float64

	for _, wakeTime := range p.Schedule {
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

	now := time.Now()
	var currentScheduledWakeTime time.Time
	for _, wakeTime := range p.Schedule {
		if !wakeTime.After(now) {
			currentScheduledWakeTime = wakeTime
		} else {
			break
		}
	}

	if currentScheduledWakeTime.IsZero() {
		currentScheduledWakeTime = p.Schedule[0]
	}

	commonDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	initialTime := time.Date(commonDate.Year(), commonDate.Month(), commonDate.Day(), p.InitialWakeTime.Hour(), p.InitialWakeTime.Minute(), 0, 0, time.UTC)
	targetTime := time.Date(commonDate.Year(), commonDate.Month(), commonDate.Day(), p.TargetWakeTime.Hour(), p.TargetWakeTime.Minute(), 0, 0, time.UTC)
	currentTime := time.Date(commonDate.Year(), commonDate.Month(), commonDate.Day(), currentScheduledWakeTime.Hour(), currentScheduledWakeTime.Minute(), 0, 0, time.UTC)

	totalAdjustment := initialTime.Sub(targetTime)
	adjustedSoFar := initialTime.Sub(currentTime)

	progress := 0.0
	if totalAdjustment > 0 {
		progress = (float64(adjustedSoFar) / float64(totalAdjustment)) * 100
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	data := TemplateData{
		DaysToTarget: len(p.Schedule),
		Adjustment:   p.Adjustment.String(),
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
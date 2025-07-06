<!--
SPDX-FileCopyrightText: 2025 Christina Sørensen

SPDX-License-Identifier: EUPL-1.2
-->

# eepy

`eepy` is a simple command-line tool to help you calibrate your sleep schedule.

It generates a day-by-day plan to gradually adjust your wake-up time to meet your target, ensuring you get an ideal amount of sleep each night.

## Usage

```bash
eepy [your-current-wake-time] [flags]
```

### Arguments

-   `your-current-wake-time`: Your current wake-up time in HH:MM format.

### Flags

-   `--target`: Your target wake-up time in HH:MM format (default: "05:00").
-   `--adjustment`: The amount of time to adjust your wake-up time by each day (default: "1h30m").

## Example

If you currently wake up at 10:00 and want to start waking up at 05:00, you can run:

```bash
eepy 10:00 --target 05:00
```

`eepy` will then print a plan for you to follow.

## Installation

You can build `eepy` from source:

```bash
go build
```

#  Register
## Description

**Register** is an easy-to-use console app for managing attendant table located in google sheet.
Since it supports only one command `backfill` for filling table with random start and finish times.
**Register** generates values based on yaml-config with the following structure:
```yaml
# file with holidays
holidays_file: supplementary/uk_holidays.txt
# file with vacations
vacations_file: supplementary/vacations.txt

credentials: <path to json with google service account key>
spread_sheet_id: ...
sheet_name: ...
table_date_layout: <time layout used compatible with go time.Parse function>

# meaning of columns in sheet
columns:
  date: "Date"
  started: "Started"
  finished: "Finished"
  comment: "Comment"
  duration: "Hours"
  month: "Month"

# a set to started time values choose from
started_options:
- "10:00:00"
- "10:30:00"
- "11:00:00"
- "11:30:00"

# length of working day
workday_duration: "8h"
```

## Build
**Register** can be compiled with go (minimum version 1.18) via command
```bash
go build -o bin/register
```

## Usage
```bash
bin/register backfill --start <your start date> --end <your end date>
```
This generates attendance records between start and end dates (end date is not included).
The provided left-closed interval will be checked on overlapping with data presented in sheets
and in case of collision the error is thrown. A default value for `start` flag is the day after maximum
date in google sheet, default for `end` is today. 

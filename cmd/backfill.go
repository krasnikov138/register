package cmd

import (
	"log"
	"time"

	"github.com/krasnikov138/register/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func maxTime(values []time.Time) time.Time {
	var max time.Time
	if len(values) == 0 {
		return max
	}

	max = values[0]
	for _, v := range values[1:] {
		if v.Compare(max) > 0 {
			max = v
		}
	}
	return max
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func areDatesOverlapped(dates []time.Time, sheetDates []time.Time) bool {
	dm := make(map[time.Time]bool, len(dates))

	for _, date := range dates {
		dm[date] = true
	}

	for _, date := range sheetDates {
		if dm[date] {
			return true
		}
	}
	return false
}

func getStartedOptions(options []string) []time.Time {
	parsed := make([]time.Time, len(options))

	var err error
	for i, opt := range options {
		parsed[i], err = time.Parse("15:04:05", opt)
		if err != nil {
			log.Fatalf("Can not parse time '%s' started option", opt)
		}
	}
	return parsed
}

// backfillCmd represents the backfill command
var backfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Backfill google table with random working day durations.",
	Run: func(cmd *cobra.Command, args []string) {
		// Use Viper to get the value of the "config" flag
		initConfig(viper.GetString("config"))

		srv, err := app.CreateSheetsService(viper.GetString("credentials"))
		if err != nil {
			log.Fatal(err)
		}

		SpreadSheetID := viper.GetString("spread_sheet_id")
		SheetName := viper.GetString("sheet_name")

		table, err := app.GetSheet(srv, SpreadSheetID, SheetName)
		if err != nil {
			log.Fatalf("Unable to retrieve Google Sheet table: %v", err)
		}
		checkColumns(table)

		formats, err := app.GetCellsFormatting(srv, SpreadSheetID, SheetName)
		if err != nil {
			log.Printf("WARNING: unable to retrieve cells formats from sheets: %v", err)
		}

		sheetDates, err := app.ParseDateSlice(table.GetColumn("Date"), viper.GetString("table_date_layout"))
		if err != nil {
			log.Fatalf("Can not parse Date column in sheet '%s': %s", SheetName, err)
		}

		startDate := parseCmdDate(cmd, "start", maxTime(sheetDates).AddDate(0, 0, 1))
		endDate := parseCmdDate(cmd, "end", truncateToDay(time.Now()))

		log.Printf("Start backfilling google sheet '%s' with parameters from %s till %s\n",
			SheetName, startDate.Format(app.DateLayout), endDate.Format(app.DateLayout))

		holidays := app.ReadDates(viper.GetString("holidays_file"))
		dates := app.GenerateWorkingDays(startDate, endDate, holidays)

		if areDatesOverlapped(dates, sheetDates) {
			log.Fatal("Overlapping with existing in shreadsheet values is detected. Check input date range.")
		}

		vacations := []time.Time{}
		VacationsFile := viper.GetString("vacations_file")
		if len(VacationsFile) != 0 {
			log.Printf("Use vacations file: %v\n", VacationsFile)
			vacations = app.ReadDates(VacationsFile)
		} else {
			log.Println("Vacations file is not used")
		}

		workdayDuration, err := time.ParseDuration(viper.GetString("workday_duration"))
		if err != nil {
			log.Fatal("Can not parse workday duration in config file")
		}

		records := app.GenerateTable(
			dates,
			vacations,
			getStartedOptions(viper.GetStringSlice("started_options")),
			workdayDuration,
			table.Columns,
			app.ColumnNames{
				Date:     viper.GetString("columns.date"),
				Started:  viper.GetString("columns.started"),
				Finished: viper.GetString("columns.finished"),
				Duration: viper.GetString("columns.duration"),
				Comment:  viper.GetString("columns.comment"),
				Month:    viper.GetString("columns.month"),
			},
		)

		err = app.AppendCells(srv, SpreadSheetID, SheetName, records.Values, formats)
		if err != nil {
			log.Fatalf("Unable to insert new records to google sheet: %s", err)
		}

		if recslen := records.NRows(); recslen != 0 {
			log.Printf("%d new records were successfully added to google sheet\n", recslen)
		} else {
			log.Println("No new records were added")
		}
	},
}

func init() {
	rootCmd.AddCommand(backfillCmd)

	backfillCmd.Flags().StringP("start", "s", "", "Backfilling start date (last date from sheet is default)")
	backfillCmd.Flags().StringP("end", "e", "", "Backfilling end date (today is default)")
}

package cmd

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/krasnikov138/register/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printTable(table *app.Table[string], stream io.Writer) error {
	writer := csv.NewWriter(stream)
	err := writer.Write(table.Columns)
	if err != nil {
		return err
	}

	row := make([]string, table.NCols())
	for i := 0; i < table.NRows(); i += 1 {
		for j := 0; j < table.NCols(); j += 1 {
			row[j] = table.Values[j][i]
		}

		err = writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump content of the google sheet into console or csv file",
	Run: func(cmd *cobra.Command, args []string) {
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

		out, _ := cmd.Flags().GetString("output")

		var stream io.Writer
		if len(out) != 0 {
			f, err := os.Create(out)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			stream = bufio.NewWriter(f)
		} else {
			stream = os.Stdout
		}

		err = printTable(table, stream)
		if err != nil {
			log.Fatalf("Can not dump spread sheet: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringP("output", "o", "", "output file (stdout if not provided)")
}

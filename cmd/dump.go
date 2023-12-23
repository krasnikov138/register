package cmd

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/krasnikov138/register/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func tableCsvPrinter(table *app.Table[string], stream io.Writer) error {
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

func justify(str string, targetLength int, builder *strings.Builder) {
	padlen := targetLength - utf8.RuneCountInString(str)

	builder.Grow(padlen + len(str))

	for i := 0; i < padlen; i += 1 {
		builder.WriteRune(' ')
	}
	builder.Write([]byte(str))
}

func prettyRow(row []string, targetLengths []int, builder *strings.Builder) []byte {
	builder.WriteRune('|')
	for i, word := range row {
		justify(word, targetLengths[i], builder)
		builder.WriteRune('|')
	}
	builder.WriteRune('\n')

	result := builder.String()
	builder.Reset()

	return []byte(result)
}

func tablePrettyPrint(table *app.Table[string], stream io.Writer) error {
	lengths := make([]int, table.NCols())

	for i, col := range table.Columns {
		lengths[i] = utf8.RuneCountInString(col)
	}

	for i := 0; i < table.NRows(); i += 1 {
		for j, vals := range table.Values {
			strlen := utf8.RuneCountInString(vals[i])

			if strlen > lengths[j] {
				lengths[j] = strlen
			}
		}
	}

	rowSize := table.NCols() + 1
	for _, l := range lengths {
		rowSize += l
	}

	var builder strings.Builder
	builder.Grow(rowSize + 1)

	// display table header
	stream.Write([]byte(strings.Repeat("-", rowSize) + "\n"))
	stream.Write(prettyRow(table.Columns, lengths, &builder))
	stream.Write([]byte(strings.Repeat("-", rowSize) + "\n"))

	row := make([]string, table.NCols())
	for i := 0; i < table.NRows(); i += 1 {
		stream.Write([]byte(strings.Repeat("-", rowSize) + "\n"))

		for c := 0; c < table.NCols(); c += 1 {
			row[c] = table.Values[c][i]
		}

		stream.Write(prettyRow(row, lengths, &builder))
	}

	stream.Write([]byte(strings.Repeat("-", rowSize) + "\n"))

	return nil
}

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump content of the google sheet into console or csv file",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig(viper.GetString("config"))

		// add your formatter here
		formatters := map[string]func(*app.Table[string], io.Writer) error{
			"pretty": tablePrettyPrint,
			"csv":    tableCsvPrinter,
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, ok := formatters[format]
		if !ok {
			log.Fatalf("Wrong output format is provided: %s", format)
		}

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

		var file *os.File
		if len(out) != 0 {
			file, err = os.Create(out)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
		} else {
			file = os.Stdout
		}

		stream := bufio.NewWriter(file)
		defer stream.Flush()

		err = formatter(table, stream)
		if err != nil {
			log.Fatalf("Can not dump spread sheet: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringP("output", "o", "", "output file (stdout if not provided)")
	dumpCmd.Flags().StringP("format", "f", "pretty", "format for data representation (csv or pretty)")
}

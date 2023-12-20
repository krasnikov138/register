package app

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func CreateSheetsService(credentials string) (*sheets.Service, error) {
	creds, err := os.ReadFile(credentials)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(creds, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to create JWT config: %v", err)
	}

	client := config.Client(context.Background())
	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to create Google Sheets service: %v", err)
	}

	return srv, nil
}

func GetSheet(srv *sheets.Service, spreadsheetID, sheetName string) (*Table[string], error) {
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, sheetName+"!A:F").ValueRenderOption("FORMATTED_VALUE").Do()
	if err != nil {
		return nil, err
	}

	data := resp.Values
	var table *Table[string]

	if len(data) != 0 {
		table = NewTable[string](len(data)-1, len(data[0]))

		// get column names
		for i, col := range data[0] {
			table.Columns[i] = col.(string)
		}

		// get rest values
		for i, row := range data[1:] {
			for j, val := range row {
				table.Values[j][i] = val.(string)
			}
		}
	} else {
		table = NewEmptyTable[string]()
	}

	return table, nil
}

func GetCellsFormatting(srv *sheets.Service, spreadsheetID, sheetName string) ([]*sheets.CellFormat, error) {
	resp, err := srv.Spreadsheets.Get(spreadsheetID).Ranges(sheetName + "!A2:F3").IncludeGridData(true).Do()
	if err != nil {
		return nil, err
	}

	var formats []*sheets.CellFormat

	data := resp.Sheets[0].Data[0].RowData
	if len(data) > 0 {
		formats = make([]*sheets.CellFormat, len(data[0].Values))
		for i, val := range data[0].Values {
			formats[i] = val.UserEnteredFormat
		}
	}

	return formats, nil
}

func getFormat(formats []*sheets.CellFormat, idx int) *sheets.CellFormat {
	if formats == nil {
		return nil
	}
	return formats[idx]
}

func getExtendedValue(value interface{}) *sheets.ExtendedValue {
	switch v := value.(type) {
	case float64:
		return &sheets.ExtendedValue{
			NumberValue: &v,
		}
	case string:
		return &sheets.ExtendedValue{
			StringValue: &v,
		}
	}
	return nil
}

func PrepareCells(records [][]interface{}, formats []*sheets.CellFormat) []*sheets.RowData {
	rows := make([]*sheets.RowData, len(records[0]))

	for i := range rows {
		rows[i] = &sheets.RowData{
			Values: make([]*sheets.CellData, len(records)),
		}

		for j := range records {
			rows[i].Values[j] = &sheets.CellData{
				UserEnteredValue:  getExtendedValue(records[j][i]),
				UserEnteredFormat: getFormat(formats, j),
			}
		}
	}

	return rows
}

func GetSheetIDBySheetName(srv *sheets.Service, spreadsheetID, sheetName string) (int64, error) {
	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return 0, err
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet.Properties.SheetId, nil
		}
	}

	return 0, fmt.Errorf("Sheet with name %s not found", sheetName)
}

func AppendCells(srv *sheets.Service, spreadsheetID, sheetName string, records [][]interface{}, formats []*sheets.CellFormat) error {
	sheetID, err := GetSheetIDBySheetName(srv, spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("Can not get sheet id for '%s': %v", sheetName, err)
	}

	batchUpdateRequest := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AppendCells: &sheets.AppendCellsRequest{
					Fields:  "*",
					Rows:    PrepareCells(records, formats),
					SheetId: sheetID,
				},
			},
		},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, &batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("Can not perform spread sheet append: %v", err)
	}

	return nil
}

package service

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsService wraps Google Sheets API v4 operations for multi-sheet sync.
type SheetsService struct {
	srv           *sheets.Service
	spreadsheetID string
}

// NewSheetsService initializes the Google Sheets client using service account credentials JSON.
func NewSheetsService(credentialsJSON, spreadsheetID string) (*SheetsService, error) {
	ctx := context.Background()

	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
	if err != nil {
		return nil, fmt.Errorf("unable to create sheets service: %w", err)
	}

	return &SheetsService{
		srv:           srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

// UpsertRow finds a row by ID (column A) in the given sheet and updates it, or appends a new row.
func (s *SheetsService) UpsertRow(sheetName string, colRange string, id string, values []interface{}) error {
	rowIndex, err := s.FindRowByID(sheetName, id)
	if err != nil {
		return err
	}

	if rowIndex == -1 {
		return s.AppendRow(sheetName, colRange, values)
	}

	return s.UpdateRow(sheetName, colRange, rowIndex, values)
}

// FindRowByID scans the ID column (A) of the specified sheet to find the row index.
// Returns -1 if not found.
func (s *SheetsService) FindRowByID(sheetName string, id string) (int, error) {
	readRange := fmt.Sprintf("'%s'!A:A", sheetName)

	resp, err := s.srv.Spreadsheets.Values.Get(s.spreadsheetID, readRange).Do()
	if err != nil {
		return -1, fmt.Errorf("unable to read ID column from '%s': %w", sheetName, err)
	}

	for i, row := range resp.Values {
		if len(row) > 0 && fmt.Sprintf("%v", row[0]) == id {
			return i + 1, nil // 1-indexed (row 1 is header)
		}
	}

	return -1, nil
}

// AppendRow adds a new row at the end of the specified sheet.
func (s *SheetsService) AppendRow(sheetName string, colRange string, values []interface{}) error {
	writeRange := fmt.Sprintf("'%s'!%s", sheetName, colRange)

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := s.srv.Spreadsheets.Values.Append(
		s.spreadsheetID,
		writeRange,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("unable to append row to '%s': %w", sheetName, err)
	}

	log.Printf("[SheetsService] Appended new row to '%s' for ID %s", sheetName, values[0])
	return nil
}

// UpdateRow updates an existing row at the given 1-indexed row number.
func (s *SheetsService) UpdateRow(sheetName string, colRange string, rowIndex int, values []interface{}) error {
	// Parse colRange to get start and end columns (e.g. "A:L" -> row range "A{n}:L{n}")
	cols := fmt.Sprintf("'%s'!%s", sheetName, colRange)
	// We need to construct the range with the row number
	// colRange is like "A:L", we need "A{rowIndex}:L{rowIndex}"
	startCol := colRange[0:1]
	endCol := colRange[len(colRange)-1:]
	writeRange := fmt.Sprintf("'%s'!%s%d:%s%d", sheetName, startCol, rowIndex, endCol, rowIndex)

	_ = cols // unused, just for clarity

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := s.srv.Spreadsheets.Values.Update(
		s.spreadsheetID,
		writeRange,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("unable to update row %d in '%s': %w", rowIndex, sheetName, err)
	}

	log.Printf("[SheetsService] Updated row %d in '%s' for ID %s", rowIndex, sheetName, values[0])
	return nil
}

// BatchRead reads all data rows from the specified sheet (skips header row).
func (s *SheetsService) BatchRead(sheetName string, colRange string) ([][]interface{}, error) {
	// colRange is like "A:L", we start at row 2 to skip header
	startCol := colRange[0:1]
	endCol := colRange[len(colRange)-1:]
	readRange := fmt.Sprintf("'%s'!%s2:%s", sheetName, startCol, endCol)

	resp, err := s.srv.Spreadsheets.Values.Get(s.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to read sheet data from '%s': %w", sheetName, err)
	}

	return resp.Values, nil
}

// BatchWrite writes all data to the specified sheet, overwriting existing content (starting from row 2).
// It first writes the header row, then appends all data rows.
func (s *SheetsService) BatchWrite(sheetName string, colRange string, headers []string, rows [][]interface{}) error {
	startCol := colRange[0:1]
	endCol := colRange[len(colRange)-1:]

	// Write header
	headerRange := fmt.Sprintf("'%s'!%s1:%s1", sheetName, startCol, endCol)
	headerValues := make([]interface{}, len(headers))
	for i, h := range headers {
		headerValues[i] = h
	}

	_, err := s.srv.Spreadsheets.Values.Update(
		s.spreadsheetID,
		headerRange,
		&sheets.ValueRange{Values: [][]interface{}{headerValues}},
	).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("unable to write headers to '%s': %w", sheetName, err)
	}

	if len(rows) == 0 {
		return nil
	}

	// Write data rows starting from row 2
	dataRange := fmt.Sprintf("'%s'!%s2:%s", sheetName, startCol, endCol)
	_, err = s.srv.Spreadsheets.Values.Update(
		s.spreadsheetID,
		dataRange,
		&sheets.ValueRange{Values: rows},
	).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("unable to write data to '%s': %w", sheetName, err)
	}

	log.Printf("[SheetsService] Batch wrote %d rows to '%s'", len(rows), sheetName)
	return nil
}

// BatchReadWithHeaders reads all rows from the specified sheet INCLUDING the header row.
// Returns [][]string where index 0 is header, 1+ is data.
func (s *SheetsService) BatchReadWithHeaders(sheetName string) ([][]string, error) {
	readRange := fmt.Sprintf("'%s'!A1:Z1000", sheetName)

	resp, err := s.srv.Spreadsheets.Values.Get(s.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to read sheet data from '%s': %w", sheetName, err)
	}

	var result [][]string
	for _, row := range resp.Values {
		strRow := make([]string, len(row))
		for i, cell := range row {
			strRow[i] = fmt.Sprintf("%v", cell)
		}
		result = append(result, strRow)
	}

	return result, nil
}

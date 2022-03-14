package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/coder/flog"
	"github.com/samber/lo"
)

type engine struct {
	Log       *flog.Logger
	Enrichers []enricher
}

// Run is the main enrichment loop
func (eng engine) Run(w io.Writer, r io.Reader) error {
	var (
		csvReader = csv.NewReader(r)
		csvWriter = csv.NewWriter(w)
	)

	inputHeader, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	// Form output header from input header and additional possible enrichers
	outputHeader := append([]string(nil), inputHeader...)
	for _, enricher := range eng.Enrichers {
		if lo.Contains(inputHeader, enricher.FieldName) {
			return fmt.Errorf("enrich column %q already exists in input", enricher.FieldName)
		}
		// Fail if dependencies don't exist
		for _, dep := range enricher.Deps {
			if !lo.Contains(inputHeader, dep) && !lo.Contains(outputHeader, dep) {
				return fmt.Errorf("enrich column %v has unmet dependency %q (try rearranging?)", enricher.FieldName, dep)
			}
		}
		outputHeader = append(outputHeader, enricher.FieldName)
	}

	err = csvWriter.Write(outputHeader)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	csvWriter.Flush()

	for i := 0; ; i++ {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read row: %w", err)
		}
		row, err = eng.processRow(row, inputHeader)
		if err != nil {
			return err
		}
		err = csvWriter.Write(row)
		if err != nil {
			return fmt.Errorf("write row: %w", err)
		}
		csvWriter.Flush()
	}
}

func (eng *engine) processRow(row []string, inputHeader []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	rowMap := make(map[string]string, len(row))
	for i, v := range row {
		rowMap[inputHeader[i]] = v
	}
	for _, e := range eng.Enrichers {
		v, err := e.Run(ctx, rowMap)
		row = append(row, v)
		rowMap[e.FieldName] = v
		if err != nil {
			eng.Log.Error("%q enrich failed: %+v", e.FieldName, err)
		}
	}
	return row, nil
}

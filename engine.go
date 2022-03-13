package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/coder/flog"
	"github.com/go-redis/redis/v8"
	"github.com/samber/lo"
)

type engine struct {
	Log   *flog.Logger
	Redis *redis.Client
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
	var usedEnrichers []enricher
findEnrichers:
	for _, enricher := range enrichers {
		if lo.Contains(inputHeader, enricher.FieldName) {
			continue
		}
		// Break if dependencies don't exist in input
		for _, dep := range enricher.Deps {
			if !lo.Contains(inputHeader, dep) {
				continue findEnrichers
			}
		}
		outputHeader = append(outputHeader, enricher.FieldName)
		usedEnrichers = append(usedEnrichers, eng.cachedEnricher(enricher))
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
		row, err = eng.processRow(row, inputHeader, usedEnrichers)
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

func (eng *engine) processRow(row []string, inputHeader []string, usedEnrichers []enricher) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	rowMap := make(map[string]string, len(row))
	for i, v := range row {
		rowMap[inputHeader[i]] = v
	}
	for _, e := range usedEnrichers {
		v, err := e.Run(ctx, rowMap)
		row = append(row, v)
		if err != nil {
			eng.Log.Error("%q enrich failed: %+v", e.FieldName, err)
		}
	}
	return row, nil
}

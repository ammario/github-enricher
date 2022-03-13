package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/coder/flog"
)

type enricherEngine struct {
	Log *flog.Logger
}

// An enricher modifies the row before output.
type enricher struct {
	FieldName string
	Run       func(ctx context.Context, row map[string]string) (string, error)
}

func (e enricherEngine) Run(w io.Writer, r io.Reader) error {
	var (
		csvReader = csv.NewReader(r)
		csvWriter = csv.NewWriter(w)
	)

	inputHeader, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	outputHeader := append([]string(nil), inputHeader...)
	err = csvWriter.Write(outputHeader)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for i := 0; ; i++ {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read row: %w", err)
		}
		err = csvWriter.Write(row)
		if err != nil {
			return fmt.Errorf("write row: %w", err)
		}
		csvWriter.Flush()
	}
}

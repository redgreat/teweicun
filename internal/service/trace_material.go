package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/db"
)

type TraceMaterialResult struct {
	SerialInfo *db.SerialCodeInfo     `json:"serial_info,omitempty"`
	Traces     []db.SerialTraceRecord `json:"traces"`
}

func TraceMaterialBySerial(ctx context.Context, serialCode string) (*TraceMaterialResult, error) {
	info, traces, err := db.QueryTraceBySerial(ctx, serialCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &TraceMaterialResult{Traces: []db.SerialTraceRecord{}}, nil
		}
		return nil, err
	}

	return &TraceMaterialResult{
		SerialInfo: info,
		Traces:     traces,
	}, nil
}


/**
 * 功能：trace.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func TraceForward(ctx context.Context, key string) ([]response.TraceForwardResp, error) {
	return db.TraceForward(ctx, key)
}

func TraceBackward(ctx context.Context, usageDesc string) ([]response.TraceBackwardResp, error) {
	return db.TraceBackward(ctx, usageDesc)
}

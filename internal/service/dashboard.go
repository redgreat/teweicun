package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
)

func GetBigscreenDashboard(ctx context.Context, rangeKey string) (*response.DashboardBigscreenResp, error) {
	if err := db.ValidateDashboardRange(rangeKey); err != nil {
		return nil, errcode.NewAppError(errcode.ErrInvalidParam.Code, "range 参数仅支持 7d/30d/mtd", errcode.ErrInvalidParam.HTTPCode)
	}
	return db.QueryDashboardBigscreen(ctx, rangeKey)
}


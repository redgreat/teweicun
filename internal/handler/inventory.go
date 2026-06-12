/**
 * 功能：inventory.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
	"github.com/xuri/excelize/v2"
)

func bindInventoryMaterialLedgerQuery(c *gin.Context, q *request.InventoryMaterialLedgerQuery) error {
	values := c.Request.URL.Query()
	if strings.TrimSpace(values.Get("material_name")) == "" && strings.TrimSpace(values.Get("sku_name")) != "" {
		values.Set("material_name", strings.TrimSpace(values.Get("sku_name")))
		c.Request.URL.RawQuery = values.Encode()
	}
	return c.ShouldBindQuery(q)
}

func ListInventoryDetail(c *gin.Context) {
	var q request.InventoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, err := service.ListInventoryDetail(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func ListInventorySummary(c *gin.Context) {
	var q request.InventoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, err := service.ListInventorySummary(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func ListInventoryAvailable(c *gin.Context) {
	var q request.InventoryAvailableQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, err := service.ListInventoryAvailable(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func ListInventoryIssued(c *gin.Context) {
	var q request.InventoryIssuedQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, err := service.ListInventoryIssued(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func ListInventoryMaterialLedger(c *gin.Context) {
	var q request.InventoryMaterialLedgerQuery
	if err := bindInventoryMaterialLedgerQuery(c, &q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, stats, err := service.ListInventoryMaterialLedger(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  q.Page,
		"size":  q.PageSize,
		"stats": stats,
	})
}

func ListInventorySKULedger(c *gin.Context) {
	ListInventoryMaterialLedger(c)
}

func ListInventoryMaterialLedgerSerials(c *gin.Context) {
	var q request.InventoryMaterialLedgerSerialQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	list, err := service.ListInventoryMaterialLedgerSerials(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

func ListInventorySKUSerials(c *gin.Context) {
	ListInventoryMaterialLedgerSerials(c)
}

func ExportInventoryMaterialLedger(c *gin.Context) {
	var q request.InventoryMaterialLedgerQuery
	if err := bindInventoryMaterialLedgerQuery(c, &q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	rows, stats, err := service.ExportInventoryMaterialLedger(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}

	f := excelize.NewFile()
	sheet := "库存台账"
	f.SetSheetName("Sheet1", sheet)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "#0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#DBEAFE"}, Pattern: 1},
	})
	subTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#4B5563"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F8FAFC"}, Pattern: 1},
	})
	summaryLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "#334155"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FEF3C7"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#F59E0B", Style: 1},
			{Type: "right", Color: "#F59E0B", Style: 1},
			{Type: "top", Color: "#F59E0B", Style: 1},
			{Type: "bottom", Color: "#F59E0B", Style: 1},
		},
	})
	summaryValueStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &[]string{"¥#,##0.00"}[0],
		Font:         &excelize.Font{Bold: true, Size: 11, Color: "#065F46"},
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"#ECFCCB"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#84CC16", Style: 1},
			{Type: "right", Color: "#84CC16", Style: 1},
			{Type: "top", Color: "#84CC16", Style: 1},
			{Type: "bottom", Color: "#84CC16", Style: 1},
		},
	})
	summaryQtyValueStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &[]string{"#,##0.000"}[0],
		Font:         &excelize.Font{Bold: true, Size: 11, Color: "#1E40AF"},
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"#E0F2FE"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#38BDF8", Style: 1},
			{Type: "right", Color: "#38BDF8", Style: 1},
			{Type: "top", Color: "#38BDF8", Style: 1},
			{Type: "bottom", Color: "#38BDF8", Style: 1},
		},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#0F766E"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#D1D5DB", Style: 1},
			{Type: "right", Color: "#D1D5DB", Style: 1},
			{Type: "top", Color: "#D1D5DB", Style: 1},
			{Type: "bottom", Color: "#D1D5DB", Style: 1},
		},
	})
	bodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})
	numberStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &[]string{"#,##0.000"}[0],
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &[]string{"¥#,##0.00"}[0],
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})
	countStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &[]string{"#,##0"}[0],
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})

	_ = f.MergeCell(sheet, "A1", "M1")
	_ = f.SetCellValue(sheet, "A1", "库存台账报表")
	_ = f.SetCellStyle(sheet, "A1", "M1", titleStyle)

	_ = f.MergeCell(sheet, "A2", "M2")
	_ = f.SetCellValue(sheet, "A2",
		fmt.Sprintf("导出时间：%s  |  筛选：物料[%s] 仓库[%s] 价格区间[%.2f - %.2f]",
			time.Now().Format("2006-01-02 15:04:05"), q.MaterialName, q.WarehouseName, q.PriceMin, q.PriceMax))
	_ = f.SetCellStyle(sheet, "A2", "M2", subTitleStyle)

	_ = f.MergeCell(sheet, "A3", "B3")
	_ = f.SetCellValue(sheet, "A3", "在库总金额")
	_ = f.SetCellStyle(sheet, "A3", "B3", summaryLabelStyle)
	_ = f.MergeCell(sheet, "C3", "D3")
	_ = f.SetCellValue(sheet, "C3", stats.TotalAmount)
	_ = f.SetCellStyle(sheet, "C3", "D3", summaryValueStyle)

	_ = f.MergeCell(sheet, "E3", "F3")
	_ = f.SetCellValue(sheet, "E3", "编码物料总金额")
	_ = f.SetCellStyle(sheet, "E3", "F3", summaryLabelStyle)
	_ = f.MergeCell(sheet, "G3", "H3")
	_ = f.SetCellValue(sheet, "G3", stats.CodeTotalAmount)
	_ = f.SetCellStyle(sheet, "G3", "H3", summaryValueStyle)

	_ = f.SetCellValue(sheet, "I3", "无编码金额")
	_ = f.SetCellStyle(sheet, "I3", "I3", summaryLabelStyle)
	_ = f.SetCellValue(sheet, "J3", stats.NoCodeTotalAmount)
	_ = f.SetCellStyle(sheet, "J3", "J3", summaryValueStyle)
	_ = f.SetCellValue(sheet, "K3", fmt.Sprintf("总锁定数量：%.3f", stats.TotalLockedQty))
	_ = f.SetCellStyle(sheet, "K3", "K3", summaryQtyValueStyle)

	headers := []string{"物料名称", "所在仓库", "是否编码", "账面库存", "锁定数量", "在途数量", "编码备货数", "可用数量", "单价", "总价", "库存批次数", "属性", "编码"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 5)
		_ = f.SetCellValue(sheet, cell, h)
	}
	_ = f.SetCellStyle(sheet, "A5", "M5", headerStyle)

	for idx, r := range rows {
		rowNo := idx + 6
		values := []interface{}{
			r.MaterialName,
			r.WarehouseName,
			map[bool]string{true: "有编码", false: "无编码"}[r.IsCode],
			r.BookQuantity,
			r.LockedQuantity,
			r.InTransitQuantity,
			r.SerialReservedQuantity,
			r.Quantity,
			r.UnitCost,
			r.TotalAmount,
			r.InventoryCount,
			"查看属性",
			map[bool]string{true: "查看编码", false: "-"}[r.IsCode],
		}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNo)
			_ = f.SetCellValue(sheet, cell, v)
		}
		_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNo), fmt.Sprintf("C%d", rowNo), bodyStyle)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("D%d", rowNo), fmt.Sprintf("H%d", rowNo), numberStyle)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("I%d", rowNo), fmt.Sprintf("J%d", rowNo), moneyStyle)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("K%d", rowNo), fmt.Sprintf("K%d", rowNo), countStyle)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("L%d", rowNo), fmt.Sprintf("M%d", rowNo), bodyStyle)
	}

	_ = f.SetColWidth(sheet, "A", "A", 30)
	_ = f.SetColWidth(sheet, "B", "B", 14)
	_ = f.SetColWidth(sheet, "C", "C", 10)
	_ = f.SetColWidth(sheet, "D", "H", 12)
	_ = f.SetColWidth(sheet, "I", "J", 12)
	_ = f.SetColWidth(sheet, "K", "K", 12)
	_ = f.SetColWidth(sheet, "L", "M", 10)
	_ = f.SetRowHeight(sheet, 1, 30)
	_ = f.SetRowHeight(sheet, 2, 22)
	_ = f.SetRowHeight(sheet, 3, 22)
	_ = f.SetRowHeight(sheet, 5, 24)
	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 5, TopLeftCell: "A6", ActivePane: "bottomLeft"})
	orientation := "landscape"
	paperSize := 9
	fitToWidth := 1
	fitToHeight := 0
	_ = f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
		Orientation: &orientation,
		Size:        &paperSize,
		FitToWidth:  &fitToWidth,
		FitToHeight: &fitToHeight,
	})
	left, right := 0.3, 0.3
	top, bottom := 0.6, 0.6
	header, footer := 0.2, 0.2
	_ = f.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{
		Left:   &left,
		Right:  &right,
		Top:    &top,
		Bottom: &bottom,
		Header: &header,
		Footer: &footer,
	})

	filename := fmt.Sprintf("库存台账_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", filename))
	c.Header("Cache-Control", "no-cache")
	if err := f.Write(c.Writer); err != nil {
		response.Error(c, err)
		return
	}
}

func ExportInventorySKULedger(c *gin.Context) {
	ExportInventoryMaterialLedger(c)
}

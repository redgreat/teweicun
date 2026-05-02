package request

type StockOutSerialSelectionItem struct {
	StockOutItemID int64   `json:"stock_out_item_id" binding:"required,gt=0"`
	SerialCodeIDs  []int64 `json:"serial_code_ids"`
}

type StockOutSerialSelectionReq struct {
	Mode  string                        `json:"mode" binding:"required,oneof=auto_fifo manual"`
	Items []StockOutSerialSelectionItem `json:"items"`
}

type UpdateStockOutItemSerialSelectionsReq struct {
	SerialCodeIDs []int64 `json:"serial_code_ids"`
}

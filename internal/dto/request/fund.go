package request

type FundPaymentQuery struct {
	PageQuery
	StatementNo string `form:"statement_no"`
	SupplierID  int64  `form:"supplier_id"`
	Status      string `form:"status"`
}

type FundPaymentSourceQuery struct {
	SupplierID int64  `form:"supplier_id" binding:"required"`
	Keyword    string `form:"keyword"`
}

type CreateFundPaymentReq struct {
	SupplierID        int64                `json:"supplier_id" binding:"required"`
	StatementDate     string               `json:"statement_date" binding:"required"`
	BillType          string               `json:"bill_type"`
	PaymentAmount     float64              `json:"payment_amount"`
	InvoiceAmount     float64              `json:"invoice_amount"`
	ActualAmount      float64              `json:"actual_amount"`
	DiscountAmount    float64              `json:"discount_amount"`
	AdvanceAmount     float64              `json:"advance_amount"`
	SettlementMethod  string               `json:"settlement_method"`
	SettlementAccount string               `json:"settlement_account"`
	SettlementNo      string               `json:"settlement_no"`
	Remark            string               `json:"remark"`
	Items             []FundPaymentItemReq `json:"items" binding:"required"`
}

type FundPaymentItemReq struct {
	SourceDocType       string  `json:"source_doc_type"`
	SourceOrderID       int64   `json:"source_order_id" binding:"required"`
	SourceOrderNo       string  `json:"source_order_no" binding:"required"`
	BusinessType        string  `json:"business_type" binding:"required"`
	OrderDate           string  `json:"order_date" binding:"required"`
	OrderAmount         float64 `json:"order_amount"`
	VerifiedAmount      float64 `json:"verified_amount"`
	UnverifiedAmount    float64 `json:"unverified_amount"`
	CurrentVerifyAmount float64 `json:"current_verify_amount" binding:"required"`
	CustomTaxAmount     float64 `json:"custom_tax_amount"`
	Remark              string  `json:"remark"`
}

type FundCollectionQuery struct {
	PageQuery
	StatementNo string `form:"statement_no"`
	CustomerID  int64  `form:"customer_id"`
	Status      string `form:"status"`
}

type FundCollectionSourceQuery struct {
	CustomerID int64  `form:"customer_id" binding:"required"`
	Keyword    string `form:"keyword"`
}

type CreateFundCollectionReq struct {
	CustomerID        int64                   `json:"customer_id" binding:"required"`
	StatementDate     string                  `json:"statement_date" binding:"required"`
	BillType          string                  `json:"bill_type"`
	CollectionAmount  float64                 `json:"collection_amount"`
	InvoiceAmount     float64                 `json:"invoice_amount"`
	ActualAmount      float64                 `json:"actual_amount"`
	DiscountAmount    float64                 `json:"discount_amount"`
	AdvanceAmount     float64                 `json:"advance_amount"`
	SettlementMethod  string                  `json:"settlement_method"`
	SettlementAccount string                  `json:"settlement_account"`
	SettlementNo      string                  `json:"settlement_no"`
	Remark            string                  `json:"remark"`
	Items             []FundCollectionItemReq `json:"items" binding:"required"`
}

type FundCollectionItemReq struct {
	SourceDocType       string  `json:"source_doc_type"`
	SourceOrderID       int64   `json:"source_order_id" binding:"required"`
	SourceOrderNo       string  `json:"source_order_no" binding:"required"`
	BusinessType        string  `json:"business_type" binding:"required"`
	OrderDate           string  `json:"order_date" binding:"required"`
	OrderAmount         float64 `json:"order_amount"`
	VerifiedAmount      float64 `json:"verified_amount"`
	UnverifiedAmount    float64 `json:"unverified_amount"`
	CurrentVerifyAmount float64 `json:"current_verify_amount" binding:"required"`
	CustomTaxAmount     float64 `json:"custom_tax_amount"`
	Remark              string  `json:"remark"`
}

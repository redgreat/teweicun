package request


type FundPaymentQuery struct {
	PageQuery
	StatementNo string `form:"statement_no"`
	SupplierID  int64  `form:"supplier_id"`
	Status      string `form:"status"`
}

type CreateFundPaymentReq struct {
	SupplierID        int64                     `json:"supplier_id" binding:"required"`
	StatementDate     string                    `json:"statement_date" binding:"required"`
	PaymentAmount     float64                   `json:"payment_amount"`
	DiscountAmount    float64                   `json:"discount_amount"`
	AdvanceAmount     float64                   `json:"advance_amount"`
	SettlementMethod  string                    `json:"settlement_method"`
	SettlementAccount string                    `json:"settlement_account"`
	SettlementNo      string                    `json:"settlement_no"`
	Remark            string                    `json:"remark"`
	Items             []FundPaymentItemReq      `json:"items" binding:"required"`
}

type FundPaymentItemReq struct {
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

type CreateFundCollectionReq struct {
	CustomerID        int64                        `json:"customer_id" binding:"required"`
	StatementDate     string                       `json:"statement_date" binding:"required"`
	CollectionAmount  float64                      `json:"collection_amount"`
	DiscountAmount    float64                      `json:"discount_amount"`
	AdvanceAmount     float64                      `json:"advance_amount"`
	SettlementMethod  string                       `json:"settlement_method"`
	SettlementAccount string                       `json:"settlement_account"`
	SettlementNo      string                       `json:"settlement_no"`
	Remark            string                       `json:"remark"`
	Items             []FundCollectionItemReq      `json:"items" binding:"required"`
}

type FundCollectionItemReq struct {
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

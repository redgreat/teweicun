package response

import "time"

type FundPaymentResp struct {
	ID                int64     `json:"id"`
	StatementNo       string    `json:"statement_no"`
	SupplierID        int64     `json:"supplier_id"`
	SupplierName      string    `json:"supplier_name"`
	SupplierCode      string    `json:"supplier_code"`
	PayerID           int64     `json:"payer_id"`
	PayerName         string    `json:"payer_name"`
	StatementDate     string    `json:"statement_date"`
	BillType          string    `json:"bill_type"`
	PaymentAmount     float64   `json:"payment_amount"`
	InvoiceAmount     float64   `json:"invoice_amount"`
	ActualAmount      float64   `json:"actual_amount"`
	DifferenceAmount  float64   `json:"difference_amount"`
	DiscountAmount    float64   `json:"discount_amount"`
	AdvanceAmount     float64   `json:"advance_amount"`
	SettlementMethod  string    `json:"settlement_method"`
	SettlementAccount string    `json:"settlement_account"`
	SettlementNo      string    `json:"settlement_no"`
	Remark            string    `json:"remark"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type FundPaymentDetailResp struct {
	FundPaymentResp
	Items []FundPaymentItemResp `json:"items"`
}

type FundPaymentItemResp struct {
	ID                  int64     `json:"id"`
	StatementID         int64     `json:"statement_id"`
	SourceDocType       string    `json:"source_doc_type"`
	SourceOrderID       int64     `json:"source_order_id"`
	SourceOrderNo       string    `json:"source_order_no"`
	BusinessType        string    `json:"business_type"`
	OrderDate           string    `json:"order_date"`
	OrderAmount         float64   `json:"order_amount"`
	VerifiedAmount      float64   `json:"verified_amount"`
	UnverifiedAmount    float64   `json:"unverified_amount"`
	CurrentVerifyAmount float64   `json:"current_verify_amount"`
	CustomTaxAmount     float64   `json:"custom_tax_amount"`
	Remark              string    `json:"remark"`
	CreatedAt           time.Time `json:"created_at"`
}

type FundCollectionResp struct {
	ID                int64     `json:"id"`
	StatementNo       string    `json:"statement_no"`
	CustomerID        int64     `json:"customer_id"`
	CustomerName      string    `json:"customer_name"`
	CustomerCode      string    `json:"customer_code"`
	PayeeID           int64     `json:"payee_id"`
	PayeeName         string    `json:"payee_name"`
	StatementDate     string    `json:"statement_date"`
	BillType          string    `json:"bill_type"`
	CollectionAmount  float64   `json:"collection_amount"`
	InvoiceAmount     float64   `json:"invoice_amount"`
	ActualAmount      float64   `json:"actual_amount"`
	DifferenceAmount  float64   `json:"difference_amount"`
	DiscountAmount    float64   `json:"discount_amount"`
	AdvanceAmount     float64   `json:"advance_amount"`
	SettlementMethod  string    `json:"settlement_method"`
	SettlementAccount string    `json:"settlement_account"`
	SettlementNo      string    `json:"settlement_no"`
	Remark            string    `json:"remark"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type FundCollectionDetailResp struct {
	FundCollectionResp
	Items []FundCollectionItemResp `json:"items"`
}

type FundCollectionItemResp struct {
	ID                  int64     `json:"id"`
	StatementID         int64     `json:"statement_id"`
	SourceDocType       string    `json:"source_doc_type"`
	SourceOrderID       int64     `json:"source_order_id"`
	SourceOrderNo       string    `json:"source_order_no"`
	BusinessType        string    `json:"business_type"`
	OrderDate           string    `json:"order_date"`
	OrderAmount         float64   `json:"order_amount"`
	VerifiedAmount      float64   `json:"verified_amount"`
	UnverifiedAmount    float64   `json:"unverified_amount"`
	CurrentVerifyAmount float64   `json:"current_verify_amount"`
	CustomTaxAmount     float64   `json:"custom_tax_amount"`
	Remark              string    `json:"remark"`
	CreatedAt           time.Time `json:"created_at"`
}

type FundPaymentSourceResp struct {
	SourceDocType    string  `json:"source_doc_type"`
	SourceOrderID    int64   `json:"source_order_id"`
	SourceOrderNo    string  `json:"source_order_no"`
	BusinessType     string  `json:"business_type"`
	OrderDate        string  `json:"order_date"`
	SupplierID       int64   `json:"supplier_id"`
	SupplierCode     string  `json:"supplier_code"`
	SupplierName     string  `json:"supplier_name"`
	OrderAmount      float64 `json:"order_amount"`
	VerifiedAmount   float64 `json:"verified_amount"`
	UnverifiedAmount float64 `json:"unverified_amount"`
	VerifyStatus     string  `json:"verify_status"`
}

type FundCollectionSourceResp struct {
	SourceDocType    string  `json:"source_doc_type"`
	SourceOrderID    int64   `json:"source_order_id"`
	SourceOrderNo    string  `json:"source_order_no"`
	BusinessType     string  `json:"business_type"`
	OrderDate        string  `json:"order_date"`
	CustomerID       int64   `json:"customer_id"`
	CustomerCode     string  `json:"customer_code"`
	CustomerName     string  `json:"customer_name"`
	OrderAmount      float64 `json:"order_amount"`
	VerifiedAmount   float64 `json:"verified_amount"`
	UnverifiedAmount float64 `json:"unverified_amount"`
	VerifyStatus     string  `json:"verify_status"`
}

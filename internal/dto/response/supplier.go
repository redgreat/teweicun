/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import (
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type SupplierResp struct {
	ID                  int64             `json:"id"`
	SupplierCode        string            `json:"supplier_code"`
	SupplierName        string            `json:"supplier_name"`
	CreditCode          database.NullString `json:"credit_code"`
	SupplierType        string            `json:"supplier_type"`
	SupplierTypeName    string            `json:"supplier_type_name"`
	ContactPerson       string            `json:"contact_person"`
	ContactPhone        string            `json:"contact_phone"`
	Address             database.NullString `json:"address"`
	IsQualified         bool              `json:"is_qualified"`
	QualificationExpire *time.Time        `json:"qualification_expire"`
	SupplierRating      database.NullString `json:"supplier_rating"`
	SupplierRatingName  string            `json:"supplier_rating_name"`
	BankName            database.NullString `json:"bank_name"`
	BankAccount         database.NullString `json:"bank_account"`
	Remark              database.NullString `json:"remark"`
	Status              string            `json:"status"`
	StatusName          string            `json:"status_name"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

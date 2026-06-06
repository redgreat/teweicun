-- MIGRATION_ID: 021_seed_stock_in_type_dict
-- MIGRATION_APPLIED: pending

BEGIN;

INSERT INTO sys_dict_type (dict_type, dict_name, remark, status)
VALUES ('stock_in_type', '入库类型', '用于入库单 stock_in.stock_in_type 的展示与筛选', 'enabled')
ON CONFLICT (dict_type) DO UPDATE
SET dict_name = EXCLUDED.dict_name,
	status = EXCLUDED.status,
	updated_at = NOW();

INSERT INTO sys_dict_data (dict_type, dict_value, dict_label, sort_order, is_default, remark, status)
VALUES
	('stock_in_type', 'purchase', '采购入库', 10, true, '', 'enabled'),
	('stock_in_type', 'sales_return', '销售退货入库', 20, false, '', 'enabled'),
	('stock_in_type', 'reversal', '退料入库', 30, false, '', 'enabled'),
	('stock_in_type', 'production', '生产入库', 40, false, '', 'enabled')
ON CONFLICT (dict_type, dict_value) DO UPDATE
SET dict_label = EXCLUDED.dict_label,
	sort_order = EXCLUDED.sort_order,
	is_default = EXCLUDED.is_default,
	status = EXCLUDED.status,
	updated_at = NOW();

COMMIT;

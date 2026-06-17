-- TeWeiCun business aggregate validation SQL.
-- Run against the business database to compare report aggregates with API output.

-- 1. Customer reconciliation summary source aggregate.
SELECT
  src.customer_id,
  src.customer_code,
  src.customer_name,
  src.receivable_amount,
  src.verified_amount,
  src.balance_amount,
  COALESCE(pay.actual_amount, 0)::numeric(18,2) AS actual_amount,
  COALESCE(pay.invoice_amount, 0)::numeric(18,2) AS invoice_amount,
  COALESCE(pay.discount_amount, 0)::numeric(18,2) AS discount_amount
FROM (
  SELECT
    customer_id,
    COALESCE(customer_code, '') AS customer_code,
    COALESCE(customer_name, '') AS customer_name,
    COALESCE(SUM(order_amount), 0)::numeric(18,2) AS receivable_amount,
    COALESCE(SUM(verified_amount), 0)::numeric(18,2) AS verified_amount,
    COALESCE(SUM(unverified_amount), 0)::numeric(18,2) AS balance_amount
  FROM v_fund_collection_source
  GROUP BY customer_id, customer_code, customer_name
) src
LEFT JOIN (
  SELECT
    customer_id,
    COALESCE(SUM(actual_amount), 0)::numeric(18,2) AS actual_amount,
    COALESCE(SUM(invoice_amount), 0)::numeric(18,2) AS invoice_amount,
    COALESCE(SUM(discount_amount), 0)::numeric(18,2) AS discount_amount
  FROM fund_collection
  WHERE deleted_at IS NULL AND status = 'completed'
  GROUP BY customer_id
) pay ON pay.customer_id = src.customer_id
ORDER BY ABS(src.balance_amount) DESC, src.customer_code ASC, src.customer_id DESC;

-- 2. Supplier reconciliation summary source aggregate.
SELECT
  src.supplier_id,
  src.supplier_code,
  src.supplier_name,
  src.payable_amount,
  src.verified_amount,
  src.balance_amount,
  COALESCE(pay.actual_amount, 0)::numeric(18,2) AS actual_amount,
  COALESCE(pay.invoice_amount, 0)::numeric(18,2) AS invoice_amount,
  COALESCE(pay.discount_amount, 0)::numeric(18,2) AS discount_amount
FROM (
  SELECT
    supplier_id,
    COALESCE(supplier_code, '') AS supplier_code,
    COALESCE(supplier_name, '') AS supplier_name,
    COALESCE(SUM(order_amount), 0)::numeric(18,2) AS payable_amount,
    COALESCE(SUM(verified_amount), 0)::numeric(18,2) AS verified_amount,
    COALESCE(SUM(unverified_amount), 0)::numeric(18,2) AS balance_amount
  FROM v_fund_payment_source
  GROUP BY supplier_id, supplier_code, supplier_name
) src
LEFT JOIN (
  SELECT
    supplier_id,
    COALESCE(SUM(actual_amount), 0)::numeric(18,2) AS actual_amount,
    COALESCE(SUM(invoice_amount), 0)::numeric(18,2) AS invoice_amount,
    COALESCE(SUM(discount_amount), 0)::numeric(18,2) AS discount_amount
  FROM fund_payment
  WHERE deleted_at IS NULL AND status = 'completed'
  GROUP BY supplier_id
) pay ON pay.supplier_id = src.supplier_id
ORDER BY ABS(src.balance_amount) DESC, src.supplier_code ASC, src.supplier_id DESC;

-- 3. Profit report source aggregate.
WITH sales AS (
  SELECT COALESCE(SUM(COALESCE(total_amount, 0)), 0)::numeric(18,2) AS sales_amount
  FROM sales_order
  WHERE deleted_at IS NULL AND order_status NOT IN ('draft', 'cancelled')
),
cost AS (
  SELECT COALESCE(SUM(soi.quantity * COALESCE(inv.unit_cost, 0)), 0)::numeric(18,2) AS cost_amount
  FROM stock_out so
  INNER JOIN stock_out_item soi ON soi.stock_out_id = so.id
  LEFT JOIN inventory inv ON inv.id = soi.inventory_id
  WHERE so.deleted_at IS NULL
    AND so.out_type = 'sales'
    AND so.status = 'confirmed'
)
SELECT
  sales.sales_amount,
  cost.cost_amount,
  (sales.sales_amount - cost.cost_amount)::numeric(18,2) AS profit
FROM sales, cost;

-- 4. Inventory and serial-reservation consistency smoke checks.
SELECT
  'negative_inventory_component' AS issue_type,
  material_id,
  warehouse_id,
  unit_cost,
  quantity,
  locked_quantity,
  in_transit_quantity
FROM inventory
WHERE deleted_at IS NULL
  AND (
    quantity < -0.005
    OR locked_quantity < -0.005
    OR in_transit_quantity < -0.005
  );

SELECT
  'duplicate_stock_out_serial_reservation' AS issue_type,
  serial_code_id,
  COUNT(*) AS reservation_count,
  ARRAY_AGG(stock_out_item_id ORDER BY stock_out_item_id) AS stock_out_item_ids
FROM stock_out_item_serial_selection
GROUP BY serial_code_id
HAVING COUNT(*) > 1;

SELECT
  'duplicate_stock_in_serial_reservation' AS issue_type,
  serial_code_id,
  COUNT(*) AS reservation_count,
  ARRAY_AGG(stock_in_item_id ORDER BY stock_in_item_id) AS stock_in_item_ids
FROM stock_in_item_serial_selection
GROUP BY serial_code_id
HAVING COUNT(*) > 1;

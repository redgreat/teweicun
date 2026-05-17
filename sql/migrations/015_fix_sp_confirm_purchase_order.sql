-- ============================================================
-- 功能：修复 sp_confirm_purchase_order 中不存在的 sku_id 列引用
-- 创建时间：2026-05-17
-- 创建人：wangcw
-- ============================================================

BEGIN;

DROP PROCEDURE IF EXISTS sp_confirm_purchase_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_purchase_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
  v_order      RECORD;
  v_stock_in_id BIGINT;
  v_stock_in_no VARCHAR;
BEGIN
  SELECT * INTO v_order
  FROM purchase_order
  WHERE id = p_order_id AND deleted_at IS NULL
  FOR UPDATE;

  IF v_order IS NULL THEN
    RAISE EXCEPTION '采购订单不存在 [id=%]', p_order_id;
  END IF;

  IF v_order.order_status != 'draft' THEN
    RAISE EXCEPTION '只有草稿状态的订单可以确认 [当前状态=%]', v_order.order_status;
  END IF;

  UPDATE purchase_order
  SET order_status = 'ordered',
      updated_by = p_operator_id,
      updated_at = NOW()
  WHERE id = p_order_id;

  v_stock_in_no := fn_generate_serial_no('SI');

  INSERT INTO stock_in (
    stock_in_no, purchase_order_id, supplier_id, supplier_code,
    warehouse_id, warehouse_code, stock_in_date,
    stock_in_status, stock_in_type, remark,
    created_by, created_at, updated_by, updated_at
  ) VALUES (
    v_stock_in_no, p_order_id, v_order.supplier_id, COALESCE(v_order.supplier_code, ''),
    NULL, NULL, CURRENT_DATE,
    'preparing', 'purchase', '采购订单提交自动生成',
    p_operator_id, NOW(), p_operator_id, NOW()
  )
  RETURNING id INTO v_stock_in_id;

  INSERT INTO stock_in_item (
    stock_in_id, material_id, custom_attributes,
    arrived_quantity, accepted_quantity, unit, unit_cost
  )
  SELECT
    v_stock_in_id,
    poi.material_id,
    COALESCE(poi.custom_attributes, '[]'::jsonb),
    poi.quantity,
    poi.quantity,
    COALESCE(m.unit, ''),
    COALESCE(poi.unit_price, 0)
  FROM purchase_order_item poi
  JOIN material m ON m.id = poi.material_id
  WHERE poi.order_id = p_order_id;

  INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail)
  SELECT p_operator_id, u.username, 'CREATE', 'stock_in', 'stock_in', v_stock_in_id,
         jsonb_build_object('stock_in_no', v_stock_in_no, 'from_order_no', v_order.order_no)
  FROM sys_user u WHERE u.id = p_operator_id;

  INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail)
  SELECT p_operator_id, u.username, 'CONFIRM', 'purchase_order', 'purchase_order', p_order_id,
         jsonb_build_object('order_no', v_order.order_no)
  FROM sys_user u WHERE u.id = p_operator_id;
END;
$procedure$;

COMMIT;

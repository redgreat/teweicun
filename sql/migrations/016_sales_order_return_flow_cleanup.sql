-- ============================================================
-- MIGRATION_ID: 016_sales_order_return_flow_cleanup
-- MIGRATION_APPLIED: pending
-- 功能：收敛销售订单/销售退货自动生成出入库单流程
-- 创建时间：2026-06-06
-- 创建人：Codex
-- ============================================================

BEGIN;

-- 1. 可用库存统一扣除在途量，避免采购退货等在途占用后仍被销售订单锁定。
CREATE OR REPLACE FUNCTION public.fn_get_available_stock(
    p_material_id bigint DEFAULT NULL::bigint,
    p_warehouse_id bigint DEFAULT NULL::bigint
)
RETURNS TABLE(
    material_id bigint,
    material_code character varying,
    material_name character varying,
    warehouse_id bigint,
    warehouse_code character varying,
    warehouse_name character varying,
    location_id bigint,
    location_code character varying,
    location_name character varying,
    quantity numeric,
    locked_quantity numeric,
    in_transit_quantity numeric,
    available_quantity numeric
)
LANGUAGE plpgsql
AS $function$
BEGIN
    RETURN QUERY
    SELECT
        i.material_id,
        m.material_code,
        m.material_name,
        i.warehouse_id,
        w.warehouse_code,
        w.warehouse_name,
        NULL::bigint AS location_id,
        ''::character varying AS location_code,
        ''::character varying AS location_name,
        i.quantity,
        i.locked_quantity,
        COALESCE(i.in_transit_quantity, 0) AS in_transit_quantity,
        GREATEST(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0), 0) AS available_quantity
    FROM inventory i
    JOIN material m ON m.id = i.material_id
    JOIN warehouse w ON w.id = i.warehouse_id
    WHERE (p_material_id IS NULL OR i.material_id = p_material_id)
      AND (p_warehouse_id IS NULL OR i.warehouse_id = p_warehouse_id);
END;
$function$;

COMMENT ON FUNCTION public.fn_get_available_stock(bigint, bigint) IS '查询可用库存：库存数量-锁定数量-在途数量';

-- 2. 销售订单提交只由数据库过程生成待出库单，并在锁定库存时扣除在途量。
DROP PROCEDURE IF EXISTS public.sp_confirm_sales_order(bigint, bigint);
CREATE OR REPLACE PROCEDURE public.sp_confirm_sales_order(IN p_order_id bigint, IN p_operator_id bigint)
LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_order         RECORD;
    v_item          RECORD;
    v_batch         RECORD;
    v_available     NUMERIC(18,3);
    v_remaining     NUMERIC(18,3);
    v_lock_qty      NUMERIC(18,3);
    v_stock_out_id  BIGINT;
    v_stock_out_no  VARCHAR(20);
    v_operator_name VARCHAR(100);
BEGIN
    SELECT *
    INTO v_order
    FROM sales_order
    WHERE id = p_order_id
      AND order_status = 'draft'
      AND deleted_at IS NULL
    FOR UPDATE;

    IF v_order IS NULL THEN
        RAISE EXCEPTION '销售订单不存在或状态不允许确认';
    END IF;

    IF COALESCE(v_order.stock_out_id, 0) <> 0 THEN
        RAISE EXCEPTION '该销售订单已生成出库单，不可重复提交';
    END IF;

    SELECT COALESCE(u.real_name, u.username, '')
    INTO v_operator_name
    FROM sys_user u
    WHERE u.id = p_operator_id;

    v_stock_out_no := fn_generate_serial_no('SO');

    INSERT INTO stock_out (
        stock_out_no, out_type, ref_doc_type, ref_doc_id,
        stock_out_date, customer_code, customer_name, receiver,
        status, remark, created_by, created_at, updated_by, updated_at
    ) VALUES (
        v_stock_out_no, 'sales', 'sales_order', p_order_id,
        COALESCE(v_order.delivery_date, v_order.order_date, CURRENT_DATE),
        NULLIF(v_order.customer_code, ''), NULLIF(v_order.customer_name, ''),
        COALESCE(v_order.receiver_name, v_order.customer_name, ''),
        'pending',
        '销售订单提交自动生成', p_operator_id, NOW(), p_operator_id, NOW()
    )
    RETURNING id INTO v_stock_out_id;

    FOR v_item IN
        SELECT soi.id,
               soi.material_id,
               soi.quantity,
               soi.unit,
               soi.unit_price,
               m.material_code,
               m.material_name,
               soi.remark
        FROM sales_order_item soi
        INNER JOIN material m ON m.id = soi.material_id
        WHERE soi.order_id = p_order_id
        ORDER BY soi.id
    LOOP
        SELECT COALESCE(SUM(s.available_quantity), 0)
          INTO v_available
          FROM fn_get_available_stock(v_item.material_id, NULL) s;

        IF v_available < v_item.quantity THEN
            RAISE EXCEPTION '物料[%-%]可用库存不足，需要:%, 可用:%',
                v_item.material_code, v_item.material_name,
                v_item.quantity, v_available;
        END IF;

        v_remaining := v_item.quantity;

        FOR v_batch IN
            SELECT i.id,
                   GREATEST(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0), 0) AS available_qty
            FROM inventory i
            WHERE i.material_id = v_item.material_id
              AND GREATEST(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0), 0) > 0
            ORDER BY i.stock_in_date ASC NULLS LAST, i.id ASC
            FOR UPDATE
        LOOP
            EXIT WHEN v_remaining <= 0;

            v_lock_qty := LEAST(v_remaining, v_batch.available_qty);
            IF v_lock_qty <= 0 THEN
                CONTINUE;
            END IF;

            UPDATE inventory
            SET locked_quantity = locked_quantity + v_lock_qty,
                updated_at = NOW()
            WHERE id = v_batch.id;

            INSERT INTO stock_out_item (
                stock_out_id, material_id, inventory_id, quantity, unit, remark, created_at
            ) VALUES (
                v_stock_out_id, v_item.material_id, v_batch.id,
                v_lock_qty, COALESCE(NULLIF(TRIM(v_item.unit), ''), '件'),
                COALESCE(v_item.remark, ''), NOW()
            );

            v_remaining := v_remaining - v_lock_qty;
        END LOOP;

        IF v_remaining > 0 THEN
            RAISE EXCEPTION '物料[%-%]可用库存不足，需要:%, 未分配:%',
                v_item.material_code, v_item.material_name,
                v_item.quantity, v_remaining;
        END IF;
    END LOOP;

    UPDATE sales_order
    SET order_status = 'confirmed',
        stock_out_id  = v_stock_out_id,
        updated_by    = p_operator_id,
        updated_at    = NOW()
    WHERE id = p_order_id;

    INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail)
    VALUES (p_operator_id, v_operator_name, 'CONFIRM', 'SALES', 'sales_order', p_order_id,
            jsonb_build_object('order_no', v_order.order_no, 'stock_out_no', v_stock_out_no));
END;
$procedure$;

COMMENT ON PROCEDURE public.sp_confirm_sales_order(bigint, bigint) IS '销售订单提交：校验并锁定可用库存，自动生成待出库单';

-- 3. 销售退货的 confirmed 表示已提交且待入库；入库单完成后自动置为 completed。
COMMENT ON COLUMN return_order.return_status IS '退货状态：draft-待提交，confirmed-待出/入库，pending_out-待出库，completed-已完成，closed-已关闭';

CREATE OR REPLACE FUNCTION public.trf_complete_sales_return_after_stock_in()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    IF NEW.stock_in_status = 'passed'
       AND OLD.stock_in_status IS DISTINCT FROM NEW.stock_in_status
       AND COALESCE(NEW.stock_in_type, '') = 'sales_return' THEN
        UPDATE return_order
        SET return_status = 'completed',
            updated_by = NEW.updated_by,
            updated_at = NOW()
        WHERE stock_in_id = NEW.id
          AND return_type = 'sales_return'
          AND deleted_at IS NULL
          AND return_status <> 'completed';
    END IF;
    RETURN NEW;
END;
$function$;

COMMENT ON FUNCTION public.trf_complete_sales_return_after_stock_in() IS '销售退货入库单完成后自动将销售退货单置为已完成';

DROP TRIGGER IF EXISTS trg_complete_sales_return_after_stock_in ON stock_in;
CREATE TRIGGER trg_complete_sales_return_after_stock_in
AFTER UPDATE OF stock_in_status ON stock_in
FOR EACH ROW
EXECUTE FUNCTION public.trf_complete_sales_return_after_stock_in();

UPDATE return_order ro
SET return_status = 'completed',
    updated_at = NOW()
FROM stock_in si
WHERE ro.stock_in_id = si.id
  AND ro.return_type = 'sales_return'
  AND ro.deleted_at IS NULL
  AND si.deleted_at IS NULL
  AND si.stock_in_status = 'passed'
  AND ro.return_status <> 'completed';

COMMIT;

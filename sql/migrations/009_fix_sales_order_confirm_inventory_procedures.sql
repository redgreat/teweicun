-- 功能：修复销售订单确认依赖旧 location 表及库存锁定逻辑
-- 创建时间：2026-05-17
-- 创建人：GPT-5.4

BEGIN;

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
        GREATEST(i.quantity - i.locked_quantity, 0) AS available_quantity
    FROM inventory i
    JOIN material m ON m.id = i.material_id
    JOIN warehouse w ON w.id = i.warehouse_id
    WHERE (p_material_id IS NULL OR i.material_id = p_material_id)
      AND (p_warehouse_id IS NULL OR i.warehouse_id = p_warehouse_id);
END;
$function$;

CREATE OR REPLACE PROCEDURE public.sp_confirm_sales_order(IN p_order_id bigint, IN p_operator_id bigint)
LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_item RECORD;
    v_batch RECORD;
    v_available numeric(18,3);
    v_remaining numeric(18,3);
    v_lock_qty numeric(18,3);
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM sales_order
        WHERE id = p_order_id
          AND order_status = 'draft'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION '销售订单不存在或状态不允许确认';
    END IF;

    FOR v_item IN
        SELECT soi.id,
               soi.material_id,
               soi.quantity,
               m.material_code,
               m.material_name
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
                   GREATEST(i.quantity - i.locked_quantity, 0) AS available_qty
            FROM inventory i
            WHERE i.material_id = v_item.material_id
              AND GREATEST(i.quantity - i.locked_quantity, 0) > 0
            ORDER BY i.stock_in_date ASC NULLS LAST, i.id ASC
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

            v_remaining := v_remaining - v_lock_qty;
        END LOOP;
    END LOOP;

    UPDATE sales_order
    SET order_status = 'confirmed',
        updated_by = p_operator_id,
        updated_at = NOW()
    WHERE id = p_order_id;
END;
$procedure$;

COMMIT;

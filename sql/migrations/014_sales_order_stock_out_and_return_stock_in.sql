-- ============================================================
-- 功能：销售订单提交生成出库单 + 销售退货提交生成入库单
--       所有订单表增加关联出入库单ID字段
-- 创建时间：2026-05-17
-- 创建人：wangcw
-- ============================================================

BEGIN;

-- 1. 销售订单增加 stock_out_id 关联字段
ALTER TABLE sales_order ADD COLUMN IF NOT EXISTS stock_out_id BIGINT;
COMMENT ON COLUMN sales_order.stock_out_id IS '关联的出库单ID';

-- 2. 退货单增加 stock_in_id 关联字段（用于销售退货）
ALTER TABLE return_order ADD COLUMN IF NOT EXISTS stock_in_id BIGINT;
COMMENT ON COLUMN return_order.stock_in_id IS '关联的入库单ID（销售退货时使用）';

-- 3. 重写 sp_confirm_sales_order：提交后生成 stock_out
DROP PROCEDURE IF EXISTS sp_confirm_sales_order CASCADE;
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

    SELECT u.real_name INTO v_operator_name FROM sys_user u WHERE u.id = p_operator_id;

    v_stock_out_no := fn_generate_serial_no('SO');

    INSERT INTO stock_out (
        stock_out_no, out_type, ref_doc_type, ref_doc_id,
        stock_out_date, receiver, status, remark, created_by, created_at, updated_by, updated_at
    ) VALUES (
        v_stock_out_no, 'sales', 'sales_order', p_order_id,
        COALESCE(v_order.delivery_date, v_order.order_date, CURRENT_DATE),
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

            INSERT INTO stock_out_item (
                stock_out_id, material_id, inventory_id, quantity, unit, remark, created_at
            ) VALUES (
                v_stock_out_id, v_item.material_id, v_batch.id,
                v_lock_qty, COALESCE(NULLIF(TRIM(v_item.unit), ''), '件'),
                COALESCE(v_item.remark, ''), NOW()
            );

            v_remaining := v_remaining - v_lock_qty;
        END LOOP;
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

-- 4. 重写 sp_confirm_return_order（销售退货分支：生成 stock_in 而非直接操作库存）
DROP PROCEDURE IF EXISTS sp_confirm_return_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_confirm_return_order(IN p_return_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_ro           RECORD;
    v_item         RECORD;
    v_stock_in_id  BIGINT;
    v_stock_in_no  VARCHAR(20);
    v_operator_name VARCHAR(100);
    v_supplier_name VARCHAR(200);
    v_customer_name VARCHAR(200);
BEGIN
    SELECT *
    INTO v_ro
    FROM return_order
    WHERE id = p_return_id
      AND deleted_at IS NULL
    FOR UPDATE;

    IF v_ro IS NULL THEN
        RAISE EXCEPTION '退货单不存在';
    END IF;

    IF v_ro.return_status <> 'draft' THEN
        RAISE EXCEPTION '退货单当前状态[%]不允许确认', v_ro.return_status;
    END IF;

    SELECT u.real_name INTO v_operator_name FROM sys_user u WHERE u.id = p_operator_id;

    IF v_ro.return_type = 'sales_return' THEN
        IF COALESCE(v_ro.stock_in_id, 0) <> 0 THEN
            RAISE EXCEPTION '该销售退货单已生成入库单，不可重复提交';
        END IF;

        IF v_ro.warehouse_id IS NULL OR v_ro.warehouse_id = 0 THEN
            RAISE EXCEPTION '销售退货必须指定入库仓库';
        END IF;

        SELECT COALESCE(c.customer_name, '')
        INTO v_customer_name
        FROM customer c
        WHERE c.customer_code = v_ro.customer_code
          AND c.deleted_at IS NULL;

        v_stock_in_no := fn_generate_serial_no('SI');

        INSERT INTO stock_in (
            stock_in_no, stock_in_date, stock_in_status, stock_in_type,
            warehouse_id, warehouse_code,
            remark, created_by, created_at, updated_by, updated_at
        ) VALUES (
            v_stock_in_no, COALESCE(v_ro.return_date, CURRENT_DATE),
            'preparing', 'sales_return',
            v_ro.warehouse_id, v_ro.warehouse_code,
            '销售退货提交自动生成', p_operator_id, NOW(), p_operator_id, NOW()
        )
        RETURNING id INTO v_stock_in_id;

        FOR v_item IN
            SELECT roi.*, m.material_code, m.material_name, COALESCE(m.unit, roi.unit) AS final_unit
            FROM return_order_item roi
            INNER JOIN material m ON m.id = roi.material_id
            WHERE roi.return_id = p_return_id
        LOOP
            INSERT INTO stock_in_item (
                stock_in_id, material_id,
                arrived_quantity, accepted_quantity, unit, created_at
            ) VALUES (
                v_stock_in_id, v_item.material_id,
                v_item.quantity, v_item.quantity,
                v_item.final_unit, NOW()
            );
        END LOOP;

        UPDATE return_order
        SET return_status = 'confirmed',
            stock_in_id   = v_stock_in_id,
            updated_at    = NOW(),
            updated_by    = p_operator_id
        WHERE id = p_return_id;

        INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail)
        VALUES (p_operator_id, v_operator_name, 'CONFIRM', 'RETURN', 'return_order', p_return_id,
                jsonb_build_object('return_no', v_ro.return_no, 'stock_in_no', v_stock_in_no));

    ELSIF v_ro.return_type = 'purchase_return' THEN
        RAISE EXCEPTION '采购退货请使用采购退货提交流程';
    ELSE
        RAISE EXCEPTION '未知的退货类型: %', v_ro.return_type;
    END IF;
END;
$procedure$;

-- 5. 重写 sp_cancel_sales_order：取消时删除关联的出库单并释放库存锁定
DROP PROCEDURE IF EXISTS sp_cancel_sales_order CASCADE;
CREATE OR REPLACE PROCEDURE public.sp_cancel_sales_order(IN p_order_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_order        RECORD;
    v_operator_name VARCHAR(100);
BEGIN
    SELECT so.*
    INTO v_order
    FROM sales_order so
    WHERE so.id = p_order_id
      AND so.order_status = 'confirmed'
      AND so.deleted_at IS NULL
    FOR UPDATE;

    IF v_order IS NULL THEN
        RAISE EXCEPTION '销售订单不存在或状态不允许取消';
    END IF;

    SELECT u.real_name INTO v_operator_name FROM sys_user u WHERE u.id = p_operator_id;

    IF COALESCE(v_order.stock_out_id, 0) <> 0 THEN
        UPDATE inventory
        SET locked_quantity = GREATEST(locked_quantity - src.qty, 0),
            updated_at = NOW()
        FROM (
            SELECT soi.inventory_id, SUM(soi.quantity) AS qty
            FROM stock_out_item soi
            WHERE soi.stock_out_id = v_order.stock_out_id
              AND COALESCE(soi.inventory_id, 0) <> 0
            GROUP BY soi.inventory_id
        ) src
        WHERE inventory.id = src.inventory_id;

        DELETE FROM stock_out_item WHERE stock_out_id = v_order.stock_out_id;
        DELETE FROM stock_out WHERE id = v_order.stock_out_id;
    END IF;

    UPDATE sales_order
    SET order_status = 'cancelled',
        updated_by   = p_operator_id,
        updated_at   = NOW()
    WHERE id = p_order_id;

    INSERT INTO sys_audit_log (user_id, username, action, module, target_type, target_id, detail)
    VALUES (p_operator_id, v_operator_name, 'CANCEL', 'SALES', 'sales_order', p_order_id,
            jsonb_build_object('order_no', v_order.order_no));
END;
$procedure$;

COMMIT;

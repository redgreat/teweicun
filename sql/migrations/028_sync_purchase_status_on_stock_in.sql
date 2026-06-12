-- MIGRATION_ID: 028_sync_purchase_status_on_stock_in
-- MIGRATION_APPLIED: pending
-- 功能：采购入库确认后同步采购明细已收数量和采购单状态，并回填历史已确认入库单

DROP PROCEDURE IF EXISTS sp_confirm_stock_in CASCADE;

CREATE OR REPLACE PROCEDURE public.sp_confirm_stock_in(IN p_stock_in_id bigint, IN p_operator_id bigint)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_stock_in RECORD;
    v_item RECORD;
    v_serial_code VARCHAR;
    v_inv_id BIGINT;
    v_operator_name VARCHAR;
    v_i INT;
BEGIN
    SELECT si.* INTO v_stock_in
    FROM stock_in si
    WHERE si.id = p_stock_in_id
      AND si.deleted_at IS NULL
    FOR UPDATE;

    IF v_stock_in IS NULL THEN
        RAISE EXCEPTION '入库单不存在 [id=%]', p_stock_in_id;
    END IF;

    IF v_stock_in.stock_in_status = 'passed' THEN
        RAISE EXCEPTION '入库单已全部入库，无需重复确认';
    END IF;

    SELECT u.real_name INTO v_operator_name
    FROM sys_user u
    WHERE u.id = p_operator_id;

    FOR v_item IN
        SELECT sii.*, m.is_code, m.material_code, m.material_name
        FROM stock_in_item sii
        JOIN material m ON m.id = sii.material_id
        WHERE sii.stock_in_id = p_stock_in_id
    LOOP
        IF COALESCE(v_item.accepted_quantity, 0) <= 0 THEN
            CONTINUE;
        END IF;

        INSERT INTO inventory (
            material_id, warehouse_id, quantity, locked_quantity, unit, stock_in_date, unit_cost, material_inspection_no
        ) VALUES (
            v_item.material_id, v_stock_in.warehouse_id, v_item.accepted_quantity, 0, v_item.unit,
            v_stock_in.stock_in_date, COALESCE(v_item.unit_cost, 0), NULL
        )
        ON CONFLICT (material_id, warehouse_id, COALESCE(material_inspection_no, ''))
        DO UPDATE SET
            quantity = inventory.quantity + v_item.accepted_quantity,
            unit_cost = CASE WHEN v_item.unit_cost > 0 THEN v_item.unit_cost ELSE inventory.unit_cost END,
            updated_at = NOW()
        RETURNING id INTO v_inv_id;

        INSERT INTO inventory_transaction (
            material_id, warehouse_id, trans_type, quantity, balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id, created_at
        ) VALUES (
            v_item.material_id, v_stock_in.warehouse_id, 'in', v_item.accepted_quantity,
            (SELECT quantity FROM inventory WHERE id = v_inv_id), 'stock_in', v_stock_in.stock_in_no, p_stock_in_id, p_operator_id, NOW()
        );

        IF v_item.is_code THEN
            FOR v_i IN 1..floor(v_item.accepted_quantity) LOOP
                v_serial_code := fn_generate_serial_no('MAT');

                INSERT INTO material_serial_code (
                    serial_code,
                    material_id, material_code, material_name,
                    stock_in_id, stock_in_item_id,
                    inventory_id, warehouse_id,
                    status, created_at, updated_at
                ) VALUES (
                    v_serial_code,
                    v_item.material_id, v_item.material_code, v_item.material_name,
                    p_stock_in_id, v_item.id,
                    v_inv_id, v_stock_in.warehouse_id,
                    'in_stock', NOW(), NOW()
                );

                INSERT INTO material_serial_trace (
                    serial_code_id, serial_code, action, ref_doc_type, ref_doc_no, ref_doc_id,
                    to_warehouse_id, operator_id, operator_name, remark, created_at
                ) VALUES (
                    (SELECT id FROM material_serial_code WHERE serial_code = v_serial_code),
                    v_serial_code,
                    'stock_in', 'stock_in', v_stock_in.stock_in_no, p_stock_in_id,
                    v_stock_in.warehouse_id, p_operator_id, COALESCE(v_operator_name, ''), '入库确认生成编码', NOW()
                );
            END LOOP;
        END IF;
    END LOOP;

    IF COALESCE(v_stock_in.stock_in_type, '') = 'purchase'
       AND COALESCE(v_stock_in.purchase_order_id, 0) > 0 THEN
        UPDATE purchase_order_item poi
        SET received_quantity = LEAST(
                COALESCE(poi.quantity, 0),
                COALESCE(poi.received_quantity, 0) + COALESCE(siq.accepted_quantity, 0)
            ),
            updated_at = NOW()
        FROM (
            SELECT material_id, SUM(COALESCE(accepted_quantity, 0)) AS accepted_quantity
            FROM stock_in_item
            WHERE stock_in_id = p_stock_in_id
            GROUP BY material_id
        ) siq
        WHERE poi.order_id = v_stock_in.purchase_order_id
          AND poi.material_id = siq.material_id;

        UPDATE purchase_order po
        SET order_status = CASE
                WHEN NOT EXISTS (
                    SELECT 1
                    FROM purchase_order_item poi
                    WHERE poi.order_id = po.id
                      AND COALESCE(poi.received_quantity, 0) < COALESCE(poi.quantity, 0)
                ) THEN 'full_received'
                WHEN EXISTS (
                    SELECT 1
                    FROM purchase_order_item poi
                    WHERE poi.order_id = po.id
                      AND COALESCE(poi.received_quantity, 0) > 0
                ) THEN 'partial_received'
                ELSE 'ordered'
            END,
            updated_by = p_operator_id,
            updated_at = NOW()
        WHERE po.id = v_stock_in.purchase_order_id
          AND po.deleted_at IS NULL
          AND po.order_status NOT IN ('draft', 'cancelled', 'closed');
    END IF;

    UPDATE stock_in
    SET stock_in_status = 'passed',
        updated_by = p_operator_id,
        updated_at = NOW()
    WHERE id = p_stock_in_id;
END;
$procedure$;

WITH received AS (
    SELECT si.purchase_order_id,
           sii.material_id,
           SUM(COALESCE(sii.accepted_quantity, 0)) AS accepted_quantity
    FROM stock_in si
    JOIN stock_in_item sii ON sii.stock_in_id = si.id
    WHERE si.deleted_at IS NULL
      AND si.stock_in_status = 'passed'
      AND COALESCE(si.stock_in_type, '') = 'purchase'
      AND si.purchase_order_id IS NOT NULL
    GROUP BY si.purchase_order_id, sii.material_id
)
UPDATE purchase_order_item poi
SET received_quantity = LEAST(COALESCE(poi.quantity, 0), COALESCE(received.accepted_quantity, 0)),
    updated_at = NOW()
FROM received
WHERE poi.order_id = received.purchase_order_id
  AND poi.material_id = received.material_id;

UPDATE purchase_order po
SET order_status = CASE
        WHEN NOT EXISTS (
            SELECT 1
            FROM purchase_order_item poi
            WHERE poi.order_id = po.id
              AND COALESCE(poi.received_quantity, 0) < COALESCE(poi.quantity, 0)
        ) THEN 'full_received'
        WHEN EXISTS (
            SELECT 1
            FROM purchase_order_item poi
            WHERE poi.order_id = po.id
              AND COALESCE(poi.received_quantity, 0) > 0
        ) THEN 'partial_received'
        ELSE 'ordered'
    END,
    updated_at = NOW()
WHERE po.deleted_at IS NULL
  AND po.order_status NOT IN ('draft', 'cancelled', 'closed')
  AND EXISTS (
      SELECT 1
      FROM stock_in si
      WHERE si.purchase_order_id = po.id
        AND si.deleted_at IS NULL
        AND si.stock_in_status = 'passed'
  );

-- MIGRATION_ID: 030_fix_partial_stock_in_and_restore_production_generation
-- MIGRATION_APPLIED: pending
-- 功能：
-- 1. 修复采购入库分批确认后状态被直接置为 passed 的问题，恢复 pending(部分入库)
-- 2. 重建领料出库确认后自动生成生产入库/生产单的数据库过程与触发器
-- 3. 回填历史采购单已收数量、采购单状态、采购入库状态，以及漏生成的生产单

BEGIN;

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

        UPDATE stock_in si
        SET stock_in_status = CASE
                WHEN EXISTS (
                    SELECT 1
                    FROM purchase_order_item poi
                    WHERE poi.order_id = v_stock_in.purchase_order_id
                      AND COALESCE(poi.received_quantity, 0) < COALESCE(poi.quantity, 0)
                ) THEN 'pending'
                ELSE 'passed'
            END,
            updated_by = p_operator_id,
            updated_at = NOW()
        WHERE si.id = p_stock_in_id;
    ELSE
        UPDATE stock_in
        SET stock_in_status = 'passed',
            updated_by = p_operator_id,
            updated_at = NOW()
        WHERE id = p_stock_in_id;
    END IF;
END;
$procedure$;

CREATE OR REPLACE PROCEDURE sp_generate_production_from_consumption(
    IN p_consumption_order_id BIGINT,
    IN p_stock_out_id BIGINT,
    IN p_operator_id BIGINT
)
LANGUAGE plpgsql
AS $procedure$
DECLARE
    v_co RECORD;
    v_wh RECORD;
    v_mat RECORD;
    v_total_cost NUMERIC(18, 3);
    v_unit_cost NUMERIC(18, 3);
    v_stock_in_id BIGINT;
    v_stock_in_no VARCHAR(20);
    v_production_no VARCHAR(20);
    v_production_id BIGINT;
    v_target_production_order_id BIGINT;
BEGIN
    IF COALESCE(p_operator_id, 0) = 0 THEN
        RETURN;
    END IF;

    SELECT
        co.id,
        co.order_no,
        co.order_date,
        co.status,
        co.produced_material_id,
        co.produced_quantity,
        co.produced_warehouse_id,
        co.production_order_id
    INTO v_co
    FROM consumption_order co
    WHERE co.id = p_consumption_order_id
      AND co.deleted_at IS NULL
    FOR UPDATE;

    IF v_co IS NULL THEN
        RETURN;
    END IF;

    IF v_co.produced_material_id IS NULL OR v_co.produced_quantity IS NULL OR v_co.produced_warehouse_id IS NULL THEN
        RETURN;
    END IF;

    v_target_production_order_id := v_co.production_order_id;

    IF v_target_production_order_id IS NULL THEN
        SELECT id INTO v_target_production_order_id
        FROM production_order
        WHERE consumption_order_id = p_consumption_order_id
        LIMIT 1;
    END IF;

    IF v_target_production_order_id IS NOT NULL THEN
        UPDATE production_order
        SET cost_price = (
            SELECT COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)
            FROM consumption_order_item coi
            LEFT JOIN inventory inv ON inv.id = coi.inventory_id
            WHERE coi.order_id IN (
                SELECT id
                FROM consumption_order
                WHERE production_order_id = v_target_production_order_id
                  AND deleted_at IS NULL
            )
        ),
        updated_at = NOW(),
        updated_by = p_operator_id
        WHERE id = v_target_production_order_id;

        UPDATE stock_in_item sii
        SET unit_cost = (
            SELECT COALESCE(po.cost_price, po.produced_unit_cost) / NULLIF(po.produced_quantity, 0)
            FROM production_order po
            WHERE po.id = v_target_production_order_id
        )
        FROM stock_in si
        JOIN production_order po ON po.stock_in_id = si.id
        WHERE sii.stock_in_id = si.id
          AND po.id = v_target_production_order_id;

        RETURN;
    END IF;

    SELECT id, warehouse_code, warehouse_name INTO v_wh
    FROM warehouse
    WHERE id = v_co.produced_warehouse_id AND deleted_at IS NULL;

    IF v_wh IS NULL THEN
        RAISE EXCEPTION '成品入库仓库不存在 [warehouse_id=%]', v_co.produced_warehouse_id;
    END IF;

    SELECT id, material_code, material_name, COALESCE(NULLIF(TRIM(unit), ''), '件') AS unit, COALESCE(is_code, FALSE) AS is_code
    INTO v_mat
    FROM material
    WHERE id = v_co.produced_material_id AND deleted_at IS NULL;

    IF v_mat IS NULL THEN
        RAISE EXCEPTION '成品物料不存在 [material_id=%]', v_co.produced_material_id;
    END IF;

    SELECT COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)
    INTO v_total_cost
    FROM consumption_order_item coi
    LEFT JOIN inventory inv ON inv.id = coi.inventory_id
    WHERE coi.order_id = p_consumption_order_id;

    IF v_co.produced_quantity <= 0 THEN
        RAISE EXCEPTION '成品数量必须 > 0';
    END IF;

    v_unit_cost := v_total_cost / v_co.produced_quantity;
    v_stock_in_no := fn_generate_serial_no('SI');

    INSERT INTO stock_in (
        stock_in_no,
        warehouse_id, warehouse_code,
        stock_in_date,
        stock_in_status,
        stock_in_type,
        remark,
        created_by, created_at, updated_by, updated_at
    ) VALUES (
        v_stock_in_no,
        v_wh.id, v_wh.warehouse_code,
        COALESCE(v_co.order_date, CURRENT_DATE),
        'preparing',
        'production',
        '生产单自动生成（领料出库确认后）',
        p_operator_id, NOW(), p_operator_id, NOW()
    )
    RETURNING id INTO v_stock_in_id;

    INSERT INTO stock_in_item (
        stock_in_id,
        material_id,
        arrived_quantity,
        accepted_quantity,
        unit,
        unit_cost,
        custom_attributes,
        created_at
    ) VALUES (
        v_stock_in_id,
        v_mat.id,
        v_co.produced_quantity,
        v_co.produced_quantity,
        v_mat.unit,
        v_unit_cost,
        '[]'::jsonb,
        NOW()
    );

    CALL sp_confirm_stock_in(v_stock_in_id, p_operator_id);

    v_production_no := fn_generate_serial_no('PR');

    INSERT INTO production_order (
        production_no,
        consumption_order_id,
        stock_out_id,
        stock_in_id,
        produced_material_id,
        produced_warehouse_id,
        produced_quantity,
        produced_unit_cost,
        cost_price,
        status,
        remark,
        created_by,
        updated_by
    ) VALUES (
        v_production_no,
        p_consumption_order_id,
        p_stock_out_id,
        v_stock_in_id,
        v_mat.id,
        v_wh.id,
        v_co.produced_quantity,
        v_unit_cost,
        v_total_cost,
        'completed',
        '由领料出库确认自动生成',
        p_operator_id,
        p_operator_id
    )
    RETURNING id INTO v_production_id;

    UPDATE consumption_order
    SET production_order_id = v_production_id,
        updated_at = NOW(),
        updated_by = p_operator_id
    WHERE id = p_consumption_order_id;
END;
$procedure$;

CREATE OR REPLACE FUNCTION trg_stock_out_after_confirm_generate_production()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
BEGIN
    IF TG_OP = 'UPDATE'
       AND NEW.status = 'confirmed'
       AND COALESCE(OLD.status, '') <> 'confirmed'
       AND NEW.ref_doc_type = 'consumption_order'
       AND COALESCE(NEW.ref_doc_id, 0) <> 0
       AND COALESCE(NEW.updated_by, 0) <> 0
    THEN
        CALL sp_generate_production_from_consumption(NEW.ref_doc_id, NEW.id, NEW.updated_by);
    END IF;
    RETURN NEW;
END;
$function$;

DROP TRIGGER IF EXISTS stock_out_after_confirm_generate_production ON stock_out;
CREATE TRIGGER stock_out_after_confirm_generate_production
AFTER UPDATE OF status ON stock_out
FOR EACH ROW
EXECUTE FUNCTION trg_stock_out_after_confirm_generate_production();

WITH received AS (
    SELECT si.purchase_order_id,
           sii.material_id,
           SUM(COALESCE(sii.accepted_quantity, 0)) AS accepted_quantity
    FROM stock_in si
    JOIN stock_in_item sii ON sii.stock_in_id = si.id
    WHERE si.deleted_at IS NULL
      AND COALESCE(si.stock_in_type, '') = 'purchase'
      AND si.purchase_order_id IS NOT NULL
      AND si.stock_in_status IN ('pending', 'passed')
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
        AND COALESCE(si.stock_in_type, '') = 'purchase'
  );

UPDATE stock_in si
SET stock_in_status = CASE
        WHEN EXISTS (
            SELECT 1
            FROM purchase_order_item poi
            WHERE poi.order_id = si.purchase_order_id
              AND COALESCE(poi.received_quantity, 0) < COALESCE(poi.quantity, 0)
        ) THEN 'pending'
        ELSE 'passed'
    END,
    updated_at = NOW()
WHERE si.deleted_at IS NULL
  AND COALESCE(si.stock_in_type, '') = 'purchase'
  AND si.purchase_order_id IS NOT NULL
  AND si.stock_in_status IN ('pending', 'passed');

DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN
        SELECT
            so.id AS stock_out_id,
            so.ref_doc_id AS consumption_order_id,
            COALESCE(NULLIF(so.updated_by, 0), NULLIF(co.updated_by, 0), NULLIF(co.created_by, 0)) AS operator_id
        FROM stock_out so
        JOIN consumption_order co
          ON co.id = so.ref_doc_id
         AND co.deleted_at IS NULL
        WHERE so.deleted_at IS NULL
          AND so.status = 'confirmed'
          AND so.ref_doc_type = 'consumption_order'
          AND COALESCE(co.produced_material_id, 0) > 0
          AND COALESCE(co.produced_quantity, 0) > 0
          AND COALESCE(co.produced_warehouse_id, 0) > 0
          AND COALESCE(co.production_order_id, 0) = 0
    LOOP
        IF COALESCE(r.operator_id, 0) > 0 THEN
            CALL sp_generate_production_from_consumption(r.consumption_order_id, r.stock_out_id, r.operator_id);
        END IF;
    END LOOP;
END $$;

COMMIT;

-- MIGRATION_ID: 029_backfill_inventory_ledger_serial_codes
-- MIGRATION_APPLIED: pending
-- 功能：修复历史/测试数据中编码物料库存台账有库存但缺少在库编码明细的问题

DO $$
DECLARE
    v_row RECORD;
    v_i INT;
    v_serial_id BIGINT;
    v_serial_code VARCHAR;
    v_missing_count INT;
BEGIN
    FOR v_row IN
        SELECT
            i.id AS inventory_id,
            i.material_id,
            m.material_code,
            m.material_name,
            i.warehouse_id,
            GREATEST(
                floor(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0))::INT
                - COALESCE(count(sc.id) FILTER (WHERE sc.status = 'in_stock'), 0)::INT,
                0
            ) AS missing_count
        FROM inventory i
        INNER JOIN material m ON m.id = i.material_id
        LEFT JOIN material_serial_code sc ON sc.inventory_id = i.id
        WHERE COALESCE(m.is_code, false) = true
          AND (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) > 0
        GROUP BY i.id, i.material_id, m.material_code, m.material_name, i.warehouse_id,
                 i.quantity, i.locked_quantity, i.in_transit_quantity
    LOOP
        v_missing_count := COALESCE(v_row.missing_count, 0);
        IF v_missing_count <= 0 THEN
            CONTINUE;
        END IF;

        FOR v_i IN 1..v_missing_count LOOP
            v_serial_code := fn_generate_serial_no('MAT');

            INSERT INTO material_serial_code (
                serial_code,
                material_id,
                material_code,
                material_name,
                inventory_id,
                warehouse_id,
                status,
                created_at,
                updated_at
            ) VALUES (
                v_serial_code,
                v_row.material_id,
                v_row.material_code,
                v_row.material_name,
                v_row.inventory_id,
                v_row.warehouse_id,
                'in_stock',
                NOW(),
                NOW()
            )
            RETURNING id INTO v_serial_id;

            INSERT INTO material_serial_trace (
                serial_code_id,
                serial_code,
                action,
                ref_doc_type,
                ref_doc_no,
                ref_doc_id,
                to_warehouse_id,
                remark,
                created_at
            ) VALUES (
                v_serial_id,
                v_serial_code,
                'stock_in',
                'inventory_repair',
                '库存台账编码补齐',
                v_row.inventory_id,
                v_row.warehouse_id,
                '修复历史/测试数据：编码物料库存存在但缺少在库编码明细',
                NOW()
            );
        END LOOP;
    END LOOP;
END
$$;

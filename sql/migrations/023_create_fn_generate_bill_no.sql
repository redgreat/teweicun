-- ============================================================
-- 功能：创建 fn_generate_bill_no 函数，用于生成付款单/收款单单号
-- 创建时间：2026-06-06
-- 创建人：wangcw
-- MIGRATION_ID: 023_create_fn_generate_bill_no
-- MIGRATION_APPLIED: pending
-- ============================================================

BEGIN;

CREATE OR REPLACE FUNCTION public.fn_generate_bill_no(p_prefix text)
 RETURNS varchar(30)
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_date_str text;
    v_seq int;
    v_result varchar(30);
BEGIN
    v_date_str := to_char(CURRENT_DATE, 'YYYYMMDD');

    SELECT COUNT(*) + 1 INTO v_seq
    FROM (
        SELECT 1 FROM fund_payment WHERE statement_no LIKE p_prefix || v_date_str || '%'
        UNION ALL
        SELECT 1 FROM fund_collection WHERE statement_no LIKE p_prefix || v_date_str || '%'
    ) t;

    v_result := p_prefix || v_date_str || LPAD(v_seq::text, 4, '0');
    RETURN v_result;
END;
$function$;

COMMENT ON FUNCTION public.fn_generate_bill_no(text) IS '生成付款/收款单单号：前缀+YYYYMMDD+4位序号';

COMMIT;

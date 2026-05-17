-- 功能：将编码追踪底表从 sku_serial_* 重命名为 material_serial_*，并保留兼容视图
-- 创建时间：2026-05-17
-- 创建人：GPT-5.4

BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relname = 'sku_serial_code'
          AND c.relkind = 'v'
    ) THEN
        EXECUTE 'DROP VIEW public.sku_serial_code';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relname = 'sku_serial_trace'
          AND c.relkind = 'v'
    ) THEN
        EXECUTE 'DROP VIEW public.sku_serial_trace';
    END IF;
END
$$;

DO $$
BEGIN
    IF to_regclass('public.material_serial_code') IS NULL
       AND to_regclass('public.sku_serial_code') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.sku_serial_code RENAME TO material_serial_code';
    END IF;

    IF to_regclass('public.material_serial_trace') IS NULL
       AND to_regclass('public.sku_serial_trace') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.sku_serial_trace RENAME TO material_serial_trace';
    END IF;
END
$$;

DO $$
BEGIN
    IF to_regclass('public.material_serial_code_id_seq') IS NULL
       AND to_regclass('public.sku_serial_code_id_seq') IS NOT NULL THEN
        EXECUTE 'ALTER SEQUENCE public.sku_serial_code_id_seq RENAME TO material_serial_code_id_seq';
    END IF;

    IF to_regclass('public.material_serial_trace_id_seq') IS NULL
       AND to_regclass('public.sku_serial_trace_id_seq') IS NOT NULL THEN
        EXECUTE 'ALTER SEQUENCE public.sku_serial_trace_id_seq RENAME TO material_serial_trace_id_seq';
    END IF;
END
$$;

ALTER TABLE IF EXISTS public.material_serial_code
    ALTER COLUMN id SET DEFAULT nextval('public.material_serial_code_id_seq'::regclass);

ALTER TABLE IF EXISTS public.material_serial_trace
    ALTER COLUMN id SET DEFAULT nextval('public.material_serial_trace_id_seq'::regclass);

ALTER SEQUENCE IF EXISTS public.material_serial_code_id_seq
    OWNED BY public.material_serial_code.id;

ALTER SEQUENCE IF EXISTS public.material_serial_trace_id_seq
    OWNED BY public.material_serial_trace.id;

DO $$
BEGIN
    IF to_regclass('public.material_serial_code_pkey') IS NULL
       AND to_regclass('public.sku_serial_code_pkey') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.sku_serial_code_pkey RENAME TO material_serial_code_pkey';
    END IF;

    IF to_regclass('public.idx_material_serial_code_code') IS NULL
       AND to_regclass('public.idx_sku_serial_code_code') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_sku_serial_code_code RENAME TO idx_material_serial_code_code';
    END IF;

    IF to_regclass('public.idx_material_serial_code_material') IS NULL
       AND to_regclass('public.idx_sku_serial_code_material') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_sku_serial_code_material RENAME TO idx_material_serial_code_material';
    END IF;

    IF to_regclass('public.idx_material_serial_code_inventory') IS NULL
       AND to_regclass('public.idx_sku_serial_code_inventory') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_sku_serial_code_inventory RENAME TO idx_material_serial_code_inventory';
    END IF;

    IF to_regclass('public.idx_material_serial_code_status') IS NULL
       AND to_regclass('public.idx_sku_serial_code_status') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_sku_serial_code_status RENAME TO idx_material_serial_code_status';
    END IF;

    IF to_regclass('public.material_serial_trace_pkey') IS NULL
       AND to_regclass('public.sku_serial_trace_pkey') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.sku_serial_trace_pkey RENAME TO material_serial_trace_pkey';
    END IF;

    IF to_regclass('public.idx_material_serial_trace_code_id') IS NULL
       AND to_regclass('public.idx_serial_trace_code_id') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_serial_trace_code_id RENAME TO idx_material_serial_trace_code_id';
    END IF;

    IF to_regclass('public.idx_material_serial_trace_code') IS NULL
       AND to_regclass('public.idx_serial_trace_code') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_serial_trace_code RENAME TO idx_material_serial_trace_code';
    END IF;

    IF to_regclass('public.idx_material_serial_trace_action') IS NULL
       AND to_regclass('public.idx_serial_trace_action') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_serial_trace_action RENAME TO idx_material_serial_trace_action';
    END IF;

    IF to_regclass('public.idx_material_serial_trace_created') IS NULL
       AND to_regclass('public.idx_serial_trace_created') IS NOT NULL THEN
        EXECUTE 'ALTER INDEX public.idx_serial_trace_created RENAME TO idx_material_serial_trace_created';
    END IF;
END
$$;

COMMENT ON TABLE public.material_serial_code IS '物料具体编码表（每个编码对应一个具体设备/部件）';
COMMENT ON TABLE public.material_serial_trace IS '物料编码流转记录表（全生命周期追溯）';

CREATE OR REPLACE VIEW public.sku_serial_code AS
SELECT *
FROM public.material_serial_code;

CREATE OR REPLACE VIEW public.sku_serial_trace AS
SELECT *
FROM public.material_serial_trace;

COMMENT ON VIEW public.sku_serial_code IS '兼容视图：请迁移到 material_serial_code';
COMMENT ON VIEW public.sku_serial_trace IS '兼容视图：请迁移到 material_serial_trace';

COMMIT;

<!--
功能：DataGrid.svelte
创建时间：2026-04-18
创建人：CodeArts Agent
-->

<script lang="ts">
	import { browser } from '$app/environment';
	import {
		ChevronLeft,
		ChevronRight,
		Search,
		Plus,
		Filter,
		Download,
		RotateCw
	} from 'lucide-svelte';
	import { dgRowBtn, dgToolbarBtn } from '$lib/dgButtonClasses';

	interface Column {
		key: string;
		label: string;
		sortable?: boolean;
		class?: string;
		width?: string;
	}

	type RowActions = (row: any) => any;
	type CellRender = (key: string, value: any, row: any) => any;
	type OnPageChange = (page: number) => void | Promise<void>;
	type OnCreate = () => void | Promise<void>;
	type OnSearch = (term: string) => void | Promise<void>;
	type OnRefresh = () => void | Promise<void>;
	type OnExport = () => void | Promise<void>;
	type HeaderFilters = () => any;

	type DataGridProps = {
		class?: string;
		columns?: Column[];
		data?: any[];
		total?: number;
		loading?: boolean;
		page?: number;
		pageSize?: number;
		rowActions?: RowActions;
		cellRender?: CellRender;
		onPageChange?: OnPageChange;
		onCreate?: OnCreate;
		onSearch?: OnSearch;
		onRefresh?: OnRefresh;
		onExport?: OnExport;
		showDefaultSearch?: boolean;
		headerFilters?: HeaderFilters;
		actionColumnWidth?: string;
		showActions?: boolean;
	};

	let {
		class: className = '',
		columns = [],
		data = [],
		total = 0,
		loading = false,
		page = $bindable(1),
		pageSize = 10,
		rowActions,
		cellRender,
		onPageChange,
		onCreate,
		onSearch,
		onRefresh,
		onExport,
		showDefaultSearch = true,
		headerFilters,
		actionColumnWidth = '120px',
		showActions = true
	}: DataGridProps = $props();

	let searchTerm = $state('');

	let rootEl = $state<HTMLDivElement | undefined>(undefined);
	let bodyEl = $state<HTMLDivElement | undefined>(undefined);

	const totalPages = $derived(Math.max(1, Math.ceil((total || 0) / (pageSize || 10))));
	const fromRecord = $derived(total > 0 ? (page - 1) * pageSize + 1 : 0);
	const toRecord = $derived(total > 0 ? Math.min(page * pageSize, total) : 0);

	function clamp(n: number, lo: number, hi: number) {
		return Math.max(lo, Math.min(hi, n));
	}

	function applyAdaptiveLayout() {
		const el = bodyEl;
		const root = rootEl;
		if (!el || !root) return;

		const H = el.clientHeight;
		if (H < 48) return;

		// 按「当前页行数」分配垂直空间；加载骨架 5 行；无数据时仍按 pageSize 预留，避免闪动
		const rowCount = loading ? 5 : data.length === 0 ? pageSize : data.length;
		const n = clamp(rowCount, 1, Math.max(pageSize, 10));
		const theadReserve = 50;
		const avail = Math.max(H - theadReserve, n * 32);
		const perRow = avail / n;
		const ideal = 50;
		// 视口够高时 s→1（大字 + 松行距）；变矮时略缩，但 s 有下限避免「又小又挤」
		const s = clamp(perRow / ideal, 0.82, 1);

		const cellFs = 14 + (17 - 14) * s;
		const cellPy = 6 + (13 - 6) * s;
		const headFs = 14 + (16.5 - 14) * s;
		const headPy = 8 + (13 - 8) * s;

		root.style.setProperty('--dg-cell-fs', `${cellFs}px`);
		root.style.setProperty('--dg-cell-py', `${cellPy}px`);
		root.style.setProperty('--dg-head-fs', `${headFs}px`);
		root.style.setProperty('--dg-head-py', `${headPy}px`);
	}

	$effect(() => {
		if (!browser) return;
		const el = bodyEl;
		const root = rootEl;
		if (!el || !root) return;

		applyAdaptiveLayout();
		const ro = new ResizeObserver(() => applyAdaptiveLayout());
		ro.observe(el);
		return () => ro.disconnect();
	});

	// 数据 / 加载态变化后下一帧再量高度（DOM 已更新）
	$effect(() => {
		if (!browser) return;
		void loading;
		void data.length;
		void pageSize;
		requestAnimationFrame(() => applyAdaptiveLayout());
	});

	async function goToPage(nextPage: number) {
		const next = Math.min(Math.max(1, nextPage), totalPages);
		if (next === page) return;
		page = next;
		await onPageChange?.(next);
	}

	async function handleSearchEnter() {
		page = 1;
		await onSearch?.(searchTerm);
		await onPageChange?.(1);
	}

	async function handleRefresh() {
		if (onRefresh) {
			await onRefresh();
			return;
		}
		await onPageChange?.(page);
	}
</script>

<div
	bind:this={rootEl}
	class="dg-root bg-base-100 shadow-base-300/50 border-base-300 flex h-full min-h-0 flex-col overflow-hidden rounded-3xl border shadow-xl {className}"
>
	<!-- Table Header -->
	<div
		class="border-base-200 bg-base-100/50 flex shrink-0 flex-wrap items-center gap-3 border-b px-5 py-4"
	>
		<div class="flex min-w-0 flex-1 flex-wrap items-center gap-2">
			{#if headerFilters}
				{@render headerFilters()}
			{:else if showDefaultSearch}
				<div class="group relative">
					<Search
						size={18}
						class="text-base-content/40 group-focus-within:text-primary absolute top-1/2 left-3 -translate-y-1/2 transition-colors"
					/>
					<input
						type="text"
						placeholder="搜索..."
						bind:value={searchTerm}
						onkeydown={(e) => e.key === 'Enter' && handleSearchEnter()}
						class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-48 rounded-lg border-none pl-11 text-base transition-all"
					/>
				</div>
				<button type="button" class={dgToolbarBtn}>
					<Filter size={17} /> 筛选
				</button>
			{/if}
		</div>

		<div class="flex shrink-0 flex-wrap items-center gap-2">
			<button type="button" class={dgToolbarBtn} onclick={handleRefresh}>
				<RotateCw size={17} /> 刷新
			</button>

			<button type="button" class={dgToolbarBtn} onclick={onExport}>
				<Download size={17} /> 导出
			</button>

			{#if onCreate}
				<button
					class="btn btn-primary shadow-primary/20 h-10 min-h-10 gap-2 rounded-lg px-4 text-base shadow-lg"
					onclick={onCreate}
				>
					<Plus size={18} /> 新增
				</button>
			{/if}
		</div>
	</div>

	<!-- Table Body：字号/行距由 ResizeObserver 写入 CSS 变量，视口矮时自动略缩，避免 10 行反复滚 -->
	<div bind:this={bodyEl} class="min-h-0 flex-1 overflow-auto">
		<table class="table w-full table-fixed leading-normal">
			<colgroup>
				{#each columns as col}
					<col style={col.width ? `width: ${col.width}` : ''} />
				{/each}
				{#if showActions}
					<col style="width: {actionColumnWidth}" />
				{/if}
			</colgroup>
			<thead class="bg-base-200/50">
				<tr>
					{#each columns as col}
						<th
							class="text-base-content/65 font-semibold whitespace-nowrap {col.class}"
							style="font-size: var(--dg-head-fs, 16px); padding: var(--dg-head-py, 12px) 12px;"
							>{col.label}</th
						>
					{/each}
					{#if showActions}
						<th
							class="text-center font-semibold whitespace-nowrap"
							style="font-size: var(--dg-head-fs, 16px); padding: var(--dg-head-py, 12px) 8px;"
							>操作</th
						>
					{/if}
				</tr>
			</thead>
			<tbody class="divide-base-200 divide-y">
				{#if loading}
					{#each Array(5) as _}
						<tr class="animate-pulse">
							{#each columns as _}<td
									style="padding: var(--dg-cell-py, 10px) 12px; font-size: var(--dg-cell-fs, 16px);"
									><div class="bg-base-300 h-4 w-28 rounded"></div></td
								>{/each}
							{#if showActions}
								<td style="padding: var(--dg-cell-py, 10px) 4px;"
									><div class="bg-base-300 mx-auto h-8 w-20 rounded-lg"></div></td
								>
							{/if}
						</tr>
					{/each}
				{:else if data.length === 0}
					<tr>
						<td
							colspan={columns.length + (showActions ? 1 : 0)}
							class="text-base-content/35 text-center text-lg italic"
							style="padding: 3rem 12px; font-size: var(--dg-cell-fs, 16px);"
						>
							暂无数据记录
						</td>
					</tr>
				{:else}
					{#each data as row}
						<tr class="hover:bg-base-200/50 transition-colors">
							{#each columns as col}
								<td
									class="align-middle {col.class}"
									style="padding: var(--dg-cell-py, 10px) 12px; font-size: var(--dg-cell-fs, 16px);"
								>
									{#if cellRender}
										{@render cellRender(col.key, row[col.key], row)}
									{:else}
										<span class="block truncate" title={row[col.key] || ''}
											>{row[col.key] || '-'}</span
										>
									{/if}
								</td>
							{/each}
							{#if showActions}
								<td
									class="text-center align-middle"
									style="padding: var(--dg-cell-py, 10px) 6px; font-size: var(--dg-cell-fs, 16px);"
								>
									{#if rowActions}
										{@render rowActions(row)}
									{:else}
										<button type="button" class={dgRowBtn}>查看</button>
									{/if}
								</td>
							{/if}
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<!-- Table Footer / Pagination -->
	<div
		class="border-base-200 flex shrink-0 flex-wrap items-center justify-between gap-2 border-t px-5 py-3.5"
	>
		<div class="text-base-content/45 text-base">
			显示第 {fromRecord} 到 {toRecord} 条，共 {total} 条
		</div>
		<div class="join border-base-300 border shadow-sm">
			<button
				class="join-item btn bg-base-100 hover:bg-base-200 h-9 min-h-9 border-none px-3"
				disabled={page === 1}
				onclick={() => goToPage(page - 1)}
			>
				<ChevronLeft size={18} />
			</button>
			<button class="join-item btn bg-base-100 h-9 min-h-9 border-none px-4 text-base"
				>第 {page}/{totalPages} 页</button
			>
			<button
				class="join-item btn bg-base-100 hover:bg-base-200 h-9 min-h-9 border-none px-3"
				disabled={page >= totalPages}
				onclick={() => goToPage(page + 1)}
			>
				<ChevronRight size={18} />
			</button>
		</div>
	</div>
</div>

<style>
	/* 首次渲染前给默认值，接近「上上次」偏大字号 */
	.dg-root {
		--dg-cell-fs: 17px;
		--dg-cell-py: 12px;
		--dg-head-fs: 16.5px;
		--dg-head-py: 12px;
	}
</style>

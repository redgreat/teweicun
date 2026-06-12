<!--
功能：库存台账页面（物料-仓库-单价维度）
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { Hash, X, QrCode, SlidersHorizontal } from 'lucide-svelte';
	import { dgToolbarBtn } from '$lib/dgButtonClasses';
	import { formatDateTimeInCn } from '$lib/datetime';

	let items = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	let stats = $state({
		total_amount: 0,
		code_total_amount: 0,
		no_code_total_amount: 0,
		total_locked_qty: 0
	});

	let filters = $state({
		material_name: '',
		warehouse_name: '',
		price_min: '',
		price_max: ''
	});

	let serialModal = $state({
		show: false,
		title: '',
		loading: false,
		rows: [] as any[]
	});

	const columns = [
		{ key: 'material_name', label: '物料名称', width: '240px' },
		{ key: 'warehouse_name', label: '所在仓库', width: '140px' },
		{ key: 'is_code', label: '是否编码', width: '100px' },
		{ key: 'book_quantity', label: '账面库存', width: '110px', class: 'text-right font-mono' },
		{ key: 'locked_quantity', label: '锁定数量', width: '110px', class: 'text-right font-mono' },
		{
			key: 'in_transit_quantity',
			label: '在途数量',
			width: '110px',
			class: 'text-right font-mono'
		},
		{
			key: 'serial_reserved_quantity',
			label: '编码备货数',
			width: '120px',
			class: 'text-right font-mono'
		},
		{ key: 'quantity', label: '可用数量', width: '110px', class: 'text-right font-mono' },
		{ key: 'unit_cost', label: '单价', width: '110px', class: 'text-right font-mono' },
		{
			key: 'total_amount',
			label: '总价',
			width: '120px',
			class: 'text-right font-mono text-success'
		},
		{ key: 'attrs', label: '属性', width: '70px', class: 'text-center' },
		{ key: 'serials', label: '编码', width: '70px', class: 'text-center' }
	];

	function fmtMoney(v: number) {
		return (
			'¥' +
			(Number(v) || 0).toLocaleString('zh-CN', {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2
			})
		);
	}

	function fmtQty(v: number) {
		return (Number(v) || 0).toLocaleString('zh-CN', { maximumFractionDigits: 3 });
	}

	async function loadInventory(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.material_name.trim()) params.set('material_name', filters.material_name.trim());
			if (filters.warehouse_name.trim())
				params.set('warehouse_name', filters.warehouse_name.trim());
			if (filters.price_min.trim()) params.set('price_min', filters.price_min.trim());
			if (filters.price_max.trim()) params.set('price_max', filters.price_max.trim());

			const res: any = await api.get(`/inventory/material-ledger?${params.toString()}`);
			items = res.list || [];
			total = res.total || 0;
			stats = res.stats || stats;
		} finally {
			loading = false;
		}
	}

	function resetFilters() {
		filters = { material_name: '', warehouse_name: '', price_min: '', price_max: '' };
		loadInventory(1);
	}

	function search() {
		loadInventory(1);
	}

	async function exportExcel() {
		try {
			const params = new URLSearchParams();
			if (filters.material_name.trim()) params.set('material_name', filters.material_name.trim());
			if (filters.warehouse_name.trim())
				params.set('warehouse_name', filters.warehouse_name.trim());
			if (filters.price_min.trim()) params.set('price_min', filters.price_min.trim());
			if (filters.price_max.trim()) params.set('price_max', filters.price_max.trim());

			const blob: Blob = await api.get(`/inventory/material-ledger/export?${params.toString()}`, {
				responseType: 'blob'
			} as any);
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `库存台账_${formatDateTimeInCn(new Date()).replace(/[-:\s]/g, '')}.xlsx`;
			document.body.appendChild(a);
			a.click();
			a.remove();
			window.URL.revokeObjectURL(url);
			toast.success('导出成功');
		} catch (err: any) {
			toast.error('导出失败: ' + (err?.message || err));
		}
	}

	function serialStatusBadgeClass(status: string) {
		const s = String(status || '').toLowerCase();
		if (s === 'stock_out_reserved') return 'badge-warning';
		if (s === 'stock_in_reserved') return 'badge-info';
		if (s === 'in_stock') return 'badge-success';
		if (s === 'issued') return 'badge-warning';
		if (s === 'returned') return 'badge-info';
		if (s === 'scrapped') return 'badge-error';
		return 'badge-ghost';
	}

	function serialStatusLabel(row: any) {
		const s = String(row?.display_status || row?.status || '').toLowerCase();
		if (row?.display_status_name) return row.display_status_name;
		if (s === 'stock_out_reserved') return '出库备货中';
		if (s === 'stock_in_reserved') return '退料备货中';
		if (s === 'in_stock') return '在库';
		if (s === 'issued') return '已领用';
		if (s === 'returned') return '已退回';
		if (s === 'scrapped') return '已报废';
		return row?.status_name || row?.status || '-';
	}

	async function viewSerials(row: any) {
		if (!row?.is_code) return;
		serialModal = {
			show: true,
			title: `${row.material_name} / ${row.warehouse_name}`,
			loading: true,
			rows: []
		};
		try {
			const params = new URLSearchParams({
				material_id: String(row.material_id),
				warehouse_id: String(row.warehouse_id),
				unit_cost: String(row.unit_cost || 0)
			});
			const res: any = await api.get(`/inventory/material-ledger/serials?${params.toString()}`);
			serialModal = { ...serialModal, loading: false, rows: res || [] };
		} catch {
			serialModal = { ...serialModal, loading: false, rows: [] };
		}
	}

	onMount(() => loadInventory(1));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-6">
	<div class="grid grid-cols-1 gap-4 md:grid-cols-4">
		<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
			<div class="text-base-content/60 text-sm">在库总金额</div>
			<div class="text-success mt-1 text-xl font-semibold">{fmtMoney(stats.total_amount)}</div>
		</div>
		<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
			<div class="text-base-content/60 text-sm">编码物料总金额</div>
			<div class="text-success mt-1 text-xl font-semibold">{fmtMoney(stats.code_total_amount)}</div>
		</div>
		<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
			<div class="text-base-content/60 text-sm">无编码物料总金额</div>
			<div class="text-success mt-1 text-xl font-semibold">
				{fmtMoney(stats.no_code_total_amount)}
			</div>
		</div>
		<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
			<div class="text-base-content/60 text-sm">总锁定数量</div>
			<div class="mt-1 text-xl font-semibold">{fmtQty(stats.total_locked_qty)}</div>
		</div>
	</div>

	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={items}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadInventory}
		onExport={exportExcel}
		showDefaultSearch={false}
		showActions={false}
	>
		{#snippet headerFilters()}
			<div
				class="scrollbar-hide flex min-w-0 flex-nowrap items-center gap-2 overflow-x-auto py-0.5"
			>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[12rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="物料名称/编码"
					bind:value={filters.material_name}
					onkeydown={(e) => e.key === 'Enter' && search()}
				/>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="所在仓库"
					bind:value={filters.warehouse_name}
					onkeydown={(e) => e.key === 'Enter' && search()}
				/>
				<input
					type="number"
					step="0.01"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="最低价"
					bind:value={filters.price_min}
					onkeydown={(e) => e.key === 'Enter' && search()}
				/>
				<input
					type="number"
					step="0.01"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="最高价"
					bind:value={filters.price_max}
					onkeydown={(e) => e.key === 'Enter' && search()}
				/>
				<button type="button" class={dgToolbarBtn} onclick={search}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}

		{#snippet cellRender(key, value, row)}
			{#if key === 'material_name'}
				<div class="block w-full truncate" title={value || '-'}>
					{value || '-'}
				</div>
			{:else if key === 'is_code'}
				<span class="badge badge-sm {row.is_code ? 'badge-info' : 'badge-ghost'}">
					{row.is_code ? '有编码' : '无编码'}
				</span>
			{:else if
				key === 'book_quantity' ||
				key === 'locked_quantity' ||
				key === 'in_transit_quantity' ||
				key === 'serial_reserved_quantity' ||
				key === 'quantity'}
				{fmtQty(value)}
			{:else if key === 'unit_cost' || key === 'total_amount'}
				{fmtMoney(value)}
			{:else if key === 'attrs'}
				{#if row.has_custom_attrs}
					<button class="btn btn-xs btn-ghost text-primary" disabled title="存在自定义属性">
						<SlidersHorizontal size={14} />
					</button>
				{:else}
					<span class="text-base-content/30">-</span>
				{/if}
			{:else if key === 'serials'}
				{#if row.is_code}
					<button
						class="btn btn-xs btn-ghost text-primary"
						onclick={() => viewSerials(row)}
						title="查看编码"
					>
						<QrCode size={14} />
					</button>
				{:else}
					<span class="text-base-content/40">-</span>
				{/if}
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
	</DataGrid>
</div>

{#if serialModal.show}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm"
		role="button"
		tabindex="0"
		onclick={() => (serialModal.show = false)}
		onkeydown={(e) => e.key === 'Escape' && (serialModal.show = false)}
	>
		<div
			class="bg-base-100 border-base-300 flex max-h-[80vh] w-full max-w-3xl flex-col rounded-2xl border shadow-2xl"
			role="button"
			tabindex="0"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (serialModal.show = false)}
		>
			<div class="border-base-200 flex items-center justify-between border-b p-5">
				<div class="flex items-center gap-2">
					<Hash size={18} class="text-primary" />
					<h3 class="text-lg font-bold">编码明细</h3>
					<span class="text-base-content/50 text-sm">{serialModal.title}</span>
				</div>
				<button class="btn btn-sm btn-ghost btn-circle" onclick={() => (serialModal.show = false)}>
					<X size={18} />
				</button>
			</div>
			<div class="flex-1 overflow-y-auto p-5">
				{#if serialModal.loading}
					<div class="py-10 text-center">
						<span class="loading loading-spinner loading-lg"></span>
					</div>
				{:else if serialModal.rows.length === 0}
					<div class="text-base-content/40 py-10 text-center">暂无编码数据</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table-zebra table w-full text-sm">
							<thead>
								<tr><th>编码</th><th>编码状态</th></tr>
							</thead>
							<tbody>
								{#each serialModal.rows as r}
									<tr>
										<td class="font-mono">{r.serial_code}</td>
										<td>
											<span
												class="badge badge-sm whitespace-nowrap {serialStatusBadgeClass(
													r.display_status || r.status
												)}"
											>
												{serialStatusLabel(r)}
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

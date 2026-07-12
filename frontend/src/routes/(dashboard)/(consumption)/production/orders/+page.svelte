<!--
功能：生产单据列表（支持手动新增和编辑成本）
创建时间：2026-06-06 / 更新：2026-07-12
创建人：GPT-5.2 / Hermes Agent
-->

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import DataGrid from '$lib/components/DataGrid.svelte';
	import CopyableNo from '$lib/components/CopyableNo.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { formatDateInCn } from '$lib/datetime';
	import { Plus } from 'lucide-svelte';

	let rows = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	let filters = $state({
		production_no: '',
		status: ''
	});

	const columns = [
		{ key: 'production_no', label: '生产单号', class: 'font-mono text-primary', width: '16%' },
		{ key: 'consumption_order_no', label: '领料单号', class: 'font-mono', width: '16%' },
		{ key: 'stock_in_no', label: '入库单号', class: 'font-mono', width: '16%' },
		{ key: 'produced_material_name', label: '成品物料', width: '18%' },
		{ key: 'produced_quantity', label: '数量', class: 'font-mono text-right pr-6', width: '10%' },
		{ key: 'status', label: '状态', width: '10%' },
		{ key: 'created_at', label: '创建时间', width: '14%', class: 'text-center' }
	];

	function statusName(v: string) {
		const map: any = { completed: '已完成' };
		return map[v] || v || '-';
	}

	function navigateToDetail(id: number) {
		goto(`/production/orders/${id}`);
	}

	function fmtDate(v: string) {
		if (!v) return '-';
		return formatDateInCn(v);
	}

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.production_no.trim()) params.set('production_no', filters.production_no.trim());
			if (filters.status) params.set('status', filters.status);
			const res: any = await api.get(`/production/orders?${params.toString()}`);
			rows = res.list || [];
			total = res.total || 0;
		} catch (e) {
			console.error(e);
			toast.error('加载生产单列表失败');
		} finally {
			loading = false;
		}
	}

	function handleFilterSearch() {
		currentPage = 1;
		loadData(1);
	}

	function resetFilters() {
		filters = { production_no: '', status: '' };
		currentPage = 1;
		loadData(1);
	}

	onMount(() => {
		loadData(1);
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-lg font-semibold">生产单列表</h2>
		<button type="button" class="btn btn-primary btn-sm gap-1" onclick={() => goto('/production/orders/create')}>
			<Plus size={14} /> 新增生产单
		</button>
	</div>
	<DataGrid
		class="min-h-0 flex-1"
		columns={columns}
		data={rows}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		showDefaultSearch={false}
		showActions={false}
		onExport={() => {}}
	>
		{#snippet headerFilters()}
			<div class="scrollbar-hide flex min-w-0 flex-nowrap items-center gap-2 overflow-x-auto py-0.5">
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[12rem] shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="生产单号"
					bind:value={filters.production_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none py-0 pr-8 pl-2 text-base leading-tight"
					bind:value={filters.status}
				>
					<option value="">状态</option>
					<option value="completed">已完成</option>
				</select>
				<button type="button" class="btn btn-ghost h-10 min-h-10 rounded-lg" onclick={handleFilterSearch}>
					查询
				</button>
				<button type="button" class="btn btn-ghost h-10 min-h-10 rounded-lg" onclick={resetFilters}>
					重置
				</button>
			</div>
		{/snippet}

		{#snippet cellRender(key, value, row)}
			{#if key === 'production_no'}
				<CopyableNo value={value} onOpen={() => navigateToDetail(row.id)} />
			{:else if key === 'consumption_order_no'}
				{#if row.consumption_order_id}
					<CopyableNo value={value} href={`/consumption/orders/${row.consumption_order_id}`} />
				{:else}
					<span class="text-base-content/30">-</span>
				{/if}
			{:else if key === 'stock_in_no'}
				{#if row.stock_in_id}
					<CopyableNo value={value} href={`/stock/in/${row.stock_in_id}`} />
				{:else}
					<span class="text-base-content/30">-</span>
				{/if}
			{:else if key === 'produced_material_name'}
				<span class="block truncate" title={value || '-'}>
					{value || '-'}
				</span>
			{:else if key === 'produced_quantity'}
				{Number(value || 0).toFixed(3)}
			{:else if key === 'status'}
				<span class="badge badge-success badge-md">{statusName(value)}</span>
			{:else if key === 'created_at'}
				{fmtDate(value)}
			{:else}
				{value ?? '-'}
			{/if}
		{/snippet}
	</DataGrid>
</div>

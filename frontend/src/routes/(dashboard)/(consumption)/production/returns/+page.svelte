<!--
功能：生产退货单列表（预留，仅展示）
创建时间：2026-06-06
创建人：GPT-5.2
-->

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import DataGrid from '$lib/components/DataGrid.svelte';
	import CopyableNo from '$lib/components/CopyableNo.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { formatDateInCn } from '$lib/datetime';

	let rows = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	let filters = $state({
		return_no: '',
		status: ''
	});

	const columns = [
		{ key: 'return_no', label: '退货单号', class: 'font-mono text-primary', width: '18%' },
		{ key: 'production_no', label: '生产单号', class: 'font-mono', width: '18%' },
		{ key: 'produced_material_name', label: '成品物料', width: '22%' },
		{ key: 'returned_quantity', label: '退回数量', class: 'font-mono text-right pr-6', width: '12%' },
		{ key: 'status', label: '状态', width: '12%' },
		{ key: 'created_at', label: '创建时间', width: '18%', class: 'text-center' }
	];

	function navigateToDetail(id: number) {
		goto(`/production/returns/${id}`);
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
			if (filters.return_no.trim()) params.set('return_no', filters.return_no.trim());
			if (filters.status) params.set('status', filters.status);
			const res: any = await api.get(`/production/returns?${params.toString()}`);
			rows = res.list || [];
			total = res.total || 0;
		} catch (e) {
			console.error(e);
			toast.error('加载生产退货单列表失败');
		} finally {
			loading = false;
		}
	}

	function handleFilterSearch() {
		currentPage = 1;
		loadData(1);
	}

	function resetFilters() {
		filters = { return_no: '', status: '' };
		currentPage = 1;
		loadData(1);
	}

	onMount(() => {
		loadData(1);
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
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
					placeholder="退货单号"
					bind:value={filters.return_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[12rem] shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="状态"
					bind:value={filters.status}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<button type="button" class="btn btn-ghost h-10 min-h-10 rounded-lg" onclick={handleFilterSearch}>
					查询
				</button>
				<button type="button" class="btn btn-ghost h-10 min-h-10 rounded-lg" onclick={resetFilters}>
					重置
				</button>
			</div>
		{/snippet}

		{#snippet cellRender(key, value, row)}
			{#if key === 'return_no'}
				<CopyableNo value={value} onOpen={() => navigateToDetail(row.id)} />
			{:else if key === 'production_no'}
				<CopyableNo value={value} />
			{:else if key === 'returned_quantity'}
				{Number(value || 0).toFixed(3)}
			{:else if key === 'created_at'}
				{fmtDate(value)}
			{:else}
				{value ?? '-'}
			{/if}
		{/snippet}
	</DataGrid>
</div>

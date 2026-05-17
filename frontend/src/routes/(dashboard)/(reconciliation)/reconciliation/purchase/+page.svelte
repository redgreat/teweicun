<!--
功能：采购对账（付款单）列表
创建时间：2026-05-17
创建人：wangcw
-->
<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { onMount } from 'svelte';
	import { FileText } from 'lucide-svelte';
	import { dgRowBtn, dgToolbarBtn } from '$lib/dgButtonClasses';
	import { goto } from '$app/navigation';
	import { formatDateInCn } from '$lib/datetime';

	let items = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let filters = $state({ statement_no: '', supplier_id: '' });

	const columns = [
		{ key: 'statement_no', label: '对账单号', class: 'font-mono text-primary' },
		{ key: 'supplier_name', label: '供应商名称' },
		{ key: 'statement_date', label: '单据日期' },
		{ key: 'payment_amount', label: '付款金额', class: 'text-right font-mono text-success pr-6' },
		{ key: 'settlement_method', label: '结算方式' },
		{ key: 'status', label: '状态' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.statement_no) params.set('statement_no', filters.statement_no);
			const res: any = await api.get(`/fund/payments?${params.toString()}`);
			items = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function handleSearch() {
		loadData(1);
	}
	function resetFilters() {
		filters = { statement_no: '', supplier_id: '' };
		loadData(1);
	}
	function navigateToCreate() {
		goto('/reconciliation/purchase/create');
	}

	onMount(() => {
		loadData(1);
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={items}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		onCreate={navigateToCreate}
		onRefresh={() => loadData(currentPage)}
		showDefaultSearch={false}
		actionColumnWidth="120px"
	>
		{#snippet headerFilters()}
			<div class="flex items-center gap-2">
				<input
					type="text"
					class="input bg-base-200 h-10 w-48 rounded-lg"
					placeholder="单号"
					bind:value={filters.statement_no}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value, row)}
			{#if key === 'payment_amount'}
				<span class="text-success font-mono font-semibold">¥{(value || 0).toFixed(2)}</span>
			{:else if key === 'status'}
				<span class="badge badge-md {value === 'completed' ? 'badge-success' : 'badge-ghost'}">
					{value === 'completed' ? '已完成' : '草稿'}
				</span>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(row)}
			<div class="flex justify-center gap-1">
				<a class={dgRowBtn} href={`/reconciliation/purchase/${row.id}`}
					><FileText size={16} /> 详情</a
				>
			</div>
		{/snippet}
	</DataGrid>
</div>

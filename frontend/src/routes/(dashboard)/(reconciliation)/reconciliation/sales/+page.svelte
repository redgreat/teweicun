<!--
功能：销售对账（收款单）列表
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

	let items = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let filters = $state({ statement_no: '', customer_id: '' });
	let customerOptions = $state<any[]>([]);

	const columns = [
		{ key: 'statement_no', label: '对账单号', class: 'font-mono text-primary' },
		{ key: 'customer_name', label: '客户名称' },
		{ key: 'statement_date', label: '单据日期' },
		{ key: 'bill_type', label: '票款类型' },
		{ key: 'collection_amount', label: '销售/抵充', class: 'text-right font-mono pr-6' },
		{ key: 'invoice_amount', label: '票据金额', class: 'text-right font-mono pr-6' },
		{ key: 'actual_amount', label: '实际收款', class: 'text-right font-mono text-success pr-6' },
		{ key: 'difference_amount', label: '差额', class: 'text-right font-mono pr-6' },
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
			if (filters.customer_id) params.set('customer_id', filters.customer_id);
			const res: any = await api.get(`/fund/collections?${params.toString()}`);
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
		filters = { statement_no: '', customer_id: '' };
		loadData(1);
	}
	function navigateToCreate() {
		goto('/reconciliation/sales/create');
	}

	async function loadCustomers() {
		try {
			const res: any = await api.get('/base/partners/dropdown?type=customer&limit=1000&status=enabled');
			customerOptions = res || [];
		} catch (err) {
			console.error(err);
			customerOptions = [];
		}
	}

	onMount(() => {
		loadData(1);
		loadCustomers();
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
				<select class="select bg-base-200 h-10 w-56 rounded-lg" bind:value={filters.customer_id}>
					<option value="">客户</option>
					{#each customerOptions as c}
						<option value={String(c.id)}>{c.name}</option>
					{/each}
				</select>
				<button type="button" class={dgToolbarBtn} onclick={handleSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value, row)}
			{#if ['collection_amount', 'invoice_amount', 'actual_amount', 'difference_amount'].includes(key)}
				<span
					class="font-mono font-semibold {key === 'actual_amount'
						? 'text-success'
						: key === 'difference_amount' && Math.abs(value || 0) > 0.005
							? 'text-warning'
							: ''}">¥{(value || 0).toFixed(2)}</span
				>
			{:else if key === 'bill_type'}
				{value === 'invoice' ? '票据' : value === 'offset' ? '抵充' : '款项'}
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
				<a class={dgRowBtn} href={`/reconciliation/sales/${row.id}`}><FileText size={16} /> 详情</a>
			</div>
		{/snippet}
	</DataGrid>
</div>

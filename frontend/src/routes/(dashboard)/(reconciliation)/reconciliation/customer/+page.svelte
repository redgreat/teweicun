<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { onMount } from 'svelte';
	import { dgToolbarBtn } from '$lib/dgButtonClasses';

	let items = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 20;

	let filters = $state({
		start_date: '',
		end_date: '',
		keyword: '',
		customer_id: ''
	});
	let customerOptions = $state<any[]>([]);

	const columns = [
		{ key: 'customer_code', label: '客户编码', class: 'font-mono text-primary' },
		{ key: 'customer_name', label: '客户名称' },
		{ key: 'receivable_amount', label: '应收金额', class: 'text-right font-mono pr-6' },
		{ key: 'verified_amount', label: '已核销', class: 'text-right font-mono pr-6' },
		{ key: 'balance_amount', label: '应收余额', class: 'text-right font-mono pr-6' },
		{ key: 'actual_amount', label: '实际收款', class: 'text-right font-mono text-success pr-6' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			if (filters.keyword) params.set('keyword', filters.keyword);
			if (filters.customer_id) params.set('customer_id', filters.customer_id);
			const res: any = await api.get(`/reports/reconciliation/customers?${params.toString()}`);
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
		filters = { start_date: '', end_date: '', keyword: '', customer_id: '' };
		loadData(1);
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
		onRefresh={() => loadData(currentPage)}
		showDefaultSearch={false}
		showActions={false}
	>
		{#snippet headerFilters()}
			<div class="flex flex-wrap items-center gap-2">
				<input
					type="date"
					class="input bg-base-200 h-10 w-40 rounded-lg"
					bind:value={filters.start_date}
				/>
				<span class="opacity-60">至</span>
				<input
					type="date"
					class="input bg-base-200 h-10 w-40 rounded-lg"
					bind:value={filters.end_date}
				/>
				<select class="select bg-base-200 h-10 w-56 rounded-lg" bind:value={filters.customer_id}>
					<option value="">客户</option>
					{#each customerOptions as c}
						<option value={String(c.id)}>{c.name}</option>
					{/each}
				</select>
				<input
					type="text"
					class="input bg-base-200 h-10 w-56 rounded-lg"
					placeholder="客户编码/名称"
					bind:value={filters.keyword}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value)}
			{#if ['receivable_amount', 'verified_amount', 'balance_amount', 'actual_amount'].includes(key)}
				<span
					class="font-mono font-semibold {key === 'actual_amount'
						? 'text-success'
						: key === 'balance_amount' && Math.abs(value || 0) > 0.005
							? 'text-warning'
							: ''}">¥{(value || 0).toFixed(2)}</span
				>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
	</DataGrid>
</div>

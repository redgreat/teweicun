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
		supplier_id: ''
	});
	let supplierOptions = $state<any[]>([]);

	const columns = [
		{ key: 'supplier_code', label: '供应商编码', class: 'font-mono text-primary' },
		{ key: 'supplier_name', label: '供应商名称' },
		{ key: 'payable_amount', label: '应付金额', class: 'text-right font-mono pr-6' },
		{ key: 'verified_amount', label: '已核销', class: 'text-right font-mono pr-6' },
		{ key: 'balance_amount', label: '应付余额', class: 'text-right font-mono pr-6' },
		{ key: 'actual_amount', label: '实际付款', class: 'text-right font-mono text-success pr-6' }
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
			if (filters.supplier_id) params.set('supplier_id', filters.supplier_id);
			const res: any = await api.get(`/reports/reconciliation/suppliers?${params.toString()}`);
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
		filters = { start_date: '', end_date: '', keyword: '', supplier_id: '' };
		loadData(1);
	}

	async function loadSuppliers() {
		try {
			const res: any = await api.get('/base/partners/dropdown?type=supplier&limit=1000&status=enabled');
			supplierOptions = res || [];
		} catch (err) {
			console.error(err);
			supplierOptions = [];
		}
	}

	onMount(() => {
		loadData(1);
		loadSuppliers();
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
				<select class="select bg-base-200 h-10 w-56 rounded-lg" bind:value={filters.supplier_id}>
					<option value="">供应商</option>
					{#each supplierOptions as s}
						<option value={String(s.id)}>{s.name}</option>
					{/each}
				</select>
				<input
					type="text"
					class="input bg-base-200 h-10 w-56 rounded-lg"
					placeholder="供应商编码/名称"
					bind:value={filters.keyword}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value)}
			{#if ['payable_amount', 'verified_amount', 'balance_amount', 'actual_amount'].includes(key)}
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

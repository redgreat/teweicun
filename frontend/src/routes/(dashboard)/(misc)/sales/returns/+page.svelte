<!--
功能：销售退货列表页
创建时间：2026-05-10
创建人：GPT-5.4
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { FileText, CheckCircle, Trash2 } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger, dgRowBtnSuccess, dgToolbarBtn } from '$lib/dgButtonClasses';
	import { getStatusStyle } from '$lib/statusStyles';
	import { formatDateInCn } from '$lib/datetime';

	let orders = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	let filters = $state({
		return_no: '',
		customer_keyword: '',
		status: '',
		start_date: '',
		end_date: ''
	});

	const columns = [
		{ key: 'return_no', label: '退货单号', class: 'font-mono text-primary', width: '15%' },
		{ key: 'customer_name', label: '客户名称', width: '16%' },
		{ key: 'warehouse_name', label: '入库仓库', width: '14%' },
		{ key: 'return_date', label: '退货日期', class: 'text-center', width: '12%' },
		{ key: 'total_amount', label: '退货金额', class: 'text-right pr-4', width: '14%' },
		{ key: 'status', label: '状态', class: 'text-center', width: '10%' }
	];

	function formatDate(value: string) {
		if (!value) return '-';
		return formatDateInCn(value);
	}

	function formatMoney(value: number) {
		const amount = Number(value) || 0;
		return (
			'¥' + amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
		);
	}

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			params.set('return_type', 'sales_return');
			if (filters.return_no.trim()) params.set('return_no', filters.return_no.trim());
			if (filters.customer_keyword.trim())
				params.set('customer_keyword', filters.customer_keyword.trim());
			if (filters.status) params.set('status', filters.status);
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/returns?${params.toString()}`);
			orders = res.list || [];
			total = res.total || 0;
		} catch (err: any) {
			toast.error('加载销售退货失败: ' + (err?.message || err));
		} finally {
			loading = false;
		}
	}

	function handleFilterSearch() {
		loadData(1);
	}

	function resetFilters() {
		filters = {
			return_no: '',
			customer_keyword: '',
			status: '',
			start_date: '',
			end_date: ''
		};
		loadData(1);
	}

	function navigateToCreate() {
		goto('/sales/returns/create');
	}

	async function handleConfirm(row: any) {
		try {
			await api.post(`/returns/${row.id}/confirm`, {});
			toast.success('销售退货已提交，已生成待入库单');
			loadData(currentPage);
		} catch (err: any) {
			toast.error('确认失败: ' + (err?.message || err));
		}
	}

	async function handleDelete(row: any) {
		try {
			await api.delete(`/returns/${row.id}`);
			toast.success('销售退货已删除');
			loadData(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	onMount(() => {
		loadData(1);
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={orders}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		onCreate={navigateToCreate}
		onRefresh={() => loadData(currentPage)}
		showDefaultSearch={false}
		actionColumnWidth="250px"
	>
		{#snippet headerFilters()}
			<div
				class="scrollbar-hide flex min-w-0 flex-nowrap items-center gap-2 overflow-x-auto py-0.5"
			>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[9rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="退货单号"
					bind:value={filters.return_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[11rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="客户名称/编码"
					bind:value={filters.customer_keyword}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8rem] shrink-0 rounded-lg border-none py-0 pr-7 pl-2 text-base leading-tight"
					bind:value={filters.status}
				>
					<option value="">状态</option>
					<option value="draft">待提交</option>
					<option value="confirmed">待入库</option>
					<option value="completed">已完成</option>
				</select>
				<input
					type="date"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.start_date}
				/>
				<span class="text-base-content/45 shrink-0 text-sm">—</span>
				<input
					type="date"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.end_date}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleFilterSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}

		{#snippet cellRender(key, value, row)}
			{#if key === 'return_no'}
				<a
					href={`/sales/returns/${row.id}`}
					class="link link-primary font-mono no-underline hover:underline"
				>
					{value || '-'}
				</a>
			{:else if key === 'status'}
				{@const style = getStatusStyle(value, 'sales_return')}
				<span class="badge badge-md whitespace-nowrap {style.class}">{style.label}</span>
			{:else if key === 'total_amount'}
				<span class="text-success font-semibold">{formatMoney(value)}</span>
			{:else if key === 'return_date'}
				<span class="whitespace-nowrap">{formatDate(value)}</span>
			{:else}
				<span class="block truncate" title={value || '-'}>{value || '-'}</span>
			{/if}
		{/snippet}

		{#snippet rowActions(row)}
			<div class="flex flex-nowrap items-center justify-center gap-1.5">
				<a class={dgRowBtn} href={`/sales/returns/${row.id}`}>
					<FileText size={16} /> 详情
				</a>
				{#if row.status === 'draft'}
					<button type="button" class={dgRowBtnSuccess} onclick={() => handleConfirm(row)}>
						<CheckCircle size={16} /> 提交
					</button>
					<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(row)}>
						<Trash2 size={16} /> 删除
					</button>
				{/if}
			</div>
		{/snippet}
	</DataGrid>
</div>

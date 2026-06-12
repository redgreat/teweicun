<!--
功能：领料订单列表页面（设计员创建领料单）
创建时间：2026-04-23
创建人：CodeArts Agent
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import CopyableNo from '$lib/components/CopyableNo.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { FileText } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { dgRowBtn, dgToolbarBtn } from '$lib/dgButtonClasses';
	import { formatDateInCn } from '$lib/datetime';

	let orders = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let filters = $state({
		order_no: '',
		project_no: '',
		product_name: '',
		status: '',
		start_date: '',
		end_date: ''
	});

	const columns = [
		{ key: 'order_no', label: '领料单号', class: 'font-mono text-primary', width: '14%' },
		{ key: 'project_no', label: '项目编号', width: '12%' },
		{ key: 'product_name', label: '产品名称', width: '15%' },
		{ key: 'order_date', label: '领料日期', width: '11%', class: 'text-center' },
		{ key: 'stock_out_no', label: '出库单号', class: 'font-mono text-center', width: '15%' },
		{
			key: 'total_amount',
			label: '总金额',
			width: '10%',
			class: 'text-left font-mono text-success pr-6'
		},
		{ key: 'status', label: '状态', width: '9%', class: 'text-left' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.order_no.trim()) params.set('order_no', filters.order_no.trim());
			if (filters.project_no.trim()) params.set('project_no', filters.project_no.trim());
			if (filters.product_name.trim()) params.set('product_name', filters.product_name.trim());
			if (filters.status) params.set('status', filters.status);
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/consumption/orders?${params.toString()}`);
			orders = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
			toast.error('加载领料订单列表失败');
		} finally {
			loading = false;
		}
	}

	function handleFilterSearch() {
		currentPage = 1;
		loadData(1);
	}

	function resetFilters() {
		filters = {
			order_no: '',
			project_no: '',
			product_name: '',
			status: '',
			start_date: '',
			end_date: ''
		};
		currentPage = 1;
		loadData(1);
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatAmount(amount: number) {
		return '¥' + (amount || 0).toFixed(2);
	}

	function statusName(status: string) {
		const map: any = {
			draft: '待提交',
			pending: '待出库',
			confirmed: '待出库',
			completed: '已完成',
			cancelled: '已取消'
		};
		return map[status] || status || '-';
	}

	function navigateToCreate() {
		goto('/consumption/orders/create');
	}

	function navigateToDetail(id: number) {
		goto(`/consumption/orders/${id}`);
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
		showDefaultSearch={false}
	>
		{#snippet headerFilters()}
			<div
				class="scrollbar-hide flex min-w-0 flex-nowrap items-center gap-2 overflow-x-auto py-0.5"
			>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="领料单号"
					bind:value={filters.order_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="项目编号"
					bind:value={filters.project_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[12rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="产品名称"
					bind:value={filters.product_name}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8rem] shrink-0 rounded-lg border-none py-0 pr-8 pl-2 text-base leading-tight"
					bind:value={filters.status}
				>
					<option value="">状态</option>
					<option value="draft">待提交</option>
					<option value="confirmed">待出库</option>
					<option value="completed">已完成</option>
				</select>
				<input
					type="date"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-3 text-base"
					bind:value={filters.start_date}
				/>
				<input
					type="date"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-3 text-base"
					bind:value={filters.end_date}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleFilterSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value, row)}
			{#if key === 'order_no'}
				<CopyableNo value={value} onOpen={() => navigateToDetail(row.id)} />
			{:else if key === 'product_name'}
				<span class="block truncate" title={value || '-'}>
					{value || '-'}
				</span>
			{:else if key === 'stock_out_no'}
				{#if row.stock_out_id}
					<CopyableNo value={value} href={`/stock/out/${row.stock_out_id}`} title="查看关联出库单" />
				{:else}
					<span class="text-base-content/30">-</span>
				{/if}
			{:else if key === 'order_date'}
				{formatDate(value)}
			{:else if key === 'status'}
				<span
					class="badge badge-md max-w-full whitespace-nowrap {row.status === 'draft'
						? 'badge-warning'
						: row.status === 'pending'
							? 'badge-warning'
							: row.status === 'confirmed'
								? 'badge-info'
								: row.status === 'completed'
									? 'badge-success'
									: 'badge-error'}">{statusName(value)}</span
				>
			{:else if key === 'total_amount'}
				<span class="text-success font-mono font-semibold">{formatAmount(value)}</span>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(order)}
			<div class="flex flex-wrap items-center justify-center gap-1.5">
				<button type="button" class={dgRowBtn} onclick={() => navigateToDetail(order.id)}>
					<FileText size={15} /> 详情
				</button>
			</div>
		{/snippet}
	</DataGrid>
</div>

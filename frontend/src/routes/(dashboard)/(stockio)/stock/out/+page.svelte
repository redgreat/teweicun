<!--
功能：出库管理页面(库管员审核)
创建时间：2026-04-22
创建人：Trae
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { FileText, CheckCircle } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnSuccess, dgToolbarBtn } from '$lib/dgButtonClasses';

	import { getStatusStyle } from '$lib/statusStyles';
	import { formatDateInCn } from '$lib/datetime';

	let stockOuts = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	let filters = $state({
		stock_out_no: '',
		out_type: '',
		receiver: '',
		status: '',
		start_date: '',
		end_date: ''
	});

	const columns = [
		{ key: 'stock_out_no', label: '出库单号', class: 'font-mono text-primary', width: '16%' },
		{ key: 'out_type', label: '类型', width: '11%' },
		{ key: 'stock_out_date', label: '出库日期', width: '11%', class: 'text-center' },
		{ key: 'business_doc_no', label: '业务单号', class: 'font-mono text-center', width: '16%' },
		{ key: 'total_amount', label: '总金额', width: '12%', class: 'text-right pr-4' },
		{ key: 'status', label: '状态', width: '10%', class: 'text-center' }
	];

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return '¥' + x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function outTypeName(outType: string) {
		const map: any = {
			purchase_return: '采购退货',
			sales: '销售出库',
			consumption: '领料出库',
			production: '生产领料',
			transfer: '调拨出库',
			other: '其他'
		};
		return map[outType] || outType || '-';
	}

	function getBusinessDocHref(row: any) {
		if (!row?.business_doc_id || !row?.business_doc_type) return '';
		if (row.business_doc_type === 'purchase_return')
			return `/purchase/return/${row.business_doc_id}`;
		if (row.business_doc_type === 'consumption_order')
			return `/consumption/orders/${row.business_doc_id}`;
		return '';
	}

	function confirmStockOut(row: any) {
		if (!row?.id) return;
		window.location.href = `/stock/out/${row.id}?mode=confirm`;
	}

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.stock_out_no.trim()) params.set('stock_out_no', filters.stock_out_no.trim());
			if (filters.out_type) params.set('out_type', filters.out_type);
			if (filters.receiver.trim()) params.set('receiver', filters.receiver.trim());
			if (filters.status) params.set('status', filters.status);
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/stock-out?${params.toString()}`);
			stockOuts = res.list || [];
			total = res.total || 0;
		} catch (err: any) {
			toast.error('加载出库单失败: ' + (err?.message || err));
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
			stock_out_no: '',
			out_type: '',
			receiver: '',
			status: '',
			start_date: '',
			end_date: ''
		};
		currentPage = 1;
		loadData(1);
	}

	onMount(() => loadData(1));
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={stockOuts}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		showDefaultSearch={false}
		actionColumnWidth="220px"
	>
		{#snippet headerFilters()}
			<div
				class="scrollbar-hide flex min-w-0 flex-nowrap items-center gap-2 overflow-x-auto py-0.5"
			>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[9rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="出库单号"
					bind:value={filters.stock_out_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8.5rem] shrink-0 rounded-lg border-none py-0 pr-8 pl-2 text-base leading-tight"
					bind:value={filters.out_type}
				>
					<option value="">出库类型</option>
					<option value="purchase_return">采购退货</option>
					<option value="sales">销售出库</option>
					<option value="production">生产领料</option>
					<option value="transfer">调拨出库</option>
					<option value="other">其他</option>
				</select>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="收货方"
					bind:value={filters.receiver}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[7.5rem] shrink-0 rounded-lg border-none py-0 pr-7 pl-2 text-base leading-tight"
					bind:value={filters.status}
				>
					<option value="">状态</option>
					<option value="pending">待出库</option>
					<option value="confirmed">已完成</option>
				</select>
				<input
					type="date"
					title="出库日期起"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.start_date}
				/>
				<span class="text-base-content/45 shrink-0 text-sm">—</span>
				<input
					type="date"
					title="出库日期止"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.end_date}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleFilterSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}

		{#snippet cellRender(key, value, row)}
			{#if key === 'stock_out_no'}
				<a
					class="link link-primary block min-w-0 truncate font-mono no-underline hover:underline"
					href={`/stock/out/${row.id}`}
				>
					{value || '-'}
				</a>
			{:else if key === 'status'}
				{@const style = getStatusStyle(value, 'stock_out')}
				<span class="badge badge-md whitespace-nowrap {style.class}">{style.label}</span>
			{:else if key === 'out_type'}
				<span class="block min-w-0 truncate" title={outTypeName(value)}>{outTypeName(value)}</span>
			{:else if key === 'business_doc_no'}
				{@const href = getBusinessDocHref(row)}
				{#if href && value}
					<a
						class="link link-primary block min-w-0 truncate font-mono no-underline hover:underline"
						{href}
						title="查看业务单详情"
					>
						{value}
					</a>
				{:else}
					<span class="text-base-content/30">-</span>
				{/if}
			{:else if key === 'total_amount'}
				<span class="text-success font-semibold">{formatMoney(value)}</span>
			{:else if key === 'stock_out_date'}
				<span class="whitespace-nowrap">{formatDate(value)}</span>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}

		{#snippet rowActions(row)}
			<div class="flex flex-wrap items-center justify-center gap-1.5">
				{#if row.status === 'confirmed'}
					<a class={dgRowBtn} href={`/stock/out/${row.id}`}>
						<FileText size={17} /> 详情
					</a>
				{:else}
					<button type="button" class={dgRowBtnSuccess} onclick={() => confirmStockOut(row)}>
						<CheckCircle size={17} /> 确认出库
					</button>
				{/if}
			</div>
		{/snippet}
	</DataGrid>
</div>

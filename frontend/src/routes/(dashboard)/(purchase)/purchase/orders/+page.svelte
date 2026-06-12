<!--
功能：采购订货列表页面
创建时间：2026-05-16
创建人：GPT-5.4
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import CopyableNo from '$lib/components/CopyableNo.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { SquarePen, Trash2, FileText } from 'lucide-svelte';
	import {
		dgRowBtn,
		dgRowBtnDanger,
		dgRowBtnPrimary,
		dgToolbarBtn
	} from '$lib/dgButtonClasses';
	import { goto } from '$app/navigation';
	import { getStatusStyle } from '$lib/statusStyles';
	import { formatDateInCn } from '$lib/datetime';

	let orders = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);
	let filters = $state({
		order_no: '',
		supplier_keyword: '',
		order_status: '',
		start_date: '',
		end_date: ''
	});

	const columns = [
		{ key: 'order_no', label: '系统单号', class: 'font-mono text-primary', width: '14%' },
		{ key: 'supplier_name', label: '供应商', width: '16%' },
		{ key: 'expected_date', label: '预计到货', width: '10%' },
		{ key: 'stock_in_no', label: '入库单号', width: '12%' },
		{
			key: 'total_amount',
			label: '总金额',
			width: '12%',
			class: 'text-left font-mono text-success pr-6'
		},
		{ key: 'order_status_name', label: '状态', width: '8%', class: 'text-left' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			params.set('order_type', 'purchase');
			if (filters.order_no.trim()) params.set('order_no', filters.order_no.trim());
			if (filters.supplier_keyword.trim())
				params.set('supplier_keyword', filters.supplier_keyword.trim());
			if (filters.order_status) params.set('order_status', filters.order_status);
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/purchase/orders?${params.toString()}`);
			orders = res.list || [];
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
		filters = {
			order_no: '',
			supplier_keyword: '',
			order_status: '',
			start_date: '',
			end_date: ''
		};
		loadData(1);
	}

	function navigateToCreate() {
		goto('/purchase/orders/create');
	}

	function navigateToEdit(order: any) {
		goto(`/purchase/orders/${order.id}/edit`);
	}

	async function handleDelete(order: any) {
		deleteTarget = order;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/purchase/orders/${deleteTarget.id}`);
			toast.success('删除成功');
			showConfirm = false;
			loadData(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	function formatAmount(amount: number) {
		return '¥' + (amount || 0).toFixed(2);
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
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
		actionColumnWidth="320px"
	>
		{#snippet headerFilters()}
			<div
				class="scrollbar-hide flex min-w-0 flex-nowrap items-center gap-2 overflow-x-auto py-0.5"
			>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8.5rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="系统单号"
					bind:value={filters.order_no}
					onkeydown={(e) => e.key === 'Enter' && handleSearch()}
				/>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10.5rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="供应商名称/编码"
					bind:value={filters.supplier_keyword}
					onkeydown={(e) => e.key === 'Enter' && handleSearch()}
				/>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[7rem] shrink-0 rounded-lg border-none py-0 pr-7 pl-2 text-base leading-tight"
					bind:value={filters.order_status}
				>
					<option value="">状态</option>
					<option value="draft">待提交</option>
					<option value="ordered">已下单</option>
					<option value="partial_received">部分到货</option>
					<option value="full_received">已完成</option>
				</select>
				<input
					type="date"
					title="下单起"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.start_date}
				/>
				<span class="text-base-content/45 shrink-0 text-sm">—</span>
				<input
					type="date"
					title="下单止"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.end_date}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}

		{#snippet cellRender(key, value, row)}
			{#if key === 'order_no'}
				<CopyableNo value={value} href={`/purchase/orders/${row.id}`} title="查看详情" />
			{:else if key === 'order_status_name'}
				{@const style = getStatusStyle(row.order_status, 'purchase_order')}
				<span class="badge badge-md whitespace-nowrap {style.class}">{style.label}</span>
			{:else if key === 'total_amount'}
				<span class="text-success font-mono font-semibold">{formatAmount(value)}</span>
			{:else if key === 'expected_date'}
				{formatDate(value)}
			{:else if key === 'stock_in_no'}
				{#if row.stock_in_id && value}
					<CopyableNo value={value} href={`/stock/in/${row.stock_in_id}`} title="查看入库单详情" />
				{:else}
					-
				{/if}
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(order)}
			<div class="flex flex-nowrap items-center justify-center gap-1.5">
				<a class={dgRowBtn} href={`/purchase/orders/${order.id}`}>
					<FileText size={16} /> 详情
				</a>
				{#if order.order_status === 'draft' || order.order_status === 'ordered'}
					<button type="button" class={dgRowBtnPrimary} onclick={() => navigateToEdit(order)}>
						<SquarePen size={16} /> 编辑
					</button>
					<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(order)}>
						<Trash2 size={16} /> 删除
					</button>
				{/if}
			</div>
		{/snippet}
	</DataGrid>
</div>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除采购订货"
	message={`确定要删除采购订货「${deleteTarget?.order_no || ''}」吗？删除后无法恢复。`}
	onConfirm={confirmDelete}
/>

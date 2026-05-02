<!--
功能：采购退货页面
创建时间：2026-04-18
创建人：CodeArts Agent
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { Trash2, CheckCircle, FileText, Edit3 } from 'lucide-svelte';
	import {
		dgRowBtn,
		dgRowBtnDanger,
		dgRowBtnPrimary,
		dgRowBtnSuccess,
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

	let suppliers = $state<any[]>([]);

	let filters = $state({
		return_no: '',
		supplier_code: '',
		status: '',
		start_date: '',
		end_date: ''
	});

	const columns = [
		{ key: 'return_no', label: '退货单号', class: 'font-mono text-primary', width: '14%' },
		{ key: 'supplier_name', label: '供应商', width: '14%' },
		{ key: 'return_date', label: '退货日期', width: '11%', class: 'text-center' },
		{ key: 'stock_out_no', label: '出库单号', class: 'font-mono text-center', width: '15%' },
		{
			key: 'total_amount',
			label: '总金额',
			width: '12%',
			class: 'text-right font-mono text-success pr-8'
		},
		{ key: 'status', label: '状态', width: '10%', class: 'text-center' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			params.set('return_type', 'purchase_return');
			if (filters.return_no.trim()) params.set('return_no', filters.return_no.trim());
			if (filters.supplier_code) params.set('supplier_code', filters.supplier_code);
			if (filters.status) params.set('status', filters.status);
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/returns?${params.toString()}`);
			orders = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
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
			supplier_code: '',
			status: '',
			start_date: '',
			end_date: ''
		};
		loadData(1);
	}

	async function loadSuppliers() {
		try {
			const res: any = await api.get('/base/suppliers?page=1&page_size=100');
			suppliers = res.list || [];
		} catch (err) {
			console.error(err);
		}
	}

	function navigateToCreate() {
		goto('/purchase/return/create');
	}

	function navigateToEdit(order: any) {
		goto(`/purchase/return/${order.id}/edit`);
	}

	async function handleDelete(order: any) {
		deleteTarget = order;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/returns/${deleteTarget.id}`);
			toast.success('删除成功');
			showConfirm = false;
			loadData(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	async function handleConfirm(order: any) {
		try {
			await api.post(`/returns/${order.id}/confirm`);
			toast.success('提交成功，已生成出库单');
			loadData(currentPage);
		} catch (err: any) {
			toast.error('确认失败: ' + (err?.message || err));
		}
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatAmount(amount: number) {
		const x = Number(amount) || 0;
		return '¥' + x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
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
		data={orders}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		onCreate={navigateToCreate}
		onRefresh={() => loadData(currentPage)}
		showDefaultSearch={false}
		actionColumnWidth="260px"
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
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[9.5rem] shrink-0 rounded-lg border-none py-0 pr-8 pl-2 text-base leading-tight"
					bind:value={filters.supplier_code}
				>
					<option value="">供应商</option>
					{#each suppliers as s}
						<option value={s.supplier_code}>{s.supplier_name}</option>
					{/each}
				</select>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8rem] shrink-0 rounded-lg border-none py-0 pr-7 pl-2 text-base leading-tight"
					bind:value={filters.status}
				>
					<option value="">状态</option>
					<option value="draft">待提交</option>
					<option value="pending_out">待出库</option>
					<option value="completed">已完成</option>
				</select>
				<input
					type="date"
					title="退货日起"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.start_date}
				/>
				<span class="text-base-content/45 shrink-0 text-sm">—</span>
				<input
					type="date"
					title="退货日止"
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
					href={`/purchase/return/${row.id}`}
					class="link link-primary font-mono no-underline hover:underline"
					title="查看详情"
				>
					{value || '-'}
				</a>
			{:else if key === 'supplier_name'}
				<span class="block truncate" title={value || '-'}>
					{value || '-'}
				</span>
			{:else if key === 'status'}
				{@const style = getStatusStyle(value, 'purchase_return')}
				<span class="badge badge-md whitespace-nowrap {style.class}">{style.label}</span>
			{:else if key === 'return_date'}
				{formatDate(value)}
			{:else if key === 'stock_out_no'}
				{#if row.stock_out_id}
					<a
						href={`/stock/out/${row.stock_out_id}`}
						class="link link-primary font-mono no-underline hover:underline"
						title="查看关联出库单"
					>
						{value || '-'}
					</a>
				{:else}
					<span class="text-base-content/30">-</span>
				{/if}
			{:else if key === 'total_amount'}
				<span class="text-success font-mono font-semibold">{formatAmount(value)}</span>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(order)}
			<div class="flex flex-nowrap items-center justify-center gap-1.5">
				<a class={dgRowBtn} href={`/purchase/return/${order.id}`}>
					<FileText size={16} /> 详情
				</a>
				{#if order.status === 'draft'}
					<button type="button" class={dgRowBtnPrimary} onclick={() => navigateToEdit(order)}>
						<Edit3 size={16} /> 编辑
					</button>
					<button type="button" class={dgRowBtnSuccess} onclick={() => handleConfirm(order)}>
						<CheckCircle size={16} /> 提交
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
	title="删除采购退货"
	message={`确定要删除采购退货单「${deleteTarget?.return_no || ''}」吗？删除后无法恢复。`}
	onConfirm={confirmDelete}
/>

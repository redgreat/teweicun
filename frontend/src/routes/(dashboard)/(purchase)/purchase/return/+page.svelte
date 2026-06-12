<!--
功能：采购退货页面
创建时间：2026-04-18
创建人：CodeArts Agent
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import CopyableNo from '$lib/components/CopyableNo.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { Trash2, FileText, Edit3 } from 'lucide-svelte';
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

	let suppliers = $state<any[]>([]);
	let supplierOptions = $state<any[]>([]);
	let supplierOptionsPage = $state(1);
	let supplierOptionsHasMore = $state(true);
	let supplierOptionsLoading = $state(false);
	let supplierDropdownOpen = $state(false);
	let supplierSearchValue = $state('');
	let supplierSearchTimeout: ReturnType<typeof setTimeout> | null = null;

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
		supplierSearchValue = '';
		loadData(1);
	}

	function normalizeSearchTerm(value: string) {
		return String(value || '').trim();
	}

	async function loadSupplierOptions(reset = false) {
		if (supplierOptionsLoading) return;
		const nextPage = reset ? 1 : supplierOptionsPage + 1;
		supplierOptionsLoading = true;
		try {
			let url = `/base/suppliers?page=${nextPage}&page_size=20`;
			const q = normalizeSearchTerm(supplierSearchValue);
			if (q) {
				const byName: any = await api.get(`${url}&supplier_name=${encodeURIComponent(q)}`);
				const list = byName.list || [];
				const res =
					nextPage > 1 || list.length > 0
						? byName
						: await api.get(`${url}&supplier_code=${encodeURIComponent(q)}`);
				const nextList = res.list || [];
				const totalCount = Number(res.total || 0);
				supplierOptionsPage = nextPage;
				supplierOptions = reset ? nextList : [...supplierOptions, ...nextList];
				supplierOptionsHasMore = supplierOptions.length < totalCount && nextList.length > 0;
				return;
			}
			const res: any = await api.get(url);
			const list = res.list || [];
			const totalCount = Number(res.total || 0);
			supplierOptionsPage = nextPage;
			supplierOptions = reset ? list : [...supplierOptions, ...list];
			supplierOptionsHasMore = supplierOptions.length < totalCount && list.length > 0;
		} catch (err) {
			console.error(err);
		} finally {
			supplierOptionsLoading = false;
		}
	}

	function openSupplierDropdown() {
		supplierDropdownOpen = true;
		supplierOptions = [];
		supplierOptionsPage = 1;
		supplierOptionsHasMore = true;
		loadSupplierOptions(true);
	}

	function closeSupplierDropdown() {
		supplierDropdownOpen = false;
	}

	function onSupplierInput() {
		filters.supplier_code = '';
		if (supplierSearchTimeout) clearTimeout(supplierSearchTimeout);
		supplierSearchTimeout = setTimeout(() => {
			supplierOptions = [];
			supplierOptionsPage = 1;
			supplierOptionsHasMore = true;
			loadSupplierOptions(true);
		}, 250);
	}

	function selectSupplier(supplier: any) {
		filters.supplier_code = supplier.supplier_code || '';
		supplierSearchValue = supplier.supplier_name || supplier.supplier_code || '';
		closeSupplierDropdown();
	}

	function onSupplierOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 12;
		if (nearBottom && supplierOptionsHasMore && !supplierOptionsLoading) {
			loadSupplierOptions(false);
		}
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeSupplierDropdown();
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
	});
</script>

<svelte:window onkeydown={onWindowKeydown} />

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
				<div
					class="relative z-20 w-[10.5rem] min-w-0 shrink-0"
					onclick={(e) => e.stopPropagation()}
					role="presentation"
				>
					<input
						type="text"
						class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-full rounded-lg border-none px-3 text-base"
						placeholder="供应商名称/编码"
						bind:value={supplierSearchValue}
						onfocus={openSupplierDropdown}
						oninput={onSupplierInput}
						onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
					/>
					{#if supplierDropdownOpen}
						<div
							class="fixed inset-0 z-[60]"
							role="presentation"
							onclick={closeSupplierDropdown}
						></div>
						<div
							class="bg-base-100 border-base-300 absolute top-full right-0 left-0 z-[70] mt-2 overflow-hidden rounded-xl border shadow-2xl"
						>
							<div class="max-h-72 overflow-auto" onscroll={onSupplierOptionsScroll}>
								{#if supplierOptions.length === 0 && !supplierOptionsLoading}
									<div class="text-base-content/50 p-4 text-center text-sm">未找到匹配供应商</div>
								{:else}
									{#each supplierOptions as supplier}
										<button
											type="button"
											class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
											onclick={() => selectSupplier(supplier)}
										>
											<div class="text-sm font-medium">{supplier.supplier_name || '-'}</div>
											<div class="text-base-content/60 font-mono text-xs">
												{supplier.supplier_code || '-'}
											</div>
										</button>
									{/each}
								{/if}
								{#if supplierOptionsLoading}
									<div class="text-base-content/50 p-3 text-center text-xs">加载中...</div>
								{:else if supplierOptionsHasMore}
									<div class="text-base-content/50 p-3 text-center text-xs">下拉加载更多...</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>
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
				<CopyableNo value={value} href={`/purchase/return/${row.id}`} title="查看详情" />
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
					<CopyableNo value={value} href={`/stock/out/${row.stock_out_id}`} title="查看关联出库单" />
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
				{#if order.status === 'pending_out'}
					<button type="button" class={dgRowBtnPrimary} onclick={() => navigateToEdit(order)}>
						<Edit3 size={16} /> 编辑
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

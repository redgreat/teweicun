<!--
功能：销售订单列表页
创建时间：2026-05-10
创建人：GPT-5.4
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { FileText, CircleCheckBig, Truck, Ban } from 'lucide-svelte';
	import {
		dgRowBtn,
		dgRowBtnDanger,
		dgRowBtnPrimary,
		dgRowBtnSuccess,
		dgToolbarBtn
	} from '$lib/dgButtonClasses';
	import { getStatusStyle } from '$lib/statusStyles';
	import { formatDateInCn } from '$lib/datetime';

	let orders = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let customerOptions = $state<any[]>([]);
	let customerOptionsPage = $state(1);
	let customerOptionsHasMore = $state(true);
	let customerOptionsLoading = $state(false);
	let customerDropdownOpen = $state(false);
	let customerSearchTimeout: ReturnType<typeof setTimeout> | null = null;

	let filters = $state({
		order_no: '',
		customer_keyword: '',
		status: '',
		start_date: '',
		end_date: ''
	});

	const columns = [
		{ key: 'order_no', label: '销售单号', class: 'font-mono text-primary', width: '15%' },
		{ key: 'customer_name', label: '客户名称', width: '16%' },
		{ key: 'order_date', label: '下单日期', class: 'text-center', width: '12%' },
		{ key: 'delivery_date', label: '交付日期', class: 'text-center', width: '12%' },
		{ key: 'total_amount', label: '订单金额', class: 'text-right pr-4', width: '14%' },
		{ key: 'order_status', label: '状态', class: 'text-center', width: '10%' }
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

	function normalizeSearchTerm(value: string) {
		return String(value || '').trim();
	}

	async function loadCustomerOptions(reset = false) {
		if (customerOptionsLoading) return;
		const nextPage = reset ? 1 : customerOptionsPage + 1;
		customerOptionsLoading = true;
		try {
			let url = `/base/customers?page=${nextPage}&page_size=20`;
			const q = normalizeSearchTerm(filters.customer_keyword);
			if (q) {
				const byName: any = await api.get(`${url}&customer_name=${encodeURIComponent(q)}`);
				const list = byName.list || [];
				const res =
					nextPage > 1 || list.length > 0
						? byName
						: await api.get(`${url}&customer_code=${encodeURIComponent(q)}`);
				const nextList = res.list || [];
				const totalCount = Number(res.total || 0);
				customerOptionsPage = nextPage;
				customerOptions = reset ? nextList : [...customerOptions, ...nextList];
				customerOptionsHasMore = customerOptions.length < totalCount && nextList.length > 0;
				return;
			}
			const res: any = await api.get(url);
			const list = res.list || [];
			const totalCount = Number(res.total || 0);
			customerOptionsPage = nextPage;
			customerOptions = reset ? list : [...customerOptions, ...list];
			customerOptionsHasMore = customerOptions.length < totalCount && list.length > 0;
		} catch (err) {
			console.error(err);
		} finally {
			customerOptionsLoading = false;
		}
	}

	function openCustomerDropdown() {
		customerDropdownOpen = true;
		customerOptions = [];
		customerOptionsPage = 1;
		customerOptionsHasMore = true;
		loadCustomerOptions(true);
	}

	function closeCustomerDropdown() {
		customerDropdownOpen = false;
	}

	function onCustomerInput() {
		if (customerSearchTimeout) clearTimeout(customerSearchTimeout);
		customerSearchTimeout = setTimeout(() => {
			customerOptions = [];
			customerOptionsPage = 1;
			customerOptionsHasMore = true;
			loadCustomerOptions(true);
		}, 250);
	}

	function selectCustomer(customer: any) {
		filters.customer_keyword = customer.customer_name || customer.customer_code || '';
		closeCustomerDropdown();
	}

	function onCustomerOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 12;
		if (nearBottom && customerOptionsHasMore && !customerOptionsLoading) {
			loadCustomerOptions(false);
		}
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeCustomerDropdown();
	}

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.order_no.trim()) params.set('order_no', filters.order_no.trim());
			if (filters.customer_keyword.trim())
				params.set('customer_keyword', filters.customer_keyword.trim());
			if (filters.status) params.set('status', filters.status);
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/sales/orders?${params.toString()}`);
			orders = res.list || [];
			total = res.total || 0;
		} catch (err: any) {
			toast.error('加载销售订单失败: ' + (err?.message || err));
		} finally {
			loading = false;
		}
	}

	function handleFilterSearch() {
		loadData(1);
	}

	function resetFilters() {
		filters = {
			order_no: '',
			customer_keyword: '',
			status: '',
			start_date: '',
			end_date: ''
		};
		loadData(1);
	}

	function navigateToCreate() {
		goto('/sales/orders/create');
	}

	async function handleConfirm(row: any) {
		try {
			await api.post(`/sales/orders/${row.id}/confirm`, {});
			toast.success('销售订单已确认');
			loadData(currentPage);
		} catch (err: any) {
			toast.error('确认失败: ' + (err?.message || err));
		}
	}

	async function handleCancel(row: any) {
		try {
			await api.post(`/sales/orders/${row.id}/cancel`, {});
			toast.success('销售订单已取消');
			loadData(currentPage);
		} catch (err: any) {
			toast.error('取消失败: ' + (err?.message || err));
		}
	}

	async function handleCreateStockOut(row: any) {
		try {
			const res: any = await api.post(`/sales/orders/${row.id}/ship`, {});
			const stockOutID = Number(res?.stock_out_id || 0);
			if (!stockOutID) {
				toast.error('未生成销售出库单');
				return;
			}
			toast.success('已生成销售出库单');
			goto(`/stock/out/${stockOutID}?mode=confirm`);
		} catch (err: any) {
			toast.error('生成销售出库单失败: ' + (err?.message || err));
		}
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
		actionColumnWidth="300px"
	>
		{#snippet headerFilters()}
			<div
				class="scrollbar-hide flex min-w-0 flex-nowrap items-center gap-2 overflow-x-auto py-0.5"
			>
				<input
					type="text"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[9rem] min-w-0 shrink-0 rounded-lg border-none px-3 text-base"
					placeholder="销售单号"
					bind:value={filters.order_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<div
					class="relative z-20 w-[11rem] min-w-0 shrink-0"
					onclick={(e) => e.stopPropagation()}
					role="presentation"
				>
					<input
						type="text"
						class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-full rounded-lg border-none px-3 text-base"
						placeholder="客户名称/编码"
						bind:value={filters.customer_keyword}
						onfocus={openCustomerDropdown}
						oninput={onCustomerInput}
						onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
					/>
					{#if customerDropdownOpen}
						<div
							class="fixed inset-0 z-[60]"
							role="presentation"
							onclick={closeCustomerDropdown}
						></div>
						<div
							class="bg-base-100 border-base-300 absolute top-full right-0 left-0 z-[70] mt-2 overflow-hidden rounded-xl border shadow-2xl"
						>
							<div class="max-h-72 overflow-auto" onscroll={onCustomerOptionsScroll}>
								{#if customerOptions.length === 0 && !customerOptionsLoading}
									<div class="text-base-content/50 p-4 text-center text-sm">未找到匹配客户</div>
								{:else}
									{#each customerOptions as customer}
										<button
											type="button"
											class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
											onclick={() => selectCustomer(customer)}
										>
											<div class="text-sm font-medium">{customer.customer_name || '-'}</div>
											<div class="text-base-content/60 font-mono text-xs">
												{customer.customer_code || '-'}
											</div>
										</button>
									{/each}
								{/if}
								{#if customerOptionsLoading}
									<div class="text-base-content/50 p-3 text-center text-xs">加载中...</div>
								{:else if customerOptionsHasMore}
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
					<option value="confirmed">待出库</option>
					<option value="preparing">出库中</option>
					<option value="shipped">已完成</option>
					<option value="cancelled">已取消</option>
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
			{#if key === 'order_no'}
				<a
					href={`/sales/orders/${row.id}`}
					class="link link-primary font-mono no-underline hover:underline"
				>
					{value || '-'}
				</a>
			{:else if key === 'order_status'}
				{@const style = getStatusStyle(value, 'sales_order')}
				<span class="badge badge-md whitespace-nowrap {style.class}">{style.label}</span>
			{:else if key === 'total_amount'}
				<span class="text-success font-semibold">{formatMoney(value)}</span>
			{:else if key === 'order_date' || key === 'delivery_date'}
				<span class="whitespace-nowrap">{formatDate(value)}</span>
			{:else}
				<span class="block truncate" title={value || '-'}>{value || '-'}</span>
			{/if}
		{/snippet}

		{#snippet rowActions(row)}
			<div class="flex flex-nowrap items-center justify-center gap-1.5">
				<a class={dgRowBtn} href={`/sales/orders/${row.id}`}>
					<FileText size={16} /> 详情
				</a>
				{#if row.order_status === 'draft'}
					<button type="button" class={dgRowBtnSuccess} onclick={() => handleConfirm(row)}>
						<CircleCheckBig size={16} /> 提交
					</button>
				{/if}
				{#if row.order_status === 'confirmed' || row.order_status === 'preparing'}
					<button type="button" class={dgRowBtnPrimary} onclick={() => handleCreateStockOut(row)}>
						<Truck size={16} /> 销售出
					</button>
				{/if}
				{#if row.order_status !== 'shipped' && row.order_status !== 'cancelled'}
					<button type="button" class={dgRowBtnDanger} onclick={() => handleCancel(row)}>
						<Ban size={16} /> 取消
					</button>
				{/if}
			</div>
		{/snippet}
	</DataGrid>
</div>

<!--
功能：入库单管理页面(库管员审核)
创建时间：2026-04-18
创建人：CodeArts Agent
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { FileText, CheckCircle, Building2, Calendar } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnSuccess, dgToolbarBtn } from '$lib/dgButtonClasses';
	import { getStatusStyle } from '$lib/statusStyles';
	import { formatDateInCn } from '$lib/datetime';

	let stockIns = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showModal = $state(false);
	let editingId = $state<number | null>(null);
	let submitting = $state(false);
	let confirmTarget = $state<any>(null);
	let hasStockIn = $state(false);

	let suppliers = $state<any[]>([]);
	let warehouses = $state<any[]>([]);
	let stockInTypes = $state<any[]>([]);
	let filters = $state({
		stock_in_no: '',
		stock_in_type: '',
		supplier_code: '',
		status: '',
		start_date: '',
		end_date: ''
	});
	let form = $state({
		warehouse_code: '',
		stock_in_date: '',
		items: [] as any[]
	});

	const columns = [
		{ key: 'stock_in_no', label: '入库单号', class: 'font-mono text-primary', width: '13%' },
		{ key: 'stock_in_type', label: '入库类型', width: '11%' },
		{ key: 'warehouse_name', label: '仓库', width: '13%' },
		{ key: 'stock_in_date', label: '入库日期', width: '11%', class: 'text-center' },
		{ key: 'business_doc_no', label: '业务单号', class: 'font-mono text-center', width: '15%' },
		{ key: 'total_amount', label: '总金额', width: '11%', class: 'text-right pr-4' },
		{ key: 'stock_in_status', label: '入库状态', width: '9%', class: 'text-left' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.stock_in_no.trim()) params.set('stock_in_no', filters.stock_in_no.trim());
			if (filters.stock_in_type) params.set('stock_in_type', filters.stock_in_type);
			if (filters.supplier_code) params.set('supplier_code', filters.supplier_code);
			if (filters.status) params.set('status', filters.status);
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/stock-in?${params.toString()}`);
			stockIns = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadSuppliers() {
		try {
			const acc: any[] = [];
			let page = 1;
			const ps = 100;
			while (true) {
				const res: any = await api.get(`/base/suppliers?page=${page}&page_size=${ps}`);
				const list = res?.list || [];
				acc.push(...list);
				const total = Number(res?.total ?? 0);
				if (list.length < ps || acc.length >= total) break;
				page++;
			}
			suppliers = acc;
		} catch (err) {
			console.error(err);
			toast.error('加载供应商列表失败，请稍后重试');
		}
	}

	function handleFilterSearch() {
		currentPage = 1;
		loadData(1);
	}

	function resetFilters() {
		filters = {
			stock_in_no: '',
			stock_in_type: '',
			supplier_code: '',
			status: '',
			start_date: '',
			end_date: ''
		};
		currentPage = 1;
		loadData(1);
	}

	async function loadWarehouses() {
		try {
			const res: any = await api.get('/base/warehouses?page=1&page_size=100');
			warehouses = res.list || [];
		} catch (err) {
			console.error(err);
		}
	}

	async function loadStockInTypes() {
		try {
			const res: any = await api.get('/system/dict/stock_in_type/data');
			stockInTypes = Array.isArray(res) ? res : [];
		} catch (err) {
			console.error(err);
			stockInTypes = [];
		}
	}

	async function openEditModal(stockIn: any) {
		if (stockIn?.stock_in_status === 'passed') {
			toast.error('该单据已完成，不能重复确认');
			return;
		}
		if (stockIn?.stock_in_type === 'reversal') return;
		editingId = stockIn.id;

		try {
			const detail: any = await api.get(`/stock-in/${stockIn.id}`);
			hasStockIn = detail.has_stock_in || false;
			form = {
				warehouse_code: detail.warehouse_code || '',
				stock_in_date: formatDateInCn(detail.stock_in_date),
				items: detail.items.map((item: any) => ({
					id: item.id,
					material_id: item.material_id,
					material_code: item.material_code,
					material_name: item.material_name,
					is_code: Boolean(item.is_code),
					purchase_quantity: Number(item.purchase_quantity || 0),
					received_quantity: Number(item.received_quantity || 0),
					arrived_quantity: item.arrived_quantity,
					accepted_quantity: Number(item.pending_quantity || 0),
					pending_quantity: item.pending_quantity,
					unit_cost: item.unit_cost || 0
				}))
			};
			confirmTarget = stockIn;
			showModal = true;
		} catch (err: any) {
			toast.error('加载详情失败: ' + (err?.message || err));
		}
	}

	async function handleSubmit() {
		if (form.items.length === 0) {
			toast.error('请添加入库明细');
			return;
		}
		if (!form.warehouse_code) {
			toast.error('请选择入库仓库');
			return;
		}
		for (const item of form.items) {
			const purchaseQty = Number(item.purchase_quantity || 0);
			const receivedQty = Number(item.received_quantity || 0);
			const currentQty = Number(item.accepted_quantity || 0);
			const maxAllowed = purchaseQty - receivedQty;
			if (currentQty <= 0) {
				toast.error(`${item.material_name} 的本次入库数量必须大于0`);
				return;
			}
			if (currentQty > maxAllowed) {
				toast.error(
					`${item.material_name} 本次入库数量不能超过待入库数量（采购${purchaseQty} - 累计入库${receivedQty} = ${Math.max(maxAllowed, 0)}）`
				);
				return;
			}
			if (item.is_code && !Number.isInteger(currentQty)) {
				toast.error(`${item.material_name} 需要生成编码，本次入库数量必须为整数`);
				return;
			}
		}

		submitting = true;
		try {
			await api.put(`/stock-in/${editingId}`, {
				warehouse_code: form.warehouse_code,
				items: form.items.map((item: any) => ({
					id: item.id,
					material_id: item.material_id,
					arrived_quantity: Number(item.accepted_quantity || 0),
					accepted_quantity: Number(item.accepted_quantity || 0),
					unit_cost: Number(item.unit_cost || 0)
				}))
			});
			if (confirmTarget?.id) {
				await api.post(`/stock-in/${confirmTarget.id}/confirm`);
			}
			toast.success('确认入库成功');
			showModal = false;
			loadData(currentPage);
		} catch (err: any) {
			toast.error(err?.response?.data?.message || err?.message || '确认入库失败');
		} finally {
			submitting = false;
		}
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return '¥' + x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function getTypeName(type: string) {
		const hit = stockInTypes.find((x: any) => x?.dict_value === type);
		if (hit?.dict_label) return hit.dict_label;
		const map: Record<string, string> = {
			purchase: '采购入库',
			return: '销售退货入库',
			sales_return: '销售退货入库',
			reversal: '退料入库',
			production: '生产入库'
		};
		return map[type] || type || '-';
	}

	function getBusinessDocHref(row: any) {
		if (!row?.business_doc_id || !row?.business_doc_type) return '';
		if (row.business_doc_type === 'purchase_order')
			return `/purchase/orders/${row.business_doc_id}`;
		if (row.business_doc_type === 'reversal_order')
			return `/reversal/orders/${row.business_doc_id}`;
		return '';
	}

	onMount(() => {
		loadData(1);
		loadWarehouses();
		loadSuppliers();
		loadStockInTypes();
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={stockIns}
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
					placeholder="入库单号"
					bind:value={filters.stock_in_no}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<select
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8.5rem] shrink-0 rounded-lg border-none py-0 pr-8 pl-2 text-base leading-tight"
					bind:value={filters.stock_in_type}
				>
					<option value="">入库类型</option>
					{#if stockInTypes.length > 0}
						{#each stockInTypes as t}
							<option value={t.dict_value}>{t.dict_label}</option>
						{/each}
					{:else}
						<option value="purchase">采购入库</option>
						<option value="sales_return">销售退货入库</option>
						<option value="reversal">退料入库</option>
						<option value="production">生产入库</option>
					{/if}
				</select>
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
					class="select bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[8.5rem] shrink-0 rounded-lg border-none py-0 pr-7 pl-2 text-base leading-tight"
					bind:value={filters.status}
				>
					<option value="">入库状态</option>
					<option value="preparing">待入库</option>
					<option value="pending">部分入库</option>
					<option value="passed">已完成</option>
				</select>
				<input
					type="date"
					title="入库日期起"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.start_date}
				/>
				<span class="text-base-content/45 shrink-0 text-sm">—</span>
				<input
					type="date"
					title="入库日期止"
					class="input bg-base-200 focus:bg-base-100 h-10 min-h-10 w-[10rem] shrink-0 rounded-lg border-none px-2 text-base"
					bind:value={filters.end_date}
				/>
				<button type="button" class={dgToolbarBtn} onclick={handleFilterSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value, row)}
			{#if key === 'stock_in_no'}
				<a
					href={`/stock/in/${row.id}`}
					class="link link-primary block min-w-0 truncate font-mono no-underline hover:underline"
					title="查看详情"
				>
					{value || '-'}
				</a>
			{:else if key === 'warehouse_name'}
				<span class="block min-w-0 truncate" title={value || ''}>{value || '-'}</span>
			{:else if key === 'stock_in_date'}
				<span class="whitespace-nowrap">{formatDate(value)}</span>
			{:else if key === 'business_doc_no'}
				{@const href = getBusinessDocHref(row)}
				{#if href && value}
					<a
						{href}
						class="link link-primary block min-w-0 truncate font-mono no-underline hover:underline"
						title="查看业务单详情"
					>
						{value}
					</a>
				{:else}
					<span class="text-base-content/30">-</span>
				{/if}
			{:else if key === 'total_amount'}
				<span class="text-success font-semibold">{formatMoney(value)}</span>
			{:else if key === 'stock_in_status'}
				{@const style = getStatusStyle(value, 'stock_in')}
				<span class="badge badge-md whitespace-nowrap {style.class}">
					{style.label}
				</span>
			{:else if key === 'stock_in_type'}
				<span class="whitespace-nowrap">{getTypeName(value)}</span>
			{:else}
				<span class="block min-w-0 truncate" title={value || ''}>{value || '-'}</span>
			{/if}
		{/snippet}
		{#snippet rowActions(stockIn)}
			<div class="flex flex-nowrap items-center justify-center gap-1.5">
				<a class={dgRowBtn} href={`/stock/in/${stockIn.id}`}>
					<FileText size={16} /> 详情
				</a>
				{#if stockIn.stock_in_status === 'preparing' || stockIn.stock_in_status === 'pending'}
					{#if stockIn.stock_in_type === 'reversal'}
						<a class={dgRowBtnSuccess} href={`/stock/in/${stockIn.id}?mode=confirm`}>
							<CheckCircle size={16} /> 确认入库
						</a>
					{:else}
						<button type="button" class={dgRowBtnSuccess} onclick={() => openEditModal(stockIn)}>
							<CheckCircle size={16} /> 确认入库
						</button>
					{/if}
				{/if}
			</div>
		{/snippet}
	</DataGrid>
</div>

<Modal
	bind:show={showModal}
	title="确认入库"
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-7xl"
>
	<div class="max-h-[72vh] space-y-4 overflow-y-auto pr-1">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><Building2 size={14} /> 仓库</span
					></label
				>
				<select
					bind:value={form.warehouse_code}
					class="select select-bordered bg-base-200/50 w-full"
					disabled={hasStockIn}
				>
					<option value="">请选择仓库</option>
					{#each warehouses as w}
						<option value={w.warehouse_code}>{w.warehouse_name}</option>
					{/each}
				</select>
				{#if hasStockIn}
					<div class="text-warning mt-1 text-xs">已部分入库，仓库不可修改</div>
				{/if}
			</div>
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><Calendar size={14} /> 入库日期</span
					></label
				>
				<input
					type="date"
					bind:value={form.stock_in_date}
					class="input input-bordered bg-base-200/50 w-full"
					disabled
				/>
			</div>
		</div>

		<div class="divider">入库明细</div>

		<div class="overflow-x-auto">
			<table class="table-zebra table w-full text-[15px]">
				<thead>
					<tr>
						<th>物料</th>
						<th>采购数量</th>
						<th>累计入库数量</th>
						<th>本次入库数量</th>
					</tr>
				</thead>
				<tbody>
					{#each form.items as item}
						<tr>
							<td>
								<input
									type="text"
									value="{item.material_name} ({item.material_code})"
									class="input input-bordered input-sm bg-base-200/50 w-full max-w-xs"
									disabled
								/>
							</td>
							<td>
								<input
									type="number"
									value={item.purchase_quantity}
									class="input input-bordered input-sm bg-base-300/50 w-20"
									disabled
								/>
							</td>
							<td>
								<input
									type="number"
									value={item.received_quantity}
									class="input input-bordered input-sm bg-base-300/50 w-20"
									disabled
								/>
							</td>
							<td>
								<input
									type="number"
									bind:value={item.accepted_quantity}
									class="input input-bordered input-sm w-20"
									min="1"
									max={Math.max(
										Number(item.purchase_quantity || 0) - Number(item.received_quantity || 0),
										0
									)}
									step="1"
								/>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</Modal>

import os

def write_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        f.write(content.strip() + "\n")
    print(f"Created {path}")

# ======================= PURCHASE (PAYMENT) =======================

purchase_list = """
<!--
功能：采购对账（付款单）列表
创建时间：2026-05-17
创建人：wangcw
-->
<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { onMount } from 'svelte';
	import { FileText } from 'lucide-svelte';
	import { dgRowBtn, dgToolbarBtn } from '$lib/dgButtonClasses';
	import { goto } from '$app/navigation';
	import { formatDateInCn } from '$lib/datetime';

	let items = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let filters = $state({ statement_no: '', supplier_id: '' });

	const columns = [
		{ key: 'statement_no', label: '对账单号', class: 'font-mono text-primary' },
		{ key: 'supplier_name', label: '供应商名称' },
		{ key: 'statement_date', label: '单据日期' },
		{ key: 'payment_amount', label: '付款金额', class: 'text-right font-mono text-success pr-6' },
		{ key: 'settlement_method', label: '结算方式' },
		{ key: 'status', label: '状态' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams();
			params.set('page', String(page));
			params.set('page_size', String(pageSize));
			if (filters.statement_no) params.set('statement_no', filters.statement_no);
			const res: any = await api.get(`/fund/payments?${params.toString()}`);
			items = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function handleSearch() { loadData(1); }
	function resetFilters() { filters = { statement_no: '', supplier_id: '' }; loadData(1); }
	function navigateToCreate() { goto('/reconciliation/purchase/create'); }

	onMount(() => { loadData(1); });
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
		onCreate={navigateToCreate}
		onRefresh={() => loadData(currentPage)}
		showDefaultSearch={false}
		actionColumnWidth="120px"
	>
		{#snippet headerFilters()}
			<div class="flex items-center gap-2">
				<input type="text" class="input bg-base-200 h-10 w-48 rounded-lg" placeholder="单号" bind:value={filters.statement_no} />
				<button type="button" class={dgToolbarBtn} onclick={handleSearch}>查询</button>
				<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value, row)}
			{#if key === 'payment_amount'}
				<span class="text-success font-mono font-semibold">¥{(value||0).toFixed(2)}</span>
			{:else if key === 'status'}
				<span class="badge badge-md {value === 'completed' ? 'badge-success' : 'badge-ghost'}">
					{value === 'completed' ? '已完成' : '草稿'}
				</span>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(row)}
			<div class="flex justify-center gap-1">
				<a class={dgRowBtn} href={`/reconciliation/purchase/${row.id}`}><FileText size={16} /> 详情</a>
			</div>
		{/snippet}
	</DataGrid>
</div>
"""

purchase_create = """
<!--
功能：新建采购付款单
创建时间：2026-05-17
创建人：wangcw
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { toast } from '$lib/store/toast';
	import api from '$lib/api/client';
	import { Plus, Trash2, Save, ChevronLeft } from 'lucide-svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { onMount } from 'svelte';

	let form = $state({
		supplier_id: 0,
		statement_date: new Date().toISOString().split('T')[0],
		payment_amount: 0,
		settlement_method: 'bank_transfer',
		settlement_account: '',
		settlement_no: '',
		remark: '',
		items: [] as any[]
	});

	let submitting = $state(false);
	let supplierOptions = $state<any[]>([]);
	let sourceOrders = $state<any[]>([]);
	let showOrderModal = $state(false);

	async function loadSuppliers() {
		const res: any = await api.get('/base/suppliers?page_size=100');
		supplierOptions = res.list || [];
	}

	async function loadSourceOrders() {
		if (!form.supplier_id) return;
		const res: any = await api.get(`/purchase/orders?supplier_id=${form.supplier_id}&page_size=100`);
		// Filter orders that have unverified amounts
		sourceOrders = (res.list || []).filter((o: any) => (o.total_amount - (o.verified_amount || 0)) > 0);
	}

	function handleSupplierChange() {
		form.items = [];
		loadSourceOrders();
	}

	function selectOrder(order: any) {
		const unverified = order.total_amount - (order.verified_amount || 0);
		form.items.push({
			source_order_id: order.id,
			source_order_no: order.order_no,
			business_type: '采购入库',
			order_date: order.order_date,
			order_amount: order.total_amount,
			verified_amount: order.verified_amount || 0,
			unverified_amount: unverified,
			current_verify_amount: unverified,
			custom_tax_amount: 0,
			remark: ''
		});
		showOrderModal = false;
		calculateTotal();
	}

	function removeItem(index: number) {
		form.items.splice(index, 1);
		calculateTotal();
	}

	function calculateTotal() {
		form.payment_amount = form.items.reduce((sum, item) => sum + Number(item.current_verify_amount || 0), 0);
	}

	async function submit() {
		if (!form.supplier_id) return toast.warning('请选择供应商');
		if (form.items.length === 0) return toast.warning('请添加对账单据');
		
		submitting = true;
		try {
			await api.post('/fund/payments', form);
			toast.success('保存成功');
			goto('/reconciliation/purchase');
		} catch (err: any) {
			toast.error('保存失败: ' + err.message);
		} finally {
			submitting = false;
		}
	}

	onMount(() => {
		loadSuppliers();
	});
</script>

<div class="flex h-full flex-col bg-base-100">
	<div class="flex items-center justify-between border-b border-base-300 p-4">
		<div class="flex items-center gap-4">
			<button class="btn btn-ghost btn-sm" onclick={() => goto('/reconciliation/purchase')}><ChevronLeft size={20} /> 返回</button>
			<h1 class="text-xl font-bold">新建采购付款单</h1>
		</div>
		<button class="btn btn-primary btn-sm" onclick={submit} disabled={submitting}><Save size={16} /> 保存并确认</button>
	</div>

	<div class="flex-1 overflow-y-auto p-4 space-y-6">
		<div class="grid grid-cols-4 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text text-error">* 供应商</span></label>
				<select class="select select-bordered" bind:value={form.supplier_id} onchange={handleSupplierChange}>
					<option value={0}>请选择</option>
					{#each supplierOptions as s}
						<option value={s.id}>{s.supplier_name}</option>
					{/each}
				</select>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">* 单据日期</span></label>
				<input type="date" class="input input-bordered" bind:value={form.statement_date} />
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">结算方式</span></label>
				<select class="select select-bordered" bind:value={form.settlement_method}>
					<option value="bank_transfer">银行转账</option>
					<option value="cash">现金</option>
					<option value="wechat">微信</option>
					<option value="alipay">支付宝</option>
				</select>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">付款金额</span></label>
				<input type="number" class="input input-bordered font-mono text-success" bind:value={form.payment_amount} readonly />
			</div>
		</div>

		<div>
			<div class="flex items-center justify-between mb-2">
				<h3 class="font-bold">源单据明细</h3>
				<button class="btn btn-outline btn-sm" disabled={!form.supplier_id} onclick={() => showOrderModal = true}>
					<Plus size={16} /> 添加源单据
				</button>
			</div>
			<div class="overflow-x-auto border border-base-300 rounded-lg">
				<table class="table table-sm">
					<thead class="bg-base-200">
						<tr>
							<th>源单编号</th>
							<th>业务类型</th>
							<th>单据日期</th>
							<th>单据金额</th>
							<th>未核销金额</th>
							<th>本次核销金额</th>
							<th>自定义税额</th>
							<th>操作</th>
						</tr>
					</thead>
					<tbody>
						{#each form.items as item, i}
							<tr>
								<td>{item.source_order_no}</td>
								<td>{item.business_type}</td>
								<td>{item.order_date}</td>
								<td>{item.order_amount}</td>
								<td>{item.unverified_amount}</td>
								<td><input type="number" class="input input-bordered input-sm w-24" bind:value={item.current_verify_amount} oninput={calculateTotal} /></td>
								<td><input type="number" class="input input-bordered input-sm w-24" bind:value={item.custom_tax_amount} /></td>
								<td>
									<button class="btn btn-ghost btn-xs text-error" onclick={() => removeItem(i)}><Trash2 size={14} /></button>
								</td>
							</tr>
						{/each}
						{#if form.items.length === 0}
							<tr><td colspan="8" class="text-center py-4 text-base-content/50">暂无单据</td></tr>
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>

<Modal bind:show={showOrderModal} title="选择源单据" maxWidth="max-w-3xl">
	<div class="overflow-y-auto max-h-[60vh]">
		<table class="table table-sm">
			<thead><tr><th>单号</th><th>日期</th><th>总金额</th><th>已核销</th><th>未核销</th><th>操作</th></tr></thead>
			<tbody>
				{#each sourceOrders as order}
					<tr>
						<td>{order.order_no}</td>
						<td>{order.order_date}</td>
						<td>{order.total_amount}</td>
						<td>{order.verified_amount || 0}</td>
						<td class="text-error">{order.total_amount - (order.verified_amount || 0)}</td>
						<td><button class="btn btn-primary btn-xs" onclick={() => selectOrder(order)}>选择</button></td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</Modal>
"""

purchase_detail = """
<!--
功能：采购付款单详情
创建时间：2026-05-17
创建人：wangcw
-->
<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import api from '$lib/api/client';
	import { ChevronLeft } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	let orderId = $state(page.params.id);
	let order = $state<any>(null);

	async function load() {
		const res = await api.get(`/fund/payments/${orderId}`);
		order = res;
	}

	onMount(() => { load(); });
</script>

<div class="flex h-full flex-col bg-base-100">
	<div class="flex items-center gap-4 border-b border-base-300 p-4">
		<button class="btn btn-ghost btn-sm" onclick={() => goto('/reconciliation/purchase')}><ChevronLeft size={20} /> 返回</button>
		<h1 class="text-xl font-bold">付款单详情：{order?.statement_no || '-'}</h1>
	</div>

	{#if order}
		<div class="p-6 space-y-6">
			<div class="grid grid-cols-4 gap-4">
				<div><span class="text-base-content/60 text-sm">供应商：</span><br/><b>{order.supplier_name}</b></div>
				<div><span class="text-base-content/60 text-sm">单据日期：</span><br/><b>{order.statement_date}</b></div>
				<div><span class="text-base-content/60 text-sm">付款金额：</span><br/><b class="text-success font-mono">¥{order.payment_amount}</b></div>
				<div><span class="text-base-content/60 text-sm">状态：</span><br/><b>{order.status === 'completed' ? '已完成' : '草稿'}</b></div>
			</div>

			<div>
				<h3 class="font-bold mb-2 border-b pb-2">源单据明细</h3>
				<table class="table table-sm border border-base-300">
					<thead class="bg-base-200">
						<tr><th>源单编号</th><th>业务类型</th><th>单据日期</th><th>单据金额</th><th>本次核销金额</th><th>自定义税额</th></tr>
					</thead>
					<tbody>
						{#each order.items || [] as item}
							<tr>
								<td>{item.source_order_no}</td>
								<td>{item.business_type}</td>
								<td>{item.order_date}</td>
								<td>{item.order_amount}</td>
								<td class="text-success">{item.current_verify_amount}</td>
								<td>{item.custom_tax_amount}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
"""

# ======================= SALES (COLLECTION) =======================
# (Very similar to purchase, substituting keywords)
sales_list = purchase_list.replace("purchase", "sales").replace("采购对账", "销售对账").replace("付款单", "收款单").replace("supplier", "customer").replace("供应商", "客户").replace("payment", "collection").replace("付款", "收款")
sales_create = purchase_create.replace("purchase", "sales").replace("采购对账", "销售对账").replace("付款单", "收款单").replace("supplier", "customer").replace("供应商", "客户").replace("payment", "collection").replace("付款", "收款").replace("采购入库", "销售出库")
sales_detail = purchase_detail.replace("purchase", "sales").replace("采购对账", "销售对账").replace("付款单", "收款单").replace("supplier", "customer").replace("供应商", "客户").replace("payment", "collection").replace("付款", "收款")

write_file('frontend/src/routes/(dashboard)/(reconciliation)/reconciliation/purchase/+page.svelte', purchase_list)
write_file('frontend/src/routes/(dashboard)/(reconciliation)/reconciliation/purchase/create/+page.svelte', purchase_create)
write_file('frontend/src/routes/(dashboard)/(reconciliation)/reconciliation/purchase/[id]/+page.svelte', purchase_detail)

write_file('frontend/src/routes/(dashboard)/(reconciliation)/reconciliation/sales/+page.svelte', sales_list)
write_file('frontend/src/routes/(dashboard)/(reconciliation)/reconciliation/sales/create/+page.svelte', sales_create)
write_file('frontend/src/routes/(dashboard)/(reconciliation)/reconciliation/sales/[id]/+page.svelte', sales_detail)

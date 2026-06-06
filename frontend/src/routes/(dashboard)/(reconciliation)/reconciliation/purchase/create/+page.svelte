<!--
功能：新建采购付款单
创建时间：2026-05-17
创建人：wangcw
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { toast } from '$lib/store/toast';
	import api from '$lib/api/client';
	import { Plus, Trash2, Save, ChevronLeft, RefreshCcw } from 'lucide-svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { onMount } from 'svelte';

	let form = $state({
		supplier_id: 0,
		statement_date: new Date().toISOString().split('T')[0],
		bill_type: 'cash',
		payment_amount: 0,
		invoice_amount: 0,
		actual_amount: 0,
		discount_amount: 0,
		advance_amount: 0,
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
	let sourceKeyword = $state('');
	let actualTouched = $state(false);

	const differenceAmount = $derived(Number(form.payment_amount || 0) - Number(form.actual_amount || 0));

	async function loadSuppliers() {
		const res: any = await api.get('/base/suppliers?page_size=100');
		supplierOptions = res.list || [];
	}

	async function loadSourceOrders() {
		if (!form.supplier_id) {
			sourceOrders = [];
			return;
		}
		const params = new URLSearchParams();
		params.set('supplier_id', String(form.supplier_id));
		if (sourceKeyword) params.set('keyword', sourceKeyword);
		const res: any = await api.get(`/fund/payment-sources?${params.toString()}`);
		sourceOrders = Array.isArray(res) ? res : res.list || [];
	}

	function handleSupplierChange() {
		form.items = [];
		sourceKeyword = '';
		actualTouched = false;
		calculateTotal();
		loadSourceOrders();
	}

	function selectOrder(order: any) {
		if (
			form.items.some(
				(item) =>
					item.source_doc_type === order.source_doc_type &&
					item.source_order_id === order.source_order_id
			)
		) {
			toast.warning('该源单据已添加');
			return;
		}

		form.items.push({
			source_doc_type: order.source_doc_type,
			source_order_id: order.source_order_id,
			source_order_no: order.source_order_no,
			business_type: order.business_type,
			order_date: order.order_date,
			order_amount: Number(order.order_amount || 0),
			verified_amount: Number(order.verified_amount || 0),
			unverified_amount: Number(order.unverified_amount || 0),
			current_verify_amount: Number(order.unverified_amount || 0),
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
		const total = form.items.reduce(
			(sum, item) => sum + Number(item.current_verify_amount || 0),
			0
		);
		form.payment_amount = Number(total.toFixed(2));
		if (!actualTouched) form.actual_amount = form.payment_amount;
		if (form.bill_type === 'invoice' && !form.invoice_amount) form.invoice_amount = form.payment_amount;
	}

	function handleBillTypeChange() {
		if (form.bill_type === 'offset') {
			form.actual_amount = 0;
			form.invoice_amount = 0;
			form.settlement_method = 'offset';
			actualTouched = true;
			return;
		}
		actualTouched = false;
		form.actual_amount = form.payment_amount;
		if (form.bill_type === 'invoice') form.invoice_amount = form.payment_amount;
	}

	function formatMoney(value: number) {
		return `¥${Number(value || 0).toFixed(2)}`;
	}

	async function submit() {
		if (!form.supplier_id) return toast.warning('请选择供应商');
		if (form.items.length === 0) return toast.warning('请添加对账单据');
		if (form.items.some((item) => Number(item.current_verify_amount || 0) === 0)) {
			return toast.warning('本次核销金额不能为 0');
		}

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

<div class="bg-base-100 flex h-full flex-col">
	<div class="border-base-300 flex items-center justify-between border-b p-4">
		<div class="flex items-center gap-4">
			<button class="btn btn-ghost btn-sm" onclick={() => goto('/reconciliation/purchase')}
				><ChevronLeft size={20} /> 返回</button
			>
			<h1 class="text-xl font-bold">新建采购付款单</h1>
		</div>
		<button class="btn btn-primary btn-sm" onclick={submit} disabled={submitting}
			><Save size={16} /> 保存并确认</button
		>
	</div>

	<div class="flex-1 space-y-6 overflow-y-auto p-4">
		<div class="grid grid-cols-4 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text text-error">* 供应商</span></label>
				<select
					class="select select-bordered"
					bind:value={form.supplier_id}
					onchange={handleSupplierChange}
				>
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
				<label class="label"><span class="label-text">票款类型</span></label>
				<select class="select select-bordered" bind:value={form.bill_type} onchange={handleBillTypeChange}>
					<option value="cash">款项</option>
					<option value="invoice">票据</option>
					<option value="offset">抵充</option>
				</select>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">结算方式</span></label>
				<select class="select select-bordered" bind:value={form.settlement_method}>
					<option value="bank_transfer">银行转账</option>
					<option value="cash">现金</option>
					<option value="bill">承兑/票据</option>
					<option value="wechat">微信</option>
					<option value="alipay">支付宝</option>
					<option value="offset">抵充</option>
				</select>
			</div>
		</div>

		<div class="grid grid-cols-5 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text">应付/抵充金额</span></label>
				<input class="input input-bordered font-mono" value={formatMoney(form.payment_amount)} readonly />
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">票据金额</span></label>
				<input type="number" class="input input-bordered" bind:value={form.invoice_amount} />
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">实际付款金额</span></label>
				<input
					type="number"
					class="input input-bordered"
					bind:value={form.actual_amount}
					oninput={() => (actualTouched = true)}
				/>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">整单折扣</span></label>
				<input type="number" class="input input-bordered" bind:value={form.discount_amount} />
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">差额</span></label>
				<input
					class="input input-bordered font-mono {Math.abs(differenceAmount) > 0.005
						? 'text-warning'
						: 'text-success'}"
					value={formatMoney(differenceAmount)}
					readonly
				/>
			</div>
		</div>

		<div class="grid grid-cols-3 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text">付款账户</span></label>
				<input class="input input-bordered" bind:value={form.settlement_account} />
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">结算号</span></label>
				<input class="input input-bordered" bind:value={form.settlement_no} />
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">备注</span></label>
				<input class="input input-bordered" bind:value={form.remark} />
			</div>
		</div>

		<div>
			<div class="mb-2 flex items-center justify-between">
				<h3 class="font-bold">源单据明细</h3>
				<button
					class="btn btn-outline btn-sm"
					disabled={!form.supplier_id}
					onclick={() => {
						loadSourceOrders();
						showOrderModal = true;
					}}
				>
					<Plus size={16} /> 添加源单据
				</button>
			</div>
			<div class="border-base-300 overflow-x-auto rounded-lg border">
				<table class="table-sm table">
					<thead class="bg-base-200">
						<tr>
							<th>源单编号</th>
							<th>业务类型</th>
							<th>单据日期</th>
							<th class="text-right">单据金额</th>
							<th class="text-right">已核销</th>
							<th class="text-right">未核销</th>
							<th>本次核销金额</th>
							<th>税额/费用</th>
							<th>操作</th>
						</tr>
					</thead>
					<tbody>
						{#each form.items as item, i}
							<tr>
								<td class="font-mono">{item.source_order_no}</td>
								<td>{item.business_type}</td>
								<td>{item.order_date}</td>
								<td class="text-right font-mono">{formatMoney(item.order_amount)}</td>
								<td class="text-right font-mono">{formatMoney(item.verified_amount)}</td>
								<td class="text-right font-mono">{formatMoney(item.unverified_amount)}</td>
								<td>
									<input
										type="number"
										class="input input-bordered input-sm w-28"
										bind:value={item.current_verify_amount}
										oninput={calculateTotal}
									/>
								</td>
								<td>
									<input
										type="number"
										class="input input-bordered input-sm w-24"
										bind:value={item.custom_tax_amount}
									/>
								</td>
								<td>
									<button class="btn btn-ghost btn-xs text-error" onclick={() => removeItem(i)}
										><Trash2 size={14} /></button
									>
								</td>
							</tr>
						{/each}
						{#if form.items.length === 0}
							<tr><td colspan="9" class="text-base-content/50 py-4 text-center">暂无单据</td></tr>
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>

<Modal bind:show={showOrderModal} title="选择源单据" maxWidth="max-w-5xl">
	<div class="mb-3 flex items-center gap-2">
		<input
			class="input input-bordered h-9 w-72"
			placeholder="单号/业务类型"
			bind:value={sourceKeyword}
		/>
		<button class="btn btn-sm" onclick={loadSourceOrders}><RefreshCcw size={14} /> 刷新</button>
	</div>
	<div class="max-h-[60vh] overflow-y-auto">
		<table class="table-sm table">
			<thead>
				<tr>
					<th>单号</th>
					<th>业务类型</th>
					<th>日期</th>
					<th class="text-right">单据金额</th>
					<th class="text-right">已核销</th>
					<th class="text-right">未核销</th>
					<th>状态</th>
					<th>操作</th>
				</tr>
			</thead>
			<tbody>
				{#each sourceOrders as order}
					<tr>
						<td class="font-mono">{order.source_order_no}</td>
						<td>{order.business_type}</td>
						<td>{order.order_date}</td>
						<td class="text-right font-mono">{formatMoney(order.order_amount)}</td>
						<td class="text-right font-mono">{formatMoney(order.verified_amount)}</td>
						<td class="text-right font-mono">{formatMoney(order.unverified_amount)}</td>
						<td>
							<span class="badge badge-sm {order.unverified_amount < 0 ? 'badge-warning' : 'badge-info'}">
								{order.unverified_amount < 0 ? '抵充' : '待付'}
							</span>
						</td>
						<td>
							<button class="btn btn-primary btn-xs" onclick={() => selectOrder(order)}>选择</button>
						</td>
					</tr>
				{/each}
				{#if sourceOrders.length === 0}
					<tr><td colspan="8" class="text-base-content/50 py-4 text-center">暂无可核销单据</td></tr>
				{/if}
			</tbody>
		</table>
	</div>
</Modal>

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
		const res: any = await api.get(
			`/purchase/orders?supplier_id=${form.supplier_id}&page_size=100`
		);
		// Filter orders that have unverified amounts
		sourceOrders = (res.list || []).filter(
			(o: any) => o.total_amount - (o.verified_amount || 0) > 0
		);
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
		form.payment_amount = form.items.reduce(
			(sum, item) => sum + Number(item.current_verify_amount || 0),
			0
		);
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
				<input
					type="number"
					class="input input-bordered text-success font-mono"
					bind:value={form.payment_amount}
					readonly
				/>
			</div>
		</div>

		<div>
			<div class="mb-2 flex items-center justify-between">
				<h3 class="font-bold">源单据明细</h3>
				<button
					class="btn btn-outline btn-sm"
					disabled={!form.supplier_id}
					onclick={() => (showOrderModal = true)}
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
								<td
									><input
										type="number"
										class="input input-bordered input-sm w-24"
										bind:value={item.current_verify_amount}
										oninput={calculateTotal}
									/></td
								>
								<td
									><input
										type="number"
										class="input input-bordered input-sm w-24"
										bind:value={item.custom_tax_amount}
									/></td
								>
								<td>
									<button class="btn btn-ghost btn-xs text-error" onclick={() => removeItem(i)}
										><Trash2 size={14} /></button
									>
								</td>
							</tr>
						{/each}
						{#if form.items.length === 0}
							<tr><td colspan="8" class="text-base-content/50 py-4 text-center">暂无单据</td></tr>
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>

<Modal bind:show={showOrderModal} title="选择源单据" maxWidth="max-w-3xl">
	<div class="max-h-[60vh] overflow-y-auto">
		<table class="table-sm table">
			<thead
				><tr
					><th>单号</th><th>日期</th><th>总金额</th><th>已核销</th><th>未核销</th><th>操作</th></tr
				></thead
			>
			<tbody>
				{#each sourceOrders as order}
					<tr>
						<td>{order.order_no}</td>
						<td>{order.order_date}</td>
						<td>{order.total_amount}</td>
						<td>{order.verified_amount || 0}</td>
						<td class="text-error">{order.total_amount - (order.verified_amount || 0)}</td>
						<td
							><button class="btn btn-primary btn-xs" onclick={() => selectOrder(order)}
								>选择</button
							></td
						>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</Modal>

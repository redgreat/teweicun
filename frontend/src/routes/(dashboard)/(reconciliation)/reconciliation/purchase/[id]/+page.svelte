<!--
功能：采购付款单详情
创建时间：2026-05-17
创建人：wangcw
-->
<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import api from '$lib/api/client';
	import { ChevronLeft, Printer } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	let orderId = $state(page.params.id);
	let order = $state<any>(null);

	async function load() {
		const res = await api.get(`/fund/payments/${orderId}`);
		order = res;
	}

	onMount(() => {
		load();
	});

	function billTypeLabel(value: string) {
		return value === 'invoice' ? '票据' : value === 'offset' ? '抵充' : '款项';
	}
</script>

<div class="bg-base-100 flex h-full flex-col">
	<div class="border-base-300 flex items-center gap-4 border-b p-4 print:hidden">
		<button class="btn btn-ghost btn-sm" onclick={() => goto('/reconciliation/purchase')}>
			<ChevronLeft size={20} /> 返回
		</button>
		<h1 class="text-xl font-bold flex-1">付款单详情：{order?.statement_no || '-'}</h1>
		<button class="btn btn-outline btn-sm" onclick={() => window.print()} disabled={!order}>
			<Printer size={18} /> 打印
		</button>
	</div>

	{#if order}
		<div class="space-y-6 p-6">
			<div class="grid grid-cols-4 gap-4">
				<div>
					<span class="text-base-content/60 text-sm">供应商：</span><br /><b
						>{order.supplier_name}</b
					>
				</div>
				<div>
					<span class="text-base-content/60 text-sm">单据日期：</span><br /><b
						>{order.statement_date}</b
					>
				</div>
				<div>
					<span class="text-base-content/60 text-sm">应付/抵充：</span><br /><b
						class="text-success font-mono">¥{order.payment_amount}</b
					>
				</div>
				<div>
					<span class="text-base-content/60 text-sm">票款类型：</span><br /><b
						>{billTypeLabel(order.bill_type)}</b
					>
				</div>
				<div>
					<span class="text-base-content/60 text-sm">票据金额：</span><br /><b
						class="font-mono">¥{order.invoice_amount}</b
					>
				</div>
				<div>
					<span class="text-base-content/60 text-sm">实际付款：</span><br /><b
						class="text-success font-mono">¥{order.actual_amount}</b
					>
				</div>
				<div>
					<span class="text-base-content/60 text-sm">差额：</span><br /><b
						class="font-mono">¥{order.difference_amount}</b
					>
				</div>
				<div>
					<span class="text-base-content/60 text-sm">状态：</span><br /><b
						>{order.status === 'completed' ? '已完成' : '草稿'}</b
					>
				</div>
			</div>

			<div>
				<h3 class="mb-2 border-b pb-2 font-bold">源单据明细</h3>
				<table class="table-sm border-base-300 table border">
					<thead class="bg-base-200">
						<tr
							><th>源单编号</th><th>来源类型</th><th>业务类型</th><th>单据日期</th><th>单据金额</th><th
								>本次核销金额</th
							><th>自定义税额</th></tr
						>
					</thead>
					<tbody>
						{#each order.items || [] as item}
							<tr>
								<td>{item.source_order_no}</td>
								<td>{item.source_doc_type}</td>
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

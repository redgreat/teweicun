<!--
功能：新增生产退货单
创建时间：2026-07-12
创建人：Hermes Agent
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { ArrowLeft, Hash, FileText } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	let productionOrders = $state<any[]>([]);
	let selectedProd = $state<any>({});

	let form = $state({
		production_order_id: 0,
		returned_quantity: 1,
		remark: ''
	});

	let submitting = $state(false);

	onMount(async () => {
		try {
			const res: any = await api.get('/production/orders?page=1&page_size=100');
			productionOrders = res.list || [];
		} catch (e) { toast.error('加载生产单失败'); }
	});

	function onProdChange() {
		const p = productionOrders.find((po: any) => po.id === Number(form.production_order_id));
		selectedProd = p || {};
		if (p && form.returned_quantity > p.produced_quantity) {
			form.returned_quantity = p.produced_quantity;
		}
	}

	async function handleSubmit() {
		if (!form.production_order_id) { toast.error('请选择生产单'); return; }
		if (form.returned_quantity <= 0) { toast.error('退回数量必须大于0'); return; }
		if (form.returned_quantity > (selectedProd.produced_quantity || 0)) {
			toast.error('退回数量不能超过生产数量'); return;
		}
		submitting = true;
		try {
			await api.post('/production/returns', form);
			toast.success('生产退货单创建成功');
			goto('/production/returns');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally { submitting = false; }
	}
</script>

<div class="space-y-6">
	<div class="flex items-center gap-3">
		<button type="button" class="btn btn-ghost btn-circle" onclick={() => goto('/production/returns')}>
			<ArrowLeft size={20} />
		</button>
		<div class="h-8 w-1.5 rounded-full bg-orange-500"></div>
		<h1 class="text-2xl font-bold tracking-tight">新增生产退货单</h1>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							生产单 <span class="text-error">*</span>
						</span>
					</label>
					<select bind:value={form.production_order_id} onchange={onProdChange}
						class="select select-bordered h-11 w-full">
						<option value={0}>选择生产单</option>
						{#each productionOrders as po}
							<option value={po.id}>
								{po.production_no} - {po.produced_material_name} (x{po.produced_quantity})
							</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<Hash size={14} /> 退回数量 <span class="text-error">*</span>
						</span>
					</label>
					<input type="number" bind:value={form.returned_quantity} min="1"
						max={selectedProd.produced_quantity || 9999} step="1"
						class="input input-bordered h-11 w-full" />
					{#if selectedProd.produced_quantity}
						<label class="label"><span class="label-text-alt">生产数量: {selectedProd.produced_quantity}，可退回最多 {selectedProd.produced_quantity} 件</span></label>
					{/if}
				</div>

				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<FileText size={14} /> 备注
						</span>
					</label>
					<input type="text" bind:value={form.remark}
						class="input input-bordered h-11 w-full" placeholder="备注信息" />
				</div>
			</div>

			{#if selectedProd.production_no}
				<div class="alert alert-info text-sm">
					<span>关联生产单: {selectedProd.production_no} | 成品: {selectedProd.produced_material_name} | 
					数量: {selectedProd.produced_quantity} | 成本: ¥{Number(selectedProd.cost_price || 0).toFixed(2)}</span>
				</div>
			{/if}

			<div class="border-base-300 flex flex-wrap items-center justify-end gap-4 border-t pt-6">
				<button type="button" class="btn" onclick={() => goto('/production/returns')} disabled={submitting}>取消</button>
				<button type="button" class="btn btn-primary" onclick={handleSubmit} disabled={submitting}>
					{#if submitting}<span class="loading loading-spinner loading-sm"></span>{/if}
					提交退货单
				</button>
			</div>
		</div>
	</div>
</div>

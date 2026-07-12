<!--
功能：手动新增生产单
创建时间：2026-07-12
创建人：Hermes Agent
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { ArrowLeft, Package, Warehouse, Hash, DollarSign, FileText } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	let materials = $state<any[]>([]);
	let warehouses = $state<any[]>([]);
	let consumptionOrders = $state<any[]>([]);

	let form = $state({
		produced_material_id: 0,
		produced_warehouse_id: 0,
		produced_quantity: 1,
		cost_price: 0,
		consumption_order_id: 0,
		remark: ''
	});

	let submitting = $state(false);

	onMount(async () => {
		try {
			const [matRes, whRes, coRes] = await Promise.all([
				api.get('/base/materials?page=1&page_size=200'),
				api.get('/base/warehouses?page=1&page_size=50'),
				api.get('/consumption/orders?page=1&page_size=100')
			]);
			materials = (matRes as any).list || [];
			warehouses = (whRes as any).list || [];
			consumptionOrders = (coRes as any).list || [];
		} catch (e) {
			toast.error('加载基础数据失败');
		}
	});

	async function handleSubmit() {
		if (!form.produced_material_id) { toast.error('请选择成品物料'); return; }
		if (!form.produced_warehouse_id) { toast.error('请选择入库仓库'); return; }
		if (form.produced_quantity <= 0) { toast.error('数量必须大于0'); return; }

		submitting = true;
		try {
			await api.post('/production/orders', form);
			toast.success('生产单创建成功');
			goto('/production/orders');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center gap-3">
		<button type="button" class="btn btn-ghost btn-circle" onclick={() => goto('/production/orders')}>
			<ArrowLeft size={20} />
		</button>
		<div class="h-8 w-1.5 rounded-full bg-purple-500"></div>
		<h1 class="text-2xl font-bold tracking-tight">新增生产单</h1>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<Package size={14} /> 成品物料 <span class="text-error">*</span>
						</span>
					</label>
					<select bind:value={form.produced_material_id} class="select select-bordered h-11 w-full">
						<option value={0}>选择成品物料</option>
						{#each materials as m}
							<option value={m.id}>{m.material_display_name || m.material_name} [{m.material_code}]</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<Warehouse size={14} /> 入库仓库 <span class="text-error">*</span>
						</span>
					</label>
					<select bind:value={form.produced_warehouse_id} class="select select-bordered h-11 w-full">
						<option value={0}>选择入库仓库</option>
						{#each warehouses as wh}
							<option value={wh.id}>{wh.warehouse_name}</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<Hash size={14} /> 生产数量 <span class="text-error">*</span>
						</span>
					</label>
					<input type="number" bind:value={form.produced_quantity} min="1" step="1"
						class="input input-bordered h-11 w-full" />
				</div>

				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<DollarSign size={14} /> 成本价格（元）
						</span>
					</label>
					<input type="number" bind:value={form.cost_price} min="0" step="0.01"
						class="input input-bordered h-11 w-full" placeholder="总成本" />
				</div>

				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<FileText size={14} /> 关联领料单
						</span>
					</label>
					<select bind:value={form.consumption_order_id} class="select select-bordered h-11 w-full">
						<option value={0}>不关联（手动生产单）</option>
						{#each consumptionOrders as co}
							<option value={co.id}>{co.order_no} - {co.product_name || co.project_no}</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label">
						<span class="label-text font-medium">备注</span>
					</label>
					<input type="text" bind:value={form.remark}
						class="input input-bordered h-11 w-full" placeholder="备注信息" />
				</div>
			</div>

			{#if form.produced_quantity > 0 && form.cost_price > 0}
				<div class="alert alert-info">
					<span>单位成本：¥{(form.cost_price / form.produced_quantity).toFixed(2)} / 件</span>
				</div>
			{/if}

			<div class="border-base-300 flex flex-wrap items-center justify-end gap-4 border-t pt-6">
				<button type="button" class="btn" onclick={() => goto('/production/orders')} disabled={submitting}>取消</button>
				<button type="button" class="btn btn-primary" onclick={handleSubmit} disabled={submitting}>
					{#if submitting}<span class="loading loading-spinner loading-sm"></span>{/if}
					提交生产单
				</button>
			</div>
		</div>
	</div>
</div>

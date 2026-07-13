<!--
功能：新增生产单（参考采购单样式，支持多领料单关联，手动调整成本）
创建时间：2026-07-12 / 更新：2026-07-13
创建人：Hermes Agent
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { ArrowLeft, Package, Warehouse, Hash, DollarSign, FileText, Plus, Trash2 } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	let materials = $state<any[]>([]);
	let warehouses = $state<any[]>([]);
	let consumptionOrders = $state<any[]>([]);

	let form = $state({
		produced_material_id: 0,
		produced_warehouse_id: 0,
		produced_quantity: 1,
		cost_price: 0,
		consumption_order_ids: [] as number[],
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
		} catch (e) { toast.error('加载基础数据失败'); }
	});

	function addConsumptionOrder() { form.consumption_order_ids = [...form.consumption_order_ids, 0]; }
	function removeConsumptionOrder(idx: number) {
		form.consumption_order_ids = form.consumption_order_ids.filter((_, i) => i !== idx);
	}

	async function handleSubmit() {
		if (!form.produced_material_id) { toast.error('请选择成品物料'); return; }
		if (!form.produced_warehouse_id) { toast.error('请选择入库仓库'); return; }
		if (form.produced_quantity <= 0) { toast.error('数量必须大于0'); return; }
		submitting = true;
		try {
			const payload: any = {
				produced_material_id: form.produced_material_id,
				produced_warehouse_id: form.produced_warehouse_id,
				produced_quantity: form.produced_quantity,
				cost_price: form.cost_price,
				remark: form.remark
			};
			// 去重有效的领料单ID
			const validIds = [...new Set(form.consumption_order_ids.filter(id => id > 0))];
			if (validIds.length > 0) payload.consumption_order_ids = validIds;

			await api.post('/production/orders', payload);
			toast.success('生产单创建成功');
			goto('/production/orders');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally { submitting = false; }
	}
</script>

<svelte:window />

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<button type="button" class="btn btn-ghost btn-circle" onclick={() => goto('/production/orders')}>
				<ArrowLeft size={20} />
			</button>
			<div class="h-8 w-1.5 rounded-full bg-purple-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">新增生产单</h1>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<!-- 基本信息 -->
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label" for="prod-material">
						<span class="label-text flex items-center gap-2 font-medium">
							<Package size={14} /> 成品物料 <span class="text-error">*</span>
						</span>
					</label>
					<select id="prod-material" bind:value={form.produced_material_id}
						class="select select-bordered bg-base-200/50 h-11 w-full text-base" required>
						<option value={0}>选择成品物料</option>
						{#each materials as m}
							<option value={m.id}>{m.material_display_name || m.material_name} [{m.material_code}]</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label" for="prod-warehouse">
						<span class="label-text flex items-center gap-2 font-medium">
							<Warehouse size={14} /> 入库仓库 <span class="text-error">*</span>
						</span>
					</label>
					<select id="prod-warehouse" bind:value={form.produced_warehouse_id}
						class="select select-bordered bg-base-200/50 h-11 w-full text-base" required>
						<option value={0}>选择入库仓库</option>
						{#each warehouses as wh}
							<option value={wh.id}>{wh.warehouse_name}</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label" for="prod-qty">
						<span class="label-text flex items-center gap-2 font-medium">
							<Hash size={14} /> 生产数量 <span class="text-error">*</span>
						</span>
					</label>
					<input id="prod-qty" type="number" bind:value={form.produced_quantity} min="1" step="1"
						class="input input-bordered bg-base-200/50 h-11 w-full text-base" required />
				</div>

				<div class="form-control">
					<label class="label" for="prod-cost">
						<span class="label-text flex items-center gap-2 font-medium">
							<DollarSign size={14} /> 成本价格（元）
						</span>
					</label>
					<input id="prod-cost" type="number" bind:value={form.cost_price} min="0" step="0.01"
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="总成本（自动计算单位成本）" />
					{#if form.cost_price > 0 && form.produced_quantity > 0}
						<label class="label"><span class="label-text-alt text-info">单位成本：¥{(form.cost_price / form.produced_quantity).toFixed(2)}</span></label>
					{/if}
				</div>

				<div class="form-control xl:col-span-2">
					<label class="label" for="prod-remark">
						<span class="label-text flex items-center gap-2 font-medium">
							<FileText size={14} /> 备注
						</span>
					</label>
					<input id="prod-remark" type="text" bind:value={form.remark}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base" placeholder="备注信息" />
				</div>
			</div>

			<!-- 关联领料单 -->
			<div class="divider">关联领料单（可选，可多选）</div>

			<div class="flex flex-wrap items-center justify-end gap-2">
				<button type="button" class="btn btn-sm btn-outline gap-2" onclick={addConsumptionOrder}>
					<Plus size={14} /> 添加领料单
				</button>
			</div>

			{#if form.consumption_order_ids.length === 0}
				<div class="text-base-content/50 py-6 text-center text-sm">无需关联领料单（手动生产单）</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-zebra table w-full text-base">
						<thead>
							<tr>
								<th>领料单号</th>
								<th>产品名称</th>
								<th>状态</th>
								<th class="w-14">操作</th>
							</tr>
						</thead>
						<tbody>
							{#each form.consumption_order_ids as coId, i}
								<tr>
									<td>
										<select bind:value={form.consumption_order_ids[i]}
											class="select select-bordered bg-base-200/50 h-10 w-full text-sm">
											<option value={0}>选择领料单</option>
											{#each consumptionOrders as co}
												{@const alreadyPicked = form.consumption_order_ids.some((id: number, j: number) => j !== i && id === co.id)}
												<option value={co.id} disabled={alreadyPicked}>
													{co.order_no} - {co.product_name || co.project_no} ({co.status})
												</option>
											{/each}
										</select>
									</td>
									<td class="text-sm">
										{#if coId > 0}
											{consumptionOrders.find((c: any) => c.id === coId)?.product_name || '-'}
										{:else}-{/if}
									</td>
									<td>
										{#if coId > 0}
											<span class="badge badge-sm">{consumptionOrders.find((c: any) => c.id === coId)?.status || '-'}</span>
										{:else}-{/if}
									</td>
									<td>
										<button type="button" class="btn btn-xs btn-ghost text-error"
											onclick={() => removeConsumptionOrder(i)}>
											<Trash2 size={12} />
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}

			<!-- 提交 -->
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

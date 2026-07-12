<!--
功能：编辑生产单（成本价格、备注）
创建时间：2026-07-12
创建人：Hermes Agent
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { ArrowLeft, Package, Hash, DollarSign } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';

	let id = $state(0);
	let data = $state<any>({});
	let loading = $state(true);
	let form = $state({ cost_price: 0, remark: '' });
	let submitting = $state(false);

	onMount(async () => {
		id = Number($page.params.id);
		try {
			const res: any = await api.get(`/production/orders/${id}`);
			data = res;
			form.cost_price = Number(res.cost_price) || 0;
			form.remark = res.remark || '';
		} catch (e) {
			toast.error('加载生产单失败');
		} finally {
			loading = false;
		}
	});

	async function handleSubmit() {
		submitting = true;
		try {
			await api.put(`/production/orders/${id}`, form);
			toast.success('保存成功');
			goto(`/production/orders/${id}`);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center gap-3">
		<button type="button" class="btn btn-ghost btn-circle" onclick={() => goto(`/production/orders/${id}`)}>
			<ArrowLeft size={20} />
		</button>
		<div class="h-8 w-1.5 rounded-full bg-purple-500"></div>
		<h1 class="text-2xl font-bold tracking-tight">编辑生产单</h1>
	</div>

	{#if loading}
		<div class="flex justify-center py-20"><span class="loading loading-spinner loading-lg"></span></div>
	{:else}
	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<!-- 只读信息 -->
			<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
				<div class="form-control">
					<label class="label"><span class="label-text font-medium">生产单号</span></label>
					<div class="input input-bordered bg-base-200 flex h-11 items-center font-mono">{data.production_no || '-'}</div>
				</div>
				<div class="form-control">
					<label class="label"><span class="label-text font-medium">成品物料</span></label>
					<div class="input input-bordered bg-base-200 flex h-11 items-center">
						<Package size={14} class="mr-2 text-base-content/50" />
						{data.produced_material_name || '-'}
					</div>
				</div>
				<div class="form-control">
					<label class="label"><span class="label-text font-medium">生产数量</span></label>
					<div class="input input-bordered bg-base-200 flex h-11 items-center font-mono">
						<Hash size={14} class="mr-2 text-base-content/50" />
						{data.produced_quantity || 0}
					</div>
				</div>
			</div>

			<div class="divider">可编辑字段</div>

			<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
				<div class="form-control">
					<label class="label">
						<span class="label-text flex items-center gap-2 font-medium">
							<DollarSign size={14} /> 成本价格（元）
						</span>
					</label>
					<input type="number" bind:value={form.cost_price} min="0" step="0.01"
						class="input input-bordered h-11 w-full" />
					{#if form.cost_price > 0 && data.produced_quantity > 0}
						<label class="label"><span class="label-text-alt text-info">单位成本：¥{(form.cost_price / data.produced_quantity).toFixed(2)}</span></label>
					{/if}
				</div>

				<div class="form-control">
					<label class="label"><span class="label-text font-medium">备注</span></label>
					<input type="text" bind:value={form.remark}
						class="input input-bordered h-11 w-full" placeholder="备注信息" />
				</div>
			</div>

			<div class="border-base-300 flex flex-wrap items-center justify-end gap-4 border-t pt-6">
				<button type="button" class="btn" onclick={() => goto(`/production/orders/${id}`)} disabled={submitting}>取消</button>
				<button type="button" class="btn btn-primary" onclick={handleSubmit} disabled={submitting}>
					{#if submitting}<span class="loading loading-spinner loading-sm"></span>{/if}
					保存修改
				</button>
			</div>
		</div>
	</div>
	{/if}
</div>

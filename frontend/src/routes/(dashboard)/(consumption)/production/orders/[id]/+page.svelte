<!--
功能：生产单据详情（可编辑成本价格，查看关联领料/退料单）
创建时间：2026-06-06
创建人：GPT-5.2
-->

<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { formatDateTimeInCn } from '$lib/datetime';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Printer } from 'lucide-svelte';

	let data = $state<any>(null);
	let consumptionOrders = $state<any[]>([]);
	let reversalOrders = $state<any[]>([]);
	let loading = $state(true);
	let editCostPrice = $state(0);
	let editRemark = $state('');
	let saving = $state(false);
	let showEdit = $state(false);

	function id() {
		return Number(page.params.id || 0);
	}

	async function load() {
		loading = true;
		try {
			data = await api.get(`/production/orders/${id()}`);
			editCostPrice = Number(data?.cost_price || 0);
			editRemark = String(data?.remark || '');
		} catch (e) {
			console.error(e);
			toast.error('加载生产单详情失败');
		} finally {
			loading = false;
		}
	}

	async function loadRelated() {
		try {
			const coRes: any = await api.get(`/production/orders/${id()}/consumption-orders`);
			consumptionOrders = coRes || [];
		} catch (e) {
			console.error(e);
		}
		try {
			const roRes: any = await api.get(`/production/orders/${id()}/reversal-orders`);
			reversalOrders = roRes || [];
		} catch (e) {
			console.error(e);
		}
	}

	async function handleSave() {
		saving = true;
		try {
			await api.put(`/production/orders/${id()}`, {
				cost_price: Number(editCostPrice),
				remark: editRemark
			});
			toast.success('保存成功');
			data = { ...data, cost_price: Number(editCostPrice), remark: editRemark };
			showEdit = false;
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			saving = false;
		}
	}

	function fmt(v: any) {
		if (v === null || v === undefined || v === '') return '-';
		return String(v);
	}

	function fmtDT(v: any) {
		if (!v) return '-';
		return formatDateTimeInCn(String(v));
	}

	function fmtMoney(v: any) {
		const n = Number(v || 0);
		return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function statusName(v: string) {
		const map: any = { completed: '已完成' };
		return map[v] || v || '-';
	}

	onMount(load);
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
	<div class="flex items-center justify-between gap-2 print:hidden">
		<div class="flex items-center gap-2">
			<button class="btn btn-ghost gap-2" onclick={() => goto('/production/orders')}>
				<ArrowLeft size={18} /> 返回
			</button>
			<h1 class="text-xl font-bold">生产单详情</h1>
		</div>
		{#if data}
			<button class="btn btn-outline btn-sm gap-1" onclick={() => window.print()}>
				<Printer size={16} /> 打印
			</button>
		{/if}
	</div>

	{#if loading}
		<div class="skeleton h-40 w-full"></div>
	{:else if !data}
		<div class="text-base-content/60">无数据</div>
	{:else}
		<div class="card bg-base-100 border-base-300 border">
			<div class="card-body grid grid-cols-1 gap-4 md:grid-cols-2">
				<div>
					<div class="text-base-content/60 text-sm">生产单号</div>
					<div class="font-mono text-lg">{fmt(data.production_no)}</div>
				</div>
				<div>
					<div class="text-base-content/60 text-sm">状态</div>
					<div class="badge badge-success">{statusName(data.status)}</div>
				</div>

				<div>
					<div class="text-base-content/60 text-sm">成品物料</div>
					<div>{fmt(data.produced_material_name || data.produced_material_code)}</div>
				</div>

				<div>
					<div class="text-base-content/60 text-sm">成品仓库</div>
					<div>{fmt(data.produced_warehouse_name || data.produced_warehouse_code)}</div>
				</div>

				<div>
					<div class="text-base-content/60 text-sm">成品数量</div>
					<div class="font-mono">{Number(data.produced_quantity || 0).toFixed(3)}</div>
				</div>

				<div>
					<div class="text-base-content/60 text-sm">单位成本</div>
					<div class="font-mono">{Number(data.produced_unit_cost || 0).toFixed(3)}</div>
				</div>

				<!-- 成本总价（可编辑） -->
				<div class="md:col-span-2">
					<div class="text-base-content/60 text-sm flex items-center gap-2">
						成本总价
						{#if !showEdit}
							<button class="btn btn-xs btn-ghost" onclick={() => (showEdit = true)}>
								编辑
							</button>
						{/if}
					</div>
					{#if showEdit}
						<div class="flex items-center gap-2 mt-1">
							<input
								type="number"
								bind:value={editCostPrice}
								min="0"
								step="0.01"
								class="input input-bordered h-10 w-48 text-base font-mono"
							/>
							<button class="btn btn-sm btn-primary" onclick={handleSave} disabled={saving}>
								{saving ? '保存中...' : '保存'}
							</button>
							<button class="btn btn-sm" onclick={() => { showEdit = false; editCostPrice = Number(data?.cost_price || 0); }} disabled={saving}>取消</button>
						</div>
					{:else}
						<div class="text-success font-mono text-lg font-semibold">¥{fmtMoney(data.cost_price)}</div>
					{/if}
				</div>

				<div>
					<div class="text-base-content/60 text-sm">领料出库单</div>
					{#if data.stock_out_id}
						<a class="link link-primary font-mono" href={`/stock/out/${data.stock_out_id}`}>
							{fmt(data.stock_out_no)}</a>
					{:else}
						<div class="text-base-content/60">-</div>
					{/if}
				</div>

				<div>
					<div class="text-base-content/60 text-sm">成品入库单</div>
					{#if data.stock_in_id}
						<a class="link link-primary font-mono" href={`/stock/in/${data.stock_in_id}`}>
							{fmt(data.stock_in_no)}</a>
					{:else}
						<div class="text-base-content/60">-</div>
					{/if}
				</div>

				<div class="md:col-span-2">
					<div class="text-base-content/60 text-sm">备注</div>
					<div>{fmt(data.remark)}</div>
				</div>

				<div>
					<div class="text-base-content/60 text-sm">创建时间</div>
					<div class="font-mono">{fmtDT(data.created_at)}</div>
				</div>
				<div>
					<div class="text-base-content/60 text-sm">更新时间</div>
					<div class="font-mono">{fmtDT(data.updated_at)}</div>
				</div>
			</div>
		</div>

		<!-- 关联领料单列表 -->
		<div class="card bg-base-100 border-base-300 border mt-4">
			<div class="card-body">
				<h2 class="card-title text-lg flex items-center gap-2">
					<span class="h-5 w-1 rounded bg-orange-500"></span>
					关联领料单 ({consumptionOrders.length})
				</h2>
				{#if consumptionOrders.length === 0}
					<div class="text-base-content/50 py-6 text-center text-sm">暂无关联领料单</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table-zebra table w-full text-sm">
							<thead>
								<tr>
									<th>领料单号</th>
									<th>项目编号</th>
									<th>产品名称</th>
									<th class="text-right">数量</th>
									<th class="text-right">金额</th>
									<th>状态</th>
									<th>日期</th>
								</tr>
							</thead>
							<tbody>
								{#each consumptionOrders as co}
									<tr>
										<td>
											<a class="link link-primary font-mono" href={`/consumption/orders/${co.id}`}>
												{co.order_no}
											</a>
										</td>
										<td class="font-mono">{co.project_no || '-'}</td>
										<td>{co.product_name || '-'}</td>
										<td class="text-right font-mono">{Number(co.total_quantity || 0).toFixed(3)}</td>
										<td class="text-success text-right font-mono font-semibold">¥{fmtMoney(co.total_amount)}</td>
										<td>
											<span class="badge badge-sm">{co.status_name || co.status}</span>
										</td>
										<td class="font-mono text-xs">{fmtDT(co.order_date || co.created_at)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>

		<!-- 关联退料单列表 -->
		<div class="card bg-base-100 border-base-300 border mt-4">
			<div class="card-body">
				<h2 class="card-title text-lg flex items-center gap-2">
					<span class="h-5 w-1 rounded bg-blue-500"></span>
					关联退料单 ({reversalOrders.length})
				</h2>
				{#if reversalOrders.length === 0}
					<div class="text-base-content/50 py-6 text-center text-sm">暂无关联退料单</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table-zebra table w-full text-sm">
							<thead>
								<tr>
									<th>退料单号</th>
									<th>项目编号</th>
									<th>产品名称</th>
									<th class="text-right">数量</th>
									<th class="text-right">金额</th>
									<th>状态</th>
									<th>日期</th>
								</tr>
							</thead>
							<tbody>
								{#each reversalOrders as ro}
									<tr>
										<td>
											<a class="link link-primary font-mono" href={`/reversal/orders/${ro.id}`}>
												{ro.order_no}
											</a>
										</td>
										<td class="font-mono">{ro.project_no || '-'}</td>
										<td>{ro.product_name || '-'}</td>
										<td class="text-right font-mono">{Number(ro.total_quantity || 0).toFixed(3)}</td>
										<td class="text-success text-right font-mono font-semibold">¥{fmtMoney(ro.total_amount)}</td>
										<td>
											<span class="badge badge-sm">{ro.status_name || ro.status}</span>
										</td>
										<td class="font-mono text-xs">{fmtDT(ro.order_date || ro.created_at)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	@page {
		size: A4;
		margin: 12mm;
	}
</style>

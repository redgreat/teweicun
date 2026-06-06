<!--
功能：采购退货详情页
创建时间：2026-05-16
创建人：GPT-5.4
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { ArrowLeft, FileText, Building2, Calendar, BadgeCheck, Package, ClipboardList, Printer } from 'lucide-svelte';
	import { formatDateInCn, formatDateTimeInCn } from '$lib/datetime';

	let loading = $state(true);
	let order = $state<any | null>(null);

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatDateTime(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateTimeInCn(dateStr);
	}

	function statusBadgeClass(status: string) {
		if (status === 'draft') return 'badge-warning';
		if (status === 'pending_out') return 'badge-info';
		if (status === 'completed') return 'badge-success';
		if (status === 'confirmed') return 'badge-info';
		return 'badge-ghost';
	}

	function statusName(status: string) {
		const map: Record<string, string> = {
			draft: '待提交',
			pending_out: '待出库',
			completed: '已完成',
			confirmed: '待出库'
		};
		return map[status] || status || '-';
	}

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function grandTotal() {
		let s = 0;
		for (const it of order?.items || [])
			s += (Number(it?.quantity) || 0) * (Number(it?.unit_cost) || 0);
		return s;
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			const detail: any = await api.get(`/returns/${id}`);
			order = detail;
		} catch (err: any) {
			toast.error('加载退货单详情失败: ' + (err?.message || err));
			order = null;
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		const id = Number(page.params.id);
		if (!id) {
			toast.error('无效的退货单 ID');
			loading = false;
			return;
		}
		loadDetail(id);
	});
</script>

<div class="space-y-6 text-base">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-amber-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">采购退货单详情</h1>
		</div>

		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>采购管理</li>
				<li><a class="text-primary" href="/purchase/return">采购退货</a></li>
				<li>详情</li>
			</ul>
		</div>
	</div>

	<div class="flex flex-wrap items-center justify-between gap-3 print:hidden">
		<div class="flex flex-wrap items-center gap-2">
			<a href="/purchase/return" class="btn btn-ghost gap-1 text-base">
				<ArrowLeft size={18} /> 返回列表
			</a>
			{#if order?.status === 'pending_out'}
				<a href={`/purchase/return/${order.id}/edit`} class="btn btn-primary btn-sm text-base"
					>编辑</a
				>
			{/if}
		</div>
		{#if order}
			<div class="flex flex-wrap items-center gap-2">
				<button class="btn btn-outline btn-sm gap-1" onclick={() => window.print()}>
					<Printer size={16} /> 打印
				</button>
				<span class="badge badge-lg {statusBadgeClass(order.status)}"
					>{statusName(order.status)}</span
				>
				<span class="text-base-content/70 font-mono">{order.return_no}</span>
			</div>
		{/if}
	</div>

	{#if loading}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center text-lg"
		>
			正在加载...
		</div>
	{:else if !order}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center text-lg"
		>
			未找到退货单信息
		</div>
	{:else}
		<div class="bg-base-100 border-base-300 space-y-5 rounded-2xl border p-5">
			<div class="flex items-center gap-2 text-lg font-semibold">
				<FileText size={18} /> 单据信息
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<FileText size={16} /> 退货单号
					</div>
					<div class="mt-1 font-mono text-base font-medium">{order.return_no}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<Building2 size={16} /> 供应商
					</div>
					<div class="mt-1 text-base font-semibold">{order.supplier_name || '-'}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<Calendar size={16} /> 退货日期
					</div>
					<div class="mt-1 text-base">{formatDate(order.return_date)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<BadgeCheck size={16} /> 状态
					</div>
					<div class="mt-1">
						<span class="badge badge-lg {statusBadgeClass(order.status)}"
							>{statusName(order.status)}</span
						>
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">退货金额合计</div>
					<div class="text-success mt-1 font-mono text-base font-semibold">
						¥{formatMoney(grandTotal())}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<ClipboardList size={16} /> 关联出库单
					</div>
					<div class="mt-1">
						{#if order.stock_out_id && order.stock_out_no}
							<a
								class="link link-primary font-mono text-base no-underline hover:underline"
								href={`/stock/out/${order.stock_out_id}`}
							>
								{order.stock_out_no}
							</a>
						{:else}
							<span class="text-base-content/50">-</span>
						{/if}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">创建时间</div>
					<div class="mt-1 text-base">{formatDateTime(order.created_at)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-4">
					<div class="text-base-content/55 text-sm">备注</div>
					<div class="mt-1 text-base break-words whitespace-pre-wrap">{order.remark || '-'}</div>
				</div>
			</div>
		</div>

		<div class="bg-base-100 border-base-300 space-y-4 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="flex items-center gap-2 text-lg font-semibold">
					<Package size={18} /> 退货明细
				</div>
				<div class="text-base-content/50 text-sm">共 {order.items?.length || 0} 行</div>
			</div>

			<div class="overflow-x-auto">
				<table class="table-zebra table w-full min-w-[720px] table-fixed text-base">
					<thead>
						<tr>
							<th class="w-[46%] min-w-[320px]">物料名称</th>
							<th class="w-24 whitespace-nowrap">属性</th>
							<th class="text-right">数量</th>
							<th>单位</th>
							<th class="min-w-[110px] text-right">单价</th>
							<th class="min-w-[120px] text-right">物料总价</th>
						</tr>
					</thead>
					<tbody>
						{#each order.items || [] as it}
							<tr>
								<td>
									<div class="space-y-1 pr-3">
										<div class="group/line relative">
											<div class="truncate font-medium">{it.material_name || '-'}</div>
											<div
												class="border-base-300 bg-base-100 pointer-events-none absolute top-full left-0 z-20 mt-1 hidden max-w-[42rem] rounded-md border px-2 py-1 text-sm leading-5 whitespace-normal shadow-xl group-hover/line:block"
											>
												{it.material_name || '-'}
											</div>
										</div>
										<div class="text-base-content/60 truncate font-mono text-base">
											{it.material_code || '-'}
										</div>
									</div>
								</td>
								<td class="whitespace-nowrap">
									<span class="text-sm">{(it.custom_attributes || []).length}项</span>
								</td>
								<td class="text-right font-mono">{it.quantity ?? '-'}</td>
								<td>{it.unit || '-'}</td>
								<td class="text-right font-mono">
									{#if it.inventory_id}
										¥{(Number(it.unit_cost) || 0).toLocaleString('zh-CN', {
											minimumFractionDigits: 2,
											maximumFractionDigits: 2
										})}
									{:else}
										—
									{/if}
								</td>
								<td class="text-success text-right font-mono font-semibold">
									{#if it.inventory_id}
										¥{((Number(it.quantity) || 0) * (Number(it.unit_cost) || 0)).toLocaleString(
											'zh-CN',
											{
												minimumFractionDigits: 2,
												maximumFractionDigits: 2
											}
										)}
									{:else}
										—
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

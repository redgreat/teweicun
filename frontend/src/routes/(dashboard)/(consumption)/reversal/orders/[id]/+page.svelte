<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { ArrowLeft, FileText, Calendar, BadgeCheck, ClipboardList } from 'lucide-svelte';
	import { getStatusStyle } from '$lib/statusStyles';
	import { formatDateInCn, formatDateTimeInCn } from '$lib/datetime';

	let loading = $state(true);
	let order = $state<any | null>(null);

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function lineAmount(it: any) {
		return (Number(it?.quantity) || 0) * (Number(it?.unit_cost) || 0);
	}

	function grandTotal() {
		let s = 0;
		for (const it of order?.items || []) s += lineAmount(it);
		return s;
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatDateTime(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateTimeInCn(dateStr);
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			order = await api.get(`/reversal/orders/${id}`);
		} catch (err: any) {
			toast.error('加载退料单详情失败: ' + (err?.message || err));
			order = null;
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		const id = Number($page.params.id);
		if (!id) {
			toast.error('无效的订单ID');
			loading = false;
			return;
		}
		loadDetail(id);
	});
</script>

<div class="space-y-6 text-base">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-cyan-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">退料单详情</h1>
		</div>

		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>出入库管理</li>
				<li><a class="text-primary" href="/reversal/orders">退料</a></li>
				<li>详情</li>
			</ul>
		</div>
	</div>

	<div class="flex flex-wrap items-center justify-between gap-3">
		<div class="flex flex-wrap items-center gap-2">
			<a href="/reversal/orders" class="btn btn-ghost gap-1 text-base">
				<ArrowLeft size={18} /> 返回列表
			</a>
		</div>

		{#if order}
			{@const style = getStatusStyle(order.status, 'reversal_order')}
			<div class="flex flex-wrap items-center gap-2">
				<span class="badge badge-lg {style.class}">{style.label}</span>
				<span class="text-base-content/70 font-mono">{order.order_no}</span>
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
			未找到退料单信息
		</div>
	{:else}
		<div class="bg-base-100 border-base-300 space-y-5 rounded-2xl border p-5">
			<div class="flex items-center gap-2 text-lg font-semibold">
				<FileText size={18} /> 单据信息
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<FileText size={16} /> 退料单号
					</div>
					<div class="mt-1 font-mono text-base font-medium">{order.order_no}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<ClipboardList size={16} /> 项目 / 产品
					</div>
					<div class="mt-1 text-base">
						<div class="font-mono">{order.project_no || '-'}</div>
						<div class="text-base-content/70">{order.product_name || '-'}</div>
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<Calendar size={16} /> 退料日期
					</div>
					<div class="mt-1 text-base">{formatDate(order.order_date)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<BadgeCheck size={16} /> 状态
					</div>
					<div class="mt-1">
						<span class="badge badge-lg {getStatusStyle(order.status, 'reversal_order').class}">
							{getStatusStyle(order.status, 'reversal_order').label}
						</span>
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">退料金额合计</div>
					<div class="text-success mt-1 font-mono text-base font-semibold">
						¥{formatMoney(grandTotal())}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">入库单号</div>
					<div class="mt-1">
						{#if order.stock_in_id && order.stock_in_no}
							<a
								class="link link-primary font-mono text-base no-underline hover:underline"
								href={`/stock/in/${order.stock_in_id}`}
							>
								{order.stock_in_no}
							</a>
						{:else}
							-
						{/if}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">创建时间</div>
					<div class="mt-1 text-base">{formatDateTime(order.created_at)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">更新时间</div>
					<div class="mt-1 text-base">{formatDateTime(order.updated_at)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-4">
					<div class="text-base-content/55 text-sm">备注</div>
					<div class="mt-1 text-base break-words whitespace-pre-wrap">{order.remark || '-'}</div>
				</div>
			</div>
		</div>

		<div class="bg-base-100 border-base-300 space-y-4 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="text-lg font-semibold">退料明细</div>
				<div class="text-base-content/50 text-base">
					共 {order.items?.length || 0} 行 · 合计数量 {order.total_quantity ?? '-'}
				</div>
			</div>

			<div class="overflow-x-auto">
				<table class="table-zebra table w-full min-w-[860px] table-fixed text-base">
					<thead>
						<tr>
							<th class="w-[46%] min-w-[320px]">物料名称</th>
							<th class="w-24 whitespace-nowrap">属性</th>
							<th class="min-w-[9rem]">仓库</th>
							<th class="text-right">数量</th>
							<th class="min-w-[5rem]">单位</th>
							<th class="min-w-[8rem] text-right">单价</th>
							<th class="min-w-[9rem] text-right">金额</th>
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
										<div
											class="text-base-content/60 truncate font-mono text-base"
											title={it.material_code || ''}
										>
											{it.material_code || '-'}
										</div>
									</div>
								</td>
								<td class="whitespace-nowrap">
									<span class="text-sm">{(it.custom_attributes || []).length}项</span>
								</td>
								<td>{it.warehouse_name || '-'}</td>
								<td class="text-right font-mono">{it.quantity ?? '-'}</td>
								<td>{it.unit || '-'}</td>
								<td class="text-right font-mono">¥{formatMoney(it.unit_cost)}</td>
								<td class="text-success text-right font-mono font-semibold">
									¥{formatMoney(lineAmount(it))}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

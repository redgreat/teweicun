<!--
功能：销售订单详情页
创建时间：2026-05-10
创建人：GPT-5.4
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { ArrowLeft, FileText, Calendar, User, Truck, CircleCheckBig, Ban, ClipboardList, Pencil, Printer } from 'lucide-svelte';
	import { getStatusStyle } from '$lib/statusStyles';
	import { formatDateInCn, formatDateTimeInCn } from '$lib/datetime';

	let loading = $state(true);
	let detail = $state<any | null>(null);
	let acting = $state(false);
	let showEdit = $state(false);

	function formatMoney(value: number) {
		const amount = Number(value) || 0;
		return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function formatDate(value: string) {
		if (!value) return '-';
		return formatDateInCn(value);
	}

	function formatDateTime(value: string) {
		if (!value) return '-';
		return formatDateTimeInCn(value);
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			detail = await api.get(`/sales/orders/${id}`);
		} catch (err: any) {
			toast.error('加载销售订单详情失败: ' + (err?.message || err));
			detail = null;
		} finally {
			loading = false;
		}
	}

	async function handleConfirm() {
		if (!detail?.id) return;
		acting = true;
		try {
			await api.post(`/sales/orders/${detail.id}/confirm`, {});
			toast.success('销售订单已确认');
			await loadDetail(detail.id);
		} catch (err: any) {
			toast.error('确认失败: ' + (err?.message || err));
		} finally {
			acting = false;
		}
	}

	async function handleCancel() {
		if (!detail?.id) return;
		acting = true;
		try {
			await api.post(`/sales/orders/${detail.id}/cancel`, {});
			toast.success('销售订单已取消');
			await loadDetail(detail.id);
		} catch (err: any) {
			toast.error('取消失败: ' + (err?.message || err));
		} finally {
			acting = false;
		}
	}

	onMount(() => {
		const id = Number(page.params.id);
		if (!id) {
			loading = false;
			toast.error('无效的销售订单ID');
			return;
		}
		loadDetail(id);
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-emerald-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">销售订单详情</h1>
		</div>
		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>销售管理</li>
				<li><a class="text-primary" href="/sales/orders">销售订单</a></li>
				<li>详情</li>
			</ul>
		</div>
	</div>

	<div class="flex flex-wrap items-center justify-between gap-3 print:hidden">
		<a href="/sales/orders" class="btn btn-ghost btn-sm gap-1">
			<ArrowLeft size={14} /> 返回列表
		</a>
		{#if detail}
			<div class="flex flex-wrap items-center gap-2">
				<button class="btn btn-outline btn-sm gap-1" onclick={() => window.print()}>
					<Printer size={16} /> 打印
				</button>
				{#if detail.order_status === 'draft'}
					<button
						class="btn btn-sm btn-success text-white"
						onclick={handleConfirm}
						disabled={acting}
					>
						<CircleCheckBig size={16} /> 提交订单
					</button>
				{/if}
				{#if detail.order_status === 'confirmed' || detail.order_status === 'preparing'}
					{#if detail.stock_out_id && detail.stock_out_no}
						<a
							href="/stock/out/{detail.stock_out_id}"
							class="btn btn-sm btn-outline gap-1"
						>
							<Truck size={14} /> {detail.stock_out_no}
						</a>
					{/if}
					<button
						class="btn btn-sm btn-primary btn-outline gap-1"
						onclick={() => (showEdit = true)}
					>
						<Pencil size={14} /> 编辑
					</button>
				{/if}
				{#if detail.order_status !== 'shipped' && detail.order_status !== 'cancelled'}
					<button class="btn btn-sm btn-error btn-outline" onclick={handleCancel} disabled={acting}>
						<Ban size={16} /> 取消订单
					</button>
				{/if}
			</div>
		{/if}
	</div>

	{#if loading}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center text-lg"
		>
			正在加载...
		</div>
	{:else if !detail}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center text-lg"
		>
			未找到销售订单信息
		</div>
	{:else}
		<div class="bg-base-100 border-base-300 space-y-5 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="flex items-center gap-2 text-lg font-semibold">
					<FileText size={18} /> 单据信息
				</div>
				<div class="flex items-center gap-2">
					<span class="badge badge-lg {getStatusStyle(detail.order_status, 'sales_order').class}">
						{getStatusStyle(detail.order_status, 'sales_order').label}
					</span>
					<span class="text-base-content/70 font-mono">{detail.order_no}</span>
				</div>
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<FileText size={16} /> 销售单号
					</div>
					<div class="mt-1 font-mono text-base font-medium">{detail.order_no}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<User size={16} /> 客户
					</div>
					<div class="mt-1 text-base">{detail.customer_name || '-'}</div>
					<div class="text-base-content/60 font-mono text-sm">{detail.customer_code || '-'}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<Calendar size={16} /> 下单日期
					</div>
					<div class="mt-1 text-base">{formatDate(detail.order_date)}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<Calendar size={16} /> 交付日期
					</div>
					<div class="mt-1 text-base">{formatDate(detail.delivery_date)}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">合同编号</div>
					<div class="mt-1 text-base">{detail.contract_no || '-'}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">付款方式</div>
					<div class="mt-1 text-base">{detail.payment_method || '-'}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">收货联系人</div>
					<div class="mt-1 text-base">{detail.receiver_name || '-'}</div>
					<div class="text-base-content/60 text-sm">{detail.receiver_phone || '-'}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">订单金额</div>
					<div class="text-success mt-1 font-mono text-base font-semibold">
						¥{formatMoney(detail.total_amount)}
					</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-2">
					<div class="text-base-content/55 text-sm">收货地址</div>
					<div class="mt-1 text-base break-words whitespace-pre-wrap">
						{detail.receiver_address || '-'}
					</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">创建时间</div>
					<div class="mt-1 text-base">{formatDateTime(detail.created_at)}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-4">
					<div class="text-base-content/55 text-sm">备注</div>
					<div class="mt-1 text-base break-words whitespace-pre-wrap">{detail.remark || '-'}</div>
				</div>
			</div>
		</div>

		<div class="bg-base-100 border-base-300 space-y-4 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="flex items-center gap-2 text-lg font-semibold">
					<ClipboardList size={18} /> 销售明细
				</div>
				<div class="text-base-content/50 text-base">共 {detail.items?.length || 0} 行</div>
			</div>
			<div class="overflow-x-auto">
				<table class="table-zebra table w-full min-w-[860px] table-fixed text-base">
					<thead>
						<tr>
							<th class="w-[36%] min-w-[320px]">物料名称</th>
							<th class="text-right">数量</th>
							<th class="text-right">已出库</th>
							<th>单位</th>
							<th class="text-right">销售单价</th>
							<th class="text-right">金额</th>
							<th>备注</th>
						</tr>
					</thead>
					<tbody>
						{#each detail.items || [] as item}
							<tr>
								<td>
									<div class="space-y-1 pr-3">
										<div class="truncate font-medium">{item.material_name || '-'}</div>
										<div class="text-base-content/60 truncate font-mono text-sm">
											{item.material_code || '-'}
										</div>
									</div>
								</td>
								<td class="text-right font-mono">{item.quantity ?? 0}</td>
								<td class="text-right font-mono">{item.shipped_quantity ?? 0}</td>
								<td>{item.unit || '-'}</td>
								<td class="text-right font-mono">¥{formatMoney(item.unit_price)}</td>
								<td class="text-success text-right font-mono font-semibold">
									¥{formatMoney(item.amount)}
								</td>
								<td>{item.remark || '-'}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

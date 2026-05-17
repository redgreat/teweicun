<!--
功能：销售退货详情页
创建时间：2026-05-10
创建人：GPT-5.4
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import {
		ArrowLeft,
		FileText,
		Calendar,
		User,
		CheckCircle,
		Trash2,
		ClipboardList,
		PackageOpen,
		Pencil
	} from 'lucide-svelte';
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

	function lineAmount(item: any) {
		return (Number(item.quantity) || 0) * (Number(item.unit_cost) || 0);
	}

	function totalAmount() {
		return (detail?.items || []).reduce((sum: number, item: any) => sum + lineAmount(item), 0);
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			detail = await api.get(`/returns/${id}`);
		} catch (err: any) {
			toast.error('加载销售退货详情失败: ' + (err?.message || err));
			detail = null;
		} finally {
			loading = false;
		}
	}

	async function handleConfirm() {
		if (!detail?.id) return;
		acting = true;
		try {
			await api.post(`/returns/${detail.id}/confirm`, {});
			toast.success('销售退货已确认入库');
			await loadDetail(detail.id);
		} catch (err: any) {
			toast.error('确认失败: ' + (err?.message || err));
		} finally {
			acting = false;
		}
	}

	async function handleDelete() {
		if (!detail?.id) return;
		acting = true;
		try {
			await api.delete(`/returns/${detail.id}`);
			toast.success('销售退货已删除');
			window.location.href = '/sales/returns';
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		} finally {
			acting = false;
		}
	}

	onMount(() => {
		const id = Number(page.params.id);
		if (!id) {
			loading = false;
			toast.error('无效的销售退货ID');
			return;
		}
		loadDetail(id);
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-cyan-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">销售退货单详情</h1>
		</div>
		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>销售管理</li>
				<li><a class="text-primary" href="/sales/returns">销售退货单</a></li>
				<li>详情</li>
			</ul>
		</div>
	</div>

	<div class="flex flex-wrap items-center justify-between gap-3">
		<a href="/sales/returns" class="btn btn-ghost btn-sm gap-1">
			<ArrowLeft size={14} /> 返回列表
		</a>
		{#if detail}
			<div class="flex flex-wrap items-center gap-2">
				{#if detail.status === 'draft'}
					<button
						class="btn btn-sm btn-success text-white"
						onclick={handleConfirm}
						disabled={acting}
					>
						<CheckCircle size={16} /> 确认入库
					</button>
					<button class="btn btn-sm btn-error btn-outline" onclick={handleDelete} disabled={acting}>
						<Trash2 size={16} /> 删除
					</button>
				{/if}
				{#if detail.status === 'confirmed'}
					{#if detail.stock_in_id && detail.stock_in_no}
						<a
							href="/stock/in/{detail.stock_in_id}"
							class="btn btn-sm btn-outline gap-1"
						>
							<PackageOpen size={14} /> {detail.stock_in_no}
						</a>
					{/if}
					<button
						class="btn btn-sm btn-primary btn-outline gap-1"
						onclick={() => (showEdit = true)}
					>
						<Pencil size={14} /> 编辑
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
			未找到销售退货信息
		</div>
	{:else}
		<div class="bg-base-100 border-base-300 space-y-5 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="flex items-center gap-2 text-lg font-semibold">
					<FileText size={18} /> 单据信息
				</div>
				<div class="flex items-center gap-2">
					<span class="badge badge-lg {getStatusStyle(detail.status, 'sales_return').class}">
						{getStatusStyle(detail.status, 'sales_return').label}
					</span>
					<span class="text-base-content/70 font-mono">{detail.return_no}</span>
				</div>
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<FileText size={16} /> 退货单号
					</div>
					<div class="mt-1 font-mono text-base font-medium">{detail.return_no}</div>
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
						<Calendar size={16} /> 退货日期
					</div>
					<div class="mt-1 text-base">{formatDate(detail.return_date)}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">入库仓库</div>
					<div class="mt-1 text-base">{detail.warehouse_name || '-'}</div>
				</div>
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">退货金额</div>
					<div class="text-success mt-1 font-mono text-base font-semibold">
						¥{formatMoney(totalAmount())}
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
					<ClipboardList size={18} /> 退货明细
				</div>
				<div class="text-base-content/50 text-base">共 {detail.items?.length || 0} 行</div>
			</div>
			<div class="overflow-x-auto">
				<table class="table-zebra table w-full min-w-[860px] table-fixed text-base">
					<thead>
						<tr>
							<th class="w-[40%] min-w-[320px]">物料名称</th>
							<th>仓库</th>
							<th class="text-right">数量</th>
							<th>单位</th>
							<th class="text-right">成本单价</th>
							<th class="text-right">金额</th>
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
								<td>{item.warehouse_name || detail.warehouse_name || '-'}</td>
								<td class="text-right font-mono">{item.quantity ?? 0}</td>
								<td>{item.unit || '-'}</td>
								<td class="text-right font-mono">¥{formatMoney(item.unit_cost)}</td>
								<td class="text-success text-right font-mono font-semibold">
									¥{formatMoney(lineAmount(item))}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

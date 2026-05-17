<!--
功能：物料详情页面
创建时间：2026-05-09
创建人：GPT-5.3-Codex
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { formatDateInCn, formatDateTimeInCn } from '$lib/datetime';
	import { ArrowLeft, FileText, Calendar, BadgeCheck, ClipboardList, Pencil } from 'lucide-svelte';

	let loading = $state(true);
	let material = $state<any>(null);

	function text(v: any) {
		return v === null || v === undefined || v === '' ? '-' : String(v);
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			material = await api.get(`/base/materials/${id}`);
		} catch (err: any) {
			toast.error('加载详情失败: ' + (err?.message || err));
			material = null;
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		const id = Number(page.params.id);
		if (!id) return;
		loadDetail(id);
	});
</script>

<div class="space-y-6 text-base">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-violet-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">物料详情</h1>
		</div>

		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>基础数据</li>
				<li><a class="text-primary" href="/materials">物料管理</a></li>
				<li>详情</li>
			</ul>
		</div>
	</div>

	<div class="flex flex-wrap items-center justify-between gap-3">
		<div class="flex flex-wrap items-center gap-2">
			<a href="/materials" class="btn btn-ghost btn-sm gap-1">
				<ArrowLeft size={14} /> 返回列表
			</a>
			<a href={`/materials/${page.params.id}/edit`} class="btn btn-primary btn-sm gap-1">
				<Pencil size={14} /> 编辑
			</a>
		</div>
		{#if material}
			<div class="flex flex-wrap items-center gap-2">
				<span
					class="badge badge-lg {material.status === 'enabled' ? 'badge-success' : 'badge-error'}"
					>{text(material.status_name || material.status)}</span
				>
				<span class="text-base-content/70 font-mono">{text(material.material_code)}</span>
			</div>
		{/if}
	</div>

	{#if loading}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center text-lg"
		>
			正在加载...
		</div>
	{:else if !material}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center text-lg"
		>
			未找到物料信息
		</div>
	{:else}
		<div class="bg-base-100 border-base-300 space-y-5 rounded-2xl border p-5">
			<div class="flex items-center gap-2 text-lg font-semibold">
				<FileText size={18} /> 物料信息
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<FileText size={16} /> 物料编码
					</div>
					<div class="mt-1 font-mono text-base font-medium">{text(material.material_code)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-2">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<ClipboardList size={16} /> 物料名称
					</div>
					<div class="mt-1 text-base font-semibold break-words">
						{text(material.material_display_name || material.material_name)}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<BadgeCheck size={16} /> 状态
					</div>
					<div class="mt-1">
						<span
							class="badge badge-lg {material.status === 'enabled'
								? 'badge-success'
								: 'badge-error'}">{text(material.status_name || material.status)}</span
						>
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">物料分类</div>
					<div class="mt-1 text-base">{text(material.category_name)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">计量单位</div>
					<div class="mt-1 text-base">{text(material.unit_name || material.unit)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">安全库存</div>
					<div class="mt-1 text-base">{text(material.safety_stock)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">最大库存</div>
					<div class="mt-1 text-base">{text(material.max_stock)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 text-sm">独立编码</div>
					<div class="mt-1 text-base">{material.is_code ? '是' : '否'}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<Calendar size={16} /> 创建时间
					</div>
					<div class="mt-1 text-base">
						{material.created_at ? formatDateTimeInCn(material.created_at) : '-'}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/55 flex items-center gap-1 text-sm">
						<Calendar size={16} /> 更新时间
					</div>
					<div class="mt-1 text-base">
						{material.updated_at ? formatDateTimeInCn(material.updated_at) : '-'}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-4">
					<div class="text-base-content/55 text-sm">备注</div>
					<div class="mt-1 text-base break-words whitespace-pre-wrap">{text(material.remark)}</div>
				</div>
			</div>
		</div>

		<div class="bg-base-100 border-base-300 space-y-4 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="text-lg font-semibold">自定义属性</div>
				<div class="text-base-content/50 text-base">
					共 {material.custom_attributes?.length || 0} 项
				</div>
			</div>

			{#if !material.custom_attributes || material.custom_attributes.length === 0}
				<div class="text-base-content/50 py-10 text-center text-base">暂无自定义属性</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-zebra table w-full min-w-[600px] table-fixed text-base">
						<thead>
							<tr>
								<th class="w-20">#</th>
								<th class="min-w-[200px]">属性名</th>
								<th>属性值</th>
							</tr>
						</thead>
						<tbody>
							{#each material.custom_attributes as attr, idx}
								<tr>
									<td>{idx + 1}</td>
									<td class="font-medium">{text(attr?.attr_name)}</td>
									<td>{text(attr?.attr_value)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}
</div>

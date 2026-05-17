<!--
功能：物料管理列表页
创建时间：2026-05-09
创建人：GPT-5.3-Codex
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Search, FunnelX } from 'lucide-svelte';
	import { dgRowBtn } from '$lib/dgButtonClasses';

	let materials = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let currentPage = $state(1);
	const pageSize = 10;
	let categories = $state<any[]>([]);

	let filters = $state({
		material_code: '',
		material_name: '',
		category_id: '',
		unit: '',
		status: ''
	});

	let unitOptions = $state<Array<{ value: string; label: string }>>([]);

	const columns = [
		{ key: 'material_code', label: '物料编码', class: 'font-mono text-primary', width: '180px' },
		{ key: 'material_display_name', label: '名称', class: 'font-semibold', width: '360px' },
		{ key: 'category_name', label: '分类', width: '160px' },
		{ key: 'unit_name', label: '单位', width: '90px' },
		{ key: 'status_name', label: '状态', width: '90px' }
	];

	async function loadCategories() {
		try {
			const res: any = await api.get('/base/categories/tree');
			const flat: any[] = [];
			const walk = (nodes: any[], depth = 0) => {
				for (const n of nodes || []) {
					const indent = depth > 0 ? `${'　'.repeat(depth)}└ ` : '';
					flat.push({
						...n,
						display_name: `${indent}${n.category_code} ${n.category_name}`
					});
					if (n.children?.length) walk(n.children, depth + 1);
				}
			};
			walk(res || [], 0);
			categories = flat;
		} catch (err) {
			console.error(err);
		}
	}

	async function loadUnitOptions() {
		try {
			const res: any = await api.get('/system/dict/unit/data');
			const items = Array.isArray(res) ? res : [];
			unitOptions = items.map((d: any) => ({
				value: String(d.dict_value || ''),
				label: String(d.dict_label || d.dict_value || '')
			}));
		} catch (_err) {
			unitOptions = [
				{ value: 'kg', label: '千克(kg)' },
				{ value: 't', label: '吨(t)' },
				{ value: 'm', label: '米(m)' },
				{ value: 'm2', label: '平方米(m²)' },
				{ value: 'pcs', label: '件' }
			];
		}
	}

	async function loadMaterials(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const params = new URLSearchParams({
				page: String(page),
				page_size: String(pageSize)
			});
			if (filters.material_code.trim()) params.set('material_code', filters.material_code.trim());
			if (filters.material_name.trim()) params.set('material_name', filters.material_name.trim());
			if (filters.category_id) params.set('category_id', filters.category_id);
			if (filters.unit) params.set('unit', filters.unit);
			if (filters.status) params.set('status', filters.status);

			const res: any = await api.get(`/base/materials?${params.toString()}`);
			materials = res.list || [];
			total = Number(res.total || 0);
		} catch (err: any) {
			toast.error('加载物料失败: ' + (err?.message || err));
		} finally {
			loading = false;
		}
	}

	async function search() {
		currentPage = 1;
		await loadMaterials(1);
	}

	function resetFilters() {
		filters = { material_code: '', material_name: '', category_id: '', unit: '', status: '' };
		search();
	}

	onMount(async () => {
		await loadCategories();
		await loadUnitOptions();
		await loadMaterials(1);
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={materials}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadMaterials}
		onCreate={() => goto('/materials/create')}
		onRefresh={() => loadMaterials(currentPage)}
		showDefaultSearch={false}
		actionColumnWidth="180px"
	>
		{#snippet headerFilters()}
			<div class="flex flex-wrap items-center gap-2">
				<input
					class="input input-sm input-bordered w-40"
					placeholder="编码"
					bind:value={filters.material_code}
					onkeydown={(e) => e.key === 'Enter' && search()}
				/>
				<input
					class="input input-sm input-bordered w-48"
					placeholder="名称/属性"
					bind:value={filters.material_name}
					onkeydown={(e) => e.key === 'Enter' && search()}
				/>
				<div class="relative w-40">
					<select
						class="select select-sm select-bordered [&_option]:text-base-content w-full text-transparent"
						bind:value={filters.category_id}
					>
						<option value="">全部分类</option>
						{#each categories as c}
							<option value={String(c.id)}>{c.display_name}</option>
						{/each}
					</select>
					<span
						class="text-base-content pointer-events-none absolute inset-y-0 right-7 left-3 flex items-center truncate text-xs"
					>
						{#if !filters.category_id}
							<span class="text-base-content/60">全部分类</span>
						{:else}
							{categories.find((c) => String(c.id) === filters.category_id)?.category_name || ''}
						{/if}
					</span>
				</div>
				<select class="select select-sm select-bordered w-28" bind:value={filters.unit}>
					<option value="">全部单位</option>
					{#each unitOptions as unit}
						<option value={unit.value}>{unit.label}</option>
					{/each}
				</select>
				<select class="select select-sm select-bordered w-28" bind:value={filters.status}>
					<option value="">全部状态</option>
					<option value="enabled">可用</option>
					<option value="disabled">禁用</option>
				</select>
				<button type="button" class="btn btn-sm btn-primary gap-1" onclick={search}>
					<Search size={14} /> 查询
				</button>
				<button type="button" class="btn btn-sm btn-ghost gap-1" onclick={resetFilters}>
					<FunnelX size={14} /> 重置
				</button>
			</div>
		{/snippet}
		{#snippet cellRender(key, value, row)}
			{#if key === 'status_name'}
				<span class="badge badge-sm {row.status === 'enabled' ? 'badge-success' : 'badge-error'}"
					>{value || '-'}</span
				>
			{:else if key === 'material_display_name'}
				<span
					class="block w-[340px] truncate"
					title={row.material_display_name || row.material_name || ''}
				>
					{row.material_display_name || row.material_name || '-'}
				</span>
			{:else}
				<span class="block truncate" title={value || ''}>{value || '-'}</span>
			{/if}
		{/snippet}
		{#snippet rowActions(row)}
			<div class="flex items-center justify-center gap-1">
				<a href={`/materials/${row.id}`} class={dgRowBtn}>查看详情</a>
				<a href={`/materials/${row.id}/edit`} class={dgRowBtn}>编辑</a>
			</div>
		{/snippet}
	</DataGrid>
</div>

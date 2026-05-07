<!--
功能：materials页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { Package, Edit3, Trash2, ChevronRight, FolderTree, HelpCircle, X } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger } from '$lib/dgButtonClasses';

	let materials = $state([]);
	let categories = $state<any[]>([]);
	let attributeDefs = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showModal = $state(false);
	let editingId = $state<number | null>(null);
	let submitting = $state(false);
	let expandedCatIds = $state<Set<number>>(new Set());
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);

	let showHelpModal = $state('');
	let attrSearchTerm = $state('');

	let form = $state({
		category_id: 0,
		material_code: '',
		material_name: '',
		unit: '件',
		safety_stock: 0,
		max_stock: 0,
		is_code: true,
		custom_attributes: [] as any[],
		remark: '',
		status: 'enabled'
	});

	const columns = [
		{ key: 'material_code', label: '物料编码', class: 'font-mono text-primary', width: '180px' },
		{ key: 'material_name', label: '名称', class: 'font-bold', width: '260px' },
		{ key: 'category_name', label: '分类', width: '220px' },
		{ key: 'unit_name', label: '单位', width: '90px' },
		{ key: 'status_name', label: '状态', width: '90px' }
	];

	async function loadMaterials(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/base/materials?page=${page}&page_size=${pageSize}`);
			materials = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadCategories() {
		try {
			const res: any = await api.get('/base/categories/tree');
			categories = res || [];
			if (expandedCatIds.size === 0) {
				expandedCatIds = new Set(categories.map((c: any) => c.id));
			}
		} catch (err) {
			console.error(err);
		}
	}

	async function loadAttributeDefs() {
		try {
			const res: any = await api.get('/base/material-attributes?page=1&page_size=100');
			attributeDefs = res.list || [];
		} catch (err) {
			console.error(err);
		}
	}

	function toggleCatExpand(id: number) {
		const next = new Set(expandedCatIds);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		expandedCatIds = next;
	}

	function resetForm() {
		form = {
			category_id: 0,
			material_code: '',
			material_name: '',
			unit: '件',
			safety_stock: 0,
			max_stock: 0,
			is_code: true,
			custom_attributes: [],
			remark: '',
			status: 'enabled'
		};
		attrSearchTerm = '';
	}

	function openCreateModal() {
		editingId = null;
		resetForm();
		showModal = true;
	}

	function openEditModal(mat: any) {
		editingId = mat.id;
		form = {
			category_id: mat.category_id,
			material_code: mat.material_code,
			material_name: mat.material_name,
			unit: mat.unit || '件',
			safety_stock: mat.safety_stock || 0,
			max_stock: mat.max_stock || 0,
			is_code: mat.is_code || false,
			custom_attributes: (mat.custom_attributes || []).map((attr: any) => {
				const def = attributeDefs.find((a: any) => a.id === attr.attr_id);
				return {
					...attr,
					select_options: def?.select_options || ''
				};
			}),
			remark: mat.remark || '',
			status: mat.status || 'enabled'
		};
		attrSearchTerm = '';
		showModal = true;
	}

	function toggleAttribute(attr: any) {
		const index = form.custom_attributes.findIndex((a: any) => a.attr_id === attr.id);
		if (index > -1) {
			form.custom_attributes = form.custom_attributes.filter((a: any) => a.attr_id !== attr.id);
		} else {
			form.custom_attributes = [
				...form.custom_attributes,
				{
					attr_id: attr.id,
					attr_code: attr.attr_code,
					attr_name: attr.attr_name,
					attr_type: attr.attr_type,
					attr_unit: attr.attr_unit || '',
					attr_value: ''
				}
			];
		}
	}

	function isAttributeSelected(attrId: number) {
		return form.custom_attributes.some((a: any) => a.attr_id === attrId);
	}

	async function handleSubmit() {
		if (form.category_id === 0) {
			toast.warning('请选择物料分类');
			return;
		}
		submitting = true;
		try {
			const submitData = {
				...form,
				custom_attributes: form.custom_attributes.map((attr) => ({
					...attr,
					attr_value: String(attr.attr_value)
				}))
			};
			if (editingId) {
				await api.put(`/base/materials/${editingId}`, submitData);
			} else {
				await api.post('/base/materials', submitData);
			}
			showModal = false;
			loadMaterials(1);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function handleDelete(mat: any) {
		deleteTarget = mat;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/base/materials/${deleteTarget.id}`);
			showConfirm = false;
			loadMaterials(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	function getCategoryName(id: number, nodes: any[] = categories): string {
		for (const node of nodes) {
			if (node.id === id) return node.category_name;
			if (node.children && node.children.length > 0) {
				const found = getCategoryName(id, node.children);
				if (found) return found;
			}
		}
		return '';
	}

	onMount(() => {
		loadMaterials(1);
		loadCategories();
		loadAttributeDefs();
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
		onCreate={openCreateModal}
	>
		{#snippet cellRender(key, value, row)}
			{#if key === 'status_name'}
				<span class="badge badge-sm {row.status === 'enabled' ? 'badge-success' : 'badge-error'}"
					>{value || '-'}</span
				>
			{:else if key === 'material_name' || key === 'category_name'}
				<span class="block truncate" title={value || ''}>{value || '-'}</span>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(mat)}
			<div class="flex flex-wrap items-center justify-center gap-1">
				<button type="button" class={dgRowBtn} onclick={() => openEditModal(mat)}>
					<Edit3 size={15} /> 编辑
				</button>
				<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(mat)}>
					<Trash2 size={15} /> 删除
				</button>
			</div>
		{/snippet}
	</DataGrid>
</div>

<Modal
	bind:show={showModal}
	title={editingId ? '编辑物料' : '新增物料'}
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-5xl"
	fillBodyHeight={true}
>
	<div class="flex min-h-0 flex-1 flex-col gap-6 lg:flex-row lg:items-stretch">
		<div class="flex min-h-0 w-full min-w-0 flex-1 basis-0 flex-col space-y-5 overflow-y-auto">
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><FolderTree size={14} /> 物料分类</span
					></label
				>
				{#if form.category_id > 0}
					<div class="mb-2 flex items-center gap-2">
						<span class="badge badge-primary badge-lg gap-1">
							<FolderTree size={12} />
							{getCategoryName(form.category_id)}
						</span>
						<button class="btn btn-xs btn-ghost" onclick={() => (form.category_id = 0)}>清除</button
						>
					</div>
				{/if}
				<div
					class="bg-base-200/30 border-base-300 scrollbar-hide max-h-56 overflow-y-auto rounded-xl border p-2"
				>
					{#each categories as cat}
						{@render catOption(cat, 0)}
					{/each}
				</div>
			</div>

			<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
				<div class="form-control">
					<label class="label"
						><span class="label-text flex items-center gap-2"><Package size={14} /> 物料编码</span
						></label
					>
					<input
						type="text"
						bind:value={form.material_code}
						class="input input-bordered bg-base-200/50 w-full"
						placeholder={editingId ? '' : '留空自动生成'}
					/>
				</div>
				<div class="form-control">
					<label class="label"><span class="label-text">物料名称</span></label>
					<input
						type="text"
						bind:value={form.material_name}
						class="input input-bordered bg-base-200/50 w-full"
						placeholder="如 Q235B钢板（12mm/2000*6000）"
					/>
				</div>
			</div>

			<div class="grid min-w-0 grid-cols-1 gap-4 sm:grid-cols-3">
				<div class="form-control min-w-0">
					<label class="label"><span class="label-text">计量单位</span></label>
					<select bind:value={form.unit} class="select select-bordered bg-base-200/50 w-full">
						<option value="件">件</option>
						<option value="kg">kg</option>
						<option value="t">t (吨)</option>
						<option value="m">m (米)</option>
						<option value="m2">m² (平方米)</option>
						<option value="m3">m³ (立方米)</option>
						<option value="个">个</option>
						<option value="根">根</option>
						<option value="卷">卷</option>
						<option value="套">套</option>
					</select>
				</div>
				<div class="form-control min-w-0">
					<label class="label"><span class="label-text">安全库存</span></label>
					<input
						type="number"
						bind:value={form.safety_stock}
						class="input input-bordered bg-base-200/50 w-full"
						min="0"
						step="0.01"
					/>
				</div>
				<div class="form-control min-w-0">
					<label class="label"><span class="label-text">最大库存</span></label>
					<input
						type="number"
						bind:value={form.max_stock}
						class="input input-bordered bg-base-200/50 w-full"
						min="0"
						step="0.01"
					/>
				</div>
			</div>

			<div class="flex items-center gap-8 py-1">
				<label class="flex cursor-pointer items-center gap-2">
					<input
						type="checkbox"
						bind:checked={form.is_code}
						class="checkbox checkbox-sm checkbox-primary"
					/>
					<span class="label-text">独立编码</span>
					<button
						type="button"
						class="btn btn-ghost btn-xs btn-circle h-5 min-h-0 w-5 p-0 opacity-40 hover:opacity-100"
						onclick={() => (showHelpModal = 'is_code')}
					>
						<HelpCircle size={14} />
					</button>
				</label>
			</div>

			<div class="form-control">
				<label class="label"><span class="label-text">备注</span></label>
				<textarea
					bind:value={form.remark}
					class="textarea textarea-bordered bg-base-200/50 h-24"
					placeholder="补充说明"
				></textarea>
			</div>

			{#if editingId}
				<div class="form-control">
					<label class="label"><span class="label-text">状态</span></label>
					<select bind:value={form.status} class="select select-bordered bg-base-200/50 w-full">
						<option value="enabled">可用</option>
						<option value="disabled">禁用</option>
					</select>
				</div>
			{/if}
		</div>

		<div class="flex min-h-0 w-full min-w-0 flex-1 basis-0 flex-col">
			<div class="form-control flex min-h-0 flex-1 flex-col">
				<label class="label shrink-0"><span class="label-text">物料自定义属性</span></label>
				<div
					class="bg-base-200/30 border-base-300 flex min-h-0 flex-1 flex-col rounded-xl border p-3"
				>
					<div class="mb-2 flex min-h-[32px] shrink-0 flex-wrap gap-1.5">
						{#if form.custom_attributes.length === 0}
							<span class="text-base-content/30 py-1 text-xs">可不选，后续按需补充</span>
						{:else}
							{#each form.custom_attributes as attr}
								<span class="badge badge-sm badge-primary badge-outline gap-1">
									{attr.attr_name}{attr.attr_unit ? `(${attr.attr_unit})` : ''}
									<button
										type="button"
										class="hover:text-error"
										onclick={() => {
											form.custom_attributes = form.custom_attributes.filter(
												(a: any) => a.attr_id !== attr.attr_id
											);
										}}
									>
										<X size={10} />
									</button>
								</span>
							{/each}
						{/if}
					</div>
					<input
						type="text"
						bind:value={attrSearchTerm}
						placeholder="搜索属性..."
						class="input input-sm input-bordered bg-base-100/50 mb-2 w-full shrink-0"
					/>
					<div class="scrollbar-hide min-h-0 flex-1 space-y-0.5 overflow-y-auto">
						{#each attributeDefs.filter((a: any) => !attrSearchTerm || a.attr_name.includes(attrSearchTerm) || a.attr_code.includes(attrSearchTerm)) as attr}
							{@const selected = isAttributeSelected(attr.id)}
							<button
								type="button"
								class="flex w-full items-center gap-2 rounded px-2 py-2 text-left text-sm transition-colors {selected
									? 'bg-primary/10 text-primary'
									: 'hover:bg-base-200/50'}"
								onclick={() => toggleAttribute(attr)}
							>
								<input
									type="checkbox"
									checked={selected}
									class="checkbox checkbox-xs checkbox-primary"
									tabindex={-1}
								/>
								<span class="flex-1">{attr.attr_name}</span>
								<span class="text-base-content/50 font-mono text-xs">{attr.attr_code}</span>
								{#if attr.attr_unit}
									<span class="text-base-content/40 text-xs">({attr.attr_unit})</span>
								{/if}
							</button>
						{/each}
						{#if attributeDefs.length === 0}
							<div class="text-base-content/40 py-3 text-center text-xs">
								暂无属性定义，请先在物料属性页面添加
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>
	</div>
</Modal>

{#if showHelpModal}
	<div class="modal modal-open" onclick={() => (showHelpModal = '')}>
		<div class="modal-box max-w-sm" onclick={(e: Event) => e.stopPropagation()}>
			<h3 class="mb-3 text-lg font-bold">独立编码</h3>
			<p class="text-base-content/70 text-sm leading-relaxed">
				启用后，入库时系统将为该物料自动生成材料检号（独立编码），用于单件追溯。适用于受压元件、关键安全部件等需要独立跟踪的物料。
			</p>
			<div class="modal-action">
				<button class="btn btn-sm btn-primary" onclick={() => (showHelpModal = '')}>知道了</button>
			</div>
		</div>
	</div>
{/if}

{#snippet catOption(node: any, depth: number)}
	<div
		class="hover:bg-base-200 flex cursor-pointer items-center gap-1 rounded-lg px-2 py-1 transition-colors {form.category_id ===
		node.id
			? 'bg-primary/10 text-primary font-bold'
			: ''}"
		style="padding-left: {8 + depth * 20}px"
	>
		{#if node.children && node.children.length > 0}
			<button
				class="btn btn-xs btn-ghost btn-square h-4 min-h-0 w-4 p-0"
				onclick={(e: Event) => {
					e.stopPropagation();
					toggleCatExpand(node.id);
				}}
			>
				<ChevronRight
					size={10}
					class="transition-transform {expandedCatIds.has(node.id) ? 'rotate-90' : ''}"
				/>
			</button>
		{:else}
			<div class="w-3"></div>
		{/if}
		<span class="text-base-content/40 w-8 font-mono text-xs">{node.category_code}</span>
		<span class="flex-1 text-sm" onclick={() => (form.category_id = node.id)}
			>{node.category_name}</span
		>
	</div>
	{#if node.children && node.children.length > 0 && expandedCatIds.has(node.id)}
		{#each node.children as child}
			{@render catOption(child, depth + 1)}
		{/each}
	{/if}
{/snippet}

<ConfirmDialog
	bind:show={showConfirm}
	title="删除物料"
	message={`确定要删除物料「${deleteTarget?.material_name || ''}」吗？删除后该物料的库存及关联数据将无法恢复。`}
	onConfirm={confirmDelete}
/>

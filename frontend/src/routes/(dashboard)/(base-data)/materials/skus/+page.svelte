<!--
功能：SKU管理页面
创建时间：2026-04-19
创建人：wangcw
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Edit3, Trash2, Tag, Package, ArrowLeft } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger } from '$lib/dgButtonClasses';

	let skus = $state([]);
	let materials = $state<any[]>([]);
	let attributeDefs = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showModal = $state(false);
	let editingId = $state<number | null>(null);
	let submitting = $state(false);
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);

	// 筛选条件
	let filterMaterialId = $state(0);
	let filterSKUCode = $state('');
	let filterStatus = $state('');

	let form = $state({
		material_id: 0,
		sku_name: '',
		reference_price: 0,
		custom_attributes: [] as any[],
		remark: '',
		status: 'enabled'
	});

	// 当前选中物料的可用属性定义
	let availableAttrs = $state<any[]>([]);

	function normalizeSelectOptions(options: any): string[] {
		if (!options) return [];
		return String(options)
			.split(/\r?\n/)
			.map((s) => s.trim())
			.filter(Boolean);
	}

	function enrichAttrWithDef(attr: any) {
		const def = attributeDefs.find((a: any) => a.id === attr.attr_id);
		if (!def) return attr;
		return {
			...attr,
			is_required: def.is_required ?? attr.is_required ?? false,
			attr_type: def.attr_type || attr.attr_type,
			attr_unit: def.attr_unit ?? attr.attr_unit ?? '',
			select_options: def.select_options ?? attr.select_options ?? ''
		};
	}

	function syncFormAttrsWithDefs() {
		if (attributeDefs.length === 0) return;
		if (form.custom_attributes?.length) {
			form.custom_attributes = form.custom_attributes.map(enrichAttrWithDef);
		}
		if (availableAttrs?.length) {
			availableAttrs = availableAttrs.map(enrichAttrWithDef);
		}
	}

	const columns = [
		{ key: 'sku_code', label: 'SKU编码', class: 'font-mono text-primary', width: '14%' },
		{ key: 'material_name', label: '物料名称', width: '16%' },
		{ key: 'sku_name', label: 'SKU名称', class: 'font-semibold', width: '20%' },
		{ key: 'reference_price', label: '参考价格', class: 'font-mono text-success', width: '10%' },
		{ key: 'attr_summary', label: '属性详情', width: '22%' },
		{ key: 'status_name', label: '状态', width: '8%' }
	];

	async function loadSKUs(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			let url = `/base/skus?page=${page}&page_size=${pageSize}`;
			if (filterMaterialId) url += `&material_id=${filterMaterialId}`;
			if (filterSKUCode) url += `&sku_code=${encodeURIComponent(filterSKUCode)}`;
			if (filterStatus) url += `&status=${filterStatus}`;
			const res: any = await api.get(url);
			skus = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadMaterials() {
		try {
			// 只加载启用了SKU管理的物料
			const res: any = await api.get('/base/materials?page=1&page_size=100&status=enabled');
			materials = (res.list || []).filter((m: any) => m.sku_managed);
		} catch (err) {
			console.error(err);
		}
	}

	async function loadAttributeDefs() {
		try {
			const res: any = await api.get('/base/material-attributes?page=1&page_size=100');
			attributeDefs = res.list || [];
			syncFormAttrsWithDefs();
		} catch (err) {
			console.error(err);
		}
	}

	function resetForm() {
		form = {
			material_id: 0,
			sku_name: '',
			reference_price: 0,
			custom_attributes: [],
			remark: '',
			status: 'enabled'
		};
		availableAttrs = [];
	}

	function onMaterialChange() {
		// 根据选择的物料加载其已定义的可用属性
		const mat = materials.find((m: any) => m.id === form.material_id);
		if (mat && mat.custom_attributes && mat.custom_attributes.length > 0) {
			// 获取属性定义的详细信息（包括是否必填、单位、类型等）
			availableAttrs = mat.custom_attributes.map((attr: any) => {
				const def = attributeDefs.find((a: any) => a.id === attr.attr_id);
				return {
					...attr,
					is_required: def?.is_required || false,
					attr_type: def?.attr_type || attr.attr_type,
					attr_unit: def?.attr_unit ?? attr.attr_unit ?? '',
					select_options: def?.select_options || ''
				};
			});
			// 初始化已选属性
			form.custom_attributes = availableAttrs.map((attr: any) => ({
				attr_id: attr.attr_id,
				attr_code: attr.attr_code,
				attr_name: attr.attr_name,
				attr_type: attr.attr_type,
				attr_unit: attr.attr_unit || '',
				attr_value: '',
				is_required: attr.is_required,
				select_options: attr.select_options || ''
			}));
		} else {
			availableAttrs = [];
			form.custom_attributes = [];
		}
		generateName();
	}

	// 自动生成SKU名称
	function generateName() {
		if (editingId) return; // 编辑模式不自动覆盖

		const mat = materials.find((m: any) => m.id === form.material_id);
		if (!mat) {
			form.sku_name = '';
			return;
		}

		const attrParts = form.custom_attributes
			.filter((a) => a.attr_value)
			.map((a) => `${a.attr_value}${a.attr_unit || ''}`);

		if (attrParts.length > 0) {
			form.sku_name = `${mat.material_name} (${attrParts.join(' / ')})`;
		} else {
			form.sku_name = mat.material_name;
		}
	}

	// 监听属性值变化
	$effect(() => {
		if (!editingId && form.material_id > 0) {
			generateName();
		}
	});

	async function openCreateModal() {
		editingId = null;
		resetForm();
		if (materials.length === 0) {
			await loadMaterials();
		}
		if (attributeDefs.length === 0) {
			await loadAttributeDefs();
		}
		// 如果有筛选的物料，自动选中
		if (filterMaterialId) {
			form.material_id = filterMaterialId;
			onMaterialChange();
		}
		showModal = true;
	}

	function openEditModal(sku: any) {
		editingId = sku.id;
		form = {
			material_id: sku.material_id,
			sku_name: sku.sku_name || '',
			reference_price: Number(sku.reference_price || 0),
			custom_attributes: (sku.custom_attributes || []).map((attr: any) => enrichAttrWithDef(attr)),
			remark: sku.remark || '',
			status: sku.status || 'enabled'
		};
		// 加载可用属性
		const mat = materials.find((m: any) => m.id === sku.material_id);
		if (mat && mat.custom_attributes && mat.custom_attributes.length > 0) {
			availableAttrs = mat.custom_attributes.map((a: any) => enrichAttrWithDef(a));
		} else {
			availableAttrs = attributeDefs.map((a: any) => ({
				attr_id: a.id,
				attr_code: a.attr_code,
				attr_name: a.attr_name,
				attr_type: a.attr_type,
				attr_unit: a.attr_unit || '',
				select_options: a.select_options || ''
			}));
		}
		syncFormAttrsWithDefs();
		showModal = true;
	}

	async function handleSubmit() {
		if (form.material_id === 0) {
			toast.warning('请选择物料');
			return;
		}
		if (form.custom_attributes.length === 0) {
			toast.warning('请至少选择一个属性');
			return;
		}
		// 检查必填属性
		const missingRequired = form.custom_attributes.filter(
			(a: any) => a.is_required && !a.attr_value
		);
		if (missingRequired.length > 0) {
			toast.warning(`请填写必填属性：${missingRequired.map((a: any) => a.attr_name).join('、')}`);
			return;
		}

		submitting = true;
		try {
			const submitData = {
				...form,
				reference_price: Number(form.reference_price || 0),
				custom_attributes: form.custom_attributes.map((attr) => ({
					attr_id: attr.attr_id,
					attr_code: attr.attr_code,
					attr_name: attr.attr_name,
					attr_type: attr.attr_type,
					attr_unit: attr.attr_unit,
					attr_value: String(attr.attr_value)
				}))
			};
			if (editingId) {
				await api.put(`/base/skus/${editingId}`, submitData);
			} else {
				await api.post('/base/skus', submitData);
			}
			showModal = false;
			loadSKUs(1);
			toast.success(editingId ? 'SKU更新成功' : 'SKU创建成功');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	async function handleDelete(sku: any) {
		deleteTarget = sku;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/base/skus/${deleteTarget.id}`);
			showConfirm = false;
			loadSKUs(currentPage);
			toast.success('删除成功');
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	onMount(() => {
		// 检查URL参数中是否有material_id
		const urlParams = new URLSearchParams($page.url.search);
		const matId = urlParams.get('material_id');
		if (matId) {
			filterMaterialId = parseInt(matId);
		}
		loadMaterials();
		loadAttributeDefs();
		loadSKUs(1);
	});
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
	<!-- 筛选器 -->
	<div class="bg-base-100 border-base-300 flex items-center gap-4 rounded-2xl border p-4">
		<a href="/materials" class="btn btn-ghost btn-sm gap-1">
			<ArrowLeft size={14} /> 返回物料
		</a>
		<div class="divider divider-horizontal mx-0"></div>
		<select
			bind:value={filterMaterialId}
			onchange={() => loadSKUs(1)}
			class="select select-bordered select-sm bg-base-200/50 min-w-[200px]"
		>
			<option value={0}>全部物料</option>
			{#each materials as mat}
				<option value={mat.id}>{mat.material_code} - {mat.material_name}</option>
			{/each}
		</select>
		<input
			type="text"
			bind:value={filterSKUCode}
			oninput={() => loadSKUs(1)}
			placeholder="搜索SKU编码/名称..."
			class="input input-bordered input-sm bg-base-200/50 w-48"
		/>
		<select
			bind:value={filterStatus}
			onchange={() => loadSKUs(1)}
			class="select select-bordered select-sm bg-base-200/50"
		>
			<option value="">全部状态</option>
			<option value="enabled">启用</option>
			<option value="disabled">禁用</option>
		</select>
	</div>

	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={skus}
		{total}
		{loading}
		{pageSize}
		actionColumnWidth="160px"
		bind:page={currentPage}
		onPageChange={loadSKUs}
		onCreate={openCreateModal}
	>
		{#snippet cellRender(key, value, row)}
			{#if key === 'status_name'}
				<span class="badge badge-sm {row.status === 'enabled' ? 'badge-success' : 'badge-error'}"
					>{value || '-'}</span
				>
			{:else if key === 'reference_price'}
				<span class="text-success font-mono font-semibold">{Number(value || 0).toFixed(2)}</span>
			{:else if key === 'attr_summary'}
				<span class="text-base-content/70 block truncate" title={value || ''}>{value || '-'}</span>
			{:else if key === 'sku_code'}
				<div class="flex items-center gap-1.5">
					<Tag size={12} class="text-secondary" />
					<span class="text-primary font-mono">{value}</span>
				</div>
			{:else if key === 'material_name' || key === 'sku_name'}
				<span class="block truncate" title={value || ''}>{value || '-'}</span>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(sku)}
			<div class="flex flex-nowrap items-center justify-center gap-1">
				<button type="button" class={dgRowBtn} onclick={() => openEditModal(sku)}>
					<Edit3 size={15} /> 编辑
				</button>
				<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(sku)}>
					<Trash2 size={15} /> 删除
				</button>
			</div>
		{/snippet}
	</DataGrid>
</div>

<!-- 新增/编辑SKU弹窗 -->
<Modal
	bind:show={showModal}
	title={editingId ? '编辑SKU' : '新增SKU'}
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-5xl"
	bodyMaxHeight="max-h-[82vh]"
>
	<div class="space-y-4">
		<!-- 物料选择 -->
		<div class="form-control">
			<label class="label"
				><span class="label-text flex items-center gap-2"><Package size={14} /> 所属物料</span
				></label
			>
			{#if editingId}
				<input
					type="text"
					value={materials.find((m: any) => m.id === form.material_id)?.material_name || ''}
					class="input input-bordered bg-base-200/50 w-full"
					disabled
				/>
			{:else}
				<select
					bind:value={form.material_id}
					onchange={onMaterialChange}
					class="select select-bordered bg-base-200/50 w-full"
				>
					<option value={0}>-- 请选择物料 --</option>
					{#each materials as mat}
						<option value={mat.id}>
							{mat.material_code} - {mat.material_name}
						</option>
					{/each}
				</select>
				{#if materials.length === 0}
					<p class="text-warning mt-1 text-xs">⚠ 暂无启用SKU管理的物料，请先在物料管理中开启</p>
				{/if}
			{/if}
		</div>

		<!-- SKU名称 -->
		<div class="form-control">
			<label class="label"
				><span class="label-text"
					>SKU名称 <span class="text-base-content/40">(可留空,将由属性值自动生成)</span></span
				></label
			>
			<input
				type="text"
				bind:value={form.sku_name}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="如 12mm×2000×6000"
			/>
		</div>

		<div class="form-control">
			<label class="label"><span class="label-text">参考价格(元)</span></label>
			<input
				type="number"
				bind:value={form.reference_price}
				class="input input-bordered bg-base-200/50 w-full"
				min="0"
				step="0.01"
			/>
		</div>

		<!-- 属性填写 -->
		{#if form.material_id > 0}
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><Tag size={14} /> 规格属性详情</span
					></label
				>

				{#if availableAttrs.length > 0}
					<div class="bg-base-200/30 border-base-300 space-y-4 rounded-xl border p-4">
						<div class="grid grid-cols-2 gap-x-6 gap-y-4">
							{#each form.custom_attributes as attr, i}
								<div class="form-control">
									<label class="label py-1">
										<span class="label-text text-sm font-medium">
											{attr.attr_name}
											{#if attr.is_required}<span class="text-error ml-0.5">*</span>{/if}
											{#if attr.attr_unit}<span class="text-base-content/40 ml-1 text-xs"
													>({attr.attr_unit})</span
												>{/if}
										</span>
									</label>

									{#if attr.attr_type === 'select'}
										<select
											bind:value={form.custom_attributes[i].attr_value}
											required={attr.is_required}
											class="select select-bordered bg-base-100 focus:select-primary shadow-sm transition-all"
										>
											<option value="">-- 请选择 --</option>
											{#if attr.select_options}
												{#each normalizeSelectOptions(attr.select_options) as opt}
													<option value={opt}>{opt}</option>
												{/each}
											{/if}
										</select>
									{:else if attr.attr_type === 'number'}
										<input
											type="number"
											bind:value={form.custom_attributes[i].attr_value}
											required={attr.is_required}
											class="input input-bordered bg-base-100 focus:input-primary shadow-sm transition-all"
											step="any"
											placeholder="输入数值"
										/>
									{:else if attr.attr_type === 'date'}
										<input
											type="date"
											bind:value={form.custom_attributes[i].attr_value}
											required={attr.is_required}
											class="input input-bordered bg-base-100 focus:input-primary shadow-sm transition-all"
										/>
									{:else}
										<input
											type="text"
											bind:value={form.custom_attributes[i].attr_value}
											required={attr.is_required}
											class="input input-bordered bg-base-100 focus:input-primary shadow-sm transition-all"
											placeholder="输入文本内容"
										/>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<div
						class="text-base-content/50 bg-base-200/30 border-base-300 rounded-xl border border-dashed py-8 text-center"
					>
						<p>该物料尚未在物料管理中配置SKU属性</p>
						<a href="/materials" class="link link-primary mt-2 block text-xs"
							>前往物料管理配置属性</a
						>
					</div>
				{/if}
			</div>
		{/if}

		<!-- 备注 -->
		<div class="form-control">
			<label class="label"><span class="label-text">备注</span></label>
			<textarea
				bind:value={form.remark}
				class="textarea textarea-bordered bg-base-200/50 h-16"
				placeholder="补充说明"
			></textarea>
		</div>

		<!-- 编辑时显示状态 -->
		{#if editingId}
			<div class="form-control">
				<label class="label"><span class="label-text">状态</span></label>
				<select bind:value={form.status} class="select select-bordered bg-base-200/50 w-full">
					<option value="enabled">启用</option>
					<option value="disabled">禁用</option>
				</select>
			</div>
		{/if}
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除SKU"
	message={`确定要删除SKU「${deleteTarget?.sku_code || ''}」吗？此操作不可恢复。`}
	onConfirm={confirmDelete}
/>

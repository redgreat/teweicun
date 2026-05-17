<!--
功能：dicts页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import Modal from '$lib/components/Modal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { BookOpen, Plus, Edit3, Trash2, List } from 'lucide-svelte';

	let dictTypes = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let selectedType = $state<any>(null);
	let dictDataList = $state<any[]>([]);
	let dataLoading = $state(false);

	// Dict Type Modal
	let showTypeModal = $state(false);
	let editingTypeId = $state<number | null>(null);
	let submitting = $state(false);
	let typeForm = $state({ dict_type: '', dict_name: '', remark: '' });

	// Dict Data Modal
	let showDataModal = $state(false);
	let editingDataId = $state<number | null>(null);
	let dataForm = $state({
		dict_type: '',
		dict_label: '',
		dict_value: '',
		sort_order: 1,
		remark: ''
	});

	// Delete
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);
	let deleteKind = $state<'type' | 'data'>('type');

	async function loadDictTypes() {
		loading = true;
		try {
			const res: any = await api.get('/system/dict-types?page=1&page_size=100');
			dictTypes = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function selectType(item: any) {
		selectedType = item;
		await loadDictData(item.dict_type);
	}

	async function loadDictData(dictType: string) {
		dataLoading = true;
		try {
			const res: any = await api.get(`/system/dict/${dictType}/data`);
			dictDataList = res || [];
		} catch (err) {
			console.error(err);
		} finally {
			dataLoading = false;
		}
	}

	// --- Dict Type CRUD ---
	function openCreateType() {
		editingTypeId = null;
		typeForm = { dict_type: '', dict_name: '', remark: '' };
		showTypeModal = true;
	}

	function openEditType(item: any) {
		editingTypeId = item.id;
		typeForm = { dict_type: item.dict_type, dict_name: item.dict_name, remark: item.remark || '' };
		showTypeModal = true;
	}

	async function handleTypeSubmit() {
		submitting = true;
		try {
			if (editingTypeId) {
				await api.put(`/system/dict-types/${editingTypeId}`, typeForm);
			} else {
				await api.post('/system/dict-types', typeForm);
			}
			showTypeModal = false;
			loadDictTypes();
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	// --- Dict Data CRUD ---
	function openCreateData() {
		if (!selectedType) {
			toast.warning('请先选择一个字典类型');
			return;
		}
		editingDataId = null;
		dataForm = {
			dict_type: selectedType.dict_type,
			dict_label: '',
			dict_value: '',
			sort_order: dictDataList.length + 1,
			remark: ''
		};
		showDataModal = true;
	}

	function openEditData(item: any) {
		editingDataId = item.id;
		dataForm = {
			dict_type: item.dict_type,
			dict_label: item.dict_label,
			dict_value: item.dict_value,
			sort_order: item.sort_order,
			remark: item.remark || ''
		};
		showDataModal = true;
	}

	async function handleDataSubmit() {
		submitting = true;
		try {
			if (editingDataId) {
				await api.put(`/system/dict-data/${editingDataId}`, dataForm);
			} else {
				await api.post('/system/dict-data', dataForm);
			}
			showDataModal = false;
			if (selectedType) loadDictData(selectedType.dict_type);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	// --- Delete ---
	function handleDeleteType(item: any) {
		deleteTarget = item;
		deleteKind = 'type';
		showConfirm = true;
	}

	function handleDeleteData(item: any) {
		deleteTarget = item;
		deleteKind = 'data';
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			if (deleteKind === 'type') {
				await api.delete(`/system/dict-types/${deleteTarget.id}`);
				showConfirm = false;
				if (selectedType?.id === deleteTarget.id) {
					selectedType = null;
					dictDataList = [];
				}
				loadDictTypes();
			} else {
				await api.delete(`/system/dict-data/${deleteTarget.id}`);
				showConfirm = false;
				if (selectedType) loadDictData(selectedType.dict_type);
			}
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	onMount(loadDictTypes);
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-violet-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">数据字典管理</h1>
		</div>
		<div class="breadcrumbs text-sm opacity-60">
			<ul>
				<li>首页</li>
				<li>系统管理</li>
				<li>数据字典</li>
			</ul>
		</div>
	</div>

	<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
		<!-- Left: Dict Types -->
		<div
			class="bg-base-100 shadow-base-300/50 border-base-300 overflow-hidden rounded-3xl border shadow-xl"
		>
			<div class="border-base-200 bg-base-100/50 flex items-center justify-between border-b p-5">
				<div class="flex items-center gap-2">
					<BookOpen size={18} class="text-violet-500" />
					<h2 class="text-xl font-bold">字典类型</h2>
					<span class="badge badge-sm badge-ghost">{total}</span>
				</div>
				<button class="btn btn-sm btn-primary gap-1 rounded-lg shadow-sm" onclick={openCreateType}>
					<Plus size={15} /> 新增
				</button>
			</div>

			<div class="max-h-[65vh] overflow-y-auto">
				{#if loading}
					<div class="p-8 text-center">
						<span class="loading loading-spinner loading-md"></span>
					</div>
				{:else if dictTypes.length === 0}
					<div class="text-base-content/30 p-8 text-center italic">暂无字典类型</div>
				{:else}
					<div class="divide-base-200 divide-y">
						{#each dictTypes as item}
							<div
								class="hover:bg-base-200/50 w-full cursor-pointer px-5 py-3.5 text-left transition-colors {selectedType?.id ===
								item.id
									? 'border-l-4 border-l-violet-500 bg-violet-500/10'
									: ''}"
								role="button"
								tabindex="0"
								onclick={() => selectType(item)}
								onkeydown={(e) => {
									if (e.key === 'Enter' || e.key === ' ') {
										e.preventDefault();
										selectType(item);
									}
								}}
							>
								<div class="flex items-center justify-between">
									<div>
										<p class="text-base font-bold">{item.dict_name}</p>
										<p class="text-base-content/50 font-mono text-sm">{item.dict_type}</p>
									</div>
									<div class="flex items-center gap-1">
										<button
											class="btn btn-sm btn-ghost text-primary hover:bg-primary/10 rounded-md"
											onclick={(e) => {
												e.stopPropagation();
												openEditType(item);
											}}
										>
											<Edit3 size={15} />
										</button>
										<button
											class="btn btn-sm btn-ghost text-error hover:bg-error/10 rounded-md"
											onclick={(e) => {
												e.stopPropagation();
												handleDeleteType(item);
											}}
										>
											<Trash2 size={15} />
										</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Right: Dict Data -->
		<div
			class="bg-base-100 shadow-base-300/50 border-base-300 overflow-hidden rounded-3xl border shadow-xl lg:col-span-2"
		>
			<div class="border-base-200 bg-base-100/50 flex items-center justify-between border-b p-5">
				<div class="flex items-center gap-2">
					<List size={18} class="text-violet-500" />
					{#if selectedType}
						<h2 class="text-xl font-bold">{selectedType.dict_name}</h2>
						<span class="badge badge-sm badge-ghost font-mono">{selectedType.dict_type}</span>
					{:else}
						<h2 class="text-base-content/40 text-xl font-bold">字典数据</h2>
					{/if}
				</div>
				{#if selectedType}
					<button
						class="btn btn-sm btn-primary gap-1 rounded-lg shadow-sm"
						onclick={openCreateData}
					>
						<Plus size={15} /> 新增
					</button>
				{/if}
			</div>

			{#if !selectedType}
				<div class="text-base-content/30 p-16 text-center italic">
					<BookOpen size={48} class="mx-auto mb-4 opacity-30" />
					<p>请从左侧选择一个字典类型</p>
				</div>
			{:else if dataLoading}
				<div class="p-8 text-center"><span class="loading loading-spinner loading-md"></span></div>
			{:else if dictDataList.length === 0}
				<div class="text-base-content/30 p-12 text-center italic">暂无字典数据</div>
			{:else}
				<table class="table-md table w-full text-[15px]">
					<thead class="bg-base-200/50">
						<tr>
							<th class="text-base-content/70 text-[16px] font-bold">排序</th>
							<th class="text-base-content/70 text-[16px] font-bold">显示标签</th>
							<th class="text-base-content/70 text-[16px] font-bold">数据值</th>
							<th class="text-base-content/70 text-[16px] font-bold">备注</th>
							<th class="w-20 text-center text-[16px]">操作</th>
						</tr>
					</thead>
					<tbody class="divide-base-200 divide-y">
						{#each dictDataList as item}
							<tr class="hover:bg-base-200/50 transition-colors">
								<td class="text-base-content/50 font-mono text-base">{item.sort_order}</td>
								<td class="font-medium">{item.dict_label}</td>
								<td class="text-primary font-mono text-base">{item.dict_value}</td>
								<td class="text-base-content/50 text-base">{item.remark || '-'}</td>
								<td class="text-center">
									<div class="flex items-center justify-center gap-1">
										<button
											class="btn btn-sm btn-ghost text-primary hover:bg-primary/10 rounded-md"
											onclick={() => openEditData(item)}
										>
											<Edit3 size={15} />
										</button>
										<button
											class="btn btn-sm btn-ghost text-error hover:bg-error/10 rounded-md"
											onclick={() => handleDeleteData(item)}
										>
											<Trash2 size={15} />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	</div>
</div>

<!-- Dict Type Modal -->
<Modal
	bind:show={showTypeModal}
	title={editingTypeId ? '编辑字典类型' : '新增字典类型'}
	onConfirm={handleTypeSubmit}
	loading={submitting}
>
	<div class="space-y-4">
		<div class="form-control">
			<label class="label" for="dict-type-code"><span class="label-text">类型编码</span></label>
			<input
				id="dict-type-code"
				type="text"
				bind:value={typeForm.dict_type}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="如 supplier_type"
			/>
		</div>
		<div class="form-control">
			<label class="label" for="dict-type-name"><span class="label-text">类型名称</span></label>
			<input
				id="dict-type-name"
				type="text"
				bind:value={typeForm.dict_name}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="如 供应商类型"
			/>
		</div>
		<div class="form-control">
			<label class="label" for="dict-type-remark"><span class="label-text">备注</span></label>
			<input
				id="dict-type-remark"
				type="text"
				bind:value={typeForm.remark}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="备注说明"
			/>
		</div>
	</div>
</Modal>

<!-- Dict Data Modal -->
<Modal
	bind:show={showDataModal}
	title={editingDataId ? '编辑字典数据' : '新增字典数据'}
	onConfirm={handleDataSubmit}
	loading={submitting}
>
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label" for="dict-data-label"><span class="label-text">显示标签</span></label>
				<input
					id="dict-data-label"
					type="text"
					bind:value={dataForm.dict_label}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="如 物料供应商"
				/>
			</div>
			<div class="form-control">
				<label class="label" for="dict-data-value"><span class="label-text">数据值</span></label>
				<input
					id="dict-data-value"
					type="text"
					bind:value={dataForm.dict_value}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="如 material"
				/>
			</div>
		</div>
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label" for="dict-data-sort-order"
					><span class="label-text">排序号</span></label
				>
				<input
					id="dict-data-sort-order"
					type="number"
					bind:value={dataForm.sort_order}
					class="input input-bordered bg-base-200/50 w-full"
					min="1"
				/>
			</div>
			<div class="form-control">
				<label class="label" for="dict-data-remark"><span class="label-text">备注</span></label>
				<input
					id="dict-data-remark"
					type="text"
					bind:value={dataForm.remark}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="备注"
				/>
			</div>
		</div>
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title={deleteKind === 'type' ? '删除字典类型' : '删除字典数据'}
	message={deleteKind === 'type'
		? `确定要删除字典类型「${deleteTarget?.dict_name || ''}」吗？其下所有字典数据也将被删除。`
		: `确定要删除字典数据「${deleteTarget?.dict_label || ''}」吗？`}
	onConfirm={confirmDelete}
/>

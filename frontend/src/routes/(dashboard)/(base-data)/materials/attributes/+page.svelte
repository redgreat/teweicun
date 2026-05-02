<!--
功能：attributes页面
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
	import { Edit3, Trash2 } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger } from '$lib/dgButtonClasses';

	let attributes = $state([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showModal = $state(false);
	let editingId = $state<number | null>(null);
	let submitting = $state(false);
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);

	let form = $state({
		attr_code: '',
		attr_name: '',
		attr_type: 'text',
		attr_unit: '',
		select_options: '',
		is_required: false,
		sort_order: 0,
		remark: '',
		status: 'enabled'
	});

	const columns = [
		{ key: 'attr_code', label: '属性编码', class: 'font-mono text-primary', width: '12%' },
		{ key: 'attr_name', label: '属性名称', class: 'font-bold', width: '18%' },
		{ key: 'attr_type', label: '属性类型', width: '10%' },
		{ key: 'attr_unit', label: '单位', width: '8%' },
		{ key: 'is_required', label: '必填', class: 'text-center', width: '8%' },
		{ key: 'sort_order', label: '排序', class: 'text-center', width: '8%' },
		{ key: 'status', label: '状态', width: '8%' }
	];

	const typeMap: Record<string, string> = {
		text: '文本',
		number: '数字',
		select: '下拉选择',
		date: '日期'
	};

	async function loadAttributes(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(
				`/base/material-attributes?page=${page}&page_size=${pageSize}`
			);
			attributes = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		form = {
			attr_code: '',
			attr_name: '',
			attr_type: 'text',
			attr_unit: '',
			select_options: '',
			is_required: false,
			sort_order: 0,
			remark: '',
			status: 'enabled'
		};
	}

	function openCreateModal() {
		editingId = null;
		resetForm();
		showModal = true;
	}

	function openEditModal(attr: any) {
		editingId = attr.id;
		form = {
			attr_code: attr.attr_code,
			attr_name: attr.attr_name,
			attr_type: attr.attr_type || 'text',
			attr_unit: attr.attr_unit || '',
			select_options: attr.select_options || '',
			is_required: attr.is_required || false,
			sort_order: attr.sort_order || 0,
			remark: attr.remark || '',
			status: attr.status || 'enabled'
		};
		showModal = true;
	}

	async function handleSubmit() {
		submitting = true;
		try {
			if (editingId) {
				await api.put(`/base/material-attributes/${editingId}`, form);
				toast.success('更新成功');
			} else {
				await api.post('/base/material-attributes', form);
				toast.success('创建成功');
			}
			showModal = false;
			loadAttributes(1);
		} catch (err: any) {
			toast.error(err.message || '操作失败');
		} finally {
			submitting = false;
		}
	}

	function openDeleteConfirm(attr: any) {
		deleteTarget = attr;
		showConfirm = true;
	}

	async function handleDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/base/material-attributes/${deleteTarget.id}`);
			toast.success('删除成功');
			showConfirm = false;
			deleteTarget = null;
			loadAttributes(currentPage);
		} catch (err: any) {
			toast.error(err.message || '删除失败');
		}
	}

	function renderCell(row: any, key: string) {
		if (key === 'attr_type') {
			return typeMap[row[key]] || row[key];
		}
		if (key === 'is_required') {
			return row[key] ? '是' : '否';
		}
		if (key === 'status') {
			return row[key] === 'enabled' ? '启用' : '禁用';
		}
		return row[key] || '-';
	}

	onMount(() => {
		loadAttributes(1);
	});
</script>

<svelte:head>
	<title>物料属性管理 - 特维存</title>
</svelte:head>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={attributes}
		{total}
		{loading}
		{pageSize}
		actionColumnWidth="160px"
		bind:page={currentPage}
		onPageChange={loadAttributes}
		onCreate={openCreateModal}
	>
		{#snippet cellRender(key, value, row)}
			{#if key === 'status'}
				<span class="badge badge-sm {row.status === 'enabled' ? 'badge-success' : 'badge-error'}"
					>{renderCell(row, key)}</span
				>
			{:else if key === 'is_required'}
				<span class="badge badge-sm {value ? 'badge-warning' : 'badge-ghost'}"
					>{renderCell(row, key)}</span
				>
			{:else if key === 'attr_type'}
				<span class="badge badge-sm badge-info">{renderCell(row, key)}</span>
			{:else}
				<span class="block max-w-full truncate" title={value || ''}>{value || '-'}</span>
			{/if}
		{/snippet}
		{#snippet rowActions(attr)}
			<div class="flex flex-nowrap items-center justify-center gap-1">
				<button type="button" class={dgRowBtn} onclick={() => openEditModal(attr)}>
					<Edit3 size={15} /> 编辑
				</button>
				<button type="button" class={dgRowBtnDanger} onclick={() => openDeleteConfirm(attr)}>
					<Trash2 size={15} /> 删除
				</button>
			</div>
		{/snippet}
	</DataGrid>
</div>

<Modal
	bind:show={showModal}
	title={editingId ? '编辑属性' : '新增属性'}
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-2xl"
>
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text">属性编码</span></label>
				{#if editingId}
					<input
						type="text"
						bind:value={form.attr_code}
						class="input input-bordered bg-base-200/50 w-full"
						disabled
					/>
				{:else}
					<input
						type="text"
						value="系统自动生成"
						class="input input-bordered bg-base-200/50 w-full"
						disabled
					/>
				{/if}
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">属性名称</span></label>
				<input
					type="text"
					bind:value={form.attr_name}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="如 长度"
				/>
			</div>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text">属性类型</span></label>
				<select bind:value={form.attr_type} class="select select-bordered bg-base-200/50 w-full">
					<option value="text">文本</option>
					<option value="number">数字</option>
					<option value="select">下拉选择</option>
					<option value="date">日期</option>
				</select>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">单位</span></label>
				<select bind:value={form.attr_unit} class="select select-bordered bg-base-200/50 w-full">
					<option value="">无单位</option>
					<option value="mm">mm (毫米)</option>
					<option value="cm">cm (厘米)</option>
					<option value="m">m (米)</option>
					<option value="kg">kg (千克)</option>
					<option value="t">t (吨)</option>
					<option value="℃">℃ (摄氏度)</option>
					<option value="MPa">MPa (兆帕)</option>
					<option value="%">% (百分比)</option>
					<option value="个">个</option>
					<option value="件">件</option>
					<option value="根">根</option>
				</select>
			</div>
		</div>

		{#if form.attr_type === 'select'}
			<div class="form-control">
				<label class="label"><span class="label-text">下拉选项值</span></label>
				<textarea
					bind:value={form.select_options}
					class="textarea textarea-bordered bg-base-200/50 h-20"
					placeholder="每行一个选项值，如：&#10;优质&#10;合格&#10;不合格"
				></textarea>
				<label class="label"
					><span class="label-text-alt text-base-content/50">每行输入一个选项值</span></label
				>
			</div>
		{/if}

		<div class="grid grid-cols-3 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text">是否必填</span></label>
				<select bind:value={form.is_required} class="select select-bordered bg-base-200/50 w-full">
					<option value={false}>否</option>
					<option value={true}>是</option>
				</select>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">排序</span></label>
				<input
					type="number"
					bind:value={form.sort_order}
					class="input input-bordered bg-base-200/50 w-full"
					min="0"
				/>
			</div>
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

		<div class="form-control">
			<label class="label"><span class="label-text">备注</span></label>
			<textarea
				bind:value={form.remark}
				class="textarea textarea-bordered bg-base-200/50 h-16"
				placeholder="补充说明"
			></textarea>
		</div>
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除属性"
	message={`确定要删除属性「${deleteTarget?.attr_name || ''}」吗？删除后该属性将无法恢复。`}
	onConfirm={handleDelete}
/>

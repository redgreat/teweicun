<!--
功能：warehouses页面
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

	// 仓库列表
	let warehouses = $state<any[]>([]);
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
		warehouse_code: '',
		warehouse_name: '',
		warehouse_type: '',
		manager_id: 0,
		status: 'enabled'
	});

	// 字典选项
	let warehouseTypeOptions = $state<{ value: string; label: string }[]>([]);
	let managerOptions = $state<{ value: number; label: string }[]>([]);

	const columns = [
		{ key: 'warehouse_code', label: '仓库编码', class: 'font-mono text-primary', width: '12%' },
		{ key: 'warehouse_name', label: '仓库名称', class: 'font-bold', width: '24%' },
		{ key: 'warehouse_type_name', label: '仓库类型', width: '18%' },
		{ key: 'manager_name', label: '负责人', width: '12%' },
		{ key: 'status_name', label: '状态', width: '8%' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/base/warehouses?page=${page}&page_size=${pageSize}`);
			warehouses = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadDictData() {
		try {
			const typeRes: any = await api.get('/system/dict/warehouse_type/data');
			warehouseTypeOptions = (typeRes || []).map((d: any) => ({
				value: d.dict_value,
				label: d.dict_label
			}));
		} catch (err) {
			console.error('load warehouse_type dict failed', err);
		}

		try {
			const userRes: any = await api.get('/system/users?page=1&page_size=100');
			const userList = userRes?.list || [];
			managerOptions = userList.map((u: any) => ({
				value: u.id,
				label: u.real_name || u.username
			}));
		} catch (err) {
			console.error('load users failed', err);
		}
	}

	// 仓库 CRUD
	function resetForm() {
		form = {
			warehouse_code: '',
			warehouse_name: '',
			warehouse_type: '',
			manager_id: 0,
			status: 'enabled'
		};
	}

	function openCreateModal() {
		editingId = null;
		resetForm();
		showModal = true;
	}

	function openEditModal(wh: any) {
		editingId = wh.id;
		form = {
			warehouse_code: wh.warehouse_code || '',
			warehouse_name: wh.warehouse_name || '',
			warehouse_type: wh.warehouse_type || '',
			manager_id: wh.manager_id || 0,
			status: wh.status || 'enabled'
		};
		showModal = true;
	}

	async function handleSubmit() {
		submitting = true;
		try {
			if (editingId) {
				await api.put(`/base/warehouses/${editingId}`, form);
			} else {
				await api.post('/base/warehouses', form);
			}
			showModal = false;
			loadData(1);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	async function handleDelete(wh: any) {
		deleteTarget = wh;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/base/warehouses/${deleteTarget.id}`);
			showConfirm = false;
			loadData(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	onMount(() => {
		loadData(1);
		loadDictData();
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<div class="min-h-0 flex-1">
		<DataGrid
			class="h-full min-h-0"
			{columns}
			data={warehouses}
			{total}
			{loading}
			{pageSize}
			actionColumnWidth="160px"
			bind:page={currentPage}
			onPageChange={loadData}
			onCreate={openCreateModal}
		>
			{#snippet cellRender(key, value, row)}
				{#if key === 'status_name'}
					<span class="badge badge-sm {row.status === 'enabled' ? 'badge-success' : 'badge-error'}"
						>{value || '-'}</span
					>
				{:else if key === 'warehouse_name' || key === 'warehouse_type_name' || key === 'manager_name'}
					<span class="block max-w-full truncate" title={value || ''}>{value || '-'}</span>
				{:else}
					{value || '-'}
				{/if}
			{/snippet}
			{#snippet rowActions(wh)}
				<div class="flex flex-nowrap items-center justify-center gap-1">
					<button type="button" class={dgRowBtn} onclick={() => openEditModal(wh)}>
						<Edit3 size={15} /> 编辑
					</button>
					<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(wh)}>
						<Trash2 size={15} /> 删除
					</button>
				</div>
			{/snippet}
		</DataGrid>
	</div>
</div>

<!-- 仓库新增/编辑弹窗 -->
<Modal
	bind:show={showModal}
	title={editingId ? '编辑仓库' : '新增仓库'}
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-2xl"
>
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text font-bold">仓库编码</span></label>
				{#if editingId}
					<input
						type="text"
						bind:value={form.warehouse_code}
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
				<label class="label"><span class="label-text font-bold">仓库名称</span></label>
				<input
					type="text"
					bind:value={form.warehouse_name}
					class="input input-bordered w-full"
					placeholder="仓库名称"
					required
				/>
			</div>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text font-bold">仓库类型</span></label>
				<select bind:value={form.warehouse_type} class="select select-bordered w-full" required>
					<option value="" disabled selected>选择类型</option>
					{#each warehouseTypeOptions as opt}
						<option value={opt.value}>{opt.label}</option>
					{/each}
				</select>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text font-bold">负责人</span></label>
				<select bind:value={form.manager_id} class="select select-bordered w-full" required>
					<option value={0} disabled selected>选择负责人</option>
					{#each managerOptions as opt}
						<option value={opt.value}>{opt.label}</option>
					{/each}
				</select>
			</div>
		</div>

		{#if editingId}
			<div class="form-control">
				<label class="label"><span class="label-text font-bold">状态</span></label>
				<select bind:value={form.status} class="select select-bordered w-full">
					<option value="enabled">可用</option>
					<option value="disabled">禁用</option>
				</select>
			</div>
		{/if}
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除仓库"
	message={`确定要删除仓库「${deleteTarget?.warehouse_name || ''}」吗？删除后该仓库的所有关联数据将无法恢复。`}
	onConfirm={confirmDelete}
/>

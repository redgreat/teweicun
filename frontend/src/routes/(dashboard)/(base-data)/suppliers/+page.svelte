<!--
功能：suppliers页面
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
	import { Building2, Phone, User, MapPin, Edit3, Trash2 } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger } from '$lib/dgButtonClasses';

	let suppliers = $state([]);
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
		supplier_code: '',
		supplier_name: '',
		supplier_type: '',
		contact_person: '',
		contact_phone: '',
		address: '',
		status: 'enabled'
	});

	let supplierTypeOptions = $state<{ value: string; label: string }[]>([]);

	const columns = [
		{ key: 'supplier_code', label: '供应商编码', class: 'font-mono text-primary', width: '12%' },
		{ key: 'supplier_name', label: '企业名称', class: 'font-bold', width: '25%' },
		{ key: 'supplier_type_name', label: '供应商类型', width: '12%' },
		{ key: 'contact_person', label: '联系人', width: '10%' },
		{ key: 'contact_phone', label: '联系电话', width: '15%' },
		{ key: 'status_name', label: '状态', width: '8%' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/base/suppliers?page=${page}&page_size=${pageSize}`);
			suppliers = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadDictData() {
		try {
			const res: any = await api.get('/system/dict/supplier_type/data');
			const items = res || [];
			supplierTypeOptions = items.map((d: any) => ({ value: d.dict_value, label: d.dict_label }));
		} catch (_err) {
			// fallback
			supplierTypeOptions = [
				{ value: 'material', label: '物料供应商' },
				{ value: 'equipment', label: '设备供应商' },
				{ value: 'service', label: '服务供应商' },
				{ value: 'other', label: '其他' }
			];
		}
	}

	async function handleSubmit() {
		submitting = true;
		try {
			if (editingId) {
				await api.put(`/base/suppliers/${editingId}`, form);
			} else {
				await api.post('/base/suppliers', form);
			}
			showModal = false;
			loadData(1);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function resetForm() {
		form = {
			supplier_code: '',
			supplier_name: '',
			supplier_type: '',
			contact_person: '',
			contact_phone: '',
			address: '',
			status: 'enabled'
		};
	}

	function openCreateModal() {
		editingId = null;
		resetForm();
		showModal = true;
	}

	function openEditModal(sup: any) {
		editingId = sup.id;
		form = {
			supplier_code: sup.supplier_code || '',
			supplier_name: sup.supplier_name,
			supplier_type: sup.supplier_type || '',
			contact_person: sup.contact_person || '',
			contact_phone: sup.contact_phone || '',
			address: sup.address || '',
			status: sup.status || 'enabled'
		};
		showModal = true;
	}

	async function handleDelete(sup: any) {
		deleteTarget = sup;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/base/suppliers/${deleteTarget.id}`);
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
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={suppliers}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		onCreate={openCreateModal}
	>
		{#snippet cellRender(key, value, row)}
			{#if key === 'status_name'}
				<span class="badge badge-sm {row.status === 'enabled' ? 'badge-success' : 'badge-error'}"
					>{value || '-'}</span
				>
			{:else}
				{value || '-'}
			{/if}
		{/snippet}
		{#snippet rowActions(sup)}
			<div class="flex flex-wrap items-center justify-center gap-1">
				<button type="button" class={dgRowBtn} onclick={() => openEditModal(sup)}>
					<Edit3 size={15} /> 编辑
				</button>
				<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(sup)}>
					<Trash2 size={15} /> 删除
				</button>
			</div>
		{/snippet}
	</DataGrid>
</div>

<Modal
	bind:show={showModal}
	title={editingId ? '编辑供应商' : '新增供应商'}
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-5xl"
>
	<div class="space-y-6">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><Building2 size={14} /> 供应商编码</span
					></label
				>
				{#if editingId}
					<input
						type="text"
						bind:value={form.supplier_code}
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
				<label class="label"
					><span class="label-text flex items-center gap-2"><Building2 size={14} /> 供应商类型</span
					></label
				>
				<select
					bind:value={form.supplier_type}
					class="select select-bordered bg-base-200/50 w-full"
					required
				>
					<option value="" disabled selected>选择类型</option>
					{#each supplierTypeOptions as opt}
						<option value={opt.value}>{opt.label}</option>
					{/each}
				</select>
			</div>
		</div>

		<div class="form-control">
			<label class="label"
				><span class="label-text flex items-center gap-2"><Building2 size={14} /> 企业名称</span
				></label
			>
			<input
				type="text"
				bind:value={form.supplier_name}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="全称"
			/>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><User size={14} /> 联系人</span></label
				>
				<input
					type="text"
					bind:value={form.contact_person}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="姓名"
				/>
			</div>
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><Phone size={14} /> 联系电话</span
					></label
				>
				<input
					type="text"
					bind:value={form.contact_phone}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="手机/座机"
				/>
			</div>
		</div>

		<div class="form-control">
			<label class="label"
				><span class="label-text flex items-center gap-2"><MapPin size={14} /> 企业地址</span
				></label
			>
			<textarea
				bind:value={form.address}
				class="textarea textarea-bordered bg-base-200/50 h-28"
				placeholder="详细地址"
			></textarea>
		</div>

		{#if editingId}
			<div class="form-control">
				<label class="label" for="supplier-status"><span class="label-text">状态</span></label>
				<select
					id="supplier-status"
					bind:value={form.status}
					class="select select-bordered bg-base-200/50 w-full"
				>
					<option value="enabled">可用</option>
					<option value="disabled">禁用</option>
				</select>
			</div>
		{/if}
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除供应商"
	message={`确定要删除供应商「${deleteTarget?.supplier_name || ''}」吗？删除后该供应商的所有关联数据将无法恢复。`}
	onConfirm={confirmDelete}
/>

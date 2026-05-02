<!--
功能：customers页面
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
	import { Users, Phone, User, Landmark, Edit3, Trash2 } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger } from '$lib/dgButtonClasses';

	let customers = $state([]);
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
		customer_code: '',
		customer_name: '',
		contact_person: '',
		contact_phone: '',
		address: '',
		credit_code: '',
		status: 'enabled'
	});

	const columns = [
		{ key: 'customer_code', label: '客户编号', class: 'font-mono text-emerald-600', width: '12%' },
		{ key: 'customer_name', label: '客户名称', class: 'font-bold', width: '25%' },
		{ key: 'contact_person', label: '主要联系人', width: '15%' },
		{ key: 'contact_phone', label: '联系方式', width: '15%' },
		{ key: 'status_name', label: '信用状态', width: '10%' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/base/customers?page=${page}&page_size=${pageSize}`);
			customers = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		form = {
			customer_code: '',
			customer_name: '',
			contact_person: '',
			contact_phone: '',
			address: '',
			credit_code: '',
			status: 'enabled'
		};
	}

	function openCreateModal() {
		editingId = null;
		resetForm();
		showModal = true;
	}

	function openEditModal(cus: any) {
		editingId = cus.id;
		form = {
			customer_code: cus.customer_code,
			customer_name: cus.customer_name,
			contact_person: cus.contact_person || '',
			contact_phone: cus.contact_phone || '',
			address: cus.address || '',
			credit_code: cus.credit_code || '',
			status: cus.status || 'enabled'
		};
		showModal = true;
	}

	async function handleSubmit() {
		submitting = true;
		try {
			if (editingId) {
				await api.put(`/base/customers/${editingId}`, form);
			} else {
				await api.post('/base/customers', form);
			}
			showModal = false;
			loadData(1);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	async function handleDelete(cus: any) {
		deleteTarget = cus;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/base/customers/${deleteTarget.id}`);
			showConfirm = false;
			loadData(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	onMount(() => loadData(1));
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={customers}
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
		{#snippet rowActions(cus)}
			<div class="flex flex-wrap items-center justify-center gap-1">
				<button type="button" class={dgRowBtn} onclick={() => openEditModal(cus)}>
					<Edit3 size={15} /> 编辑
				</button>
				<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(cus)}>
					<Trash2 size={15} /> 删除
				</button>
			</div>
		{/snippet}
	</DataGrid>
</div>

<Modal
	bind:show={showModal}
	title={editingId ? '编辑客户' : '开通新客户'}
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-5xl"
>
	<div class="space-y-6">
		<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
			<div class="form-control">
				<label class="label"><span class="label-text">客户编号</span></label>
				{#if editingId}
					<input
						type="text"
						bind:value={form.customer_code}
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
					><span class="label-text flex items-center gap-2"><Landmark size={14} /> 信用代码</span
					></label
				>
				<input
					type="text"
					bind:value={form.credit_code}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="统一社会信用代码"
				/>
			</div>
		</div>

		<div class="form-control">
			<label class="label"
				><span class="label-text flex items-center gap-2"><Users size={14} /> 客户名称</span></label
			>
			<input
				type="text"
				bind:value={form.customer_name}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="客户企业全称"
			/>
		</div>

		<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
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
					placeholder="联系手机"
				/>
			</div>
		</div>

		<div class="form-control">
			<label class="label"><span class="label-text">地址/备注</span></label>
			<textarea
				bind:value={form.address}
				class="textarea textarea-bordered bg-base-200/50 h-28"
				placeholder="详细地址或备注"
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
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除客户"
	message={`确定要删除客户「${deleteTarget?.customer_name || ''}」吗？删除后该客户的所有关联数据将无法恢复。`}
	onConfirm={confirmDelete}
/>

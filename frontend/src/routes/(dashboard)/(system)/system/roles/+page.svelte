<!--
功能：roles页面
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
	import { Shield, KeyRound } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger, dgRowBtnPrimary } from '$lib/dgButtonClasses';

	let roles = $state<any[]>([]);
	let permissions = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showModal = $state(false);
	let showPermModal = $state(false);
	let submitting = $state(false);
	let selectedRole: any = $state(null);
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);

	let form = $state({
		role_code: '',
		role_name: '',
		description: ''
	});

	let selectedPermIds = $state<number[]>([]);

	const columns = [
		{ key: 'role_code', label: '角色编码', class: 'font-mono text-primary', width: '16%' },
		{ key: 'role_name', label: '角色名称', class: 'font-bold', width: '18%' },
		{ key: 'description', label: '描述', width: '26%' },
		{ key: 'user_count', label: '用户数', width: '10%' },
		{ key: 'status_name', label: '状态', width: '10%' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/system/roles?page=${page}&page_size=${pageSize}`);
			roles = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadPermissions() {
		try {
			const res: any = await api.get('/system/permissions/tree');
			permissions = res || [];
		} catch (err) {
			console.error(err);
		}
	}

	async function handleSubmit() {
		submitting = true;
		try {
			await api.post('/system/roles', form);
			showModal = false;
			resetForm();
			loadData(1);
		} catch (err: any) {
			toast.error('创建失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	async function handleToggleStatus(role: any) {
		const newStatus = role.status === 'enabled' ? 'disabled' : 'enabled';
		try {
			await api.put(`/system/roles/${role.id}`, {
				role_name: role.role_name,
				description: role.description || '',
				status: newStatus
			});
			loadData(currentPage);
		} catch (err: any) {
			toast.error('操作失败: ' + (err?.message || err));
		}
	}

	async function handleDelete(role: any) {
		deleteTarget = role;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/system/roles/${deleteTarget.id}`);
			showConfirm = false;
			loadData(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	function openPermModal(role: any) {
		selectedRole = role;
		selectedPermIds = [...(role.permission_ids || [])];
		showPermModal = true;
	}

	async function handleSetPermissions() {
		if (!selectedRole) return;
		submitting = true;
		try {
			await api.post(`/system/roles/${selectedRole.id}/permissions`, {
				permission_ids: selectedPermIds
			});
			showPermModal = false;
			loadData();
		} catch (err: any) {
			toast.error('设置失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function resetForm() {
		form = { role_code: '', role_name: '', description: '' };
	}

	function togglePerm(permId: number) {
		const idx = selectedPermIds.indexOf(permId);
		if (idx >= 0) {
			selectedPermIds.splice(idx, 1);
		} else {
			selectedPermIds.push(permId);
		}
		selectedPermIds = [...selectedPermIds];
	}

	onMount(() => {
		loadData(1);
		loadPermissions();
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={roles}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		actionColumnWidth="180px"
		onCreate={() => {
			resetForm();
			showModal = true;
		}}
	>
		{#snippet rowActions(role)}
			<div class="flex flex-nowrap items-center justify-center gap-1">
				<button type="button" class={dgRowBtnPrimary} onclick={() => openPermModal(role)}>
					<KeyRound size={15} /> 权限
				</button>
				<button type="button" class={dgRowBtn} onclick={() => handleToggleStatus(role)}>
					{role.status === 'enabled' ? '停用' : '启用'}
				</button>
				<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(role)}>删除</button
				>
			</div>
		{/snippet}
	</DataGrid>
</div>

<!-- 新建角色弹窗 -->
<Modal bind:show={showModal} title="新建角色" onConfirm={handleSubmit} loading={submitting}>
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2"><Shield size={14} /> 角色编码</span
					></label
				>
				<input
					type="text"
					bind:value={form.role_code}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="如 purchaser"
				/>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">角色名称</span></label>
				<input
					type="text"
					bind:value={form.role_name}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="如 采购员"
				/>
			</div>
		</div>

		<div class="form-control">
			<label class="label"><span class="label-text">角色描述</span></label>
			<textarea
				bind:value={form.description}
				class="textarea textarea-bordered bg-base-200/50 h-20"
				placeholder="描述该角色的职责范围"
			></textarea>
		</div>
	</div>
</Modal>

<!-- 权限分配弹窗 -->
<Modal
	bind:show={showPermModal}
	title="权限配置 - {selectedRole?.role_name || ''}"
	onConfirm={handleSetPermissions}
	loading={submitting}
	maxWidth="max-w-3xl"
>
	<div class="space-y-3">
		<p class="text-base-content/60 text-sm">
			为角色 <strong>{selectedRole?.role_name}</strong> 配置菜单和操作权限：
		</p>
		<div
			class="bg-base-200/30 scrollbar-hide border-base-300 max-h-[50vh] overflow-y-auto rounded-xl border p-4"
		>
			{#each permissions as perm}
				<div class="mb-1">
					<label
						class="hover:bg-base-200 flex cursor-pointer items-center gap-2 rounded-lg px-2 py-2 transition-colors"
					>
						<input
							type="checkbox"
							class="checkbox checkbox-sm checkbox-primary"
							checked={selectedPermIds.includes(perm.id)}
							onchange={() => togglePerm(perm.id)}
						/>
						<span class="text-sm font-bold">📁 {perm.perm_name}</span>
						<span class="text-base-content/40 text-xs">{perm.perm_code}</span>
					</label>

					{#if perm.children && perm.children.length > 0}
						<div class="ml-8 space-y-0.5">
							{#each perm.children as child}
								<label
									class="hover:bg-base-200 flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1.5 transition-colors"
								>
									<input
										type="checkbox"
										class="checkbox checkbox-sm checkbox-primary"
										checked={selectedPermIds.includes(child.id)}
										onchange={() => togglePerm(child.id)}
									/>
									<span class="text-sm font-medium">📁 {child.perm_name}</span>
									<span class="text-base-content/40 text-xs">{child.perm_code}</span>
								</label>

								{#if child.children && child.children.length > 0}
									<div class="ml-8 space-y-0.5">
										{#each child.children as grandChild}
											<label
												class="hover:bg-base-200 flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1 transition-colors"
											>
												<input
													type="checkbox"
													class="checkbox checkbox-xs checkbox-primary"
													checked={selectedPermIds.includes(grandChild.id)}
													onchange={() => togglePerm(grandChild.id)}
												/>
												<span class="text-xs">🔘 {grandChild.perm_name}</span>
												<span class="text-base-content/30 text-xs">{grandChild.perm_code}</span>
											</label>
										{/each}
									</div>
								{/if}
							{/each}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除角色"
	message={`确定要删除角色「${deleteTarget?.role_name || ''}」吗？删除后已分配该角色的用户将失去对应权限。`}
	onConfirm={confirmDelete}
/>

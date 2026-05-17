<!--
功能：users页面
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
	import { User, Phone, Mail, Building2, KeyRound, Shield } from 'lucide-svelte';
	import { dgRowBtn, dgRowBtnDanger, dgRowBtnPrimary } from '$lib/dgButtonClasses';

	let users = $state<any[]>([]);
	let roles = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showModal = $state(false);
	let showRoleModal = $state(false);
	let submitting = $state(false);
	let selectedUser: any = $state(null);
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);

	let form = $state({
		username: '',
		password: '',
		real_name: '',
		phone: '',
		email: '',
		department: '',
		role_ids: [] as number[]
	});

	const columns = [
		{ key: 'username', label: '用户名', class: 'font-mono text-primary', width: '13%' },
		{ key: 'real_name', label: '真实姓名', class: 'font-bold', width: '11%' },
		{ key: 'department', label: '部门', width: '11%' },
		{ key: 'phone', label: '手机号', width: '13%' },
		{ key: 'status_name', label: '状态', width: '8%' },
		{ key: 'role_names', label: '角色', width: '13%' }
	];

	async function loadData(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/system/users?page=${page}&page_size=${pageSize}`);
			users = (res.list || []).map((u: any) => ({
				...u,
				role_names: (u.role_names || []).join(', ') || '-'
			}));
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadRoles() {
		try {
			const res: any = await api.get('/system/roles?page=1&page_size=100');
			roles = res.list || [];
		} catch (err) {
			console.error(err);
		}
	}

	async function handleSubmit() {
		submitting = true;
		try {
			await api.post('/system/users', form);
			showModal = false;
			resetForm();
			loadData(1);
		} catch (err: any) {
			toast.error('创建失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	async function handleToggleStatus(user: any) {
		const newStatus = user.status === 'enabled' ? 'disabled' : 'enabled';
		try {
			await api.put(`/system/users/${user.id}`, {
				real_name: user.real_name,
				phone: user.phone || '',
				email: user.email || '',
				department: user.department || '',
				status: newStatus,
				role_ids: user.role_ids || []
			});
			loadData(currentPage);
		} catch (err: any) {
			toast.error('操作失败: ' + (err?.message || err));
		}
	}

	async function handleDelete(user: any) {
		deleteTarget = user;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/system/users/${deleteTarget.id}`);
			showConfirm = false;
			loadData(currentPage);
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	function openRoleModal(user: any) {
		selectedUser = user;
		form.role_ids = [...(user.role_ids || [])];
		showRoleModal = true;
	}

	async function handleAssignRoles() {
		if (!selectedUser) return;
		submitting = true;
		try {
			await api.post(`/system/users/${selectedUser.id}/roles`, {
				role_ids: form.role_ids
			});
			showRoleModal = false;
			loadData();
		} catch (err: any) {
			toast.error('分配失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function resetForm() {
		form = {
			username: '',
			password: '',
			real_name: '',
			phone: '',
			email: '',
			department: '',
			role_ids: []
		};
	}

	function toggleRole(roleId: number) {
		const idx = form.role_ids.indexOf(roleId);
		if (idx >= 0) {
			form.role_ids.splice(idx, 1);
		} else {
			form.role_ids.push(roleId);
		}
		form.role_ids = [...form.role_ids];
	}

	onMount(() => {
		loadData(1);
		loadRoles();
	});
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={users}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadData}
		actionColumnWidth="190px"
		onCreate={() => {
			resetForm();
			showModal = true;
		}}
	>
		{#snippet rowActions(user)}
			<div class="flex flex-nowrap items-center justify-center gap-0.5">
				<button type="button" class={dgRowBtnPrimary} onclick={() => openRoleModal(user)}>
					<Shield size={15} /> 分配角色
				</button>
				<button type="button" class={dgRowBtn} onclick={() => handleToggleStatus(user)}>
					{user.status === 'enabled' ? '停用' : '启用'}
				</button>
				<button type="button" class={dgRowBtnDanger} onclick={() => handleDelete(user)}>删除</button
				>
			</div>
		{/snippet}
	</DataGrid>
</div>

<!-- 新建用户弹窗 -->
<Modal
	bind:show={showModal}
	title="新建用户"
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-2xl"
>
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label" for="system-user-username"
					><span class="label-text flex items-center gap-2"><User size={14} /> 用户名</span></label
				>
				<input
					id="system-user-username"
					type="text"
					bind:value={form.username}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="登录用户名"
				/>
			</div>
			<div class="form-control">
				<label class="label" for="system-user-password"
					><span class="label-text flex items-center gap-2"><KeyRound size={14} /> 初始密码</span
					></label
				>
				<input
					id="system-user-password"
					type="password"
					bind:value={form.password}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="至少6位"
				/>
			</div>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label" for="system-user-real-name"
					><span class="label-text">真实姓名</span></label
				>
				<input
					id="system-user-real-name"
					type="text"
					bind:value={form.real_name}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="姓名"
				/>
			</div>
			<div class="form-control">
				<label class="label" for="system-user-department"
					><span class="label-text flex items-center gap-2"><Building2 size={14} /> 部门</span
					></label
				>
				<input
					id="system-user-department"
					type="text"
					bind:value={form.department}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="所属部门"
				/>
			</div>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label" for="system-user-phone"
					><span class="label-text flex items-center gap-2"><Phone size={14} /> 手机号</span></label
				>
				<input
					id="system-user-phone"
					type="text"
					bind:value={form.phone}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="手机号"
				/>
			</div>
			<div class="form-control">
				<label class="label" for="system-user-email"
					><span class="label-text flex items-center gap-2"><Mail size={14} /> 邮箱</span></label
				>
				<input
					id="system-user-email"
					type="email"
					bind:value={form.email}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="邮箱地址"
				/>
			</div>
		</div>

		<div class="form-control">
			<label class="label"
				><span class="label-text flex items-center gap-2"><Shield size={14} /> 分配角色</span
				></label
			>
			<div class="mt-1 flex flex-wrap gap-2">
				{#each roles as role}
					<label class="flex cursor-pointer items-center gap-2">
						<input
							type="checkbox"
							class="checkbox checkbox-sm checkbox-primary"
							checked={form.role_ids.includes(role.id)}
							onchange={() => toggleRole(role.id)}
						/>
						<span class="text-sm">{role.role_name}</span>
					</label>
				{/each}
			</div>
		</div>
	</div>
</Modal>

<!-- 角色分配弹窗 -->
<Modal
	bind:show={showRoleModal}
	title="分配角色 - {selectedUser?.real_name || ''}"
	onConfirm={handleAssignRoles}
	loading={submitting}
>
	<div class="space-y-3">
		<p class="text-base-content/60 text-sm">
			为用户 <strong>{selectedUser?.username}</strong> 分配系统角色：
		</p>
		<div class="flex flex-wrap gap-3">
			{#each roles as role}
				<label
					class="hover:bg-base-200 flex cursor-pointer items-center gap-2 rounded-lg p-2 transition-colors"
				>
					<input
						type="checkbox"
						class="checkbox checkbox-sm checkbox-primary"
						checked={form.role_ids.includes(role.id)}
						onchange={() => toggleRole(role.id)}
					/>
					<div>
						<span class="text-sm font-medium">{role.role_name}</span>
						<span class="text-base-content/40 ml-1 text-xs">({role.role_code})</span>
					</div>
				</label>
			{/each}
		</div>
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除用户"
	message={`确定要删除用户「${deleteTarget?.real_name || ''}」吗？删除后该用户将无法登录系统，所有操作记录将保留。`}
	onConfirm={confirmDelete}
/>

<!--
功能：布局组件
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { LogOut, KeyRound, UserCircle, ChevronDown } from 'lucide-svelte';
	import { auth } from '$lib/store/auth';
	import { toast } from '$lib/store/toast';
	import api from '$lib/api/client';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fly } from 'svelte/transition';

	let { children } = $props();

	let showUserMenu = $state(false);
	let showPasswordModal = $state(false);
	let showProfileModal = $state(false);
	let submitting = $state(false);

	let passwordForm = $state({ old_password: '', new_password: '', confirm_password: '' });
	let profileForm = $state({ real_name: '', phone: '', email: '', department: '' });

	function openPasswordModal() {
		showUserMenu = false;
		passwordForm = { old_password: '', new_password: '', confirm_password: '' };
		showPasswordModal = true;
	}

	function openProfileModal() {
		showUserMenu = false;
		profileForm = {
			real_name: $auth.user?.real_name || '',
			phone: $auth.user?.phone || '',
			email: $auth.user?.email || '',
			department: $auth.user?.department || ''
		};
		showProfileModal = true;
	}

	async function handlePasswordSubmit() {
		if (passwordForm.new_password !== passwordForm.confirm_password) {
			toast.warning('两次输入的新密码不一致');
			return;
		}
		if (passwordForm.new_password.length < 6) {
			toast.warning('新密码长度不能少于6位');
			return;
		}
		submitting = true;
		try {
			const userId = $auth.user?.id;
			await api.put(`/system/users/${userId}/password`, {
				old_password: passwordForm.old_password,
				new_password: passwordForm.new_password
			});
			showPasswordModal = false;
			toast.success('密码修改成功，请重新登录');
			auth.logout();
			goto('/login');
		} catch (err: any) {
			toast.error('修改失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	async function handleProfileSubmit() {
		submitting = true;
		try {
			const userId = $auth.user?.id;
			await api.put(`/system/users/${userId}`, {
				real_name: profileForm.real_name,
				phone: profileForm.phone,
				email: profileForm.email,
				department: profileForm.department,
				status: 'enabled',
				role_ids: []
			});
			auth.updateUser({
				real_name: profileForm.real_name,
				phone: profileForm.phone,
				email: profileForm.email,
				department: profileForm.department
			});
			showProfileModal = false;
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function handleLogout() {
		showUserMenu = false;
		auth.logout();
		goto('/login');
	}

	onMount(() => {
		if (!$auth.isAuthenticated && !localStorage.getItem('token')) {
			goto('/login');
		}
	});
</script>

<svelte:window
	onclick={() => {
		showUserMenu = false;
	}}
/>

<div class="bg-base-200 flex h-screen">
	<Sidebar />

	<main class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
		<!-- Topbar -->
		<header
			class="bg-base-100/50 border-base-300 z-30 flex h-16 items-center justify-end border-b px-8 backdrop-blur-md"
		>
			<div class="flex items-center gap-4">
				<!-- 用户下拉菜单 -->
				<div class="relative">
					<button
						class="hover:bg-base-200/50 flex items-center gap-3 rounded-xl py-1 pr-1 pl-2 transition-colors"
						onclick={(e: Event) => {
							e.stopPropagation();
							showUserMenu = !showUserMenu;
						}}
					>
						<div class="flex hidden flex-col items-end sm:flex">
							<span class="text-sm font-bold">{$auth.user?.real_name || '管理员'}</span>
							<span class="text-[10px] opacity-60">{$auth.user?.role || '系统角色'}</span>
						</div>
						<div class="avatar">
							<div
								class="bg-primary text-primary-content flex h-10 w-10 items-center justify-center rounded-xl font-bold"
							>
								{($auth.user?.real_name || 'A')[0]}
							</div>
						</div>
						<ChevronDown
							size={14}
							class="opacity-40 {showUserMenu ? 'rotate-180' : ''} transition-transform"
						/>
					</button>

					{#if showUserMenu}
						<div
							class="bg-base-100 border-base-300 absolute top-14 right-0 z-50 w-64 overflow-hidden rounded-2xl border shadow-2xl"
							in:fly={{ y: 10, duration: 150 }}
						>
							<!-- 用户信息头部 -->
							<div class="border-base-200 bg-base-200/30 border-b p-4">
								<div class="flex items-center gap-3">
									<div
										class="bg-primary text-primary-content flex h-10 w-10 items-center justify-center rounded-xl text-lg font-bold"
									>
										{($auth.user?.real_name || 'A')[0]}
									</div>
									<div class="min-w-0 flex-1">
										<p class="truncate text-sm font-bold">{$auth.user?.real_name || '管理员'}</p>
										<p class="text-base-content/50 truncate text-xs">
											@{$auth.user?.username || 'admin'}
										</p>
									</div>
								</div>
								{#if $auth.user?.department}
									<p class="text-base-content/40 mt-2 text-xs">{$auth.user.department}</p>
								{/if}
							</div>

							<!-- 菜单项 -->
							<div class="p-2">
								<button
									class="hover:bg-base-200 flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-colors"
									onclick={openProfileModal}
								>
									<UserCircle size={16} class="text-primary" />
									<span>个人信息</span>
								</button>
								<button
									class="hover:bg-base-200 flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-colors"
									onclick={openPasswordModal}
								>
									<KeyRound size={16} class="text-amber-500" />
									<span>修改密码</span>
								</button>
							</div>

							<div class="border-base-200 border-t p-2">
								<button
									class="hover:bg-error/10 text-error flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-colors"
									onclick={handleLogout}
								>
									<LogOut size={16} />
									<span>退出登录</span>
								</button>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</header>

		<!-- Page Content：min-h-0 + flex 子页可撑满；矮窗口时整页可滚 -->
		<div class="flex min-h-0 flex-1 flex-col overflow-y-auto px-8 pt-0 pb-8">
			{@render children()}
		</div>
	</main>
</div>

<!-- 修改密码弹窗 -->
<Modal
	bind:show={showPasswordModal}
	title="修改密码"
	onConfirm={handlePasswordSubmit}
	loading={submitting}
>
	<div class="space-y-4">
		<div class="form-control">
			<label class="label"
				><span class="label-text flex items-center gap-2"><KeyRound size={14} /> 当前密码</span
				></label
			>
			<input
				type="password"
				bind:value={passwordForm.old_password}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="输入当前密码"
			/>
		</div>
		<div class="divider divider-sm text-base-content/20 my-0">新密码</div>
		<div class="form-control">
			<label class="label"><span class="label-text">新密码</span></label>
			<input
				type="password"
				bind:value={passwordForm.new_password}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="至少6位"
			/>
		</div>
		<div class="form-control">
			<label class="label"><span class="label-text">确认新密码</span></label>
			<input
				type="password"
				bind:value={passwordForm.confirm_password}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="再次输入新密码"
			/>
		</div>
	</div>
</Modal>

<!-- 个人信息修改弹窗 -->
<Modal
	bind:show={showProfileModal}
	title="个人信息"
	onConfirm={handleProfileSubmit}
	loading={submitting}
	maxWidth="max-w-2xl"
>
	<div class="space-y-4">
		<div class="bg-base-200/30 border-base-300 flex items-center gap-4 rounded-xl border p-4">
			<div
				class="bg-primary text-primary-content flex h-14 w-14 items-center justify-center rounded-xl text-xl font-bold"
			>
				{($auth.user?.real_name || 'A')[0]}
			</div>
			<div>
				<p class="font-bold">{$auth.user?.username || '-'}</p>
				<p class="text-base-content/50 text-xs">用户名不可修改</p>
			</div>
		</div>

		<div class="form-control">
			<label class="label"
				><span class="label-text flex items-center gap-2"><UserCircle size={14} /> 真实姓名</span
				></label
			>
			<input
				type="text"
				bind:value={profileForm.real_name}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="姓名"
			/>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"><span class="label-text">手机号</span></label>
				<input
					type="text"
					bind:value={profileForm.phone}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="手机号"
				/>
			</div>
			<div class="form-control">
				<label class="label"><span class="label-text">邮箱</span></label>
				<input
					type="email"
					bind:value={profileForm.email}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="邮箱地址"
				/>
			</div>
		</div>

		<div class="form-control">
			<label class="label"><span class="label-text">部门</span></label>
			<input
				type="text"
				bind:value={profileForm.department}
				class="input input-bordered bg-base-200/50 w-full"
				placeholder="所属部门"
			/>
		</div>
	</div>
</Modal>

<!-- Toast 通知 -->
<Toast />

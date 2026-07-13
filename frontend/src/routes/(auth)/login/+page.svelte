<!--
功能：login页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import { Shield, User, Lock, ArrowRight } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import api from '$lib/api/client';
	import { auth } from '$lib/store/auth';
	import { goto } from '$app/navigation';
	import { APP_VERSION } from '$lib/version';

	let username = $state('');
	let password = $state('');
	let loading = $state(false);
	let errorMsg = $state('');

	const handleLogin = async () => {
		if (!username || !password) {
			errorMsg = '请输入用户名和密码';
			return;
		}

		loading = true;
		errorMsg = '';

		try {
			const res: any = await api.post('/auth/login', { username, password });
			// 拦截器已解包，res 即为 { token, user_id, username, real_name, roles }
			if (res && res.token) {
				const user = {
					id: res.user_id,
					username: res.username,
					real_name: res.real_name,
					role: res.roles && res.roles.length > 0 ? res.roles[0] : 'user'
				};
				auth.login(user, res.token);
				goto('/dashboard');
			} else {
				errorMsg = '返回数据格式错误';
			}
		} catch (err: any) {
			errorMsg = err.message || '登录失败，请检查网络或凭据';
		} finally {
			loading = false;
		}
	};
</script>

<div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-slate-950">
	<!-- 动态背景装饰 -->
	<div
		class="absolute -top-20 -left-20 h-96 w-96 animate-pulse rounded-full bg-blue-600/20 blur-[100px]"
	></div>
	<div
		class="absolute -right-20 -bottom-20 h-96 w-96 animate-pulse rounded-full bg-indigo-600/20 blur-[100px]"
	></div>

	<div class="z-10 w-full max-w-md px-6" in:fly={{ y: 20, duration: 800 }}>
		<div
			class="glass-card relative flex flex-col items-center overflow-hidden rounded-3xl p-10 shadow-2xl"
		>
			<div
				class="pointer-events-none absolute inset-0 bg-gradient-to-br from-white/5 to-transparent"
			></div>

			<div
				class="mb-8 flex h-16 w-16 items-center justify-center rounded-2xl bg-blue-600 shadow-lg shadow-blue-500/50"
			>
				<Shield size={32} class="text-white" />
			</div>

			<h1 class="mb-2 text-3xl font-bold tracking-tight text-white">特维存 TeWeiCun</h1>
			<p class="mb-8 text-slate-400">特种设备专用进销存管理系统</p>

			<div class="w-full space-y-4">
				<div class="form-control">
					<label class="label mb-1" for="login-username">
						<span class="label-text ml-1 font-medium text-slate-300">用户名</span>
					</label>
					<div class="group relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
							<User
								size={18}
								class="text-slate-500 transition-colors group-focus-within:text-blue-500"
							/>
						</div>
						<input
							id="login-username"
							type="text"
							placeholder="输入账号"
							bind:value={username}
							class="input input-bordered w-full rounded-xl border-white/10 bg-white/5 pl-11 text-white transition-all focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div class="form-control">
					<label class="label mb-1" for="login-password">
						<span class="label-text ml-1 font-medium text-slate-300">密码</span>
					</label>
					<div class="group relative">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
							<Lock
								size={18}
								class="text-slate-500 transition-colors group-focus-within:text-blue-500"
							/>
						</div>
						<input
							id="login-password"
							type="password"
							placeholder="输入密码"
							bind:value={password}
							class="input input-bordered w-full rounded-xl border-white/10 bg-white/5 pl-11 text-white transition-all focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						/>
					</div>
				</div>

				{#if errorMsg}
					<div
						class="alert alert-error bg-error/20 border-error/30 text-error mt-4 rounded-xl py-2 text-sm"
						in:fade
					>
						<span>{errorMsg}</span>
					</div>
				{/if}

				<button
					class="btn btn-primary group relative mt-6 flex h-12 w-full items-center justify-center gap-2 overflow-hidden rounded-xl shadow-lg shadow-blue-600/30"
					onclick={handleLogin}
					disabled={loading}
				>
					{#if loading}
						<span class="loading loading-spinner"></span>
						验证中...
					{:else}
						<span class="z-10 flex items-center gap-2">
							进入系统 <ArrowRight
								size={18}
								class="transition-transform group-hover:translate-x-1"
							/>
						</span>
						<div
							class="absolute inset-0 bg-gradient-to-r from-blue-500 to-blue-700 opacity-0 transition-opacity group-hover:opacity-100"
						></div>
					{/if}
				</button>

				<div class="mt-8 flex justify-between px-1 text-xs text-slate-500">
					<span
						class="cursor-pointer underline-offset-4 transition-colors hover:text-slate-300 hover:underline"
						>忘记密码?</span
					>
					<span class="text-slate-600">{APP_VERSION}</span>
				</div>
			</div>
		</div>

		<div class="mt-8 text-center text-sm text-slate-600" in:fade={{ delay: 1000 }}>
			&copy; 2026 特维存管理系统 版权所有
		</div>
	</div>
</div>

<style>
</style>

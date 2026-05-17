<!--
功能：ConfirmDialog.svelte
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import { AlertTriangle, Trash2, ShieldAlert, Info } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';

	type Variant = 'danger' | 'warning' | 'info';

	let {
		show = $bindable(false),
		title = '确认操作',
		message = '',
		variant = 'danger' as Variant,
		confirmText = '确认删除',
		cancelText = '取消',
		onConfirm,
		loading = false
	} = $props();

	const variantConfig: Record<
		Variant,
		{
			icon: typeof AlertTriangle;
			iconBg: string;
			iconColor: string;
			borderAccent: string;
			btnClass: string;
			badgeClass: string;
		}
	> = {
		danger: {
			icon: Trash2,
			iconBg: 'bg-error/15',
			iconColor: 'text-error',
			borderAccent: 'border-l-error',
			btnClass: 'btn-error',
			badgeClass: 'badge-error'
		},
		warning: {
			icon: ShieldAlert,
			iconBg: 'bg-warning/15',
			iconColor: 'text-warning',
			borderAccent: 'border-l-warning',
			btnClass: 'btn-warning',
			badgeClass: 'badge-warning'
		},
		info: {
			icon: Info,
			iconBg: 'bg-info/15',
			iconColor: 'text-info',
			borderAccent: 'border-l-info',
			btnClass: 'btn-info',
			badgeClass: 'badge-info'
		}
	};

	function close() {
		show = false;
	}

	function handleConfirm() {
		onConfirm?.();
	}

	let config = $derived(variantConfig[variant]);
</script>

{#if show}
	<div
		class="fixed inset-0 z-[60] flex items-center justify-center p-4"
		in:fade={{ duration: 200 }}
		out:fade={{ duration: 150 }}
	>
		<button
			type="button"
			class="absolute inset-0 bg-slate-900/70 backdrop-blur-sm"
			aria-label="关闭确认弹窗"
			onclick={close}
		></button>

		<div
			class="bg-base-100 border-base-300 relative w-full max-w-md rounded-2xl border border-l-4 shadow-2xl {config.borderAccent} overflow-hidden"
			transition:scale={{ start: 0.9, duration: 200 }}
		>
			<!-- Warning Header Band -->
			<div class="px-6 pt-6 pb-2">
				<div class="flex items-start gap-4">
					<div class="flex-shrink-0 rounded-xl p-3 {config.iconBg} {config.iconColor}">
						<config.icon size={24} />
					</div>
					<div class="min-w-0 flex-1">
						<h3 class="text-lg font-bold tracking-tight">{title}</h3>
						<p class="text-base-content/70 mt-1.5 text-sm leading-relaxed">{message}</p>
					</div>
				</div>
			</div>

			<!-- Danger Hint -->
			{#if variant === 'danger'}
				<div class="bg-error/5 border-error/10 mx-6 mb-4 rounded-xl border px-4 py-2.5">
					<p class="text-error/80 flex items-center gap-1.5 text-xs font-medium">
						<AlertTriangle size={12} />
						此操作不可撤销，请谨慎确认
					</p>
				</div>
			{/if}

			<!-- Footer -->
			<div class="bg-base-200/30 border-base-200 flex justify-end gap-3 border-t px-6 py-4">
				<button class="btn btn-sm border-base-300 rounded-xl px-5" onclick={close}>
					{cancelText}
				</button>
				<button
					class="btn btn-sm {config.btnClass} rounded-xl px-5 shadow-lg"
					onclick={handleConfirm}
					disabled={loading}
				>
					{#if loading}
						<span class="loading loading-spinner loading-xs"></span>
					{/if}
					{confirmText}
				</button>
			</div>
		</div>
	</div>
{/if}

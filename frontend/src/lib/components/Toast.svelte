<!--
功能：Toast.svelte
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import { toast } from '$lib/store/toast';
	import type { ToastType } from '$lib/store/toast';
	import { CheckCircle2, XCircle, AlertTriangle, Info, X } from 'lucide-svelte';
	import { fly, fade } from 'svelte/transition';

	const iconMap: Record<ToastType, typeof CheckCircle2> = {
		success: CheckCircle2,
		error: XCircle,
		warning: AlertTriangle,
		info: Info
	};

	const styleMap: Record<ToastType, { bg: string; border: string; icon: string; text: string }> = {
		success: {
			bg: 'bg-emerald-500/10',
			border: 'border-emerald-500/30',
			icon: 'text-emerald-500',
			text: 'text-emerald-200'
		},
		error: {
			bg: 'bg-red-500/10',
			border: 'border-red-500/30',
			icon: 'text-red-400',
			text: 'text-red-200'
		},
		warning: {
			bg: 'bg-amber-500/10',
			border: 'border-amber-500/30',
			icon: 'text-amber-400',
			text: 'text-amber-200'
		},
		info: {
			bg: 'bg-sky-500/10',
			border: 'border-sky-500/30',
			icon: 'text-sky-400',
			text: 'text-sky-200'
		}
	};
</script>

{#if $toast.length > 0}
	<div class="pointer-events-none fixed top-6 right-6 z-[100] flex flex-col gap-3">
		{#each $toast as item (item.id)}
			{@const style = styleMap[item.type]}
			{@const Icon = iconMap[item.type]}
			<div
				class="pointer-events-auto flex items-start gap-3 rounded-2xl border px-5 py-3.5 {style.bg} {style.border} max-w-[420px] min-w-[320px] shadow-2xl backdrop-blur-md"
				in:fly={{ x: 80, duration: 300 }}
				out:fade={{ duration: 200 }}
			>
				<div class="mt-0.5 flex-shrink-0 {style.icon}">
					<Icon size={20} />
				</div>
				<p class="flex-1 text-sm leading-relaxed font-medium {style.text}">
					{item.message}
				</p>
				<button
					class="mt-0.5 flex-shrink-0 opacity-50 transition-opacity hover:opacity-100 {style.icon}"
					onclick={() => toast.dismiss(item.id)}
				>
					<X size={16} />
				</button>
			</div>
		{/each}
	</div>
{/if}

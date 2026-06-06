<!--
功能：Modal.svelte
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import { X } from 'lucide-svelte';
	import { fade, scale } from 'svelte/transition';

	let {
		show = $bindable(false),
		title = '',
		children,
		onConfirm = undefined,
		confirmText = '确定',
		cancelText = '取消',
		loading = false,
		maxWidth = 'max-w-xl',
		bodyMaxHeight = 'max-h-[70vh]',
		/** 默认整区滚动；设为 overflow-hidden 时由子级自行分配高度（如双栏表单） */
		bodyClass = 'overflow-y-auto',
		/** true：弹窗卡片 max-height + body 占满剩余高度，便于子级 flex 分栏 */
		fillBodyHeight = false
	} = $props();

	function close() {
		show = false;
	}
</script>

{#if show}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		in:fade={{ duration: 200 }}
		out:fade={{ duration: 150 }}
	>
		<button
			type="button"
			class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
			aria-label="关闭弹窗"
			onclick={close}
		></button>

		<div
			class="relative w-full {maxWidth} bg-base-100 overflow-hidden rounded-3xl border border-white/10 shadow-2xl {fillBodyHeight
				? 'flex max-h-[min(88vh,900px)] flex-col'
				: ''}"
			transition:scale={{ start: 0.95, duration: 200 }}
		>
			<!-- Header -->
			<div
				class="border-base-200 bg-base-100/50 flex shrink-0 items-center justify-between border-b px-6 py-4"
			>
				<h3 class="text-lg font-bold tracking-tight">{title}</h3>
				<button class="btn btn-sm btn-ghost btn-circle" onclick={close}>
					<X size={20} />
				</button>
			</div>

			<!-- Body -->
			<div
				class="scrollbar-hide flex min-h-0 flex-col p-6 {fillBodyHeight
					? 'min-h-0 flex-1 overflow-hidden'
					: `${bodyMaxHeight} ${bodyClass}`}"
			>
				{@render children()}
			</div>

			<!-- Footer -->
			<div
				class="border-base-200 bg-base-200/50 flex shrink-0 justify-end gap-3 border-t px-6 py-4"
			>
				<button class="btn btn-sm border-base-300 rounded-xl px-6" onclick={close}>
					{cancelText}
				</button>
				{#if onConfirm}
					<button
						class="btn btn-sm btn-primary shadow-primary/20 rounded-xl px-6 shadow-lg"
						onclick={onConfirm}
						disabled={loading}
					>
						{#if loading}
							<span class="loading loading-spinner loading-xs"></span>
						{/if}
						{confirmText}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

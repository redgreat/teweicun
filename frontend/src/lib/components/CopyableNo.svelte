<script lang="ts">
	import { Copy } from 'lucide-svelte';
	import { toast } from '$lib/store/toast';

	type Props = {
		value?: string | number | null;
		href?: string | null;
		title?: string;
		class?: string;
		onOpen?: (() => void) | null;
		truncate?: boolean;
	};

	let {
		value,
		href = null,
		title = '查看详情',
		class: className = '',
		onOpen = null,
		truncate = true
	}: Props = $props();

	const text = $derived(value === null || value === undefined || value === '' ? '' : String(value));
	const textClass = $derived(
		`${truncate ? 'min-w-0 truncate' : 'whitespace-nowrap'} font-mono ${className}`
	);

	async function copyNo(event: MouseEvent) {
		event.preventDefault();
		event.stopPropagation();
		if (!text) return;
		try {
			await navigator.clipboard.writeText(text);
			toast.success('已复制单号', 1200);
		} catch {
			const el = document.createElement('textarea');
			el.value = text;
			el.setAttribute('readonly', 'readonly');
			el.style.position = 'fixed';
			el.style.opacity = '0';
			document.body.appendChild(el);
			el.select();
			document.execCommand('copy');
			document.body.removeChild(el);
			toast.success('已复制单号', 1200);
		}
	}
</script>

{#if text}
	<span class="inline-flex max-w-full items-center gap-1 align-middle">
		{#if href}
			<a
				{href}
				class="link link-primary no-underline hover:underline {textClass}"
				{title}
			>
				{text}
			</a>
		{:else if onOpen}
			<button
				type="button"
				class="link link-primary no-underline hover:underline {textClass}"
				onclick={onOpen}
				{title}
			>
				{text}
			</button>
		{:else}
			<span class={textClass} {title}>{text}</span>
		{/if}
		<button
			type="button"
			class="btn btn-ghost text-base-content/45 hover:text-primary h-5 min-h-5 w-5 shrink-0 rounded p-0"
			title="复制单号"
			aria-label="复制单号"
			onclick={copyNo}
		>
			<Copy size={12} />
		</button>
	</span>
{:else}
	<span class="text-base-content/30">-</span>
{/if}

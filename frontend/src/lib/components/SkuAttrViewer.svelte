<script lang="ts">
	import api from '$lib/api/client';

	type AttrItem = {
		attr_id?: number;
		attr_name?: string;
		attr_value?: string;
		attr_unit?: string;
	};

	let openState = $state(false);
	let title = $state('');
	let attrs = $state<AttrItem[]>([]);
	let loading = $state(false);
	let errorMsg = $state('');

	export async function open(params: { skuId: number; title?: string; attrs?: AttrItem[] }) {
		const skuId = Number(params?.skuId || 0);
		if (!skuId) return;
		title = params?.title || `SKU #${skuId}`;
		attrs = Array.isArray(params?.attrs) ? params.attrs : [];
		errorMsg = '';
		openState = true;

		// 若已携带属性则不必请求详情
		if (attrs.length > 0) return;

		loading = true;
		try {
			const detail: any = await api.get(`/base/skus/${skuId}`);
			attrs = Array.isArray(detail?.custom_attributes) ? detail.custom_attributes : [];
		} catch (e: any) {
			errorMsg = e?.message || String(e) || '加载失败';
		} finally {
			loading = false;
		}
	}

	export function close() {
		openState = false;
	}
</script>

{#if openState}
	<div class="modal modal-open">
		<div class="modal-box max-w-2xl">
			<h3 class="text-xl font-semibold">SKU 属性</h3>
			<p class="text-base-content/60 mt-1 text-base break-words">{title}</p>

			<div class="mt-4 max-h-[60vh] space-y-2 overflow-auto">
				{#if loading}
					<div class="text-base-content/60 text-base">加载中…</div>
				{:else if errorMsg}
					<div class="text-error text-base">{errorMsg}</div>
				{:else if attrs.length === 0}
					<div class="text-base-content/60 text-base">该 SKU 没有自定义属性</div>
				{:else}
					{#each attrs as a}
						<div class="bg-base-200/40 border-base-300 rounded border px-3 py-2 text-base">
							<span class="text-base-content/60">{a.attr_name || '属性'}：</span>
							<span>{a.attr_value || '-'}</span>
							{#if a.attr_unit}
								<span class="text-base-content/60"> {a.attr_unit}</span>
							{/if}
						</div>
					{/each}
				{/if}
			</div>

			<div class="modal-action">
				<button type="button" class="btn text-base" onclick={close}>关闭</button>
			</div>
		</div>

		<button type="button" class="modal-backdrop" aria-label="关闭" onclick={close}></button>
	</div>
{/if}

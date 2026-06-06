<script lang="ts">
	import api from '$lib/api/client';
	import { onMount } from 'svelte';
	import { dgToolbarBtn } from '$lib/dgButtonClasses';

	let loading = $state(false);
	let filters = $state({ start_date: '', end_date: '' });
	let data = $state({ sales_amount: 0, cost_amount: 0, profit: 0 });

	async function load() {
		loading = true;
		try {
			const params = new URLSearchParams();
			if (filters.start_date) params.set('start_date', filters.start_date);
			if (filters.end_date) params.set('end_date', filters.end_date);
			const res: any = await api.get(`/reports/profit?${params.toString()}`);
			data = {
				sales_amount: Number(res?.sales_amount ?? 0),
				cost_amount: Number(res?.cost_amount ?? 0),
				profit: Number(res?.profit ?? 0)
			};
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function resetFilters() {
		filters = { start_date: '', end_date: '' };
		load();
	}

	onMount(load);
</script>

<div class="space-y-4">
	<div class="bg-base-100 border-base-300 shadow-base-300/50 rounded-3xl border p-5 shadow-xl">
		<div class="flex flex-wrap items-center gap-2">
			<input
				type="date"
				class="input bg-base-200 h-10 w-40 rounded-lg"
				bind:value={filters.start_date}
			/>
			<span class="opacity-60">至</span>
			<input
				type="date"
				class="input bg-base-200 h-10 w-40 rounded-lg"
				bind:value={filters.end_date}
			/>
			<button type="button" class={dgToolbarBtn} onclick={load}>查询</button>
			<button type="button" class={dgToolbarBtn} onclick={resetFilters}>重置</button>
		</div>
	</div>

	<div class="bg-base-100 border-base-300 shadow-base-300/50 overflow-hidden rounded-3xl border shadow-xl">
		<div class="border-base-200 bg-base-100/50 border-b p-5">
			<h2 class="text-lg font-bold">利润表</h2>
		</div>
		{#if loading}
			<div class="p-12 text-center"><span class="loading loading-spinner loading-lg"></span></div>
		{:else}
			<div class="overflow-x-auto">
				<table class="table-md table w-full">
					<thead class="bg-base-200/50">
						<tr>
							<th class="text-base-content/70 font-bold">项目</th>
							<th class="text-base-content/70 text-right font-bold">金额</th>
						</tr>
					</thead>
					<tbody class="divide-base-200 divide-y">
						<tr class="hover:bg-base-200/50 transition-colors">
							<td class="py-3 text-sm font-medium">销售收入</td>
							<td class="py-3 text-right font-mono text-sm">¥{data.sales_amount.toFixed(2)}</td>
						</tr>
						<tr class="hover:bg-base-200/50 transition-colors">
							<td class="py-3 text-sm font-medium">销售成本</td>
							<td class="py-3 text-right font-mono text-sm">¥{data.cost_amount.toFixed(2)}</td>
						</tr>
						<tr class="hover:bg-base-200/50 transition-colors">
							<td class="py-3 text-sm font-black">利润（销售-成本）</td>
							<td class="py-3 text-right font-mono text-sm font-black {data.profit >= 0 ? 'text-success' : 'text-error'}">
								¥{data.profit.toFixed(2)}
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>


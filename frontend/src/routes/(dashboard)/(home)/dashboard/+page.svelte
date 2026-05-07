<script lang="ts">
	import api from '$lib/api/client';
	import { onMount } from 'svelte';
	import {
		RefreshCw,
		CalendarDays,
		Package,
		ShoppingCart,
		TrendingUp,
		TrendingDown,
		Scale,
		Boxes,
		ChartColumnBig
	} from 'lucide-svelte';
	import { formatDateTimeInCn } from '$lib/datetime';

	type RangeKey = '7d' | '30d' | 'mtd';
	type MetricMode = 'qty' | 'amount';

	const rangeOptions: { key: RangeKey; label: string }[] = [
		{ key: '7d', label: '近7天' },
		{ key: '30d', label: '近30天' },
		{ key: 'mtd', label: '本月' }
	];

	let range = $state<RangeKey>('7d');
	let metricMode = $state<MetricMode>('qty');
	let loading = $state(true);
	let autoRefreshTimer: ReturnType<typeof setInterval> | null = null;
	let bigscreen = $state<any>({
		updated_at: '',
		kpi: {
			purchase_qty: 0,
			purchase_amount: 0,
			consumption_qty: 0,
			consumption_amount: 0,
			reversal_amount: 0,
			net_consumption_amount: 0,
			inventory_amount: 0
		},
		trend: [],
		top_purchase_amount: [],
		top_consumption_amount: [],
		top_consumption_qty: [],
		summary: {
			purchase_minus_consumption_amount: 0,
			active_material_count: 0,
			active_sku_count: 0,
			max_single_consumption_amount: 0
		}
	});

	function fmtMoney(v: number) {
		return (
			'¥' +
			(Number(v) || 0).toLocaleString('zh-CN', {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2
			})
		);
	}

	function fmtQty(v: number) {
		return (Number(v) || 0).toLocaleString('zh-CN', { maximumFractionDigits: 3 });
	}

	function fmtTime(v: string) {
		if (!v) return '-';
		return formatDateTimeInCn(v);
	}

	function shortDate(v: string) {
		if (!v) return '';
		return v.slice(5);
	}

	function fmtByMode(v: number, mode: MetricMode) {
		return mode === 'qty' ? fmtQty(v) : fmtMoney(v);
	}

	function niceMax(raw: number) {
		if (!raw || raw <= 0) return 1;
		const exp = Math.floor(Math.log10(raw));
		const base = Math.pow(10, exp);
		const ratio = raw / base;
		if (ratio <= 1) return 1 * base;
		if (ratio <= 2) return 2 * base;
		if (ratio <= 5) return 5 * base;
		return 10 * base;
	}

	function trendValue(point: any, side: 'purchase' | 'consumption', mode: MetricMode) {
		if (mode === 'qty') {
			return Number(side === 'purchase' ? point?.purchase_qty : point?.consumption_qty) || 0;
		}
		return Number(side === 'purchase' ? point?.purchase_amount : point?.consumption_amount) || 0;
	}

	function trendMaxByMode(side: 'purchase' | 'consumption', mode: MetricMode) {
		let m = 0;
		for (const p of bigscreen.trend || []) {
			m = Math.max(m, trendValue(p, side, mode));
		}
		return niceMax(m);
	}

	function netFlow(point: any, mode: MetricMode) {
		return trendValue(point, 'purchase', mode) - trendValue(point, 'consumption', mode);
	}

	function netFlowMaxAbs(mode: MetricMode) {
		let m = 0;
		for (const p of bigscreen.trend || []) {
			m = Math.max(m, Math.abs(netFlow(p, mode)));
		}
		return niceMax(m);
	}

	function axisTicks(max: number, steps = 4) {
		const out: number[] = [];
		for (let i = steps; i >= 0; i--) {
			out.push((max / steps) * i);
		}
		return out;
	}

	function pctY(v: number, max: number) {
		if (!max) return 100;
		return 100 - (Math.max(0, Number(v || 0)) / max) * 100;
	}

	function chartX(idx: number, total: number) {
		if (total <= 1) return 50;
		return (idx / (total - 1)) * 100;
	}

	function chartY(v: number, max: number) {
		const y = pctY(v, max);
		return Math.max(2, Math.min(98, y));
	}

	function linePointsByMode(side: 'purchase' | 'consumption', mode: MetricMode, max: number) {
		const trend = bigscreen.trend || [];
		return trend
			.map((d: any, idx: number) => {
				const x = chartX(idx, trend.length);
				const y = chartY(trendValue(d, side, mode), max);
				return `${x},${y}`;
			})
			.join(' ');
	}

	function lineYByMode(
		point: any,
		side: 'purchase' | 'consumption',
		mode: MetricMode,
		max: number
	) {
		return chartY(trendValue(point, side, mode), max);
	}

	function netBarHeightPct(v: number, maxAbs: number) {
		if (!maxAbs) return 0;
		return Math.min(50, (Math.abs(v) / maxAbs) * 50);
	}

	function recentNetFlow(mode: MetricMode) {
		const trend = bigscreen.trend || [];
		if (trend.length === 0) return 0;
		return trend.reduce((sum: number, d: any) => sum + netFlow(d, mode), 0);
	}

	function peakDay(mode: MetricMode) {
		let winner: any = null;
		for (const d of bigscreen.trend || []) {
			const p = trendValue(d, 'purchase', mode);
			const c = trendValue(d, 'consumption', mode);
			const dayPeak = Math.max(p, c);
			if (!winner || dayPeak > winner.value) {
				winner = {
					value: dayPeak,
					side: p >= c ? 'purchase' : 'consumption',
					date: d?.biz_date || ''
				};
			}
		}
		return winner;
	}

	function trendBiasLabel(mode: MetricMode) {
		const net = recentNetFlow(mode);
		if (net > 0) return '近期补库为主';
		if (net < 0) return '近期消耗偏快';
		return '近期收支平衡';
	}

	function shouldScroll(list: any[]) {
		return (list || []).length > 5;
	}

	function loopList(list: any[]) {
		const arr = list || [];
		return shouldScroll(arr) ? [...arr, ...arr] : arr;
	}

	function summaryDeltaClass(v: number) {
		if (Number(v || 0) > 0) return 'text-success';
		if (Number(v || 0) < 0) return 'text-warning';
		return 'text-base-content';
	}

	async function loadBigscreen() {
		loading = true;
		try {
			const res: any = await api.get(`/dashboard/bigscreen?range=${range}`);
			bigscreen = res || bigscreen;
		} finally {
			loading = false;
		}
	}

	function switchRange(next: RangeKey) {
		if (range === next) return;
		range = next;
		loadBigscreen();
	}

	onMount(() => {
		loadBigscreen();
		autoRefreshTimer = setInterval(() => {
			loadBigscreen();
		}, 60000);
		return () => {
			if (autoRefreshTimer) clearInterval(autoRefreshTimer);
		};
	});
</script>

<div class="flex h-[calc(100vh-7rem)] min-h-0 flex-col gap-3 overflow-hidden pb-1">
	<div
		class="bg-base-100 border-base-300 flex items-center justify-between rounded-2xl border px-4 py-2.5"
	>
		<div class="flex items-center gap-2 text-xs">
			<span class="relative inline-flex h-2 w-2">
				<span
					class="bg-success absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"
				></span>
				<span class="bg-success relative inline-flex h-2 w-2 rounded-full"></span>
			</span>
			<span class="text-base-content/55">大屏实时汇总 · 每60秒自动刷新</span>
			<span class="text-base-content/35">|</span>
			<span class="text-base-content/45">更新时间：{fmtTime(bigscreen.updated_at)}</span>
		</div>
		<div class="flex items-center gap-2">
			<div class="join">
				{#each rangeOptions as opt}
					<button
						type="button"
						class="join-item btn btn-sm {range === opt.key ? 'btn-primary' : 'btn-ghost'}"
						onclick={() => switchRange(opt.key)}
					>
						<CalendarDays size={14} />
						{opt.label}
					</button>
				{/each}
			</div>
			<button type="button" class="btn btn-ghost btn-sm" onclick={loadBigscreen}>
				<RefreshCw size={14} /> 刷新
			</button>
		</div>
	</div>

	<div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
		<div
			class="rounded-2xl border border-emerald-500/20 bg-gradient-to-br from-emerald-500/20 to-emerald-700/5 p-3"
		>
			<div class="text-xs opacity-70">近期采购数量</div>
			<div class="mt-1 flex items-end justify-between">
				<div class="text-xl font-black">{fmtQty(bigscreen.kpi.purchase_qty)}</div>
				<ShoppingCart size={18} class="text-emerald-300" />
			</div>
			<div class="mt-0.5 text-[11px] opacity-70">{fmtMoney(bigscreen.kpi.purchase_amount)}</div>
		</div>
		<div
			class="rounded-2xl border border-indigo-500/20 bg-gradient-to-br from-indigo-500/20 to-indigo-700/5 p-3"
		>
			<div class="text-xs opacity-70">近期领料数量</div>
			<div class="mt-1 flex items-end justify-between">
				<div class="text-xl font-black">{fmtQty(bigscreen.kpi.consumption_qty)}</div>
				<Package size={18} class="text-indigo-300" />
			</div>
			<div class="mt-0.5 text-[11px] opacity-70">{fmtMoney(bigscreen.kpi.consumption_amount)}</div>
		</div>
		<div
			class="rounded-2xl border border-violet-500/20 bg-gradient-to-br from-violet-500/20 to-violet-700/5 p-3"
		>
			<div class="text-xs opacity-70">净消耗金额（领料-退料）</div>
			<div class="mt-1 flex items-end justify-between">
				<div class="text-xl font-black">{fmtMoney(bigscreen.kpi.net_consumption_amount)}</div>
				<Scale size={18} class="text-violet-300" />
			</div>
			<div class="mt-0.5 text-[11px] opacity-70">
				退料 {fmtMoney(bigscreen.kpi.reversal_amount)}
			</div>
		</div>
		<div
			class="rounded-2xl border border-cyan-500/20 bg-gradient-to-br from-cyan-500/20 to-cyan-700/5 p-3"
		>
			<div class="text-xs opacity-70">当前库存总金额</div>
			<div class="mt-1 flex items-end justify-between">
				<div class="text-xl font-black">{fmtMoney(bigscreen.kpi.inventory_amount)}</div>
				<Boxes size={18} class="text-cyan-300" />
			</div>
			<div class="mt-0.5 text-[11px] opacity-70">库存快照</div>
		</div>
	</div>

	<div class="grid min-h-0 flex-1 grid-cols-1 gap-3 xl:grid-cols-[2fr_1fr]">
		<div class="grid min-h-0 grid-rows-2 gap-3">
			<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
				<div class="mb-2 flex items-center justify-between">
					<div class="flex items-center gap-2 text-sm font-bold">
						<ChartColumnBig size={16} class="text-primary" /> 净流入趋势（采购-领料）
					</div>
					<div class="text-base-content/55 text-xs">
						{metricMode === 'qty' ? '单位：数量' : '单位：金额'}
					</div>
				</div>
				<div class="border-base-300/70 bg-base-200/20 h-[calc(100%-1.75rem)] rounded-xl border p-2">
					{#if (bigscreen.trend || []).length > 0}
						<div class="relative h-full">
							<div class="border-base-300/70 absolute inset-x-1 top-1/2 border-t"></div>
							<div class="absolute inset-0 grid grid-rows-4">
								<div class="border-base-300/35 border-b"></div>
								<div class="border-base-300/35 border-b"></div>
								<div class="border-base-300/35 border-b"></div>
							</div>
							<div class="absolute inset-x-1 top-2 bottom-5 flex items-end gap-1">
								{#each bigscreen.trend as d}
									{@const flow = netFlow(d, metricMode)}
									{@const h = netBarHeightPct(flow, netFlowMaxAbs(metricMode))}
									<div class="group relative flex min-w-0 flex-1 justify-center">
										<div
											class="w-full max-w-7 rounded-sm transition-all duration-300 {flow >= 0
												? 'bg-emerald-400/85'
												: 'bg-orange-400/85'}"
											style={flow >= 0
												? `height:${h}%; align-self:center; transform:translateY(-100%);`
												: `height:${h}%; align-self:center; transform:translateY(0%);`}
											title={`${shortDate(d.biz_date)} 净流入 ${fmtByMode(flow, metricMode)}`}
										></div>
									</div>
								{/each}
							</div>
							<div class="absolute inset-x-1 bottom-0 flex items-center gap-1">
								{#each bigscreen.trend as d}
									<div class="text-base-content/45 min-w-0 flex-1 truncate text-center text-[10px]">
										{shortDate(d.biz_date)}
									</div>
								{/each}
							</div>
						</div>
					{:else}
						<div class="text-base-content/45 flex h-full items-center justify-center text-xs">
							暂无趋势数据
						</div>
					{/if}
				</div>
				<div class="mt-1 flex gap-4 text-xs">
					<span class="inline-flex items-center gap-1"
						><span class="h-2 w-2 rounded-full bg-emerald-400"></span>净流入为正（补库）</span
					>
					<span class="inline-flex items-center gap-1"
						><span class="h-2 w-2 rounded-full bg-orange-400"></span>净流入为负（消耗）</span
					>
				</div>
			</div>

			<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
				<div class="mb-2 flex items-center justify-between">
					<div class="flex items-center gap-2 text-sm font-bold">
						<Scale size={16} class="text-primary" /> 采购 vs 领料趋势（双轴）
					</div>
					<div class="join">
						<button
							type="button"
							class="join-item btn btn-xs {metricMode === 'qty' ? 'btn-primary' : 'btn-ghost'}"
							onclick={() => (metricMode = 'qty')}
						>
							数量
						</button>
						<button
							type="button"
							class="join-item btn btn-xs {metricMode === 'amount' ? 'btn-primary' : 'btn-ghost'}"
							onclick={() => (metricMode = 'amount')}
						>
							金额
						</button>
					</div>
				</div>
				<div class="grid h-[calc(100%-3.5rem)] grid-cols-[56px_1fr_56px] gap-2">
					<div class="relative">
						{#each axisTicks(trendMaxByMode('purchase', metricMode)) as t, i}
							<div class="text-base-content/55 absolute right-0 text-[10px]" style="top: {i * 25}%">
								{fmtByMode(t, metricMode)}
							</div>
						{/each}
					</div>
					<div class="relative">
						<div class="absolute inset-0">
							{#each axisTicks(trendMaxByMode('purchase', metricMode)) as _, i}
								<div
									class="border-base-300/70 absolute w-full border-t"
									style="top: {i * 25}%"
								></div>
							{/each}
						</div>
						<div class="absolute inset-x-1 top-0 bottom-5">
							{#if (bigscreen.trend || []).length > 0}
								<svg
									class="h-full w-full overflow-visible"
									viewBox="0 0 100 100"
									preserveAspectRatio="none"
								>
									<polyline
										points={linePointsByMode(
											'purchase',
											metricMode,
											trendMaxByMode('purchase', metricMode)
										)}
										fill="none"
										stroke="rgb(110 231 183)"
										stroke-width="1.4"
									/>
									<polyline
										points={linePointsByMode(
											'consumption',
											metricMode,
											trendMaxByMode('consumption', metricMode)
										)}
										fill="none"
										stroke="rgb(192 132 252)"
										stroke-width="1.4"
										stroke-opacity="0.75"
									/>
									{#each bigscreen.trend as d, idx}
										{@const px = chartX(idx, bigscreen.trend.length)}
										{@const pyPurchase = lineYByMode(
											d,
											'purchase',
											metricMode,
											trendMaxByMode('purchase', metricMode)
										)}
										{@const pyConsumption = lineYByMode(
											d,
											'consumption',
											metricMode,
											trendMaxByMode('consumption', metricMode)
										)}
										<circle cx={px} cy={pyPurchase} r="1.15" fill="rgb(110 231 183)" />
										<circle cx={px} cy={pyConsumption} r="1.15" fill="rgb(192 132 252)" />
									{/each}
								</svg>
							{:else}
								<div class="text-base-content/45 flex h-full items-center justify-center text-xs">
									暂无趋势数据
								</div>
							{/if}
						</div>
						<div class="absolute right-0 bottom-0 left-0 flex items-center gap-1 px-1">
							{#each bigscreen.trend as d}
								<div class="text-base-content/45 min-w-0 flex-1 truncate text-center text-[10px]">
									{shortDate(d.biz_date)}
								</div>
							{/each}
						</div>
					</div>
					<div class="relative">
						{#each axisTicks(trendMaxByMode('consumption', metricMode)) as t, i}
							<div class="text-base-content/55 absolute left-0 text-[10px]" style="top: {i * 25}%">
								{fmtByMode(t, metricMode)}
							</div>
						{/each}
					</div>
				</div>
				<div class="mt-1 grid grid-cols-1 gap-2 xl:grid-cols-3">
					<div class="border-base-300/70 bg-base-200/20 rounded-lg border px-2.5 py-1.5 text-xs">
						<span class="text-base-content/55">近期开口净流入：</span>
						<span class={summaryDeltaClass(recentNetFlow(metricMode))}
							>{fmtByMode(recentNetFlow(metricMode), metricMode)}</span
						>
					</div>
					<div class="border-base-300/70 bg-base-200/20 rounded-lg border px-2.5 py-1.5 text-xs">
						<span class="text-base-content/55">峰值日：</span>
						{#if peakDay(metricMode)}
							<span
								>{shortDate(peakDay(metricMode).date)} · {peakDay(metricMode).side === 'purchase'
									? '采购'
									: '领料'}
								{fmtByMode(peakDay(metricMode).value, metricMode)}</span
							>
						{:else}
							<span>暂无</span>
						{/if}
					</div>
					<div class="border-base-300/70 bg-base-200/20 rounded-lg border px-2.5 py-1.5 text-xs">
						<span class="text-base-content/55">趋势判断：</span>
						<span>{trendBiasLabel(metricMode)}</span>
					</div>
				</div>
				<div class="mt-1 flex gap-4 text-xs">
					<span class="inline-flex items-center gap-1"
						><span class="h-2 w-2 rounded-full bg-emerald-300"></span>采购（左轴）</span
					>
					<span class="inline-flex items-center gap-1"
						><span class="h-2 w-2 rounded-full bg-violet-400"></span>领料（右轴）</span
					>
				</div>
			</div>
		</div>

		<div class="grid min-h-0 grid-rows-[1fr_1fr_1fr_auto] gap-3">
			<div class="bg-base-100 border-base-300 min-h-0 overflow-hidden rounded-2xl border p-3">
				<div class="mb-2 flex items-center gap-2 text-sm font-bold">
					<TrendingUp size={16} class="text-emerald-400" /> 采购金额 TOP
				</div>
				<div class="relative h-[calc(100%-1.75rem)] overflow-hidden">
					<div class="space-y-1.5" class:top-scroll={shouldScroll(bigscreen.top_purchase_amount)}>
						{#each loopList(bigscreen.top_purchase_amount) as item, idx}
							<div
								class="bg-base-200/40 flex items-center justify-between rounded-lg px-2.5 py-1.5 text-xs"
							>
								<div class="min-w-0 flex-1">
									<div class="truncate font-medium">
										{(idx % (bigscreen.top_purchase_amount.length || 1)) + 1}. {item.material_name}
									</div>
								</div>
								<div class="text-success ml-2 font-mono">{fmtMoney(item.amount)}</div>
							</div>
						{:else}
							<div class="text-base-content/40 py-6 text-center text-xs">暂无数据</div>
						{/each}
					</div>
				</div>
			</div>
			<div class="bg-base-100 border-base-300 min-h-0 overflow-hidden rounded-2xl border p-3">
				<div class="mb-2 flex items-center gap-2 text-sm font-bold">
					<TrendingDown size={16} class="text-violet-400" /> 领料金额 TOP
				</div>
				<div class="relative h-[calc(100%-1.75rem)] overflow-hidden">
					<div
						class="space-y-1.5"
						class:top-scroll={shouldScroll(bigscreen.top_consumption_amount)}
					>
						{#each loopList(bigscreen.top_consumption_amount) as item, idx}
							<div
								class="bg-base-200/40 flex items-center justify-between rounded-lg px-2.5 py-1.5 text-xs"
							>
								<div class="min-w-0 flex-1">
									<div class="truncate font-medium">
										{(idx % (bigscreen.top_consumption_amount.length || 1)) + 1}. {item.material_name}
									</div>
								</div>
								<div class="text-warning ml-2 font-mono">{fmtMoney(item.amount)}</div>
							</div>
						{:else}
							<div class="text-base-content/40 py-6 text-center text-xs">暂无数据</div>
						{/each}
					</div>
				</div>
			</div>
			<div class="bg-base-100 border-base-300 min-h-0 overflow-hidden rounded-2xl border p-3">
				<div class="mb-2 flex items-center gap-2 text-sm font-bold">
					<Package size={16} class="text-indigo-400" /> 领料数量 TOP
				</div>
				<div class="relative h-[calc(100%-1.75rem)] overflow-hidden">
					<div class="space-y-1.5" class:top-scroll={shouldScroll(bigscreen.top_consumption_qty)}>
						{#each loopList(bigscreen.top_consumption_qty) as item, idx}
							<div
								class="bg-base-200/40 flex items-center justify-between rounded-lg px-2.5 py-1.5 text-xs"
							>
								<div class="min-w-0 flex-1">
									<div class="truncate font-medium">
										{(idx % (bigscreen.top_consumption_qty.length || 1)) + 1}. {item.material_name}
									</div>
								</div>
								<div class="text-info ml-2 font-mono">{fmtQty(item.quantity)}</div>
							</div>
						{:else}
							<div class="text-base-content/40 py-6 text-center text-xs">暂无数据</div>
						{/each}
					</div>
				</div>
			</div>
			<div class="grid grid-cols-2 gap-2">
				<div class="bg-base-100 border-base-300 rounded-xl border p-2">
					<div class="text-base-content/55 text-[11px]">采购-领料金额差</div>
					<div
						class="mt-0.5 text-sm font-bold {summaryDeltaClass(
							bigscreen.summary.purchase_minus_consumption_amount
						)}"
					>
						{fmtMoney(bigscreen.summary.purchase_minus_consumption_amount)}
					</div>
				</div>
				<div class="bg-base-100 border-base-300 rounded-xl border p-2">
					<div class="text-base-content/55 text-[11px]">最大单笔领料金额</div>
					<div class="mt-0.5 text-sm font-bold">
						{fmtMoney(bigscreen.summary.max_single_consumption_amount)}
					</div>
				</div>
				<div class="bg-base-100 border-base-300 rounded-xl border p-2">
					<div class="text-base-content/55 text-[11px]">活跃物料数</div>
					<div class="mt-0.5 text-sm font-bold">{bigscreen.summary.active_material_count || 0}</div>
				</div>
				<div class="bg-base-100 border-base-300 rounded-xl border p-2">
					<div class="text-base-content/55 text-[11px]">活跃编码数</div>
					<div class="mt-0.5 text-sm font-bold">{bigscreen.summary.active_sku_count || 0}</div>
				</div>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="text-base-content/45 py-1 text-center text-xs">数据加载中...</div>
	{/if}
</div>

<style>
	.top-scroll {
		animation: top-scroll-up 14s linear infinite;
	}

	@keyframes top-scroll-up {
		0% {
			transform: translateY(0);
		}
		100% {
			transform: translateY(-50%);
		}
	}
</style>

<!--
功能：reports页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { onMount } from 'svelte';
	import {
		Package,
		TrendingUp,
		TrendingDown,
		ArrowUpRight,
		Warehouse,
		ShoppingCart,
		Truck,
		RefreshCw
	} from 'lucide-svelte';
	import { fly } from 'svelte/transition';
	import { currentMonthInCn } from '$lib/datetime';

	let activeTab = $state('inventory');
	let loading = $state(true);
	let currentMonth = $state(currentMonthInCn());

	// 报表数据
	let inventoryData = $state<any[]>([]);
	let stockInData = $state<any[]>([]);
	let stockOutData = $state<any[]>([]);

	// 汇总指标
	let summary = $state({
		totalMaterials: 0,
		totalQuantity: 0,
		totalAvailable: 0,
		totalLocked: 0
	});

	async function loadInventoryReport() {
		loading = true;
		try {
			const res: any = await api.get('/reports/inventory');
			inventoryData = res || [];
			// 计算汇总
			summary.totalMaterials = inventoryData.length;
			summary.totalQuantity = inventoryData.reduce(
				(acc: number, r: any) => acc + (r.current_quantity || 0),
				0
			);
			summary.totalAvailable = inventoryData.reduce(
				(acc: number, r: any) => acc + (r.available_quantity || 0),
				0
			);
			summary.totalLocked = inventoryData.reduce(
				(acc: number, r: any) => acc + (r.locked_quantity || 0),
				0
			);
		} catch (err) {
			console.error(err);
			inventoryData = [];
		} finally {
			loading = false;
		}
	}

	async function loadStockInReport() {
		loading = true;
		try {
			const res: any = await api.get(`/reports/stock-in?month=${currentMonth}`);
			stockInData = res || [];
		} catch (err) {
			console.error(err);
			stockInData = [];
		} finally {
			loading = false;
		}
	}

	async function loadStockOutReport() {
		loading = true;
		try {
			const res: any = await api.get(`/reports/stock-out?month=${currentMonth}`);
			stockOutData = res || [];
		} catch (err) {
			console.error(err);
			stockOutData = [];
		} finally {
			loading = false;
		}
	}

	async function refreshData() {
		if (activeTab === 'inventory') await loadInventoryReport();
		else if (activeTab === 'stock-in') await loadStockInReport();
		else if (activeTab === 'stock-out') await loadStockOutReport();
	}

	function switchTab(tab: string) {
		activeTab = tab;
		if (tab === 'inventory') loadInventoryReport();
		else if (tab === 'stock-in') loadStockInReport();
		else if (tab === 'stock-out') loadStockOutReport();
	}

	onMount(loadInventoryReport);
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-indigo-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">运营报表</h1>
		</div>

		<div class="breadcrumbs text-sm opacity-60">
			<ul>
				<li>首页</li>
				<li>运营报表</li>
			</ul>
		</div>
	</div>

	<!-- Tab 切换 -->
	<div class="flex items-center gap-2">
		<div class="join border-base-300 border shadow-sm">
			<button
				class="join-item btn btn-sm {activeTab === 'inventory'
					? 'btn-primary'
					: 'bg-base-100 border-none'}"
				onclick={() => switchTab('inventory')}
			>
				<Warehouse size={14} /> 库存状态
			</button>
			<button
				class="join-item btn btn-sm {activeTab === 'stock-in'
					? 'btn-primary'
					: 'bg-base-100 border-none'}"
				onclick={() => switchTab('stock-in')}
			>
				<TrendingUp size={14} /> 入库汇总
			</button>
			<button
				class="join-item btn btn-sm {activeTab === 'stock-out'
					? 'btn-primary'
					: 'bg-base-100 border-none'}"
				onclick={() => switchTab('stock-out')}
			>
				<TrendingDown size={14} /> 出库汇总
			</button>
		</div>

		<div class="flex-1"></div>

		{#if activeTab !== 'inventory'}
			<input
				type="month"
				bind:value={currentMonth}
				class="input input-sm input-bordered w-40 rounded-lg"
				onchange={refreshData}
			/>
		{/if}

		<button class="btn btn-sm btn-ghost gap-1 rounded-lg" onclick={refreshData}>
			<RefreshCw size={14} /> 刷新
		</button>
	</div>

	<!-- 库存状态报表 -->
	{#if activeTab === 'inventory'}
		<!-- 汇总卡片 -->
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
			{#each [{ label: '物料种类', value: summary.totalMaterials, icon: Package, color: 'text-primary', bg: 'bg-primary/10' }, { label: '库存总量', value: summary.totalQuantity, icon: Warehouse, color: 'text-success', bg: 'bg-success/10' }, { label: '可用数量', value: summary.totalAvailable, icon: TrendingUp, color: 'text-indigo-500', bg: 'bg-indigo-500/10' }, { label: '锁定数量', value: summary.totalLocked, icon: TrendingDown, color: 'text-warning', bg: 'bg-warning/10' }] as card, i}
				<div
					class="bg-base-100 shadow-base-300/20 border-base-300 rounded-2xl border p-5 shadow-lg"
					in:fly={{ y: 10, delay: i * 50 }}
				>
					<div class="mb-3 flex items-center justify-between">
						<div class="rounded-xl p-2.5 {card.bg} {card.color}"><card.icon size={18} /></div>
						<ArrowUpRight size={12} class="opacity-20" />
					</div>
					<p class="text-xs font-bold tracking-widest uppercase opacity-40">{card.label}</p>
					<p class="mt-1 text-2xl font-black tracking-tight">{card.value.toLocaleString()}</p>
				</div>
			{/each}
		</div>

		<!-- 库存明细表 -->
		<div
			class="bg-base-100 shadow-base-300/50 border-base-300 overflow-hidden rounded-3xl border shadow-xl"
		>
			<div class="border-base-200 bg-base-100/50 border-b p-5">
				<h2 class="flex items-center gap-2 text-lg font-bold">
					<Warehouse size={18} class="text-primary" /> 库存状态明细
				</h2>
			</div>
			{#if loading}
				<div class="p-12 text-center"><span class="loading loading-spinner loading-lg"></span></div>
			{:else if inventoryData.length === 0}
				<div class="text-base-content/30 p-12 text-center italic">暂无库存数据</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-md table w-full">
						<thead class="bg-base-200/50">
							<tr>
								<th class="text-base-content/70 font-bold">仓库</th>
								<th class="text-base-content/70 font-bold">分类</th>
								<th class="text-base-content/70 font-bold">物料编码</th>
								<th class="text-base-content/70 font-bold">物料名称</th>

								<th class="text-base-content/70 font-bold">单位</th>
								<th class="text-base-content/70 text-right font-bold">当前数量</th>
								<th class="text-base-content/70 text-right font-bold">锁定数量</th>
								<th class="text-base-content/70 text-right font-bold">可用数量</th>
							</tr>
						</thead>
						<tbody class="divide-base-200 divide-y">
							{#each inventoryData as row}
								<tr class="hover:bg-base-200/50 transition-colors">
									<td class="py-3 text-sm">{row.warehouse_name || '-'}</td>
									<td class="py-3 text-sm">{row.category_name || '-'}</td>
									<td class="text-primary py-3 font-mono text-sm">{row.material_code}</td>
									<td class="py-3 text-sm font-medium">{row.material_name}</td>

									<td class="py-3 text-sm">{row.unit}</td>
									<td class="py-3 text-right font-mono text-sm">{row.current_quantity}</td>
									<td class="text-warning py-3 text-right font-mono text-sm"
										>{row.locked_quantity}</td
									>
									<td
										class="py-3 text-right font-mono text-sm {row.available_quantity < 0
											? 'text-error'
											: 'text-success'}">{row.available_quantity}</td
									>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>

		<!-- 入库汇总报表 -->
	{:else if activeTab === 'stock-in'}
		<div
			class="bg-base-100 shadow-base-300/50 border-base-300 overflow-hidden rounded-3xl border shadow-xl"
		>
			<div class="border-base-200 bg-base-100/50 flex items-center justify-between border-b p-5">
				<h2 class="flex items-center gap-2 text-lg font-bold">
					<ShoppingCart size={18} class="text-success" /> 入库汇总报表
				</h2>
				<span class="text-base-content/50 text-sm">月份: {currentMonth}</span>
			</div>
			{#if loading}
				<div class="p-12 text-center"><span class="loading loading-spinner loading-lg"></span></div>
			{:else if stockInData.length === 0}
				<div class="text-base-content/30 p-12 text-center italic">该月暂无入库数据</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-md table w-full">
						<thead class="bg-base-200/50">
							<tr>
								<th class="text-base-content/70 font-bold">供应商</th>
								<th class="text-base-content/70 font-bold">物料编码</th>
								<th class="text-base-content/70 font-bold">物料名称</th>

								<th class="text-base-content/70 font-bold">单位</th>
								<th class="text-base-content/70 text-right font-bold">入库总量</th>
								<th class="text-base-content/70 text-right font-bold">入库单数</th>
							</tr>
						</thead>
						<tbody class="divide-base-200 divide-y">
							{#each stockInData as row}
								<tr class="hover:bg-base-200/50 transition-colors">
									<td class="py-3 text-sm">{row.supplier_name || '-'}</td>
									<td class="text-primary py-3 font-mono text-sm">{row.material_code}</td>
									<td class="py-3 text-sm font-medium">{row.material_name}</td>

									<td class="py-3 text-sm">{row.unit}</td>
									<td class="text-success py-3 text-right font-mono text-sm font-bold"
										>{row.total_quantity}</td
									>
									<td class="py-3 text-right font-mono text-sm">{row.order_count}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>

		<!-- 出库汇总报表 -->
	{:else if activeTab === 'stock-out'}
		<div
			class="bg-base-100 shadow-base-300/50 border-base-300 overflow-hidden rounded-3xl border shadow-xl"
		>
			<div class="border-base-200 bg-base-100/50 flex items-center justify-between border-b p-5">
				<h2 class="flex items-center gap-2 text-lg font-bold">
					<Truck size={18} class="text-primary" /> 出库汇总报表
				</h2>
				<span class="text-base-content/50 text-sm">月份: {currentMonth}</span>
			</div>
			{#if loading}
				<div class="p-12 text-center"><span class="loading loading-spinner loading-lg"></span></div>
			{:else if stockOutData.length === 0}
				<div class="text-base-content/30 p-12 text-center italic">该月暂无出库数据</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-md table w-full">
						<thead class="bg-base-200/50">
							<tr>
								<th class="text-base-content/70 font-bold">出库类型</th>
								<th class="text-base-content/70 font-bold">物料编码</th>
								<th class="text-base-content/70 font-bold">物料名称</th>

								<th class="text-base-content/70 font-bold">单位</th>
								<th class="text-base-content/70 text-right font-bold">出库总量</th>
								<th class="text-base-content/70 text-right font-bold">出库单数</th>
							</tr>
						</thead>
						<tbody class="divide-base-200 divide-y">
							{#each stockOutData as row}
								<tr class="hover:bg-base-200/50 transition-colors">
									<td class="py-3 text-sm">
										<span
											class="badge badge-sm {row.out_type === 'sales'
												? 'badge-primary'
												: 'badge-ghost'}"
										>
											{row.out_type || '-'}
										</span>
									</td>
									<td class="text-primary py-3 font-mono text-sm">{row.material_code}</td>
									<td class="py-3 text-sm font-medium">{row.material_name}</td>

									<td class="py-3 text-sm">{row.unit}</td>
									<td class="text-primary py-3 text-right font-mono text-sm font-bold"
										>{row.total_quantity}</td
									>
									<td class="py-3 text-right font-mono text-sm">{row.order_count}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}
</div>

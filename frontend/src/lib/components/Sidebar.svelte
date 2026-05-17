<!--
功能：Sidebar.svelte
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import {
		LayoutDashboard,
		Package,
		ShoppingCart,
		Warehouse,
		TrendingUp,
		ChevronLeft,
		ClipboardCheck,
		History,
		BadgeDollarSign,
		ReceiptText
	} from 'lucide-svelte';
	import { page } from '$app/state';

	let collapsed = $state(false);
	const currentYear = new Date().getFullYear();

	const menuItems = [
		{ name: '首页大屏', icon: LayoutDashboard, path: '/dashboard' },
		{
			name: '采购管理',
			icon: ShoppingCart,
			children: [
				{ name: '采购订货', path: '/purchase/orders' },
				{ name: '采购退货', path: '/purchase/return' }
			]
		},
		{
			name: '销售管理',
			icon: BadgeDollarSign,
			children: [
				{ name: '销售订单', path: '/sales/orders' },
				{ name: '销售退货单', path: '/sales/returns' }
			]
		},
		{
			name: '领料管理',
			icon: TrendingUp,
			children: [
				{ name: '领料订单', path: '/consumption/orders' },
				{ name: '退料订单', path: '/reversal/orders' }
			]
		},
		{
			name: '出入库管理',
			icon: Warehouse,
			children: [
				{ name: '入库管理', path: '/stock/in' },
				{ name: '出库管理', path: '/stock/out' }
			]
		},
		{
			name: '对账管理',
			icon: ReceiptText,
			children: [
				{ name: '采购对账', path: '/reconciliation/purchase' },
				{ name: '销售对账', path: '/reconciliation/sales' }
			]
		},
		{ name: '库存台账', icon: ClipboardCheck, path: '/inventory/stock' },
		{ name: '物料追踪', icon: History, path: '/trace' },
		// { name: '运营报表', icon: FileText, path: '/reports' },
		{
			name: '基础数据',
			icon: Package,
			children: [
				{ name: '物料管理', path: '/materials' },
				{ name: '物料分类', path: '/materials/categories' },
				{ name: '供应商管理', path: '/suppliers' },
				{ name: '客户管理', path: '/customers' },
				{ name: '仓库管理', path: '/warehouses' },
				{ name: '用户管理', path: '/system/users' },
				{ name: '角色权限', path: '/system/roles' },
				{ name: '数据字典', path: '/system/dicts' }
			]
		}
	];

	function isActive(path: string) {
		return page.url.pathname === path || page.url.pathname.startsWith(path + '/');
	}
</script>

<aside
	class="bg-base-100 border-base-300 flex h-full flex-col border-r transition-all duration-300 {collapsed
		? 'w-20'
		: 'w-64'}"
>
	<div
		class="border-base-300 flex h-16 items-center gap-3 overflow-hidden border-b px-6 whitespace-nowrap"
	>
		<div
			class="bg-primary text-primary-content flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg font-bold"
		>
			T
		</div>
		{#if !collapsed}
			<span class="text-2xl font-bold tracking-tight"
				>特维存 <span class="text-primary text-xs opacity-70">WMS</span></span
			>
		{/if}
	</div>

	<div class="scrollbar-hide flex-1 overflow-x-hidden overflow-y-auto py-4">
		<nav class="space-y-1 px-3">
			{#each menuItems as item}
				{#if item.children}
					<details class="group" open={item.children.some((c) => isActive(c.path))}>
						<summary
							class="hover:bg-base-200 text-base-content/70 flex cursor-pointer list-none items-center gap-3 rounded-xl px-3 py-2.5 text-xl transition-colors"
						>
							<item.icon size={20} />
							{#if !collapsed}<span class="flex-1 font-medium">{item.name}</span>{/if}
							{#if !collapsed}<ChevronLeft
									size={16}
									class="transition-transform group-open:-rotate-90"
								/>{/if}
						</summary>
						{#if !collapsed}
							<div class="mt-1.5 ml-8 space-y-0.5">
								{#each item.children as child}
									<a
										href={child.path}
										class="hover:text-primary block rounded-lg px-3 py-2.5 text-[18px] transition-colors {isActive(
											child.path
										)
											? 'text-primary bg-primary/10 font-bold'
											: 'text-base-content/60'}"
									>
										{child.name}
									</a>
								{/each}
							</div>
						{/if}
					</details>
				{:else}
					<a
						href={item.path}
						class="hover:bg-base-200 flex items-center gap-3 rounded-xl px-3 py-2.5 text-xl transition-colors {isActive(
							item.path
						)
							? 'bg-primary text-primary-content shadow-primary/20 font-bold shadow-lg'
							: 'text-base-content/70'}"
					>
						<item.icon size={20} />
						{#if !collapsed}<span class="font-medium">{item.name}</span>{/if}
					</a>
				{/if}
			{/each}
		</nav>
	</div>

	<div class="border-base-300 border-t p-4">
		{#if !collapsed}
			<p class="text-base-content/30 text-center text-[10px] leading-relaxed">
				&copy; {currentYear} 特维存 ·
				<a
					class="link link-hover font-semibold tracking-wide italic"
					href="https://github.com/redgreat/TeWeiCun"
					target="_blank"
					rel="noopener noreferrer"
				>
					RedGreat
				</a><br />
				特种设备进销存管理系统
			</p>
		{:else}
			<p class="text-base-content/30 text-center text-[8px]">&copy;</p>
		{/if}
	</div>

	<button
		class="bg-base-100 border-base-300 hover:text-primary absolute top-20 -right-3 flex h-6 w-6 items-center justify-center rounded-full border transition-transform {collapsed
			? 'rotate-180'
			: ''}"
		onclick={() => (collapsed = !collapsed)}
	>
		<ChevronLeft size={14} />
	</button>
</aside>

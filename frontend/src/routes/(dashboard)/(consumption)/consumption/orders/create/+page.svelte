<!--
功能：新建领料订单页面（设计员创建领料单）
创建时间：2026-04-23
创建人：CodeArts Agent
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { ArrowLeft, Plus, Trash2, Package } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { todayDateInCn } from '$lib/datetime';

	function genProjectNo() {
		const d = new Date();
		const y = d.getFullYear();
		const m = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		const r = String(Math.floor(Math.random() * 9000) + 1000);
		return `PRJ-${y}${m}${day}-${r}`;
	}

	function todayDateString() {
		return todayDateInCn();
	}

	function parseDateOnly(input: string) {
		const text = String(input || '').trim();
		if (!text) return null;
		const normalized = text.replace(/\//g, '-');
		const parts = normalized.split('-');
		if (parts.length !== 3) return null;
		const y = Number(parts[0]);
		const m = Number(parts[1]);
		const d = Number(parts[2]);
		if (!Number.isFinite(y) || !Number.isFinite(m) || !Number.isFinite(d)) return null;
		const dt = new Date(y, m - 1, d);
		if (dt.getFullYear() !== y || dt.getMonth() !== m - 1 || dt.getDate() !== d) return null;
		return dt;
	}

	function isBeforeToday(input: string) {
		const dt = parseDateOnly(input);
		if (!dt) return false;
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		return dt.getTime() < today.getTime();
	}

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function lineAmount(item: any) {
		if (!item?.inventory_id) return 0;
		return (Number(item.quantity) || 0) * (Number(item.unit_cost) || 0);
	}

	function grandTotal() {
		let s = 0;
		for (const it of form.items) s += lineAmount(it);
		return s;
	}

	let form = $state({
		project_no: '',
		product_name: '',
		order_date: todayDateInCn(),
		remark: '',
		items: [] as any[]
	});

	let availableInventory = $state<any[]>([]);
	let invSearchTerm = $state('');
	let invOptionsPage = $state(1);
	let invOptionsHasMore = $state(true);
	let invOptionsLoading = $state(false);
	let invDropdownOpenRow = $state<number | null>(null);
	let invDropdownAnchor = $state<HTMLInputElement | null>(null);
	let invDropdownTop = $state(0);
	let invDropdownLeft = $state(0);
	let invDropdownWidth = $state(0);
	let invDropdownListMaxHeight = $state(260);
	let invDropdownRAF: number | null = null;
	let invSearchTimeout: ReturnType<typeof setTimeout> | null = null;

	let submitting = $state(false);

	function newEmptyItem() {
		return {
			inventory_id: 0,
			material_id: 0,
			material_code: '',
			material_name: '',
			is_code: false,
			sku_id: 0,
			sku_code: '',
			sku_name: '',
			warehouse_id: 0,
			warehouse_code: '',
			warehouse_name: '',
			quantity_total: 0,
			available_qty: 0,
			unit_cost: 0,
			quantity: 1,
			unit: '',
			displaySearch: ''
		};
	}

	function updateInvDropdownPosition() {
		if (!invDropdownAnchor) return;
		const rect = invDropdownAnchor.getBoundingClientRect();
		const headerHeight = 44;
		const maxListHeight = 320;
		const spaceBelow = window.innerHeight - rect.bottom;
		const spaceAbove = rect.top;
		const openUp = spaceBelow < headerHeight + 160 && spaceAbove > spaceBelow;

		const padding = 8;
		const width = Math.max(rect.width, 360);
		let left = rect.left;
		if (left + width > window.innerWidth - padding) {
			left = Math.max(padding, window.innerWidth - padding - width);
		}
		invDropdownWidth = width;
		invDropdownLeft = left;

		if (openUp) {
			const available = Math.max(
				120,
				Math.min(maxListHeight, rect.top - padding - 8 - headerHeight)
			);
			invDropdownListMaxHeight = available;
			invDropdownTop = Math.max(padding, rect.top - 8 - headerHeight - available);
		} else {
			const available = Math.max(
				120,
				Math.min(maxListHeight, window.innerHeight - padding - (rect.bottom + 8) - headerHeight)
			);
			invDropdownListMaxHeight = available;
			invDropdownTop = Math.min(
				window.innerHeight - padding - headerHeight - available,
				rect.bottom + 8
			);
		}
	}

	function startInvDropdownRAF() {
		if (invDropdownRAF) return;
		const loop = () => {
			if (invDropdownOpenRow === null || !invDropdownAnchor) {
				invDropdownRAF = null;
				return;
			}
			updateInvDropdownPosition();
			invDropdownRAF = requestAnimationFrame(loop);
		};
		invDropdownRAF = requestAnimationFrame(loop);
	}

	function stopInvDropdownRAF() {
		if (invDropdownRAF) {
			cancelAnimationFrame(invDropdownRAF);
			invDropdownRAF = null;
		}
	}

	$effect(() => {
		if (invDropdownOpenRow === null) {
			stopInvDropdownRAF();
		}
	});

	async function loadAvailableInventory(reset = false) {
		if (reset) {
			availableInventory = [];
			invOptionsPage = 1;
			invOptionsHasMore = true;
		}

		invOptionsLoading = true;
		try {
			const params = new URLSearchParams({
				page: invOptionsPage.toString(),
				page_size: '30'
			});
			const q = invSearchTerm.trim();
			if (q) params.set('q', q);

			const res: any = await api.get(`/inventory/available?${params.toString()}`);
			const newItems = res.list || [];

			if (reset) {
				availableInventory = newItems;
			} else {
				availableInventory = [...availableInventory, ...newItems];
			}

			invOptionsHasMore = newItems.length >= 30;
		} catch (err) {
			console.error(err);
			toast.error('加载有库存的 SKU 失败');
		} finally {
			invOptionsLoading = false;
		}
	}

	function addItem() {
		form.items.push(newEmptyItem());
	}

	function removeItem(index: number) {
		form.items.splice(index, 1);
	}

	function openInvDropdown(index: number, anchor: HTMLInputElement) {
		invDropdownOpenRow = index;
		invDropdownAnchor = anchor;
		invSearchTerm = form.items[index]?.displaySearch || '';
		loadAvailableInventory(true);
		updateInvDropdownPosition();
		startInvDropdownRAF();
	}

	function closeInvDropdown() {
		invDropdownOpenRow = null;
		invDropdownAnchor = null;
		stopInvDropdownRAF();
	}

	function selectInventory(index: number, inv: any) {
		const skuLabel =
			inv.sku_code || inv.sku_name
				? `${inv.sku_code || '-'} ${inv.sku_name ? `· ${inv.sku_name}` : ''}`.trim()
				: '';
		const mat = `${inv.material_code || ''} ${inv.material_name || ''}`.trim();
		form.items[index] = {
			...form.items[index],
			inventory_id: inv.inventory_id ?? inv.id,
			material_id: inv.material_id,
			material_code: inv.material_code || '',
			material_name: inv.material_name || '',
			is_code: Boolean(inv.is_code),
			sku_id: inv.sku_id || 0,
			sku_code: inv.sku_code || '',
			sku_name: inv.sku_name || '',
			warehouse_id: inv.warehouse_id,
			warehouse_code: inv.warehouse_code || '',
			warehouse_name: inv.warehouse_name || '',
			quantity_total: Number(inv.quantity ?? 0),
			available_qty: Number(inv.available_quantity ?? 0),
			unit_cost: Number(inv.unit_cost ?? 0),
			quantity: inv.is_code ? 1 : 1,
			unit: (inv.unit || '').trim() || '件',
			displaySearch: [skuLabel, mat, inv.warehouse_name].filter(Boolean).join(' · ')
		};
		closeInvDropdown();
	}

	function onInvInput(index: number) {
		if (invDropdownOpenRow !== index) return;
		const val = form.items[index]?.displaySearch || '';
		invSearchTerm = val;
		if (invSearchTimeout) clearTimeout(invSearchTimeout);
		invSearchTimeout = setTimeout(() => {
			loadAvailableInventory(true);
		}, 280);
	}

	function onInvOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		if (el.scrollTop + el.clientHeight >= el.scrollHeight - 12) {
			if (invOptionsHasMore && !invOptionsLoading) {
				invOptionsPage++;
				loadAvailableInventory(false);
			}
		}
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeInvDropdown();
	}

	async function handleSubmit() {
		if (!form.project_no.trim()) {
			toast.error('请填写项目编号');
			return;
		}
		if (!form.product_name.trim()) {
			toast.error('请输入产品名称');
			return;
		}
		if (!form.order_date) {
			toast.error('请填写领料日期');
			return;
		}
		if (isBeforeToday(form.order_date)) {
			toast.error('领料日期不能早于今天');
			return;
		}
		if (form.items.length === 0) {
			toast.error('请至少添加一条领料明细');
			return;
		}

		for (let i = 0; i < form.items.length; i++) {
			const item = form.items[i];
			if (!item.inventory_id) {
				toast.error(`第${i + 1}行：请选择 SKU / 库存批次`);
				return;
			}
			if (!item.quantity || item.quantity <= 0) {
				toast.error(`第${i + 1}行：请输入有效的领料数量`);
				return;
			}
			if (item.is_code && Math.abs(item.quantity - Math.trunc(item.quantity)) > 1e-9) {
				toast.error(`第${i + 1}行：编码管理物料数量须为整数`);
				return;
			}
			if (item.quantity > item.available_qty + 1e-9) {
				toast.error(
					`第${i + 1}行：领料数量不能超过当前可用库存（${item.available_qty} ${item.unit || ''}），提交前请刷新或改小数量`
				);
				return;
			}
		}

		submitting = true;
		try {
			const submitData = {
				project_no: form.project_no.trim(),
				product_name: form.product_name.trim(),
				order_date: form.order_date,
				remark: form.remark,
				items: form.items.map((item) => ({
					material_id: item.material_id,
					inventory_id: item.inventory_id,
					quantity: Number(item.quantity),
					unit: item.unit
				}))
			};

			await api.post('/consumption/orders', submitData);
			toast.success('领料订单提交成功，已生成出库单');
			goto('/consumption/orders');
		} catch (err: any) {
			toast.error('创建失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		goto('/consumption/orders');
	}

	onMount(() => {
		if (!form.project_no.trim()) {
			form.project_no = genProjectNo();
		}
	});
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<button type="button" class="btn btn-ghost btn-circle" onclick={goBack}>
				<ArrowLeft size={20} />
			</button>
			<div class="h-8 w-1.5 rounded-full bg-orange-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">新建领料订单</h1>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<div class="form-control">
					<label class="label" for="co-project-no">
						<span class="label-text font-medium">项目编号 <span class="text-error">*</span></span>
					</label>
					<input
						id="co-project-no"
						type="text"
						bind:value={form.project_no}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="已自动生成，可改"
					/>
				</div>

				<div class="form-control">
					<label class="label" for="co-product-name">
						<span class="label-text font-medium">产品名称 <span class="text-error">*</span></span>
					</label>
					<input
						id="co-product-name"
						type="text"
						bind:value={form.product_name}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="请输入压力容器名称"
					/>
				</div>

				<div class="form-control lg:col-span-2">
					<div class="grid grid-cols-1 gap-4 md:grid-cols-[minmax(0,14rem)_1fr]">
						<div>
							<label class="label py-0 pb-1" for="co-order-date">
								<span class="label-text font-medium"
									>领料日期 <span class="text-error">*</span></span
								>
							</label>
							<input
								id="co-order-date"
								type="date"
								bind:value={form.order_date}
								class="input input-bordered bg-base-200/50 h-11 w-full text-base"
								min={todayDateString()}
							/>
						</div>
						<div>
							<label class="label py-0 pb-1" for="co-remark">
								<span class="label-text font-medium">备注</span>
							</label>
							<input
								id="co-remark"
								type="text"
								bind:value={form.remark}
								class="input input-bordered bg-base-200/50 h-11 w-full text-base"
								placeholder="订单备注"
							/>
						</div>
					</div>
				</div>
			</div>

			<div class="divider">领料明细</div>

			<div class="space-y-4">
				<div class="flex flex-wrap items-center justify-end gap-2">
					<button type="button" class="btn btn-sm btn-primary" onclick={addItem}>
						<Plus size={16} /> 添加明细
					</button>
				</div>

				{#if form.items.length === 0}
					<div class="text-base-content/50 py-10 text-center">
						<Package size={48} class="mx-auto mb-4 opacity-30" />
						<div>暂无领料明细，请点击「添加明细」</div>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table-zebra table w-full text-base">
							<thead>
								<tr>
									<th class="w-12">#</th>
									<th class="min-w-[240px]">SKU 名称</th>
									<th class="min-w-[100px]">仓库</th>
									<th class="min-w-[72px] text-right">在库</th>
									<th class="min-w-[72px] text-right">可用</th>
									<th class="min-w-[64px]">单位</th>
									<th class="min-w-[88px] text-right">单价</th>
									<th class="min-w-[88px]">领料数量</th>
									<th class="min-w-[100px] text-right">物料总价</th>
									<th class="w-16">操作</th>
								</tr>
							</thead>
							<tbody>
								{#each form.items as item, index}
									<tr>
										<td>{index + 1}</td>
										<td>
											<div
												class="relative"
												onclick={(e) => e.stopPropagation()}
												role="presentation"
											>
												<input
													type="text"
													bind:value={item.displaySearch}
													onfocus={(e) =>
														openInvDropdown(index, e.currentTarget as HTMLInputElement)}
													oninput={() => onInvInput(index)}
													class="input input-bordered bg-base-200/50 h-10 w-full text-base"
													placeholder="搜索 SKU / 物料 / 仓库…"
												/>
											</div>
										</td>
										<td>
											{#if item.warehouse_name}
												<span class="text-base">{item.warehouse_name}</span>
											{:else}
												<span class="text-base-content/40">—</span>
											{/if}
										</td>
										<td class="text-right tabular-nums">
											{item.inventory_id ? item.quantity_total : '—'}
										</td>
										<td class="text-right font-medium tabular-nums">
											{item.inventory_id ? item.available_qty : '—'}
										</td>
										<td class="text-center">{item.unit || '—'}</td>
										<td class="text-right tabular-nums">
											{#if item.inventory_id}
												<span class="text-base-content/80">¥{formatMoney(item.unit_cost)}</span>
											{:else}
												—
											{/if}
										</td>
										<td>
											<input
												type="number"
												bind:value={item.quantity}
												min="0.001"
												step={item.is_code ? '1' : '0.001'}
												class="input input-bordered bg-base-200/50 h-10 w-full min-w-[5rem] text-base"
											/>
										</td>
										<td class="text-success text-right font-medium tabular-nums">
											{#if item.inventory_id}
												¥{formatMoney(lineAmount(item))}
											{:else}
												—
											{/if}
										</td>
										<td>
											<button
												type="button"
												class="btn btn-sm btn-ghost text-error"
												onclick={() => removeItem(index)}
											>
												<Trash2 size={16} />
											</button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>

			<div class="border-base-300 flex flex-wrap items-center justify-end gap-4 border-t pt-6">
				{#if form.items.length > 0}
					<div class="mr-auto text-base">
						<span class="text-base-content/60">合计：</span>
						<span class="text-success font-mono text-lg font-semibold"
							>¥{formatMoney(grandTotal())}</span
						>
					</div>
				{/if}
				<button type="button" class="btn" onclick={goBack} disabled={submitting}>取消</button>
				<button type="button" class="btn btn-primary" onclick={handleSubmit} disabled={submitting}>
					{#if submitting}
						<span class="loading loading-spinner loading-sm"></span>
					{/if}
					提交领料订单
				</button>
			</div>
		</div>
	</div>
</div>

{#if invDropdownOpenRow !== null}
	<div class="fixed inset-0 z-[70]" role="presentation" onclick={closeInvDropdown}></div>
	<div
		class="bg-base-100 border-base-300 fixed z-[80] overflow-hidden rounded-xl border shadow-2xl"
		style="left: {invDropdownLeft}px; top: {invDropdownTop}px; width: {invDropdownWidth}px;"
		role="presentation"
		onclick={(e) => e.stopPropagation()}
	>
		<div
			class="text-base-content/50 border-base-200 flex items-center justify-between border-b px-3 py-2 text-xs"
		>
			<span>可选库存</span>
			<button type="button" class="btn btn-xs btn-ghost" onclick={closeInvDropdown}>关闭</button>
		</div>
		<div
			class="overflow-auto"
			style="max-height: {invDropdownListMaxHeight}px"
			onscroll={onInvOptionsScroll}
		>
			{#if invOptionsLoading && invOptionsPage === 1}
				<div class="text-base-content/50 p-4 text-center">正在加载...</div>
			{:else if availableInventory.length === 0}
				<div class="text-base-content/50 p-4 text-center">暂无可用库存</div>
			{:else}
				{#each availableInventory as inv}
					<button
						type="button"
						class="hover:bg-base-200/60 border-base-200 w-full cursor-pointer border-b px-3 py-2.5 text-left last:border-b-0"
						onclick={() => selectInventory(invDropdownOpenRow as number, inv)}
					>
						<div class="text-sm font-medium">
							{inv.sku_name || inv.material_name || inv.material_code || '—'}
						</div>
						<div class="text-base-content/70 mt-0.5 text-xs">
							{inv.material_code}
							{inv.material_name}
							{#if inv.sku_code}
								<span class="text-primary ml-1 font-mono">[{inv.sku_code}]</span>
							{/if}
						</div>
						<div class="text-base-content/50 mt-0.5 flex flex-wrap gap-x-2 text-xs">
							<span>{inv.warehouse_name}</span>
							<span>可用 {inv.available_quantity} {inv.unit}</span>
						</div>
					</button>
				{/each}
				{#if invOptionsLoading && invOptionsPage > 1}
					<div class="text-base-content/50 p-2 text-center text-xs">加载更多…</div>
				{/if}
			{/if}
		</div>
	</div>
{/if}

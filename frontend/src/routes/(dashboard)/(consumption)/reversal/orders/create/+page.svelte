<!--
功能：新建退料订单页面（设计员创建退料单）
创建时间：2026-04-23
创建人：CodeArts Agent
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { ArrowLeft, Plus, Trash2, RotateCcw } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { todayDateInCn } from '$lib/datetime';
	import { buildFloatingDropdownGridLayout, calcFloatingDropdownPlacement } from '$lib/dropdown';

	function genProjectNo() {
		const d = new Date();
		const y = d.getFullYear();
		const m = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		const r = String(Math.floor(Math.random() * 9000) + 1000);
		return `PRJ-${y}${m}${day}-${r}`;
	}

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function formatQty(n: number) {
		const x = Number(n);
		if (!Number.isFinite(x)) return '—';
		if (Math.abs(x - Math.round(x)) < 1e-9) return String(Math.round(x));
		return x.toLocaleString('zh-CN', { maximumFractionDigits: 3 });
	}

	function invRowCap(inv: any) {
		const issued = Number(inv?.issued_quantity ?? 0);
		const avail = Number(inv?.available_quantity ?? 0);
		if (!Number.isFinite(issued) || !Number.isFinite(avail)) return 0;
		return Math.max(0, Math.min(issued, avail));
	}

	function buildInvOptionLabel(inv: any) {
		const matName = inv?.material_display_name || inv?.material_name || '';
		const matCode = inv?.material_code || '';
		const wh = inv?.warehouse_name || inv?.warehouse_code || '';
		const netIssued = formatQty(Number(inv?.issued_quantity ?? 0));
		const avail = formatQty(Number(inv?.available_quantity ?? 0));
		const cap = formatQty(invRowCap(inv));
		const codeTag = inv?.is_code === false ? '无编码' : '有编码';
		const base = matName && matCode ? `${matName} [${matCode}]` : matName || matCode || '';
		return `${base} | ${matCode} | ${codeTag} | 仓库:${wh} | 可退(净):${netIssued} | 可用:${avail} | 上限:${cap}`;
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
		production_order_id: 0,
		production_return_order_id: 0,
		items: [] as any[]
	});

	let invOptions = $state<any[]>([]);
	let invOptionsTotal = $state(0);
	let invOptionsPage = $state(1);
	let invOptionsHasMore = $state(true);
	let invOptionsLoading = $state(false);
	let invDropdownOpenRow = $state<number | null>(null);
	let invDropdownAnchor = $state<HTMLInputElement | null>(null);
	let invDropdownTop = $state(0);
	let invDropdownLeft = $state(0);
	let invDropdownWidth = $state(0);
	let invDropdownListMaxHeight = $state(260);
	let invDropdownGridTemplate = $state('300px 192px 192px');
	let invSearchTerm = $state('');
	let invSearchTimeout: ReturnType<typeof setTimeout> | null = null;
	let invDropdownRAF: number | null = null;

	let submitting = $state(false);

	let productionOrders = $state<any[]>([]);
	let productionReturnOrders = $state<any[]>([]);

	async function loadProductionDropdowns() {
		try {
			const poRes: any = await api.get('/production/orders/dropdown');
			productionOrders = poRes || [];
		} catch (e) {
			console.error(e);
		}
		try {
			const prRes: any = await api.get('/production/returns/dropdown');
			productionReturnOrders = prRes || [];
		} catch (e) {
			console.error(e);
		}
	}

	function newEmptyItem() {
		return {
			inventory_id: 0,
			material_id: 0,
			material_label: '',
			warehouse_name: '',
			warehouse_code: '',
			max_quantity: 0,
			issued_quantity: 0,
			available_quantity: 0,
			quantity: 1,
			unit: '',
			unit_cost: 0,
			remark: ''
		};
	}

	function updateInvDropdownPosition() {
		if (!invDropdownAnchor) return;
		const layout = buildFloatingDropdownGridLayout({
			firstColumnTexts: invOptions.map((inv) =>
				String(inv.material_display_name || inv.material_name || inv.material_code || '-')
			),
			fixedColumnWidths: [192, 192],
			minFirstColumnWidth: 300
		});
		const placement = calcFloatingDropdownPlacement({
			anchor: invDropdownAnchor,
			minWidth: Math.max(760, layout.preferredPanelWidth),
			maxWidth: 1800,
			maxListHeight: 320,
			headerHeight: 44,
			preferBelowMinSpace: 204,
			extraWidth: 150,
			contentTexts: invOptions.map((inv) =>
				[
					String(inv.material_display_name || inv.material_name || '-'),
					String(inv.material_code || '-'),
					String(inv.warehouse_name || inv.warehouse_code || '-'),
					`可用 ${formatQty(inv.available_quantity ?? 0)}`,
					`上限 ${formatQty(invRowCap(inv))}`
				].join('    ')
			)
		});
		const resolvedWidth = Math.max(placement.width, layout.preferredPanelWidth);
		const viewportWidth = typeof window === 'undefined' ? resolvedWidth : window.innerWidth;
		invDropdownWidth = resolvedWidth;
		invDropdownLeft = Math.max(8, Math.min(placement.left, viewportWidth - resolvedWidth - 8));
		invDropdownTop = placement.top;
		invDropdownListMaxHeight = placement.listMaxHeight;
		invDropdownGridTemplate = layout.gridTemplate;
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

	function normalizeInvSearchTerm(text: string) {
		return String(text || '')
			.replace(/\s*\[[^\]]*\]\s*$/, '')
			.trim();
	}

	async function loadInvOptions(params: { reset: boolean }) {
		if (invOptionsLoading) return;
		const nextPage = params.reset ? 1 : invOptionsPage + 1;

		invOptionsLoading = true;
		try {
			let url = `/inventory/issued?page=${nextPage}&page_size=50`;
			const q = invSearchTerm.trim();
			if (q) url += `&q=${encodeURIComponent(q)}`;
			const res: any = await api.get(url);
			const list = res.list || [];
			const total = Number(res.total || 0);
			invOptionsTotal = total;
			invOptionsPage = nextPage;
			invOptions = params.reset ? list : [...invOptions, ...list];
			invOptionsHasMore = invOptions.length < total && list.length > 0;
		} catch (err) {
			console.error(err);
			toast.error('加载可退物料失败');
		} finally {
			invOptionsLoading = false;
		}
	}

	function openInvDropdown(index: number, anchor: HTMLInputElement) {
		invDropdownOpenRow = index;
		invDropdownAnchor = anchor;
		invSearchTerm = normalizeInvSearchTerm(form.items[index]?.material_label || '');
		invOptions = [];
		invOptionsTotal = 0;
		invOptionsPage = 1;
		invOptionsHasMore = true;
		loadInvOptions({ reset: true });
		updateInvDropdownPosition();
		startInvDropdownRAF();
	}

	function closeInvDropdown() {
		invDropdownOpenRow = null;
		invDropdownAnchor = null;
		stopInvDropdownRAF();
	}

	function selectInv(index: number, inv: any) {
		const invId = Number(inv.inventory_id ?? inv.id ?? 0);
		const maxQty = invRowCap(inv);
		const issued = Number(inv.issued_quantity ?? 0);
		const avail = Number(inv.available_quantity ?? 0);
		let qty = 1;
		if (maxQty > 0) {
			qty = Math.min(maxQty, maxQty < 1 ? maxQty : 1);
		}
		form.items[index] = {
			...form.items[index],
			inventory_id: invId,
			material_id: Number(inv.material_id || 0),
			warehouse_name: String(inv.warehouse_name || ''),
			warehouse_code: String(inv.warehouse_code || ''),
			unit: String(inv.unit || '').trim() || '件',
			unit_cost: Number(inv.unit_cost ?? 0),
			max_quantity: maxQty,
			issued_quantity: issued,
			available_quantity: avail,
			material_label: buildInvOptionLabel(inv),
			quantity: qty
		};
		closeInvDropdown();
	}

	function onInvInput(index: number) {
		if (invDropdownOpenRow !== index) return;
		const val = form.items[index]?.material_label || '';
		invSearchTerm = normalizeInvSearchTerm(val);
		if (invSearchTimeout) clearTimeout(invSearchTimeout);
		invSearchTimeout = setTimeout(() => {
			invOptions = [];
			invOptionsTotal = 0;
			invOptionsPage = 1;
			invOptionsHasMore = true;
			loadInvOptions({ reset: true });
		}, 250);
	}

	function onInvOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		if (el.scrollTop + el.clientHeight >= el.scrollHeight - 12) {
			if (invOptionsHasMore && !invOptionsLoading) {
				loadInvOptions({ reset: false });
			}
		}
	}

	function addItem() {
		form.items.push(newEmptyItem());
	}

	function removeItem(index: number) {
		form.items.splice(index, 1);
	}

	function validateForm() {
		if (!form.project_no.trim()) {
			toast.error('请输入项目编号');
			return false;
		}
		if (!form.product_name.trim()) {
			toast.error('请输入产品名称');
			return false;
		}
		if (!form.order_date) {
			toast.error('请选择退料日期');
			return false;
		}
		if (form.items.length === 0) {
			toast.error('请至少添加一条退料明细');
			return false;
		}
		for (let i = 0; i < form.items.length; i++) {
			const item = form.items[i];
			if (!item.inventory_id) {
				toast.error(`第 ${i + 1} 行：请选择可退物料`);
				return false;
			}
			const qty = Number(item.quantity);
			if (!qty || qty <= 0) {
				toast.error(`第 ${i + 1} 行：请输入有效的退料数量`);
				return false;
			}
			const maxQ = Number(item.max_quantity || 0);
			if (maxQ > 0 && qty > maxQ + 1e-6) {
				toast.error(`第 ${i + 1} 行：退料数量不能超过允许上限（净可退与当前可用中的较小值）`);
				return false;
			}
		}
		return true;
	}

	async function handleSubmit() {
		if (!validateForm()) return;

		submitting = true;
		try {
			const submitData = {
				project_no: form.project_no.trim(),
				product_name: form.product_name.trim(),
				order_date: form.order_date,
				remark: form.remark,
				production_order_id: Number(form.production_order_id) || 0,
				production_return_order_id: Number(form.production_return_order_id) || 0,
				items: form.items.map((item) => ({
					inventory_id: item.inventory_id,
					material_id: item.material_id,
					quantity: Number(item.quantity),
					unit: item.unit,
					remark: item.remark || ''
				}))
			};

			await api.post('/reversal/orders', submitData);
			toast.success('退料订单提交成功，已生成入库单');
			goto('/reversal/orders');
		} catch (err: any) {
			toast.error('创建失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		goto('/reversal/orders');
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeInvDropdown();
	}

	onMount(() => {
		if (!form.project_no.trim()) {
			form.project_no = genProjectNo();
		}
		loadProductionDropdowns();
	});
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<button type="button" class="btn btn-ghost btn-circle" onclick={goBack}>
				<ArrowLeft size={20} />
			</button>
			<div class="h-8 w-1.5 rounded-full bg-blue-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">新建退料订单</h1>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<div class="form-control">
					<label class="label" for="ro-project-no">
						<span class="label-text font-medium">项目编号 <span class="text-error">*</span></span>
					</label>
					<input
						id="ro-project-no"
						type="text"
						bind:value={form.project_no}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="项目/图纸编号"
					/>
				</div>

				<div class="form-control">
					<label class="label" for="ro-product-name">
						<span class="label-text font-medium">产品名称 <span class="text-error">*</span></span>
					</label>
					<input
						id="ro-product-name"
						type="text"
						bind:value={form.product_name}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="请输入压力容器名称"
					/>
				</div>

				<div class="form-control">
					<label class="label" for="ro-order-date">
						<span class="label-text font-medium">退料日期 <span class="text-error">*</span></span>
					</label>
					<input
						id="ro-order-date"
						type="date"
						bind:value={form.order_date}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
					/>
				</div>

				<div class="form-control lg:col-span-2">
					<label class="label" for="ro-remark">
						<span class="label-text font-medium">备注</span>
					</label>
					<input
						id="ro-remark"
						type="text"
						bind:value={form.remark}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="订单备注"
					/>
				</div>
			</div>

			<div class="divider">关联生产单据（可选）</div>

			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<div class="form-control">
					<label class="label" for="ro-prod-order">
						<span class="label-text font-medium">关联生产单</span>
					</label>
					<select
						id="ro-prod-order"
						class="select select-bordered bg-base-200/50 h-11 w-full text-base"
						bind:value={form.production_order_id}
					>
						<option value="0">不关联（提交后自动生成新的）</option>
						{#each productionOrders as po}
							<option value={po.id}>
								{po.production_no}{po.material_name ? ` | ${po.material_name}` : ''}
							</option>
						{/each}
					</select>
					<label class="label py-0 pt-0.5">
						<span class="label-text-alt text-base-content/50">多个退料单可关联同一生产单，系统自动计算成本汇总</span>
					</label>
				</div>

				<div class="form-control">
					<label class="label" for="ro-prod-return">
						<span class="label-text font-medium">关联生产退货单</span>
					</label>
					<select
						id="ro-prod-return"
						class="select select-bordered bg-base-200/50 h-11 w-full text-base"
						bind:value={form.production_return_order_id}
					>
						<option value="0">不关联</option>
						{#each productionReturnOrders as pr}
							<option value={pr.id}>
								{pr.return_no}{pr.material_name ? ` | ${pr.material_name}` : ''}
							</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="divider">退料明细</div>

			<div class="space-y-4">
				<div class="flex flex-wrap items-center justify-end gap-2">
					<button type="button" class="btn btn-sm btn-primary" onclick={addItem}>
						<Plus size={16} /> 添加明细
					</button>
				</div>

				{#if form.items.length === 0}
					<div class="text-base-content/50 py-10 text-center">
						<RotateCcw size={48} class="mx-auto mb-4 opacity-30" />
						<div>暂无退料明细，请点击「添加明细」</div>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table-zebra table w-full text-base">
							<thead>
								<tr>
									<th class="w-10">#</th>
									<th class="min-w-[360px] lg:min-w-[480px]">物料名称</th>
									<th class="min-w-[140px]">退货仓库</th>
									<th class="min-w-[88px] text-right">库存可用</th>
									<th class="min-w-[88px] text-right">可退上限</th>
									<th class="min-w-[132px]">退料数量</th>
									<th class="min-w-[72px]">单位</th>
									<th class="min-w-[104px] text-right">单价</th>
									<th class="min-w-[120px] text-right">物料总价</th>
									<th class="w-14">操作</th>
								</tr>
							</thead>
							<tbody>
								{#each form.items as item, index}
									<tr>
										<td class="text-base-content/60">{index + 1}</td>
										<td>
											<div
												class="relative"
												onclick={(e) => e.stopPropagation()}
												role="presentation"
											>
												<input
													type="text"
													bind:value={item.material_label}
													class="input input-bordered bg-base-200/50 h-10 w-full min-w-[280px] text-base"
													placeholder="搜索在库物料名称或编码…"
													onfocus={(e) =>
														openInvDropdown(index, e.currentTarget as HTMLInputElement)}
													oninput={() => onInvInput(index)}
												/>
											</div>
										</td>
										<td>{item.warehouse_name || item.warehouse_code || '-'}</td>
										<td class="text-right text-base tabular-nums">
											{#if item.inventory_id}
												{formatQty(item.available_quantity)}
											{:else}
												—
											{/if}
										</td>
										<td class="text-right text-base tabular-nums">
											{#if item.inventory_id}
												{formatQty(item.max_quantity)}
											{:else}
												—
											{/if}
										</td>
										<td>
											<div class="flex flex-col gap-0.5">
												<input
													type="number"
													bind:value={item.quantity}
													min="0"
													step="any"
													class="input input-bordered bg-base-200/50 h-10 w-full max-w-[7.5rem] text-base"
												/>
											</div>
										</td>
										<td class="text-base">{item.unit || '-'}</td>
										<td class="text-right text-base tabular-nums">
											{#if item.inventory_id}
												¥{formatMoney(item.unit_cost)}
											{:else}
												—
											{/if}
										</td>
										<td class="text-success text-right text-base font-medium tabular-nums">
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
						<span class="text-base-content/60">退料金额合计：</span>
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
					提交退料订单
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
			<span>匹配可退物料 {invOptions.length} / {invOptionsTotal}</span>
			<button type="button" class="btn btn-xs btn-ghost" onclick={closeInvDropdown}>关闭</button>
		</div>
		<div
			class="overflow-auto"
			style="max-height: {invDropdownListMaxHeight}px"
			onscroll={onInvOptionsScroll}
		>
			<div
				class="bg-base-200/80 border-base-200 sticky top-0 z-10 border-b px-3 py-2 backdrop-blur-sm"
			>
				<div
					class="grid w-full gap-3 text-[11px] font-medium"
					style:grid-template-columns={invDropdownGridTemplate}
				>
					<div>物料</div>
					<div>仓库</div>
					<div>可退库存</div>
				</div>
			</div>
			{#if invOptions.length === 0 && !invOptionsLoading}
				<div class="text-base-content/50 p-4 text-sm">无匹配的已领料可退物料</div>
			{:else}
				{#each invOptions as invOpt}
					<button
						type="button"
						class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
						onclick={() => selectInv(invDropdownOpenRow as number, invOpt)}
					>
						<div
							class="grid w-full items-center gap-3"
							style:grid-template-columns={invDropdownGridTemplate}
						>
							<div class="min-w-0">
								<div class="text-sm font-medium whitespace-nowrap">
									{invOpt.material_display_name || invOpt.material_name || '-'}
								</div>
								<div class="text-base-content/60 font-mono text-[11px] whitespace-nowrap">
									{invOpt.material_code || '-'}
								</div>
							</div>
							<div>
								<div class="text-xs whitespace-nowrap">
									{invOpt.warehouse_name || invOpt.warehouse_code || '-'}
								</div>
								<div class="text-base-content/60 text-[11px] whitespace-nowrap">
									{invOpt.is_code === false ? '无编码' : '有编码'} / 单位 {invOpt.unit || '-'}
								</div>
							</div>
							<div class="text-left md:text-right">
								<div class="text-xs whitespace-nowrap">
									可用 {formatQty(invOpt.available_quantity ?? 0)}
								</div>
								<div class="text-[11px] whitespace-nowrap text-emerald-500">
									上限 {formatQty(invRowCap(invOpt))}
								</div>
							</div>
						</div>
					</button>
				{/each}
			{/if}
			{#if invOptionsLoading}
				<div class="text-base-content/50 p-3 text-xs">加载中...</div>
			{:else if invOptionsHasMore}
				<div class="text-base-content/50 p-3 text-xs">下拉加载更多...</div>
			{/if}
		</div>
	</div>
{/if}

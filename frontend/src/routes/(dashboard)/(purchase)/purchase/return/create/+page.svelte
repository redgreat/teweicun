<!--
功能：新增采购退货（内嵌页，替代弹窗）
创建时间：2026-05-16
创建人：GPT-5.4
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { ArrowLeft, Plus, Trash2, RotateCcw } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { todayDateInCn } from '$lib/datetime';
	import { buildFloatingDropdownGridLayout, calcFloatingDropdownPlacement } from '$lib/dropdown';

	let suppliers = $state<any[]>([]);

	let invOptions = $state<any[]>([]);
	let invOptionsTotal = $state(0);
	let invOptionsPage = $state(1);
	let invOptionsLoading = $state(false);
	let invOptionsHasMore = $state(true);
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

	let form = $state({
		supplier_code: '',
		return_date: todayDateInCn(),
		remark: '',
		items: [] as any[]
	});

	let submitting = $state(false);
	async function loadSuppliers() {
		try {
			const res: any = await api.get('/base/suppliers?page=1&page_size=100');
			suppliers = res.list || [];
		} catch (err) {
			console.error(err);
			toast.error('加载供应商失败');
		}
	}

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
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

	function lineAmount(item: any) {
		if (!item?.inventory_id) return 0;
		return (Number(item.quantity) || 0) * (Number(item.unit_cost) || 0);
	}

	function grandTotal() {
		let s = 0;
		for (const it of form.items) s += lineAmount(it);
		return s;
	}

	function buildInvOptionLabel(inv: any) {
		const matName = inv?.material_display_name || inv?.material_name || '';
		const matCode = inv?.material_code || '';
		const wh = inv?.warehouse_name || inv?.warehouse_code || '';
		const available = inv?.available_quantity ?? inv?.available ?? '';
		const base = matName && matCode ? `${matName} [${matCode}]` : matName || matCode || '';
		return `${base} | 仓库:${wh} | 可用:${available}`;
	}

	function normalizeInvSearchTerm(text: string) {
		return String(text || '')
			.replace(/\s*\[[^\]]*\]\s*$/, '')
			.trim();
	}

	async function loadInvOptions(params: { reset: boolean }) {
		if (invOptionsLoading) return;
		const nextPage = params.reset ? 1 : invOptionsPage + 1;

		if (!form.supplier_code) {
			invOptions = [];
			invOptionsTotal = 0;
			invOptionsPage = 1;
			invOptionsHasMore = false;
			return;
		}

		invOptionsLoading = true;
		try {
			let url = `/inventory/available?page=${nextPage}&page_size=50&supplier_code=${encodeURIComponent(form.supplier_code)}`;
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
			toast.error('加载在库物料失败');
		} finally {
			invOptionsLoading = false;
		}
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
					`可用 ${inv.available_quantity ?? 0}`,
					`单价 ¥${formatMoney(inv.unit_cost ?? 0)}`
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

	function openInvDropdown(index: number, anchor: HTMLInputElement) {
		if (!form.supplier_code) {
			toast.error('请先选择供应商');
			return;
		}
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
		const maxQty = Number(inv.available_quantity ?? 0);
		form.items[index].inventory_id = Number(inv.inventory_id || 0);
		form.items[index].material_id = Number(inv.material_id || 0);
		form.items[index].custom_attributes = inv.custom_attributes || [];
		form.items[index].unit = String(inv.unit || '').trim() || '件';
		form.items[index].unit_cost = Number(inv.unit_cost ?? 0);
		form.items[index].max_quantity = maxQty;
		form.items[index].material_label = buildInvOptionLabel(inv);
		form.items[index].quantity = maxQty > 0 ? maxQty : 1;
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
		form.items.push({
			inventory_id: 0,
			material_id: 0,
			material_label: '',
			quantity: 1,
			unit: '',
			unit_cost: 0,
			max_quantity: 0,
			custom_attributes: [] as any[]
		});
	}

	function removeItem(index: number) {
		form.items.splice(index, 1);
	}

	async function handleSubmit() {
		if (!form.supplier_code) {
			toast.error('请选择供应商');
			return;
		}
		if (!form.return_date) {
			toast.error('请选择退货日期');
			return;
		}
		if (isBeforeToday(form.return_date)) {
			toast.error('退货日期不能早于今天');
			return;
		}
		if (form.items.length === 0) {
			toast.error('请添加退货明细');
			return;
		}
		for (const item of form.items) {
			if (!item.inventory_id) {
				toast.error('退货明细必须选择在库物料');
				return;
			}
			const qty = Number(item.quantity || 0);
			if (qty <= 0) {
				toast.error('退货数量必须大于 0');
				return;
			}
			const maxQty = Number(item.max_quantity || 0);
			if (maxQty > 0 && qty > maxQty) {
				toast.error('退货数量超过可用数量');
				return;
			}
		}

		submitting = true;
		try {
			await api.post('/returns', {
				return_type: 'purchase_return',
				supplier_code: form.supplier_code,
				return_date: form.return_date,
				remark: form.remark,
				items: form.items.map((it: any) => ({
					inventory_id: it.inventory_id,
					quantity: Number(it.quantity || 0)
				}))
			});
			toast.success('创建成功');
			goto('/purchase/return');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		closeInvDropdown();
		goto('/purchase/return');
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeInvDropdown();
	}

	$effect(() => {
		if (invDropdownOpenRow === null) {
			stopInvDropdownRAF();
		}
	});

	onMount(() => {
		loadSuppliers();
	});
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<button type="button" class="btn btn-ghost btn-circle" onclick={goBack}>
				<ArrowLeft size={20} />
			</button>
			<div class="bg-primary h-8 w-1.5 rounded-full"></div>
			<h1 class="text-2xl font-bold tracking-tight">新增采购退货</h1>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label" for="prc-supplier"
						><span class="label-text font-medium">供应商</span></label
					>
					<select
						id="prc-supplier"
						bind:value={form.supplier_code}
						class="select select-bordered bg-base-200/50 h-11 w-full text-base"
						required
					>
						<option value="">选择供应商</option>
						{#each suppliers as sup}
							<option value={sup.supplier_code}>{sup.supplier_name}</option>
						{/each}
					</select>
				</div>
				<div class="form-control">
					<label class="label" for="prc-return-date"
						><span class="label-text font-medium">退货日期</span></label
					>
					<input
						id="prc-return-date"
						type="date"
						bind:value={form.return_date}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						min={todayDateString()}
					/>
				</div>
			</div>

			<div class="grid grid-cols-1">
				<div class="form-control">
					<label class="label" for="prc-remark"
						><span class="label-text font-medium">备注</span></label
					>
					<input
						id="prc-remark"
						type="text"
						bind:value={form.remark}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="备注信息"
					/>
				</div>
			</div>

			<div class="divider">退货明细</div>

			<div class="flex flex-wrap items-center justify-end gap-2">
				<button type="button" class="btn btn-sm btn-primary" onclick={addItem}>
					<Plus size={16} /> 添加明细
				</button>
			</div>

			{#if form.items.length === 0}
				<div class="text-base-content/50 py-10 text-center">
					<RotateCcw size={48} class="mx-auto mb-4 opacity-30" />
					<div>暂无退货明细，请点击「添加明细」</div>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-zebra table w-full text-base">
						<thead>
							<tr>
								<th class="w-10">#</th>
								<th class="min-w-[320px] lg:min-w-[420px]">物料名称</th>
								<th class="w-24">属性</th>
								<th class="min-w-[72px] text-right">可用</th>
								<th class="min-w-[120px]">退货数量</th>
								<th class="min-w-[56px] text-right">单位</th>
								<th class="min-w-[88px] text-right">单价</th>
								<th class="min-w-[100px] text-right">物料总价</th>
								<th class="w-14">操作</th>
							</tr>
						</thead>
						<tbody>
							{#each form.items as item, i}
								<tr>
									<td class="text-base-content/60">{i + 1}</td>
									<td>
										<div class="relative" onclick={(e) => e.stopPropagation()} role="presentation">
											<input
												type="text"
												bind:value={item.material_label}
												class="input input-bordered bg-base-200/50 h-10 w-full min-w-[280px] text-base"
												placeholder="搜索在库物料名称或编码…"
												onfocus={(e) => openInvDropdown(i, e.currentTarget as HTMLInputElement)}
												oninput={() => onInvInput(i)}
											/>
										</div>
									</td>
									<td>
										<span class="text-sm">{(item.custom_attributes || []).length}项</span>
									</td>
									<td class="text-right text-base tabular-nums">
										{item.inventory_id ? item.max_quantity : '—'}
									</td>
									<td>
										<div class="flex flex-col gap-0.5">
											<input
												type="number"
												bind:value={item.quantity}
												class="input input-bordered bg-base-200/50 h-10 w-full max-w-[7.5rem] text-base"
												min="0.001"
												step="0.001"
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
											class="btn btn-xs btn-ghost text-error"
											onclick={() => removeItem(i)}
										>
											<Trash2 size={12} />
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}

			<div class="border-base-300 flex flex-wrap items-center justify-end gap-4 border-t pt-6">
				{#if form.items.length > 0}
					<div class="mr-auto text-base">
						<span class="text-base-content/60">退货金额合计：</span>
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
					提交退货单
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
			<span>匹配在库物料 {invOptions.length} / {invOptionsTotal}</span>
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
					<div>库存</div>
				</div>
			</div>
			{#if invOptions.length === 0 && !invOptionsLoading}
				<div class="text-base-content/50 p-4 text-sm">无匹配在库物料</div>
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
									单位 {invOpt.unit || '-'}
								</div>
							</div>
							<div class="text-left md:text-right">
								<div class="text-xs whitespace-nowrap">可用 {invOpt.available_quantity ?? 0}</div>
								<div class="text-[11px] whitespace-nowrap text-emerald-500">
									单价 ¥{formatMoney(invOpt.unit_cost ?? 0)}
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

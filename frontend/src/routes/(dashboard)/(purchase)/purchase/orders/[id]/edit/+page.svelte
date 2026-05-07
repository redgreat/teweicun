<!--
功能：编辑采购订货（待提交，内嵌页）
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { ArrowLeft, Plus, Trash2, Calendar, Building2 } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { formatDateInCn, todayDateInCn } from '$lib/datetime';

	let suppliers = $state<any[]>([]);
	let initialLoading = $state(true);

	let skuOptions = $state<any[]>([]);
	let skuOptionsTotal = $state(0);
	let skuOptionsPage = $state(1);
	let skuOptionsLoading = $state(false);
	let skuOptionsHasMore = $state(true);
	let skuDropdownOpenRow = $state<number | null>(null);
	let skuDropdownAnchor = $state<HTMLInputElement | null>(null);
	let skuDropdownTop = $state(0);
	let skuDropdownLeft = $state(0);
	let skuDropdownWidth = $state(0);
	let skuDropdownListMaxHeight = $state(260);
	let skuSearchTerm = $state('');
	let skuSearchTimeout: ReturnType<typeof setTimeout> | null = null;
	let skuDropdownRAF: number | null = null;

	let form = $state({
		supplier_code: '',
		order_date: '',
		expected_date: '',
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

	function buildSkuOptionLabel(sku: any) {
		const name = sku?.material_name || sku?.sku_name || '';
		const code = sku?.material_code || sku?.sku_code || '';
		if (name && code) return `${name} [${code}]`;
		return name || code || '';
	}

	function todayDateString() {
		return todayDateInCn();
	}

	function normalizeDateInput(raw: string) {
		const text = String(raw || '').trim();
		if (!text) return '';
		const datePart = formatDateInCn(text).replace(/\//g, '-');
		if (datePart === '-') return '';
		const parts = datePart.split('-');
		if (parts.length !== 3) return datePart;
		const y = Number(parts[0]);
		const m = Number(parts[1]);
		const d = Number(parts[2]);
		if (!Number.isFinite(y) || !Number.isFinite(m) || !Number.isFinite(d)) return datePart;
		return `${String(y).padStart(4, '0')}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
	}

	function parseDateOnly(input: string) {
		const normalized = normalizeDateInput(input);
		if (!normalized) return null;
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

	function normalizeSkuSearchTerm(text: string) {
		return String(text || '')
			.replace(/\s*\[[^\]]*\]\s*$/, '')
			.trim();
	}

	async function loadSkuOptions(params: { reset: boolean }) {
		if (skuOptionsLoading) return;
		const nextPage = params.reset ? 1 : skuOptionsPage + 1;

		skuOptionsLoading = true;
		try {
			let url = `/base/materials?page=${nextPage}&page_size=50&status=enabled`;
			const q = skuSearchTerm.trim();
			if (q) {
				url += `&material_name=${encodeURIComponent(q)}`;
			}
			const res: any = await api.get(url);
			const list = res.list || [];
			const total = Number(res.total || 0);
			skuOptionsTotal = total;
			skuOptionsPage = nextPage;
			skuOptions = params.reset ? list : [...skuOptions, ...list];
			skuOptionsHasMore = skuOptions.length < total && list.length > 0;
		} catch (err) {
			console.error(err);
		} finally {
			skuOptionsLoading = false;
		}
	}

	function updateSkuDropdownPosition() {
		if (!skuDropdownAnchor) return;
		const rect = skuDropdownAnchor.getBoundingClientRect();
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
		skuDropdownWidth = width;
		skuDropdownLeft = left;

		if (openUp) {
			const available = Math.max(
				120,
				Math.min(maxListHeight, rect.top - padding - 8 - headerHeight)
			);
			skuDropdownListMaxHeight = available;
			skuDropdownTop = Math.max(padding, rect.top - 8 - headerHeight - available);
		} else {
			const available = Math.max(
				120,
				Math.min(maxListHeight, window.innerHeight - padding - (rect.bottom + 8) - headerHeight)
			);
			skuDropdownListMaxHeight = available;
			skuDropdownTop = Math.min(
				window.innerHeight - padding - headerHeight - available,
				rect.bottom + 8
			);
		}
	}

	function startSkuDropdownRAF() {
		if (skuDropdownRAF) return;
		const loop = () => {
			if (skuDropdownOpenRow === null || !skuDropdownAnchor) {
				skuDropdownRAF = null;
				return;
			}
			updateSkuDropdownPosition();
			skuDropdownRAF = requestAnimationFrame(loop);
		};
		skuDropdownRAF = requestAnimationFrame(loop);
	}

	function stopSkuDropdownRAF() {
		if (skuDropdownRAF) {
			cancelAnimationFrame(skuDropdownRAF);
			skuDropdownRAF = null;
		}
	}

	function buildItem() {
		return {
			material_id: 0,
			material_label: '',
			quantity: 1,
			unit: '',
			unit_price: 0,
			custom_attributes: [] as any[]
		};
	}

	function addItem() {
		form.items.push(buildItem());
	}

	function removeItem(index: number) {
		form.items.splice(index, 1);
	}

	function openSkuDropdown(index: number, anchor: HTMLInputElement) {
		skuDropdownOpenRow = index;
		skuDropdownAnchor = anchor;
		skuSearchTerm = normalizeSkuSearchTerm(form.items[index]?.material_label || '');
		skuOptions = [];
		skuOptionsTotal = 0;
		skuOptionsPage = 1;
		skuOptionsHasMore = true;
		loadSkuOptions({ reset: true });
		updateSkuDropdownPosition();
		startSkuDropdownRAF();
	}

	function closeSkuDropdown() {
		skuDropdownOpenRow = null;
		skuDropdownAnchor = null;
		stopSkuDropdownRAF();
	}

	function onSkuInput(index: number) {
		const item = form.items[index];
		skuDropdownOpenRow = index;
		skuSearchTerm = normalizeSkuSearchTerm(item?.material_label || '');
		if (item) {
			item.material_id = 0;
			item.custom_attributes = [];
		}
		if (skuSearchTimeout) clearTimeout(skuSearchTimeout);
		skuSearchTimeout = setTimeout(() => {
			skuOptions = [];
			skuOptionsTotal = 0;
			skuOptionsPage = 1;
			skuOptionsHasMore = true;
			loadSkuOptions({ reset: true });
		}, 250);
	}

	function selectSku(index: number, sku: any) {
		const item = form.items[index];
		if (!item) return;
		if (
			form.items.some(
				(it: any, idx: number) => idx !== index && it.material_id && it.material_id === sku.id
			)
		) {
			toast.warning('该物料已添加，不能重复');
			return;
		}
		item.material_id = sku.id;
		item.material_label = buildSkuOptionLabel(sku);
		item.custom_attributes = sku.custom_attributes || [];
		if (!item.unit) {
			item.unit = sku.unit_name || sku.unit || '';
		}
		if (!item.unit_price || Number(item.unit_price) === 0) {
			item.unit_price = Number(sku.reference_price || 0);
		}
		closeSkuDropdown();
	}

	function onSkuOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 24;
		if (nearBottom && skuOptionsHasMore && !skuOptionsLoading) {
			loadSkuOptions({ reset: false });
		}
	}

	function formatAmount(amount: number) {
		return '¥' + (amount || 0).toFixed(2);
	}

	async function loadOrder(id: number) {
		try {
			const detail: any = await api.get(`/purchase/orders/${id}`);
			if (detail.order_status !== 'draft') {
				toast.error('仅待提交订单可编辑');
				goto('/purchase/orders');
				return;
			}
			const itemList = (detail.items || []).map((item: any) => {
				const skuName = item.material_name || item.sku_name || '';
				const skuCode = item.material_code || item.sku_code || '';
				const skuLabel =
					skuName && skuCode
						? `${skuName} [${skuCode}]`
						: skuName || (skuCode ? `[${skuCode}]` : '');
				return {
					material_id: item.material_id || 0,
					material_label: skuLabel,
					quantity: item.quantity || 1,
					unit: item.unit || '',
					unit_price: Number(item.unit_price || 0),
					custom_attributes: item.custom_attributes || []
				};
			});
			form = {
				supplier_code: detail.supplier_code || '',
				order_date: normalizeDateInput(String(detail.order_date || '')),
				expected_date: normalizeDateInput(String(detail.expected_date || '')),
				remark: detail.remark || '',
				items: itemList
			};
		} catch (err: any) {
			toast.error('加载订单失败: ' + (err?.message || err));
			goto('/purchase/orders');
		} finally {
			initialLoading = false;
		}
	}

	async function handleSubmit() {
		const id = Number($page.params.id);
		if (!id) {
			toast.error('无效的订单');
			return;
		}
		if (!form.expected_date) {
			toast.error('请填写预计到货日期');
			return;
		}
		if (isBeforeToday(form.expected_date)) {
			toast.error('预计到货日期不能早于今天');
			return;
		}
		if (form.items.length === 0) {
			toast.error('请添加采购明细');
			return;
		}
		const materialSet = new Set<number>();
		for (const item of form.items) {
			if (!item.material_id) {
				toast.error('存在未选择物料的明细行');
				return;
			}
			if (materialSet.has(Number(item.material_id))) {
				toast.error('物料明细存在重复，请去重');
				return;
			}
			materialSet.add(Number(item.material_id));
			if (!item.unit) {
				toast.error('存在未填写单位的明细行');
				return;
			}
		}

		submitting = true;
		try {
			await api.put(`/purchase/orders/${id}`, {
				supplier_code: form.supplier_code,
				order_date: form.order_date,
				expected_date: form.expected_date,
				remark: form.remark,
				items: form.items.map((item: any) => ({
					material_id: item.material_id,
					quantity: Number(item.quantity || 0),
					unit_price: Number(item.unit_price || 0)
				}))
			});
			toast.success('更新成功');
			goto('/purchase/orders');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		closeSkuDropdown();
		goto('/purchase/orders');
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeSkuDropdown();
	}

	$effect(() => {
		if (skuDropdownOpenRow === null) {
			stopSkuDropdownRAF();
		}
	});

	onMount(async () => {
		const id = Number($page.params.id);
		if (!id) {
			toast.error('无效的订单 ID');
			goto('/purchase/orders');
			return;
		}
		await loadSuppliers();
		await loadOrder(id);
	});
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<button type="button" class="btn btn-ghost btn-circle" onclick={goBack}>
				<ArrowLeft size={20} />
			</button>
			<div class="h-8 w-1.5 rounded-full bg-amber-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">编辑采购订货</h1>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			{#if initialLoading}
				<div class="text-base-content/50 py-16 text-center">加载中…</div>
			{:else}
				<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
					<div class="form-control">
						<label class="label" for="poe-supplier"
							><span class="label-text flex items-center gap-2 font-medium"
								><Building2 size={14} /> 供应商</span
							></label
						>
						<input
							id="poe-supplier"
							type="text"
							value={suppliers.find((s) => s.supplier_code === form.supplier_code)?.supplier_name ||
								form.supplier_code ||
								''}
							class="input input-bordered bg-base-200/40 w-full cursor-not-allowed"
							readonly
						/>
					</div>
					<div class="form-control">
						<label class="label" for="poe-order-date"
							><span class="label-text flex items-center gap-2 font-medium"
								><Calendar size={14} /> 订单日期</span
							></label
						>
						<input
							id="poe-order-date"
							type="date"
							bind:value={form.order_date}
							class="input input-bordered bg-base-200/40 w-full cursor-not-allowed"
							readonly
						/>
					</div>
				</div>

				<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
					<div class="form-control">
						<label class="label" for="poe-expected"
							><span class="label-text font-medium"
								>预计到货日期 <span class="text-error">*</span></span
							></label
						>
						<input
							id="poe-expected"
							type="date"
							bind:value={form.expected_date}
							class="input input-bordered bg-base-200/50 w-full"
							min={todayDateString()}
							required
						/>
					</div>
					<div class="form-control">
						<label class="label" for="poe-remark"
							><span class="label-text font-medium">备注</span></label
						>
						<input
							id="poe-remark"
							type="text"
							bind:value={form.remark}
							class="input input-bordered bg-base-200/50 w-full"
							placeholder="备注信息"
						/>
					</div>
				</div>

				<div class="divider">采购明细</div>

				<div class="flex flex-wrap items-center justify-end gap-2">
					<button type="button" class="btn btn-sm btn-primary gap-2" onclick={addItem}>
						<Plus size={14} /> 添加明细
					</button>
				</div>

				{#if form.items.length === 0}
					<div class="text-base-content/50 py-10 text-center text-sm">
						暂无明细，请点击「添加明细」
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table-zebra table w-full text-[15px]">
							<thead>
								<tr>
									<th class="min-w-[320px] lg:min-w-[420px]">物料名称</th>
									<th class="min-w-[100px]">属性</th>
									<th class="min-w-[100px]">数量</th>
									<th class="min-w-[100px]">单位</th>
									<th class="min-w-[120px]">单价</th>
									<th class="min-w-[100px]">金额</th>
									<th class="w-14">操作</th>
								</tr>
							</thead>
							<tbody>
								{#each form.items as item, i}
									<tr>
										<td>
											<div
												class="relative"
												onclick={(e) => e.stopPropagation()}
												role="presentation"
											>
												<input
													type="text"
													bind:value={item.material_label}
													class="input input-sm input-bordered bg-base-200/50 w-full min-w-[280px]"
													placeholder="输入物料名称搜索…"
													onfocus={(e) => openSkuDropdown(i, e.currentTarget as HTMLInputElement)}
													oninput={() => onSkuInput(i)}
												/>
											</div>
										</td>
										<td>
											<span class="text-sm">{(item.custom_attributes || []).length}项</span>
										</td>
										<td>
											<input
												type="number"
												bind:value={item.quantity}
												class="input input-sm input-bordered bg-base-200/50 w-full max-w-[7rem]"
												min="1"
												step="1"
											/>
										</td>
										<td>
											<span class="text-sm">{item.unit || '-'}</span>
										</td>
										<td>
											<input
												type="number"
												bind:value={item.unit_price}
												class="input input-sm input-bordered bg-base-200/50 w-full max-w-[7.5rem]"
												min="0"
												step="0.01"
											/>
										</td>
										<td>
											<span class="font-mono text-sm font-semibold"
												>{formatAmount((item.quantity || 0) * (item.unit_price || 0))}</span
											>
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
					<button type="button" class="btn" onclick={goBack} disabled={submitting}>取消</button>
					<button
						type="button"
						class="btn btn-primary"
						onclick={handleSubmit}
						disabled={submitting}
					>
						{#if submitting}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						保存修改
					</button>
				</div>
			{/if}
		</div>
	</div>
</div>

{#if skuDropdownOpenRow !== null}
	<div class="fixed inset-0 z-[70]" role="presentation" onclick={closeSkuDropdown}></div>
	<div
		class="bg-base-100 border-base-300 fixed z-[80] overflow-hidden rounded-xl border shadow-2xl"
		style="left: {skuDropdownLeft}px; top: {skuDropdownTop}px; width: {skuDropdownWidth}px;"
		role="presentation"
		onclick={(e) => e.stopPropagation()}
	>
		<div
			class="text-base-content/50 border-base-200 flex items-center justify-between border-b px-3 py-2 text-xs"
		>
			<span>匹配 {skuOptions.length} / {skuOptionsTotal}</span>
			<button type="button" class="btn btn-xs btn-ghost" onclick={closeSkuDropdown}>关闭</button>
		</div>
		<div
			class="overflow-auto"
			style="max-height: {skuDropdownListMaxHeight}px"
			onscroll={onSkuOptionsScroll}
		>
			{#if skuOptions.length === 0 && !skuOptionsLoading}
				<div class="text-base-content/50 p-4 text-sm">无匹配物料</div>
			{:else}
				{#each skuOptions as skuOpt}
					<button
						type="button"
						class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
						onclick={() => selectSku(skuDropdownOpenRow as number, skuOpt)}
					>
						<div class="text-sm font-medium">{skuOpt.material_name || '-'}</div>
						<div class="text-base-content/60 font-mono text-xs">{buildSkuOptionLabel(skuOpt)}</div>
					</button>
				{/each}
			{/if}
			{#if skuOptionsLoading}
				<div class="text-base-content/50 p-3 text-xs">加载中...</div>
			{:else if skuOptionsHasMore}
				<div class="text-base-content/50 p-3 text-xs">下拉加载更多...</div>
			{/if}
		</div>
	</div>
{/if}

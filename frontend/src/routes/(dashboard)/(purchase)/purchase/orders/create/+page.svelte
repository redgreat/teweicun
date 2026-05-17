<!--
功能：新增采购订货（内嵌页）
创建时间：2026-05-16
创建人：GPT-5.4
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount, tick } from 'svelte';
	import { ArrowLeft, Plus, Trash2, Calendar, Building2 } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { todayDateInCn } from '$lib/datetime';
	import { buildFloatingDropdownGridLayout, calcFloatingDropdownPlacement } from '$lib/dropdown';

	let suppliers = $state<any[]>([]);

	let materialOptions = $state<any[]>([]);
	let materialOptionsTotal = $state(0);
	let materialOptionsPage = $state(1);
	let materialOptionsLoading = $state(false);
	let materialOptionsHasMore = $state(true);
	let materialDropdownOpenRow = $state<number | null>(null);
	let materialDropdownAnchor = $state<HTMLInputElement | null>(null);
	let materialDropdownTop = $state(0);
	let materialDropdownLeft = $state(0);
	let materialDropdownWidth = $state(0);
	let materialDropdownListMaxHeight = $state(260);
	let materialDropdownGridTemplate = $state('260px 192px 96px');
	let materialSearchTerm = $state('');
	let materialSearchTimeout: ReturnType<typeof setTimeout> | null = null;
	let materialDropdownRAF: number | null = null;

	let form = $state({
		supplier_code: '',
		order_date: todayDateInCn(),
		expected_date: todayDateInCn(),
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

	function buildMaterialOptionLabel(material: any) {
		const name = material?.material_display_name || material?.material_name || '';
		const code = material?.material_code || '';
		if (name && code) return `${name} [${code}]`;
		return name || code || '';
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

	function normalizeMaterialSearchTerm(text: string) {
		return String(text || '')
			.replace(/\s*\[[^\]]*\]\s*$/, '')
			.trim();
	}

	async function loadMaterialOptions(params: { reset: boolean }) {
		if (materialOptionsLoading) return;
		const nextPage = params.reset ? 1 : materialOptionsPage + 1;

		materialOptionsLoading = true;
		try {
			const baseUrl = `/base/materials?page=${nextPage}&page_size=50&status=enabled`;
			const q = materialSearchTerm.trim();
			let res: any;
			if (q) {
				const byName: any = await api.get(`${baseUrl}&material_name=${encodeURIComponent(q)}`);
				const byNameList = byName.list || [];
				res =
					nextPage > 1 || byNameList.length > 0
						? byName
						: await api.get(`${baseUrl}&material_code=${encodeURIComponent(q)}`);
			} else {
				res = await api.get(baseUrl);
			}
			const list = res.list || [];
			const total = Number(res.total || 0);
			materialOptionsTotal = total;
			materialOptionsPage = nextPage;
			materialOptions = params.reset ? list : [...materialOptions, ...list];
			materialOptionsHasMore = materialOptions.length < total && list.length > 0;
		} catch (err) {
			console.error(err);
		} finally {
			materialOptionsLoading = false;
		}
	}

	function updateMaterialDropdownPosition() {
		if (!materialDropdownAnchor) return;
		const layout = buildFloatingDropdownGridLayout({
			firstColumnTexts: materialOptions.map((material) =>
				String(material.material_display_name || material.material_name || '-')
			),
			fixedColumnWidths: [192, 96],
			minFirstColumnWidth: 260
		});
		const placement = calcFloatingDropdownPlacement({
			anchor: materialDropdownAnchor,
			minWidth: Math.max(620, layout.preferredPanelWidth),
			maxWidth: 1600,
			maxListHeight: 320,
			headerHeight: 44,
			preferBelowMinSpace: 204,
			extraWidth: 128,
			contentTexts: materialOptions.map((material) =>
				[
					String(material.material_display_name || material.material_name || '-'),
					String(material.material_code || '-'),
					String(material.unit_name || material.unit || '-')
				].join('    ')
			)
		});
		const resolvedWidth = Math.max(placement.width, layout.preferredPanelWidth);
		const viewportWidth = typeof window === 'undefined' ? resolvedWidth : window.innerWidth;
		materialDropdownWidth = resolvedWidth;
		materialDropdownLeft = Math.max(8, Math.min(placement.left, viewportWidth - resolvedWidth - 8));
		materialDropdownTop = placement.top;
		materialDropdownListMaxHeight = placement.listMaxHeight;
		materialDropdownGridTemplate = layout.gridTemplate;
	}

	function startMaterialDropdownRAF() {
		if (materialDropdownRAF) return;
		const loop = () => {
			if (materialDropdownOpenRow === null || !materialDropdownAnchor) {
				materialDropdownRAF = null;
				return;
			}
			updateMaterialDropdownPosition();
			materialDropdownRAF = requestAnimationFrame(loop);
		};
		materialDropdownRAF = requestAnimationFrame(loop);
	}

	function stopMaterialDropdownRAF() {
		if (materialDropdownRAF) {
			cancelAnimationFrame(materialDropdownRAF);
			materialDropdownRAF = null;
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

	async function addItem() {
		form.items.push(buildItem());
		// 行数较多时，新增后自动滚动到页面底部，避免输入区被遮挡。
		await tick();
		window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' });
	}

	function removeItem(index: number) {
		form.items.splice(index, 1);
	}

	function openMaterialDropdown(index: number, anchor: HTMLInputElement) {
		materialDropdownOpenRow = index;
		materialDropdownAnchor = anchor;
		materialSearchTerm = normalizeMaterialSearchTerm(form.items[index]?.material_label || '');
		materialOptions = [];
		materialOptionsTotal = 0;
		materialOptionsPage = 1;
		materialOptionsHasMore = true;
		loadMaterialOptions({ reset: true });
		updateMaterialDropdownPosition();
		startMaterialDropdownRAF();
	}

	function closeMaterialDropdown() {
		materialDropdownOpenRow = null;
		materialDropdownAnchor = null;
		stopMaterialDropdownRAF();
	}

	function onMaterialInput(index: number) {
		const item = form.items[index];
		materialDropdownOpenRow = index;
		materialSearchTerm = normalizeMaterialSearchTerm(item?.material_label || '');
		if (item) {
			item.material_id = 0;
			item.custom_attributes = [];
		}
		if (materialSearchTimeout) clearTimeout(materialSearchTimeout);
		materialSearchTimeout = setTimeout(() => {
			materialOptions = [];
			materialOptionsTotal = 0;
			materialOptionsPage = 1;
			materialOptionsHasMore = true;
			loadMaterialOptions({ reset: true });
		}, 250);
	}

	function selectMaterial(index: number, material: any) {
		const item = form.items[index];
		if (!item) return;
		if (
			form.items.some(
				(it: any, idx: number) => idx !== index && it.material_id && it.material_id === material.id
			)
		) {
			toast.warning('该物料已添加，不能重复');
			return;
		}
		item.material_id = material.id;
		item.material_label = buildMaterialOptionLabel(material);
		item.custom_attributes = material.custom_attributes || [];
		if (!item.unit) {
			item.unit = material.unit_name || material.unit || '';
		}
		if (!item.unit_price || Number(item.unit_price) === 0) {
			item.unit_price = Number(material.reference_price || 0);
		}
		closeMaterialDropdown();
	}

	function onMaterialOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 24;
		if (nearBottom && materialOptionsHasMore && !materialOptionsLoading) {
			loadMaterialOptions({ reset: false });
		}
	}

	function formatAmount(amount: number) {
		return '¥' + (amount || 0).toFixed(2);
	}

	function grandTotal() {
		return form.items.reduce(
			(sum: number, item: any) => sum + Number(item.quantity || 0) * Number(item.unit_price || 0),
			0
		);
	}

	async function handleSubmit() {
		if (!form.supplier_code) {
			toast.error('请选择供应商');
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
			await api.post('/purchase/orders', {
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
			toast.success('创建成功');
			goto('/purchase/orders');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		closeMaterialDropdown();
		goto('/purchase/orders');
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeMaterialDropdown();
	}

	$effect(() => {
		if (materialDropdownOpenRow === null) {
			stopMaterialDropdownRAF();
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
			<div class="h-8 w-1.5 rounded-full bg-amber-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">新增采购订货</h1>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label" for="po-supplier"
						><span class="label-text flex items-center gap-2 font-medium"
							><Building2 size={14} /> 供应商 <span class="text-error">*</span></span
						></label
					>
					<select
						id="po-supplier"
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
					<label class="label" for="po-order-date"
						><span class="label-text flex items-center gap-2 font-medium"
							><Calendar size={14} /> 订单日期</span
						></label
					>
					<input
						id="po-order-date"
						type="date"
						bind:value={form.order_date}
						class="input input-bordered bg-base-200/40 h-11 w-full cursor-not-allowed text-base"
						readonly
					/>
				</div>
			</div>

			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label" for="po-expected"
						><span class="label-text font-medium"
							>预计到货日期 <span class="text-error">*</span></span
						></label
					>
					<input
						id="po-expected"
						type="date"
						bind:value={form.expected_date}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						min={todayDateString()}
						required
					/>
				</div>
				<div class="form-control">
					<label class="label" for="po-remark"
						><span class="label-text font-medium">备注</span></label
					>
					<input
						id="po-remark"
						type="text"
						bind:value={form.remark}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
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
					<table class="table-zebra table w-full text-base">
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
										<div class="relative" onclick={(e) => e.stopPropagation()} role="presentation">
											<input
												type="text"
												bind:value={item.material_label}
												class="input input-bordered bg-base-200/50 h-10 w-full min-w-[280px] text-base"
												placeholder="输入物料名称或编码搜索…"
												onfocus={(e) =>
													openMaterialDropdown(i, e.currentTarget as HTMLInputElement)}
												oninput={() => onMaterialInput(i)}
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
											class="input input-bordered bg-base-200/50 h-10 w-full max-w-[7rem] text-base"
											min="1"
											step="1"
										/>
									</td>
									<td class="text-base">{item.unit || '-'}</td>
									<td>
										<input
											type="number"
											bind:value={item.unit_price}
											class="input input-bordered bg-base-200/50 h-10 w-full max-w-[7.5rem] text-base"
											min="0"
											step="0.01"
										/>
									</td>
									<td>
										<span class="font-mono text-base font-semibold"
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
				{#if form.items.length > 0}
					<div class="mr-auto text-base">
						<span class="text-base-content/60">新增采购单总金额：</span>
						<span class="text-success font-mono text-lg font-semibold"
							>{formatAmount(grandTotal())}</span
						>
					</div>
				{/if}
				<button type="button" class="btn" onclick={goBack} disabled={submitting}>取消</button>
				<button type="button" class="btn btn-primary" onclick={handleSubmit} disabled={submitting}>
					{#if submitting}
						<span class="loading loading-spinner loading-sm"></span>
					{/if}
					提交订单
				</button>
			</div>
		</div>
	</div>
</div>

{#if materialDropdownOpenRow !== null}
	<div class="fixed inset-0 z-[70]" role="presentation" onclick={closeMaterialDropdown}></div>
	<div
		class="bg-base-100 border-base-300 fixed z-[80] overflow-hidden rounded-xl border shadow-2xl"
		style="left: {materialDropdownLeft}px; top: {materialDropdownTop}px; width: {materialDropdownWidth}px;"
		role="presentation"
		onclick={(e) => e.stopPropagation()}
	>
		<div
			class="text-base-content/50 border-base-200 flex items-center justify-between border-b px-3 py-2 text-xs"
		>
			<span>匹配物料 {materialOptions.length} / {materialOptionsTotal}</span>
			<button type="button" class="btn btn-xs btn-ghost" onclick={closeMaterialDropdown}
				>关闭</button
			>
		</div>
		<div
			class="overflow-auto"
			style="max-height: {materialDropdownListMaxHeight}px"
			onscroll={onMaterialOptionsScroll}
		>
			<div
				class="bg-base-200/80 border-base-200 sticky top-0 z-10 border-b px-3 py-2 backdrop-blur-sm"
			>
				<div
					class="grid w-full gap-3 text-[11px] font-medium"
					style:grid-template-columns={materialDropdownGridTemplate}
				>
					<div>物料名称</div>
					<div>编码</div>
					<div>单位</div>
				</div>
			</div>
			{#if materialOptions.length === 0 && !materialOptionsLoading}
				<div class="text-base-content/50 p-4 text-sm">无匹配物料</div>
			{:else}
				{#each materialOptions as materialOpt}
					<button
						type="button"
						class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
						onclick={() => selectMaterial(materialDropdownOpenRow as number, materialOpt)}
					>
						<div
							class="grid w-full items-center gap-3"
							style:grid-template-columns={materialDropdownGridTemplate}
						>
							<div class="min-w-0">
								<div class="text-sm font-medium whitespace-nowrap">
									{materialOpt.material_display_name || materialOpt.material_name || '-'}
								</div>
							</div>
							<div>
								<div class="font-mono text-xs whitespace-nowrap">
									{materialOpt.material_code || '-'}
								</div>
							</div>
							<div>
								<div class="text-xs whitespace-nowrap">
									{materialOpt.unit_name || materialOpt.unit || '-'}
								</div>
							</div>
						</div>
					</button>
				{/each}
			{/if}
			{#if materialOptionsLoading}
				<div class="text-base-content/50 p-3 text-xs">加载中...</div>
			{:else if materialOptionsHasMore}
				<div class="text-base-content/50 p-3 text-xs">下拉加载更多...</div>
			{/if}
		</div>
	</div>
{/if}

<!--
功能：新增销售退货页面
创建时间：2026-05-10
创建人：GPT-5.4
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Plus, Trash2, RotateCcw } from 'lucide-svelte';
	import { todayDateInCn } from '$lib/datetime';
	import { buildFloatingDropdownGridLayout, calcFloatingDropdownPlacement } from '$lib/dropdown';

	let submitting = $state(false);
	let customerOptions = $state<any[]>([]);
	let customerOptionsPage = $state(1);
	let customerOptionsHasMore = $state(true);
	let customerOptionsLoading = $state(false);
	let customerDropdownOpen = $state(false);
	let customerSearchTerm = $state('');
	let customerSearchValue = $state('');
	let customerSearchTimeout: ReturnType<typeof setTimeout> | null = null;

	let warehouseOptions = $state<any[]>([]);
	let warehouseOptionsPage = $state(1);
	let warehouseOptionsHasMore = $state(true);
	let warehouseOptionsLoading = $state(false);
	let warehouseDropdownOpen = $state(false);
	let warehouseSearchTerm = $state('');
	let warehouseSearchValue = $state('');
	let warehouseSearchTimeout: ReturnType<typeof setTimeout> | null = null;

	let materialOptions = $state<any[]>([]);
	let materialOptionMap = $state<Record<number, any>>({});
	let materialOptionsTotal = $state(0);
	let materialOptionsPage = $state(1);
	let materialOptionsHasMore = $state(true);
	let materialOptionsLoading = $state(false);
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
		customer_code: '',
		warehouse_code: '',
		return_date: todayDateInCn(),
		remark: '',
		items: [{ material_id: 0, material_label: '', quantity: 1 }]
	});

	function normalizeSearchTerm(value: string) {
		return String(value || '').trim();
	}

	function formatMoney(value: number) {
		const amount = Number(value) || 0;
		return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function selectedMaterial(id: number) {
		return materialOptionMap[Number(id || 0)];
	}

	function lineAmount(item: any) {
		const material = selectedMaterial(Number(item.material_id));
		return (Number(item.quantity) || 0) * Number(material?.last_purchase_price || 0);
	}

	function totalAmount() {
		return form.items.reduce((sum, item) => sum + lineAmount(item), 0);
	}

	async function addItem() {
		form.items = [...form.items, { material_id: 0, material_label: '', quantity: 1 }];
		await tick();
		window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' });
	}

	function removeItem(index: number) {
		form.items = form.items.filter((_, i) => i !== index);
		if (form.items.length === 0) addItem();
	}

	async function queryLookupList(
		baseUrl: string,
		page: number,
		pageSize: number,
		status: string,
		keyword: string,
		nameKey: string,
		codeKey: string
	) {
		let url = `${baseUrl}?page=${page}&page_size=${pageSize}`;
		if (status) url += `&status=${encodeURIComponent(status)}`;
		const q = normalizeSearchTerm(keyword);
		if (!q) {
			return await api.get(url);
		}

		const byName: any = await api.get(`${url}&${nameKey}=${encodeURIComponent(q)}`);
		const nameList = byName.list || [];
		if (page > 1 || nameList.length > 0) {
			return byName;
		}
		return await api.get(`${url}&${codeKey}=${encodeURIComponent(q)}`);
	}

	async function loadCustomerOptions(reset = false) {
		if (customerOptionsLoading) return;
		const nextPage = reset ? 1 : customerOptionsPage + 1;
		customerOptionsLoading = true;
		try {
			const res: any = await queryLookupList(
				'/base/customers',
				nextPage,
				20,
				'enabled',
				customerSearchTerm,
				'customer_name',
				'customer_code'
			);
			const list = res.list || [];
			const total = Number(res.total || 0);
			customerOptionsPage = nextPage;
			customerOptions = reset ? list : [...customerOptions, ...list];
			customerOptionsHasMore = customerOptions.length < total && list.length > 0;
		} catch (err: any) {
			toast.error('加载客户失败: ' + (err?.message || err));
		} finally {
			customerOptionsLoading = false;
		}
	}

	async function loadWarehouseOptions(reset = false) {
		if (warehouseOptionsLoading) return;
		const nextPage = reset ? 1 : warehouseOptionsPage + 1;
		warehouseOptionsLoading = true;
		try {
			const res: any = await queryLookupList(
				'/base/warehouses',
				nextPage,
				20,
				'enabled',
				warehouseSearchTerm,
				'warehouse_name',
				'warehouse_code'
			);
			const list = res.list || [];
			const total = Number(res.total || 0);
			warehouseOptionsPage = nextPage;
			warehouseOptions = reset ? list : [...warehouseOptions, ...list];
			warehouseOptionsHasMore = warehouseOptions.length < total && list.length > 0;
		} catch (err: any) {
			toast.error('加载仓库失败: ' + (err?.message || err));
		} finally {
			warehouseOptionsLoading = false;
		}
	}

	async function loadMaterialOptions(params: { reset: boolean }) {
		if (materialOptionsLoading) return;
		const nextPage = params.reset ? 1 : materialOptionsPage + 1;
		materialOptionsLoading = true;
		try {
			const res: any = await queryLookupList(
				'/base/materials',
				nextPage,
				30,
				'enabled',
				materialSearchTerm,
				'material_name',
				'material_code'
			);
			const list = res.list || [];
			const total = Number(res.total || 0);
			const nextMap = { ...materialOptionMap };
			for (const item of list) {
				nextMap[Number(item.id || 0)] = item;
			}
			materialOptionMap = nextMap;
			materialOptionsTotal = total;
			materialOptionsPage = nextPage;
			materialOptions = params.reset ? list : [...materialOptions, ...list];
			materialOptionsHasMore = materialOptions.length < total && list.length > 0;
		} catch (err: any) {
			toast.error('加载物料失败: ' + (err?.message || err));
		} finally {
			materialOptionsLoading = false;
		}
	}

	function openCustomerDropdown() {
		customerDropdownOpen = true;
		customerSearchTerm = normalizeSearchTerm(customerSearchValue);
		customerOptions = [];
		customerOptionsPage = 1;
		customerOptionsHasMore = true;
		loadCustomerOptions(true);
	}

	function closeCustomerDropdown() {
		customerDropdownOpen = false;
	}

	function onCustomerInput() {
		form.customer_code = '';
		customerSearchTerm = normalizeSearchTerm(customerSearchValue);
		if (customerSearchTimeout) clearTimeout(customerSearchTimeout);
		customerSearchTimeout = setTimeout(() => {
			customerOptions = [];
			customerOptionsPage = 1;
			customerOptionsHasMore = true;
			loadCustomerOptions(true);
		}, 250);
	}

	function selectCustomer(customer: any) {
		form.customer_code = customer.customer_code;
		customerSearchValue = customer.customer_name || customer.customer_code || '';
		closeCustomerDropdown();
	}

	function onCustomerOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 12;
		if (nearBottom && customerOptionsHasMore && !customerOptionsLoading) {
			loadCustomerOptions(false);
		}
	}

	function openWarehouseDropdown() {
		warehouseDropdownOpen = true;
		warehouseSearchTerm = normalizeSearchTerm(warehouseSearchValue);
		warehouseOptions = [];
		warehouseOptionsPage = 1;
		warehouseOptionsHasMore = true;
		loadWarehouseOptions(true);
	}

	function closeWarehouseDropdown() {
		warehouseDropdownOpen = false;
	}

	function onWarehouseInput() {
		form.warehouse_code = '';
		warehouseSearchTerm = normalizeSearchTerm(warehouseSearchValue);
		if (warehouseSearchTimeout) clearTimeout(warehouseSearchTimeout);
		warehouseSearchTimeout = setTimeout(() => {
			warehouseOptions = [];
			warehouseOptionsPage = 1;
			warehouseOptionsHasMore = true;
			loadWarehouseOptions(true);
		}, 250);
	}

	function selectWarehouse(warehouse: any) {
		form.warehouse_code = warehouse.warehouse_code;
		warehouseSearchValue = warehouse.warehouse_name || warehouse.warehouse_code || '';
		closeWarehouseDropdown();
	}

	function onWarehouseOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 12;
		if (nearBottom && warehouseOptionsHasMore && !warehouseOptionsLoading) {
			loadWarehouseOptions(false);
		}
	}

	function buildMaterialOptionLabel(material: any) {
		return material.material_display_name || material.material_name || material.material_code || '';
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
			minWidth: Math.max(560, layout.preferredPanelWidth),
			maxWidth: 1600,
			maxListHeight: 320,
			headerHeight: 44,
			preferBelowMinSpace: 204,
			extraWidth: 120,
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

	function openMaterialDropdown(index: number, anchor: HTMLInputElement) {
		materialDropdownOpenRow = index;
		materialDropdownAnchor = anchor;
		materialSearchTerm = normalizeSearchTerm(form.items[index]?.material_label || '');
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
		if (!item) return;
		item.material_id = 0;
		materialSearchTerm = normalizeSearchTerm(item.material_label || '');
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
				(current: any, currentIndex: number) =>
					currentIndex !== index && Number(current.material_id || 0) === Number(material.id || 0)
			)
		) {
			toast.warning('该物料已添加，不能重复');
			return;
		}
		materialOptionMap = {
			...materialOptionMap,
			[Number(material.id || 0)]: material
		};
		item.material_id = Number(material.id || 0);
		item.material_label = buildMaterialOptionLabel(material);
		closeMaterialDropdown();
	}

	function onMaterialOptionsScroll(e: Event) {
		const el = e.currentTarget as HTMLElement;
		if (!el) return;
		const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 12;
		if (nearBottom && materialOptionsHasMore && !materialOptionsLoading) {
			loadMaterialOptions({ reset: false });
		}
	}

	function onWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeCustomerDropdown();
			closeWarehouseDropdown();
			closeMaterialDropdown();
		}
	}

	$effect(() => {
		if (materialDropdownOpenRow === null) {
			stopMaterialDropdownRAF();
		}
	});

	async function handleSubmit() {
		if (!form.customer_code) {
			toast.error('请选择客户');
			return;
		}
		if (!form.warehouse_code) {
			toast.error('请选择退货入库仓库');
			return;
		}
		if (!form.return_date) {
			toast.error('请选择退货日期');
			return;
		}
		if (form.items.length === 0) {
			toast.error('请至少添加一条退货明细');
			return;
		}
		for (let i = 0; i < form.items.length; i++) {
			const item = form.items[i];
			if (!item.material_id) {
				toast.error(`第${i + 1}行：请选择物料`);
				return;
			}
			if (!item.quantity || Number(item.quantity) <= 0) {
				toast.error(`第${i + 1}行：请输入有效数量`);
				return;
			}
		}

		submitting = true;
		try {
			const res: any = await api.post('/returns', {
				return_type: 'sales_return',
				customer_code: form.customer_code,
				warehouse_code: form.warehouse_code,
				return_date: form.return_date,
				remark: form.remark,
				items: form.items.map((item) => ({
					material_id: Number(item.material_id),
					quantity: Number(item.quantity)
				}))
			});
			toast.success('销售退货创建成功');
			goto(`/sales/returns/${res.id}`);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		closeCustomerDropdown();
		closeWarehouseDropdown();
		closeMaterialDropdown();
		goto('/sales/returns');
	}
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-cyan-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">新增销售退货单</h1>
		</div>
		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>销售管理</li>
				<li><a class="text-primary" href="/sales/returns">销售退货单</a></li>
				<li>新增</li>
			</ul>
		</div>
	</div>

	<div class="flex flex-wrap items-center justify-between gap-3">
		<button type="button" class="btn btn-ghost btn-sm gap-1" onclick={goBack}>
			<ArrowLeft size={14} /> 返回列表
		</button>
		<div class="flex items-center gap-2">
			<button type="button" class="btn btn-sm" onclick={goBack} disabled={submitting}>取消</button>
			<button
				type="button"
				class="btn btn-sm btn-primary"
				onclick={handleSubmit}
				disabled={submitting}
			>
				保存退货单
			</button>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label" for="sales-return-customer"
						><span class="label-text font-medium">客户 <span class="text-error">*</span></span
						></label
					>
					<div class="relative z-20" onclick={(e) => e.stopPropagation()} role="presentation">
						<input
							id="sales-return-customer"
							type="text"
							bind:value={customerSearchValue}
							class="input input-bordered bg-base-200/50 h-11 w-full text-base"
							placeholder="输入客户名称或编码搜索..."
							onfocus={openCustomerDropdown}
							oninput={onCustomerInput}
						/>
						{#if customerDropdownOpen}
							<div
								class="fixed inset-0 z-[60]"
								role="presentation"
								onclick={closeCustomerDropdown}
							></div>
							<div
								class="bg-base-100 border-base-300 absolute top-full right-0 left-0 z-[70] mt-2 overflow-hidden rounded-xl border shadow-2xl"
							>
								<div class="max-h-72 overflow-auto" onscroll={onCustomerOptionsScroll}>
									{#if customerOptions.length === 0 && !customerOptionsLoading}
										<div class="text-base-content/50 p-4 text-center text-sm">未找到匹配客户</div>
									{:else}
										{#each customerOptions as customer}
											<button
												type="button"
												class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
												onclick={() => selectCustomer(customer)}
											>
												<div class="text-sm font-medium">{customer.customer_name || '-'}</div>
												<div class="text-base-content/60 font-mono text-xs">
													{customer.customer_code || '-'}
													{#if customer.contact_person}
														<span class="ml-2">{customer.contact_person}</span>
													{/if}
												</div>
											</button>
										{/each}
									{/if}
									{#if customerOptionsLoading}
										<div class="text-base-content/50 p-3 text-center text-xs">加载中...</div>
									{:else if customerOptionsHasMore}
										<div class="text-base-content/50 p-3 text-center text-xs">下拉加载更多...</div>
									{/if}
								</div>
							</div>
						{/if}
					</div>
				</div>
				<div class="form-control">
					<label class="label" for="sales-return-warehouse"
						><span class="label-text font-medium">退货仓库 <span class="text-error">*</span></span
						></label
					>
					<div class="relative z-10" onclick={(e) => e.stopPropagation()} role="presentation">
						<input
							id="sales-return-warehouse"
							type="text"
							bind:value={warehouseSearchValue}
							class="input input-bordered bg-base-200/50 h-11 w-full text-base"
							placeholder="输入仓库名称或编码搜索..."
							onfocus={openWarehouseDropdown}
							oninput={onWarehouseInput}
						/>
						{#if warehouseDropdownOpen}
							<div
								class="fixed inset-0 z-[50]"
								role="presentation"
								onclick={closeWarehouseDropdown}
							></div>
							<div
								class="bg-base-100 border-base-300 absolute top-full right-0 left-0 z-[55] mt-2 overflow-hidden rounded-xl border shadow-2xl"
							>
								<div class="max-h-72 overflow-auto" onscroll={onWarehouseOptionsScroll}>
									{#if warehouseOptions.length === 0 && !warehouseOptionsLoading}
										<div class="text-base-content/50 p-4 text-center text-sm">未找到匹配仓库</div>
									{:else}
										{#each warehouseOptions as warehouse}
											<button
												type="button"
												class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
												onclick={() => selectWarehouse(warehouse)}
											>
												<div class="text-sm font-medium">{warehouse.warehouse_name || '-'}</div>
												<div class="text-base-content/60 font-mono text-xs">
													{warehouse.warehouse_code || '-'}
												</div>
											</button>
										{/each}
									{/if}
									{#if warehouseOptionsLoading}
										<div class="text-base-content/50 p-3 text-center text-xs">加载中...</div>
									{:else if warehouseOptionsHasMore}
										<div class="text-base-content/50 p-3 text-center text-xs">下拉加载更多...</div>
									{/if}
								</div>
							</div>
						{/if}
					</div>
				</div>
				<div class="form-control">
					<label class="label" for="sales-return-date"
						><span class="label-text font-medium">退货日期 <span class="text-error">*</span></span
						></label
					>
					<input
						id="sales-return-date"
						type="date"
						bind:value={form.return_date}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
					/>
				</div>
				<div class="form-control xl:col-span-2">
					<label class="label" for="sales-return-remark"
						><span class="label-text font-medium">备注</span></label
					>
					<input
						id="sales-return-remark"
						type="text"
						bind:value={form.remark}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="退货说明"
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
								<th class="min-w-[320px]">物料名称</th>
								<th class="min-w-[100px]">数量</th>
								<th class="min-w-[120px] text-right">参考金额</th>
								<th class="w-14">操作</th>
							</tr>
						</thead>
						<tbody>
							{#each form.items as item, index}
								<tr>
									<td>{index + 1}</td>
									<td>
										<div class="relative" onclick={(e) => e.stopPropagation()} role="presentation">
											<input
												type="text"
												bind:value={item.material_label}
												class="input input-bordered bg-base-200/50 h-10 w-full text-base"
												placeholder="输入物料名称或编码搜索..."
												onfocus={(e) =>
													openMaterialDropdown(index, e.currentTarget as HTMLInputElement)}
												oninput={() => onMaterialInput(index)}
											/>
										</div>
									</td>
									<td>
										<input
											type="number"
											bind:value={item.quantity}
											min="0.001"
											step="0.001"
											class="input input-bordered bg-base-200/50 h-10 w-full text-base"
										/>
									</td>
									<td class="text-success text-right font-mono font-semibold">
										¥{formatMoney(lineAmount(item))}
									</td>
									<td>
										<button
											type="button"
											class="btn btn-xs btn-ghost text-error"
											onclick={() => removeItem(index)}
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
				<div class="mr-auto text-base">
					<span class="text-base-content/60">退货金额参考合计：</span>
					<span class="text-success font-mono text-lg font-semibold"
						>¥{formatMoney(totalAmount())}</span
					>
				</div>
				<button type="button" class="btn" onclick={goBack} disabled={submitting}>取消</button>
				<button type="button" class="btn btn-primary" onclick={handleSubmit} disabled={submitting}>
					保存销售退货单
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
				<div class="text-base-content/50 p-4 text-sm">未找到匹配物料</div>
			{:else}
				{#each materialOptions as material}
					<button
						type="button"
						class="hover:bg-base-200/60 border-base-200 w-full border-b px-3 py-2.5 text-left last:border-b-0"
						onclick={() => selectMaterial(materialDropdownOpenRow as number, material)}
					>
						<div
							class="grid w-full items-center gap-3"
							style:grid-template-columns={materialDropdownGridTemplate}
						>
							<div class="min-w-0">
								<div class="text-sm font-medium whitespace-nowrap">
									{material.material_display_name || material.material_name || '-'}
								</div>
							</div>
							<div>
								<div class="font-mono text-xs whitespace-nowrap">
									{material.material_code || '-'}
								</div>
							</div>
							<div>
								<div class="text-xs whitespace-nowrap">
									{material.unit_name || material.unit || '-'}
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

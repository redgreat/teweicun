<!--
功能：新增销售订单页面
创建时间：2026-05-10
创建人：GPT-5.4
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Plus, Trash2, ShoppingCart } from 'lucide-svelte';
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

	let materialOptions = $state<any[]>([]);
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
	let materialDropdownGridTemplate = $state('minmax(18rem, 1fr) 12rem 6rem');
	let materialSearchTerm = $state('');
	let materialSearchTimeout: ReturnType<typeof setTimeout> | null = null;
	let materialDropdownRAF: number | null = null;
	const defaultOrderDate = todayDateInCn();
	const defaultDeliveryDate = addDaysToDateString(defaultOrderDate, 7);
	let lastAutoDeliveryDate = $state(defaultDeliveryDate);
	let lastOrderDate = $state(defaultOrderDate);

	let form = $state({
		customer_code: '',
		order_date: defaultOrderDate,
		delivery_date: defaultDeliveryDate,
		contract_no: '',
		payment_method: '',
		receiver_name: '',
		receiver_phone: '',
		receiver_address: '',
		remark: '',
		items: [
			{
				material_id: 0,
				material_label: '',
				quantity: 1,
				unit_price: 0,
				unit: '',
				remark: ''
			}
		]
	});

	function normalizeSearchTerm(value: string) {
		return String(value || '').trim();
	}

	function addDaysToDateString(value: string, days: number) {
		const parts = String(value || '')
			.split('-')
			.map((part) => Number(part));
		if (parts.length !== 3 || parts.some((part) => Number.isNaN(part))) {
			return '';
		}
		const [year, month, day] = parts;
		const date = new Date(Date.UTC(year, month - 1, day));
		date.setUTCDate(date.getUTCDate() + days);
		return date.toISOString().slice(0, 10);
	}

	function formatMoney(value: number) {
		const amount = Number(value) || 0;
		return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function lineAmount(item: any) {
		return (Number(item.quantity) || 0) * (Number(item.unit_price) || 0);
	}

	function totalAmount() {
		return form.items.reduce((sum, item) => sum + lineAmount(item), 0);
	}

	async function addItem() {
		form.items = [
			...form.items,
			{
				material_id: 0,
				material_label: '',
				quantity: 1,
				unit_price: 0,
				unit: '',
				remark: ''
			}
		];
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

	function syncCustomerDefaults(customer: any) {
		if (!customer) return;
		if (!form.receiver_name) form.receiver_name = customer.contact_person || '';
		if (!form.receiver_phone) form.receiver_phone = customer.contact_phone || '';
		if (!form.receiver_address) form.receiver_address = customer.address || '';
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
		syncCustomerDefaults(customer);
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

	function buildMaterialOptionLabel(material: any) {
		return material.material_display_name || material.material_name || material.material_code || '';
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

	function updateMaterialDropdownPlacement(anchor: HTMLInputElement) {
		materialDropdownAnchor = anchor;
		const layout = buildFloatingDropdownGridLayout({
			firstColumnTexts: materialOptions.map((material) =>
				String(material.material_display_name || material.material_name || '-')
			),
			fixedColumnWidths: [192, 96],
			minFirstColumnWidth: 260
		});
		const placement = calcFloatingDropdownPlacement({
			anchor,
			minWidth: Math.max(560, layout.preferredPanelWidth),
			maxWidth: 1600,
			maxListHeight: 320,
			headerHeight: 0,
			preferBelowMinSpace: 180,
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
			updateMaterialDropdownPlacement(materialDropdownAnchor);
			materialDropdownRAF = requestAnimationFrame(loop);
		};
		materialDropdownRAF = requestAnimationFrame(loop);
	}

	function stopMaterialDropdownRAF() {
		if (!materialDropdownRAF) return;
		cancelAnimationFrame(materialDropdownRAF);
		materialDropdownRAF = null;
	}

	function openMaterialDropdown(index: number, anchor: HTMLInputElement) {
		materialDropdownOpenRow = index;
		updateMaterialDropdownPlacement(anchor);
		materialSearchTerm = normalizeSearchTerm(form.items[index]?.material_label || '');
		materialOptions = [];
		materialOptionsTotal = 0;
		materialOptionsPage = 1;
		materialOptionsHasMore = true;
		loadMaterialOptions({ reset: true });
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
		item.unit = '';
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
		item.material_id = Number(material.id || 0);
		item.material_label = buildMaterialOptionLabel(material);
		item.unit = material.unit_name || material.unit || '';
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
			closeMaterialDropdown();
		}
	}

	$effect(() => {
		const orderDate = form.order_date;
		if (orderDate === lastOrderDate) {
			return;
		}
		const nextAutoDeliveryDate = addDaysToDateString(orderDate, 7);
		if (!form.delivery_date || form.delivery_date === lastAutoDeliveryDate) {
			form.delivery_date = nextAutoDeliveryDate;
		}
		lastOrderDate = orderDate;
		lastAutoDeliveryDate = nextAutoDeliveryDate;
	});

	function onWindowClick() {
		closeCustomerDropdown();
		closeMaterialDropdown();
	}

	function onWindowResize() {
		if (materialDropdownOpenRow !== null && materialDropdownAnchor) {
			updateMaterialDropdownPlacement(materialDropdownAnchor);
		}
	}

	async function handleSubmit() {
		if (!form.customer_code) {
			toast.error('请选择客户');
			return;
		}
		if (!form.order_date) {
			toast.error('请选择下单日期');
			return;
		}
		if (form.items.length === 0) {
			toast.error('请至少添加一条销售明细');
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
			if (!item.unit_price || Number(item.unit_price) <= 0) {
				toast.error(`第${i + 1}行：请输入有效单价`);
				return;
			}
		}

		submitting = true;
		try {
			const res: any = await api.post('/sales/orders', {
				...form,
				items: form.items.map((item) => ({
					material_id: Number(item.material_id),
					quantity: Number(item.quantity),
					unit_price: Number(item.unit_price),
					remark: item.remark || ''
				}))
			});
			toast.success('销售订单创建成功');
			goto(`/sales/orders/${res.id}`);
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		closeCustomerDropdown();
		closeMaterialDropdown();
		goto('/sales/orders');
	}
</script>

<svelte:window onkeydown={onWindowKeydown} onclick={onWindowClick} onresize={onWindowResize} />

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-emerald-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">新增销售订单</h1>
		</div>
		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>销售管理</li>
				<li><a class="text-primary" href="/sales/orders">销售订单</a></li>
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
				保存订单
			</button>
		</div>
	</div>

	<div class="card bg-base-100 border-base-300 border shadow-lg">
		<div class="card-body space-y-6">
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
				<div class="form-control">
					<label class="label" for="sales-order-customer">
						<span class="label-text font-medium">客户 <span class="text-error">*</span></span>
					</label>
					<div class="relative z-20" onclick={(e) => e.stopPropagation()} role="presentation">
						<input
							id="sales-order-customer"
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
					<label class="label" for="sales-order-date">
						<span class="label-text font-medium">下单日期 <span class="text-error">*</span></span>
					</label>
					<input
						id="sales-order-date"
						type="date"
						bind:value={form.order_date}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
					/>
				</div>
				<div class="form-control">
					<label class="label" for="sales-delivery-date">
						<span class="label-text font-medium">交付日期</span>
					</label>
					<input
						id="sales-delivery-date"
						type="date"
						bind:value={form.delivery_date}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
					/>
				</div>
				<div class="form-control">
					<label class="label" for="sales-contract-no">
						<span class="label-text font-medium">合同编号</span>
					</label>
					<input
						id="sales-contract-no"
						type="text"
						bind:value={form.contract_no}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="如有合同可填写"
					/>
				</div>
				<div class="form-control">
					<label class="label" for="sales-payment-method">
						<span class="label-text font-medium">付款方式</span>
					</label>
					<input
						id="sales-payment-method"
						type="text"
						bind:value={form.payment_method}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="例如：月结 / 现款"
					/>
				</div>
				<div class="form-control">
					<label class="label" for="sales-receiver-name">
						<span class="label-text font-medium">收货联系人</span>
					</label>
					<input
						id="sales-receiver-name"
						type="text"
						bind:value={form.receiver_name}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
					/>
				</div>
				<div class="form-control">
					<label class="label" for="sales-receiver-phone">
						<span class="label-text font-medium">联系电话</span>
					</label>
					<input
						id="sales-receiver-phone"
						type="text"
						bind:value={form.receiver_phone}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
					/>
				</div>
				<div class="form-control xl:col-span-2">
					<label class="label" for="sales-receiver-address">
						<span class="label-text font-medium">收货地址</span>
					</label>
					<input
						id="sales-receiver-address"
						type="text"
						bind:value={form.receiver_address}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
					/>
				</div>
				<div class="form-control xl:col-span-2">
					<label class="label" for="sales-remark">
						<span class="label-text font-medium">备注</span>
					</label>
					<input
						id="sales-remark"
						type="text"
						bind:value={form.remark}
						class="input input-bordered bg-base-200/50 h-11 w-full text-base"
						placeholder="订单备注"
					/>
				</div>
			</div>

			<div class="divider">销售明细</div>

			<div class="flex flex-wrap items-center justify-end gap-2">
				<button type="button" class="btn btn-sm btn-primary" onclick={addItem}>
					<Plus size={16} /> 添加明细
				</button>
			</div>

			{#if form.items.length === 0}
				<div class="text-base-content/50 py-10 text-center">
					<ShoppingCart size={48} class="mx-auto mb-4 opacity-30" />
					<div>暂无销售明细，请点击「添加明细」</div>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-zebra table w-full text-base">
						<thead>
							<tr>
								<th class="w-10">#</th>
								<th class="min-w-[320px]">物料名称</th>
								<th class="min-w-[100px]">数量</th>
								<th class="min-w-[120px]">销售单价</th>
								<th class="min-w-[120px] text-right">金额</th>
								<th class="min-w-[180px]">备注</th>
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
												class="input input-bordered bg-base-200/50 h-10 w-full min-w-[280px] text-base"
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
									<td>
										<input
											type="number"
											bind:value={item.unit_price}
											min="0.01"
											step="0.01"
											class="input input-bordered bg-base-200/50 h-10 w-full text-base"
										/>
									</td>
									<td class="text-success text-right font-mono font-semibold">
										¥{formatMoney(lineAmount(item))}
									</td>
									<td>
										<input
											type="text"
											bind:value={item.remark}
											class="input input-bordered bg-base-200/50 h-10 w-full text-base"
											placeholder="行备注"
										/>
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
					<span class="text-base-content/60">订单金额合计：</span>
					<span class="text-success font-mono text-lg font-semibold"
						>¥{formatMoney(totalAmount())}</span
					>
				</div>
				<button type="button" class="btn" onclick={goBack} disabled={submitting}>取消</button>
				<button type="button" class="btn btn-primary" onclick={handleSubmit} disabled={submitting}>
					保存销售订单
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
			{:else if materialOptions.length > 0}
				<div class="text-base-content/45 bg-base-100 border-base-200 border-t px-3 py-2 text-xs">
					匹配物料 {materialOptions.length} / {materialOptionsTotal}
				</div>
			{/if}
		</div>
	</div>
{/if}

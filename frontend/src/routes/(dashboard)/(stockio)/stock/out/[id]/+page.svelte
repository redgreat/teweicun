<!--
功能：出库单详情页面
创建时间：2026-04-22
创建人：Trae
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import {
		ArrowLeft,
		FileText,
		Calendar,
		BadgeCheck,
		User,
		CheckCircle,
		QrCode,
		Wand2,
		X,
		Hash
	} from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { formatDateInCn, formatDateTimeInCn } from '$lib/datetime';

	let loading = $state(true);
	let confirming = $state(false);
	let detail = $state<any | null>(null);
	let selectionSaving = $state(false);
	let pickerLoading = $state(false);
	let pickerOpen = $state(false);
	let pickerItem: any = $state(null);
	let pickerOptions = $state<any[]>([]);
	let pickerSelectedOnly = $state(false);
	let pickerSearchTerm = $state('');
	let manualSelections = $state<Record<number, number[]>>({});
	let insufficientPreparedItems = $state<Record<number, boolean>>({});

	let serialModal = $state({ show: false, items: [] as any[], loading: false, title: '' });

	const confirmMode = $derived($page.url.searchParams.get('mode') === 'confirm');
	const canConfirm = $derived(confirmMode && detail?.status !== 'confirmed');

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function lineAmount(it: any) {
		return (Number(it?.quantity) || 0) * (Number(it?.unit_cost) || 0);
	}

	function grandTotal() {
		return detail?.total_amount || 0;
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatDateTime(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateTimeInCn(dateStr);
	}

	function statusBadgeClass(status: string) {
		if (status === 'pending') return 'badge-warning';
		if (status === 'confirmed') return 'badge-success';
		if (status === 'draft') return 'badge-ghost';
		return 'badge-ghost';
	}

	function statusName(status: string) {
		const map: Record<string, string> = {
			pending: '待出库',
			confirmed: '已完成',
			draft: '待出库'
		};
		return map[status] || status || '-';
	}

	function outTypeName(outType: string) {
		const map: Record<string, string> = {
			purchase_return: '采购退货',
			sales: '销售出库',
			consumption: '领料出库',
			production: '生产领料',
			transfer: '调拨出库',
			other: '其他'
		};
		return map[outType] || outType || '-';
	}

	function businessDocLabel(detail: any) {
		if (detail?.business_doc_type === 'purchase_return') return '采购退货单号';
		if (detail?.business_doc_type === 'consumption_order') return '领料单号';
		return '业务单号';
	}

	function businessDocHref(detail: any) {
		if (!detail?.business_doc_id || !detail?.business_doc_type) return '';
		if (detail.business_doc_type === 'purchase_return')
			return `/purchase/return/${detail.business_doc_id}`;
		if (detail.business_doc_type === 'consumption_order')
			return `/consumption/orders/${detail.business_doc_id}`;
		return '';
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			const res: any = await api.get(`/stock-out/${id}`);
			detail = res;
		} catch (err: any) {
			toast.error('加载出库单详情失败: ' + (err?.message || err));
			detail = null;
		} finally {
			loading = false;
		}
	}

	async function confirmStockOut() {
		if (!detail?.id) return;
		confirming = true;
		try {
			await api.post(`/stock-out/${detail.id}/confirm`);
			toast.success('出库确认成功，库存已扣减');
			await loadDetail(detail.id);
		} catch (err: any) {
			toast.error('确认失败: ' + (err?.response?.data?.msg || err?.message || err));
		} finally {
			confirming = false;
		}
	}

	function getNeedCount(item: any) {
		return Math.floor(Number(item?.quantity || 0));
	}

	function codedItems() {
		return (detail?.items || []).filter((it: any) => it.is_code);
	}

	function hasSerialCodes(item: any) {
		return Boolean(item?.is_code);
	}

	function selectedCount(itemId: number) {
		return (manualSelections[itemId] || []).length;
	}

	function refreshInsufficientPreparedFlags() {
		const next: Record<number, boolean> = {};
		for (const item of codedItems()) {
			next[item.id] = selectedCount(item.id) < getNeedCount(item);
		}
		insufficientPreparedItems = next;
	}

	async function loadSelectionsForAllItems() {
		const next: Record<number, number[]> = {};
		for (const item of codedItems()) {
			try {
				const codes: any = await api.get(`/sku-serial/stock-out-item/${item.id}`);
				const codeList = Array.isArray(codes) ? codes : [];
				next[item.id] = codeList.filter((c: any) => c.selected).map((c: any) => Number(c.id));
			} catch (e) {
				console.error('Failed to load selections for item', item.id, e);
				next[item.id] = [];
			}
		}
		manualSelections = next;
		refreshInsufficientPreparedFlags();
	}

	async function openSerialPicker(item: any) {
		if (!canConfirm) return;
		if (detail?.status !== 'pending') return;
		if (!item?.is_code) return;
		pickerItem = item;
		pickerSelectedOnly = false;
		pickerSearchTerm = '';
		pickerOpen = true;
		pickerLoading = true;
		try {
			const options: any[] = await api.get(`/sku-serial/stock-out-item/${item.id}/available`);
			pickerOptions = options || [];
			// 每次打开都以“后端已备货”为准重置本地选择，避免残留导致勾选异常
			manualSelections[item.id] = (options || [])
				.filter((it) => it.selected)
				.map((it) => Number(it.id));
		} catch (err: any) {
			toast.error('加载可选编码失败: ' + (err?.response?.data?.msg || err?.message || err));
		} finally {
			pickerLoading = false;
		}
	}

	function toggleCode(codeId: number, checked: boolean) {
		if (!canConfirm) return;
		if (!pickerItem?.id) return;
		const itemId = pickerItem.id as number;
		const need = getNeedCount(pickerItem);
		const cid = Number(codeId);
		const current = (manualSelections[itemId] || []).map((x) => Number(x));
		if (checked) {
			if (current.length >= need) return;
			if (current.includes(cid)) return;
			manualSelections[itemId] = [...current, cid];
		} else {
			manualSelections[itemId] = current.filter((id) => Number(id) !== cid);
		}
		refreshInsufficientPreparedFlags();
	}

	function sortedPickerOptions() {
		const selectedSet = new Set((manualSelections[pickerItem?.id] || []).map((x) => Number(x)));
		const q = String(pickerSearchTerm || '')
			.trim()
			.toLowerCase();
		const sorted = [...pickerOptions].sort((a, b) => {
			const aId = Number(a?.id);
			const bId = Number(b?.id);
			const aSelected = selectedSet.has(aId) ? 1 : 0;
			const bSelected = selectedSet.has(bId) ? 1 : 0;
			if (aSelected !== bSelected) return bSelected - aSelected;
			return String(a.serial_code || '').localeCompare(String(b.serial_code || ''), 'zh-CN');
		});
		const filtered = q
			? sorted.filter((code) =>
					String(code?.serial_code || '')
						.toLowerCase()
						.includes(q)
				)
			: sorted;
		if (!pickerSelectedOnly) return filtered;
		return filtered.filter((code) => selectedSet.has(Number(code.id)));
	}

	function autoPickCurrentItem() {
		if (!pickerItem?.id) return;
		const need = getNeedCount(pickerItem);
		const next: number[] = [];
		for (const opt of sortedPickerOptions()) {
			if (next.length >= need) break;
			next.push(Number(opt.id));
		}
		manualSelections[pickerItem.id] = next;
		refreshInsufficientPreparedFlags();
	}

	async function saveCurrentItemSelections() {
		if (!canConfirm) return;
		if (!pickerItem?.id) return;
		const need = getNeedCount(pickerItem);
		const sel = (manualSelections[pickerItem.id] || []).map((x) => Number(x));
		if (sel.length !== need) {
			toast.error(`编码未选齐：需 ${need} 个，当前 ${sel.length} 个`);
			return;
		}
		selectionSaving = true;
		try {
			await api.put(`/stock-out-item/${pickerItem.id}/serial-selections`, {
				serial_code_ids: sel
			});
			toast.success('备货编码已保存');
			pickerOpen = false;
			await loadSelectionsForAllItems();
		} catch (err: any) {
			toast.error('保存备货失败: ' + (err?.response?.data?.msg || err?.message || err));
		} finally {
			selectionSaving = false;
		}
	}

	async function clearAllSelectionsForPickerItem() {
		if (!canConfirm) return;
		if (!pickerItem?.id) return;
		const itemId = pickerItem.id as number;
		selectionSaving = true;
		try {
			manualSelections[itemId] = [];
			await api.put(`/stock-out-item/${itemId}/serial-selections`, { serial_code_ids: [] });
			await loadSelectionsForAllItems();
			toast.success('已清空该行全部备货编码');
		} catch (err: any) {
			toast.error('清空失败: ' + (err?.response?.data?.msg || err?.message || err));
		} finally {
			selectionSaving = false;
		}
	}

	async function confirmWithSelections() {
		if (!canConfirm) return;
		if (!detail?.id) return;
		for (const item of codedItems()) {
			if (selectedCount(item.id) !== getNeedCount(item)) {
				toast.error(`${item.material_name} 需备货 ${getNeedCount(item)} 个编码`);
				return;
			}
		}
		await confirmStockOut();
	}

	async function viewSerialCodes(item: any) {
		if (!item.id) return;
		serialModal = {
			...serialModal,
			show: true,
			loading: true,
			title: `${item.material_name} - 编码详情`,
			items: []
		};
		try {
			const res: any = await api.get(`/sku-serial/stock-out-item/${item.id}`);
			serialModal = { ...serialModal, loading: false, items: res || [] };
		} catch (err: any) {
			toast.error('加载编码详情失败: ' + (err?.message || err));
			serialModal = { ...serialModal, loading: false };
		}
	}

	function closeSerialModal() {
		serialModal = { show: false, items: [], loading: false, title: '' };
	}

	function getStatusBadge(status: string) {
		const map: Record<string, string> = {
			in_stock: 'badge-success',
			issued: 'badge-warning',
			returned: 'badge-info',
			scrapped: 'badge-error'
		};
		return map[status] || 'badge-ghost';
	}

	function getStatusLabel(status: string) {
		const map: Record<string, string> = {
			in_stock: '在库',
			issued: '已领用',
			returned: '已退回',
			scrapped: '已报废'
		};
		return map[status] || status;
	}

	onMount(() => {
		const id = Number($page.params.id);
		if (!id) {
			toast.error('无效的出库单ID');
			loading = false;
			return;
		}
		loadDetail(id).then(async () => {
			if (codedItems().length > 0) {
				await loadSelectionsForAllItems();
			}
			// 进入详情页不触发任何确认动作；仅点击右上角“确认出库”才会提交。
		});
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-orange-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">{confirmMode ? '出库单确认' : '出库单详情'}</h1>
		</div>

		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>出入库管理</li>
				<li><a class="text-primary" href="/stock/out">出库管理</a></li>
				<li>{confirmMode ? '确认' : '详情'}</li>
			</ul>
		</div>
	</div>

	<div class="flex items-center justify-between gap-3">
		<a href="/stock/out" class="btn btn-ghost btn-sm gap-1">
			<ArrowLeft size={14} /> 返回列表
		</a>
		{#if detail}
			{#if confirmMode}
				<div class="flex items-center gap-2">
					{#if canConfirm}
						<button
							class="btn btn-sm btn-success gap-2 text-white"
							onclick={confirmWithSelections}
							disabled={confirming || selectionSaving}
						>
							<CheckCircle size={16} /> 确认出库
						</button>
					{/if}
				</div>
			{:else}
				<div class="flex items-center gap-2">
					<span class="badge badge-sm {statusBadgeClass(detail.status)}"
						>{statusName(detail.status)}</span
					>
					<span class="text-base-content/60 font-mono text-base">{detail.stock_out_no}</span>
				</div>
			{/if}
		{/if}
	</div>

	{#if loading}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center"
		>
			正在加载...
		</div>
	{:else if !detail}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center"
		>
			未找到出库单信息
		</div>
	{:else}
		<div class="bg-base-100 border-base-300 space-y-5 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="flex items-center gap-2 text-base font-semibold">
					<FileText size={16} /> 单据信息
				</div>
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<FileText size={14} /> 出库单号
					</div>
					<div class="mt-1 font-mono text-base">{detail.stock_out_no}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<BadgeCheck size={14} /> 类型
					</div>
					<div class="mt-1 text-base">{outTypeName(detail.out_type)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<Calendar size={14} /> 出库日期
					</div>
					<div class="mt-1 text-base">{formatDate(detail.stock_out_date)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<User size={14} /> 收货方
					</div>
					<div class="mt-1 text-base">{detail.receiver || '-'}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<FileText size={14} />
						{businessDocLabel(detail)}
					</div>
					<div class="mt-1 font-mono text-base">
						{#if businessDocHref(detail) && detail.business_doc_no}
							<a
								class="link link-primary font-mono no-underline hover:underline"
								href={businessDocHref(detail)}
							>
								{detail.business_doc_no}
							</a>
						{:else}
							-
						{/if}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 text-base">总计金额</div>
					<div class="text-success mt-1 font-mono text-base font-semibold">
						¥{formatMoney(grandTotal())}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 text-base">创建时间</div>
					<div class="mt-1 text-base">{formatDateTime(detail.created_at)}</div>
				</div>

				{#if detail.confirmed_at}
					<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
						<div class="text-base-content/50 text-base">确认时间</div>
						<div class="mt-1 text-base">{formatDateTime(detail.confirmed_at)}</div>
					</div>
				{/if}

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-4">
					<div class="text-base-content/50 text-base">备注</div>
					<div class="mt-1 text-base break-words whitespace-pre-wrap">{detail.remark || '-'}</div>
				</div>
			</div>
		</div>

		<div class="bg-base-100 border-base-300 space-y-4 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="text-base font-semibold">出库明细</div>
				<div class="text-base-content/50 text-base">共 {detail.items?.length || 0} 行</div>
			</div>

			<div class="overflow-x-auto">
				<table class="table-zebra table w-full table-fixed text-base">
					<thead>
						<tr>
							<th class="w-[34%] min-w-[320px]">物料名称</th>
							<th class="min-w-[160px]">出库仓库</th>
							<th class="w-24 whitespace-nowrap">属性</th>
							<th class="min-w-[120px]">数量</th>
							<th class="min-w-[100px]">单位</th>
							<th class="min-w-[110px] text-right">单价</th>
							<th class="min-w-[120px] text-right">金额</th>
							<th class="min-w-[160px]">编码</th>
						</tr>
					</thead>
					<tbody>
						{#each detail.items || [] as it}
							<tr
								class="hover:bg-base-200/50 transition-colors {detail.status === 'pending' &&
								insufficientPreparedItems[it.id]
									? 'bg-error/10'
									: ''}"
							>
								<td>
									<div class="space-y-1 pr-3">
										<div class="group/line relative">
											<div class="truncate font-semibold">
												{it.material_name || '-'}
												{#if it.material_code}
													<span class="text-base-content/50 ml-2 font-mono text-base"
														>{it.material_code}</span
													>
												{/if}
											</div>
											<div
												class="border-base-300 bg-base-100 pointer-events-none absolute top-full left-0 z-20 mt-1 hidden max-w-[42rem] rounded-md border px-2 py-1 text-sm leading-5 whitespace-normal shadow-xl group-hover/line:block"
											>
												{it.material_name || '-'}
												{#if it.material_code}
													<span class="text-base-content/60 ml-2 font-mono">{it.material_code}</span
													>
												{/if}
											</div>
										</div>
										<div class="text-base-content/60 truncate text-sm">
											属性 {(it.custom_attributes || []).length} 项
										</div>
									</div>
								</td>
								<td>{it.warehouse_name || it.warehouse_code || detail.warehouse_name || '-'}</td>
								<td class="whitespace-nowrap">
									<span class="text-sm">{(it.custom_attributes || []).length}项</span>
								</td>
								<td class="font-mono">{it.quantity ?? 0}</td>
								<td>{it.unit || '-'}</td>
								<td class="text-right font-mono">
									{#if it.inventory_id}
										¥{formatMoney(it.unit_cost)}
									{:else}
										—
									{/if}
								</td>
								<td class="text-success text-right font-mono font-semibold">
									{#if it.inventory_id}
										¥{formatMoney(lineAmount(it))}
									{:else}
										—
									{/if}
								</td>
								<td>
									{#if !hasSerialCodes(it)}
										<span class="text-base-content/40">-</span>
									{:else if canConfirm && detail.status === 'pending'}
										<button
											type="button"
											class="badge badge-sm {selectedCount(it.id) === getNeedCount(it)
												? 'badge-success'
												: 'badge-warning'} cursor-pointer"
											onclick={() => openSerialPicker(it)}
										>
											{selectedCount(it.id)} / {getNeedCount(it)}
										</button>
									{:else}
										<button
											class="btn btn-sm btn-ghost text-primary"
											onclick={() => viewSerialCodes(it)}
											title="查看实际出库编码"
										>
											<QrCode size={16} />
										</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

{#if pickerOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm"
		role="button"
		tabindex="0"
		onclick={() => (pickerOpen = false)}
		onkeydown={(e) => e.key === 'Escape' && (pickerOpen = false)}
	>
		<div
			class="bg-base-100 border-base-300 flex max-h-[80vh] w-full max-w-2xl flex-col rounded-2xl border shadow-2xl"
			role="button"
			tabindex="0"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (pickerOpen = false)}
			transition:fade
		>
			<div class="border-base-200 flex items-center justify-between border-b p-5">
				<h3 class="text-lg font-bold">{pickerItem?.material_name || ''} - 选择编码</h3>
				<div class="flex items-center gap-2">
					<button
						class="btn btn-sm btn-ghost gap-1"
						onclick={autoPickCurrentItem}
						disabled={pickerLoading}
					>
						<Wand2 size={16} /> 自动选择
					</button>
					<button
						class="btn btn-sm btn-ghost text-error"
						onclick={clearAllSelectionsForPickerItem}
						disabled={pickerLoading || selectionSaving}
					>
						清空已选
					</button>
					<button
						class="btn btn-sm btn-primary text-white"
						onclick={saveCurrentItemSelections}
						disabled={pickerLoading || selectionSaving}
					>
						保存备货
					</button>
					<button class="btn btn-sm btn-ghost btn-circle" onclick={() => (pickerOpen = false)}>
						<X size={18} />
					</button>
				</div>
			</div>
			<div class="flex-1 overflow-y-auto p-5">
				{#if pickerLoading}
					<div class="py-10 text-center">
						<span class="loading loading-spinner loading-lg"></span>
					</div>
				{:else}
					<div class="mb-3 flex items-center justify-between gap-3">
						<div class="text-base-content/60 text-xs">
							需选择 {pickerItem ? getNeedCount(pickerItem) : 0} 个，当前已选 {pickerItem
								? selectedCount(pickerItem.id)
								: 0} 个
						</div>
						<div class="flex items-center gap-3">
							<input
								type="text"
								bind:value={pickerSearchTerm}
								placeholder="搜索/扫码输入编码..."
								class="input input-bordered input-xs bg-base-200/50 w-56"
							/>
							<label class="label cursor-pointer gap-2 p-0">
								<span class="label-text text-xs">只看已选</span>
								<input
									type="checkbox"
									class="toggle toggle-xs"
									checked={pickerSelectedOnly}
									onchange={(e) =>
										(pickerSelectedOnly = (e.currentTarget as HTMLInputElement).checked)}
								/>
							</label>
						</div>
					</div>
					<div class="space-y-2">
						{#each sortedPickerOptions() as code (Number(code?.id))}
							<label
								class="border-base-300 flex cursor-pointer items-center gap-3 rounded-lg border p-3"
							>
								<input
									type="checkbox"
									class="checkbox checkbox-sm"
									checked={(manualSelections[pickerItem?.id] || [])
										.map((x) => Number(x))
										.includes(Number(code?.id))}
									onchange={(e) =>
										toggleCode(Number(code?.id), (e.currentTarget as HTMLInputElement).checked)}
								/>
								<span class="font-mono text-sm">{code.serial_code}</span>
							</label>
						{/each}
						{#if sortedPickerOptions().length === 0}
							<div class="text-base-content/50 py-6 text-center text-sm">暂无可显示编码</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

{#if serialModal.show}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm"
		role="button"
		tabindex="0"
		onclick={closeSerialModal}
		onkeydown={(e) => e.key === 'Escape' && closeSerialModal()}
	>
		<div
			class="bg-base-100 border-base-300 flex max-h-[80vh] w-full max-w-2xl flex-col rounded-2xl border shadow-2xl"
			role="button"
			tabindex="0"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeSerialModal()}
			transition:fade
		>
			<div class="border-base-200 flex items-center justify-between border-b p-5">
				<div class="flex items-center gap-2">
					<Hash size={18} class="text-primary" />
					<h3 class="text-lg font-bold">编码详情</h3>
					<span class="text-base-content/50 ml-2 text-sm">{serialModal.title}</span>
				</div>
				<button class="btn btn-sm btn-ghost btn-circle" onclick={closeSerialModal}>
					<X size={18} />
				</button>
			</div>

			<div class="flex-1 overflow-y-auto p-5">
				{#if serialModal.loading}
					<div class="py-10 text-center">
						<span class="loading loading-spinner loading-lg"></span>
						<p class="text-base-content/50 mt-3 text-sm">正在加载编码数据...</p>
					</div>
				{:else if serialModal.items.length === 0}
					<div class="text-base-content/30 py-10 text-center">
						<QrCode size={48} class="mx-auto mb-3 opacity-30" />
						<p>该物料无需独立编码或尚未生成编码</p>
					</div>
				{:else}
					<div class="space-y-3">
						{#each serialModal.items as code, i}
							<div
								class="bg-base-200/30 border-base-200 hover:border-primary/30 hover:bg-primary/5 flex items-center justify-between rounded-xl border p-4 transition-all"
							>
								<div class="flex items-center gap-3">
									<div
										class="bg-primary/10 text-primary flex h-8 w-8 items-center justify-center rounded-lg text-xs font-bold"
									>
										{i + 1}
									</div>
									<div>
										<div class="font-mono text-sm font-bold">{code.serial_code}</div>
									</div>
								</div>
								<div class="flex items-center gap-3">
									<span class="badge badge-sm {getStatusBadge(code.status)}"
										>{getStatusLabel(code.status)}</span
									>
									<a
										href={`/trace?key=${code.serial_code}`}
										class="btn btn-xs btn-ghost text-primary"
										target="_blank"
									>
										追踪
									</a>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<div class="border-base-200 flex justify-end border-t p-4">
				<button class="btn btn-sm" onclick={closeSerialModal}>关闭</button>
			</div>
		</div>
	</div>
{/if}

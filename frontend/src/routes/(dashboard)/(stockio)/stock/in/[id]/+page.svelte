<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { ArrowLeft, FileText, Warehouse, Calendar, BadgeCheck, Building2, QrCode, CheckCircle2, Wand2, X, Hash, ClipboardList, Printer } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { formatDateInCn, formatDateTimeInCn } from '$lib/datetime';

	let loading = $state(true);
	let detail = $state<any | null>(null);

	let serialModal = $state({ show: false, items: [] as any[], loading: false, title: '' });
	let confirmLogModal = $state({ show: false, loading: false, items: [] as any[] });

	const confirmMode = $derived(page.url.searchParams.get('mode') === 'confirm');
	const canConfirmReversal = $derived(
		confirmMode && detail?.stock_in_type === 'reversal' && detail?.stock_in_status !== 'passed'
	);

	let pickerOpen = $state(false);
	let pickerLoading = $state(false);
	let pickerItem: any = $state(null);
	let pickerOptions = $state<any[]>([]);
	let pickerSelectedOnly = $state(false);
	let pickerSearchTerm = $state('');
	let selectionSaving = $state(false);
	let confirming = $state(false);
	let manualSelections = $state<Record<number, number[]>>({});
	let insufficientPreparedItems = $state<Record<number, boolean>>({});

	function getNeedCount(item: any) {
		return Math.max(0, Math.floor(Number(item?.accepted_quantity || 0)));
	}

	function selectedCount(itemId: number) {
		return (manualSelections[itemId] || []).length;
	}

	function codedItems() {
		return (detail?.items || []).filter((it: any) => it.is_code);
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
				const codes: any = await api.get(`/serial-codes/stock-in-item/${item.id}/available-issued`);
				const codeList = Array.isArray(codes) ? codes : [];
				next[item.id] = codeList.filter((c: any) => c.selected).map((c: any) => c.id);
			} catch (e) {
				console.error('Failed to load selections for item', item.id, e);
				next[item.id] = [];
			}
		}
		manualSelections = next;
		refreshInsufficientPreparedFlags();
	}

	async function openSerialPicker(item: any) {
		if (!canConfirmReversal) return;
		if (detail?.stock_in_status === 'passed') return;
		if (!item?.is_code) return;
		pickerItem = item;
		pickerSelectedOnly = false;
		pickerSearchTerm = '';
		pickerOpen = true;
		pickerLoading = true;
		try {
			const options: any[] = await api.get(
				`/serial-codes/stock-in-item/${item.id}/available-issued`
			);
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
		if (!canConfirmReversal) return;
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
		// pickerOptions 已按 FIFO 排序返回；这里按当前搜索/只看已选过滤后的顺序取前 N 个
		for (const opt of sortedPickerOptions()) {
			if (next.length >= need) break;
			next.push(Number(opt.id));
		}
		manualSelections[pickerItem.id] = next;
		refreshInsufficientPreparedFlags();
	}

	async function saveCurrentItemSelections() {
		if (!canConfirmReversal) return;
		if (!pickerItem?.id) return;
		const need = getNeedCount(pickerItem);
		const sel = manualSelections[pickerItem.id] || [];
		if (sel.length !== need) {
			toast.error(`编码未选齐：需 ${need} 个，当前 ${sel.length} 个`);
			return;
		}
		selectionSaving = true;
		try {
			await api.put(`/stock-in-item/${pickerItem.id}/serial-selections`, {
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
		if (!canConfirmReversal) return;
		if (!pickerItem?.id) return;
		const itemId = pickerItem.id as number;
		selectionSaving = true;
		try {
			manualSelections[itemId] = [];
			await api.put(`/stock-in-item/${itemId}/serial-selections`, { serial_code_ids: [] });
			await loadSelectionsForAllItems();
			toast.success('已清空该行全部备货编码');
		} catch (err: any) {
			toast.error('清空失败: ' + (err?.response?.data?.msg || err?.message || err));
		} finally {
			selectionSaving = false;
		}
	}

	async function confirmReversalStockIn() {
		if (!canConfirmReversal) return;
		if (!detail?.id) return;
		for (const item of codedItems()) {
			if (selectedCount(item.id) !== getNeedCount(item)) {
				toast.error(`${item.material_name} 需备货 ${getNeedCount(item)} 个编码`);
				return;
			}
		}
		confirming = true;
		try {
			await api.post(`/stock-in/${detail.id}/confirm-reversal`);
			toast.success('确认入库成功');
			await loadDetail(detail.id);
		} catch (err: any) {
			toast.error(err?.response?.data?.msg || err?.message || '确认入库失败');
		} finally {
			confirming = false;
		}
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateInCn(dateStr);
	}

	function formatDateTime(dateStr: string) {
		if (!dateStr) return '-';
		return formatDateTimeInCn(dateStr);
	}

	function formatMoney(n: number) {
		const x = Number(n) || 0;
		return x.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function stockInStatusBadgeClass(status: string) {
		const map: Record<string, string> = {
			preparing: 'badge-info',
			pending: 'badge-warning',
			passed: 'badge-success',
			failed: 'badge-error'
		};
		return map[status] || 'badge-ghost';
	}

	function stockInStatusName(status: string) {
		const map: Record<string, string> = {
			preparing: '待入库',
			pending: '部分入库',
			passed: '已完成',
			failed: '已拒绝'
		};
		return map[status] || status;
	}

	function stockInTypeName(t: string) {
		const map: Record<string, string> = {
			purchase: '采购入库',
			return: '销售退货入库',
			sales_return: '销售退货入库',
			reversal: '退料入库',
			production: '生产入库'
		};
		return map[t] || t || '-';
	}

	function businessDocLabel(detail: any) {
		if (detail?.business_doc_type === 'purchase_order') return '采购单号';
		if (detail?.business_doc_type === 'reversal_order') return '退料单号';
		return '业务单号';
	}

	function businessDocHref(detail: any) {
		if (!detail?.business_doc_id || !detail?.business_doc_type) return '';
		if (detail.business_doc_type === 'purchase_order')
			return `/purchase/orders/${detail.business_doc_id}`;
		if (detail.business_doc_type === 'reversal_order')
			return `/reversal/orders/${detail.business_doc_id}`;
		return '';
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			const res: any = await api.get(`/stock-in/${id}`);
			detail = res;
		} catch (err: any) {
			toast.error('加载入库单详情失败: ' + (err?.message || err));
			detail = null;
		} finally {
			loading = false;
		}
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
			const res: any = await api.get(`/serial-codes/stock-in-item/${item.id}`);
			serialModal = { ...serialModal, loading: false, items: res || [] };
		} catch (err: any) {
			toast.error('加载编码详情失败: ' + (err?.message || err));
			serialModal = { ...serialModal, loading: false };
		}
	}

	function closeSerialModal() {
		serialModal = { show: false, items: [], loading: false, title: '' };
	}

	async function openConfirmLogModal() {
		if (!detail?.id) return;
		confirmLogModal = { show: true, loading: true, items: [] };
		try {
			const res: any = await api.get(`/stock-in/${detail.id}/confirm-logs`);
			confirmLogModal = { show: true, loading: false, items: res || [] };
		} catch (err: any) {
			toast.error('加载入库日志失败: ' + (err?.message || err));
			confirmLogModal = { ...confirmLogModal, loading: false };
		}
	}

	function closeConfirmLogModal() {
		confirmLogModal = { show: false, loading: false, items: [] };
	}

	function handleBackdropKeydown(event: KeyboardEvent, onClose: () => void) {
		if (event.key === 'Enter' || event.key === ' ' || event.key === 'Escape') {
			event.preventDefault();
			onClose();
		}
	}

	function stopModalKeydown(event: KeyboardEvent) {
		event.stopPropagation();
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

	function lineAmount(it: any) {
		return (Number(it?.received_quantity) || 0) * (Number(it?.unit_cost) || 0);
	}

	function hasSerialCodes(item: any) {
		return Boolean(item?.is_code);
	}

	function grandTotal() {
		let s = 0;
		for (const it of detail?.items || []) s += lineAmount(it);
		return s;
	}

	onMount(() => {
		const id = Number(page.params.id);
		if (!id) {
			toast.error('无效的入库单ID');
			loading = false;
			return;
		}
		loadDetail(id).then(async () => {
			if (canConfirmReversal && codedItems().length > 0) {
				await loadSelectionsForAllItems();
			}
		});
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-green-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">{confirmMode ? '入库单确认' : '入库单详情'}</h1>
		</div>

		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>出入库管理</li>
				<li><a class="text-primary" href="/stock/in">入库管理</a></li>
				<li>{confirmMode ? '确认' : '详情'}</li>
			</ul>
		</div>
	</div>

	<div class="flex items-center justify-between gap-3 print:hidden">
		<a href="/stock/in" class="btn btn-ghost btn-sm gap-1">
			<ArrowLeft size={14} /> 返回列表
		</a>
		{#if detail}
			{#if confirmMode}
				<div class="flex items-center gap-2">
					{#if canConfirmReversal}
						<button
							class="btn btn-sm btn-success gap-2 text-white"
							onclick={confirmReversalStockIn}
							disabled={confirming || selectionSaving}
						>
							<CheckCircle2 size={16} /> 确认入库
						</button>
					{/if}
				</div>
			{:else}
				<div class="flex items-center gap-2">
					<span class="badge badge-sm {stockInStatusBadgeClass(detail.stock_in_status)}">
						{stockInStatusName(detail.stock_in_status)}
					</span>
					<span class="text-base-content/60 font-mono text-base">{detail.stock_in_no}</span>
					<button class="btn btn-outline btn-sm gap-1" onclick={() => window.print()}>
						<Printer size={16} /> 打印
					</button>
					{#if detail.stock_in_type !== 'reversal'}
						<button class="btn btn-sm btn-ghost text-primary" onclick={openConfirmLogModal}>
							<ClipboardList size={15} /> 入库日志
						</button>
					{/if}
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
			未找到入库单信息
		</div>
	{:else}
		<div class="bg-base-100 border-base-300 space-y-5 rounded-2xl border p-5">
			<div class="flex items-center gap-2 text-base font-semibold">
				<FileText size={16} /> 单据信息
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<FileText size={14} /> 入库单号
					</div>
					<div class="mt-1 font-mono text-base">{detail.stock_in_no}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<BadgeCheck size={14} /> 入库类型
					</div>
					<div class="mt-1 text-base">{stockInTypeName(detail.stock_in_type)}</div>
				</div>

				{#if detail.stock_in_type !== 'reversal'}
					<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
						<div class="text-base-content/50 flex items-center gap-1 text-base">
							<Building2 size={14} /> 供应商
						</div>
						<div class="mt-1 text-base font-semibold">{detail.supplier_name || '-'}</div>
					</div>
				{/if}

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<Warehouse size={14} /> 仓库
					</div>
					<div class="mt-1 text-base">{detail.warehouse_name || '-'}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 flex items-center gap-1 text-base">
						<Calendar size={14} /> 入库日期
					</div>
					<div class="mt-1 text-base">{formatDate(detail.stock_in_date)}</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 text-base">入库金额合计</div>
					<div class="text-success mt-1 font-mono text-base font-semibold">
						¥{formatMoney(grandTotal())}
					</div>
				</div>

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4">
					<div class="text-base-content/50 text-base">创建时间</div>
					<div class="mt-1 text-base">{formatDateTime(detail.created_at)}</div>
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

				<div class="bg-base-200/40 border-base-300 rounded-xl border p-4 lg:col-span-4">
					<div class="text-base-content/50 text-base">备注</div>
					<div class="mt-1 text-base break-words whitespace-pre-wrap">{detail.remark || '-'}</div>
				</div>
			</div>
		</div>

		<div class="bg-base-100 border-base-300 space-y-4 rounded-2xl border p-5">
			<div class="flex items-center justify-between gap-3">
				<div class="text-base font-semibold">入库明细</div>
				<div class="text-base-content/50 text-base">共 {detail.items?.length || 0} 行</div>
			</div>

			<div class="overflow-x-auto">
				<table class="table-zebra table w-full table-fixed text-base">
					<colgroup>
						<col style="width: 36%" />
						<col style="width: 8%" />
						<col style="width: 12%" />
						<col style="width: 14%" />
						<col style="width: 12%" />
						<col style="width: 12%" />
						<col style="width: 6%" />
					</colgroup>
					<thead>
						<tr>
							<th class="text-left">物料名称</th>
							<th class="text-left whitespace-nowrap">属性</th>
							<th class="text-center">
								{detail?.stock_in_type === 'reversal' ? '入库数量' : '采购数量'}
							</th>
							<th class="text-center">
								{detail?.stock_in_type === 'reversal' ? '已入库数量' : '累计入库数量'}
							</th>
							<th class="text-right">单价</th>
							<th class="text-right">金额</th>
							<th class="text-center">编码</th>
						</tr>
					</thead>
					<tbody>
						{#each detail.items || [] as item}
							<tr
								class="hover:bg-base-200/50 transition-colors {canConfirmReversal &&
								insufficientPreparedItems[item.id]
									? 'bg-error/10'
									: ''}"
							>
								<td>
									<div class="space-y-1 pr-3">
										<div class="group/line relative">
											<div class="truncate font-semibold">
												{item.material_name || '-'}
												{#if item.material_code}
													<span class="text-base-content/50 ml-2 font-mono text-base"
														>{item.material_code}</span
													>
												{/if}
											</div>
											<div
												class="border-base-300 bg-base-100 pointer-events-none absolute top-full left-0 z-20 mt-1 hidden max-w-[42rem] rounded-md border px-2 py-1 text-sm leading-5 whitespace-normal shadow-xl group-hover/line:block"
											>
												{item.material_name || '-'}
												{#if item.material_code}
													<span class="text-base-content/60 ml-2 font-mono"
														>{item.material_code}</span
													>
												{/if}
											</div>
										</div>
										<div class="text-base-content/50 truncate text-xs">
											属性 {(item.custom_attributes || []).length} 项
										</div>
									</div>
								</td>
								<td class="whitespace-nowrap">
									<span class="text-sm">{(item.custom_attributes || []).length}项</span>
								</td>
								<td class="text-center">
									<span class="font-mono">{item.purchase_quantity ?? 0}</span>
								</td>
								<td class="text-center">
									<span class="font-mono">{item.received_quantity ?? 0}</span>
								</td>
								<td class="text-right font-mono">
									{#if item.unit_cost != null}
										¥{formatMoney(item.unit_cost)}
									{:else}
										—
									{/if}
								</td>
								<td class="text-success text-right font-mono font-semibold">
									{#if item.unit_cost != null}
										¥{formatMoney(lineAmount(item))}
									{:else}
										—
									{/if}
								</td>
								<td>
									{#if !hasSerialCodes(item)}
										<span class="text-base-content/40">-</span>
									{:else if canConfirmReversal}
										<button
											type="button"
											class="badge badge-sm {selectedCount(item.id) === getNeedCount(item)
												? 'badge-success'
												: 'badge-warning'} cursor-pointer"
											onclick={() => openSerialPicker(item)}
										>
											{selectedCount(item.id)} / {getNeedCount(item)}
										</button>
									{:else}
										<button
											class="btn btn-sm btn-ghost text-primary"
											onclick={() => viewSerialCodes(item)}
											title="查看编码详情"
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

{#if serialModal.show}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm"
		role="button"
		tabindex="0"
		aria-label="关闭编码详情弹窗"
		onclick={closeSerialModal}
		onkeydown={(e) => handleBackdropKeydown(e, closeSerialModal)}
	>
		<div
			class="bg-base-100 border-base-300 flex max-h-[80vh] w-full max-w-2xl flex-col rounded-2xl border shadow-2xl"
			role="dialog"
			aria-modal="true"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={stopModalKeydown}
			transition:fade
		>
			<div class="border-base-200 flex items-center justify-between border-b p-5">
				<div class="flex items-center gap-2">
					<Hash size={18} class="text-primary" />
					<h3 class="text-lg font-bold">编码详情</h3>
					<span class="text-base-content/50 ml-2 text-base">{serialModal.title}</span>
				</div>
				<button class="btn btn-sm btn-ghost btn-circle" onclick={closeSerialModal}>
					<X size={18} />
				</button>
			</div>

			<div class="flex-1 overflow-y-auto p-5">
				{#if serialModal.loading}
					<div class="py-10 text-center">
						<span class="loading loading-spinner loading-lg"></span>
						<p class="text-base-content/50 mt-3 text-base">正在加载编码数据...</p>
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

{#if confirmLogModal.show}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm"
		role="button"
		tabindex="0"
		aria-label="关闭入库日志弹窗"
		onclick={closeConfirmLogModal}
		onkeydown={(e) => handleBackdropKeydown(e, closeConfirmLogModal)}
	>
		<div
			class="bg-base-100 border-base-300 flex max-h-[80vh] w-full max-w-5xl flex-col rounded-2xl border shadow-2xl"
			role="dialog"
			aria-modal="true"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={stopModalKeydown}
			transition:fade
		>
			<div class="border-base-200 flex items-center justify-between border-b p-5">
				<div class="flex items-center gap-2">
					<ClipboardList size={18} class="text-primary" />
					<h3 class="text-lg font-bold">入库日志</h3>
					{#if detail}
						<span class="text-base-content/50 ml-2 font-mono text-sm">{detail.stock_in_no}</span>
					{/if}
				</div>
				<button class="btn btn-sm btn-ghost btn-circle" onclick={closeConfirmLogModal}>
					<X size={18} />
				</button>
			</div>
			<div class="flex-1 overflow-y-auto p-5">
				{#if confirmLogModal.loading}
					<div class="py-10 text-center">
						<span class="loading loading-spinner loading-lg"></span>
						<p class="text-base-content/50 mt-3 text-sm">正在加载入库日志...</p>
					</div>
				{:else if confirmLogModal.items.length === 0}
					<div class="text-base-content/30 py-10 text-center">
						<ClipboardList size={48} class="mx-auto mb-3 opacity-30" />
						<p>暂无入库日志</p>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table-zebra table w-full text-sm">
							<thead>
								<tr>
									<th>时间</th>
									<th>物料</th>
									<th>采购数量</th>
									<th>确认前累计入库</th>
									<th>本次入库</th>
									<th>确认后累计入库</th>
									<th>操作人</th>
								</tr>
							</thead>
							<tbody>
								{#each confirmLogModal.items as log}
									<tr>
										<td class="font-mono">{formatDateTime(log.created_at)}</td>
										<td>
											<div>{log.material_name || '-'}</div>
											<div class="text-base-content/50 font-mono text-xs">
												{log.material_code || '-'}
											</div>
										</td>
										<td class="font-mono">{log.purchase_quantity ?? 0}</td>
										<td class="font-mono">{log.before_received_quantity ?? 0}</td>
										<td class="text-primary font-mono font-semibold"
											>{log.current_received_quantity ?? 0}</td
										>
										<td class="font-mono">{log.after_received_quantity ?? 0}</td>
										<td>{log.operator_name || '-'}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
			<div class="border-base-200 flex justify-end border-t p-4">
				<button class="btn btn-sm" onclick={closeConfirmLogModal}>关闭</button>
			</div>
		</div>
	</div>
{/if}

<style>
	@page {
		size: A4;
		margin: 12mm;
	}
</style>

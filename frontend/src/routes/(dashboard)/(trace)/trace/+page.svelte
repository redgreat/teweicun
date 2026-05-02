<script lang="ts">
	import {
		Search,
		Magnet,
		ScanSearch,
		Clock3,
		Package,
		ArrowRightLeft,
		RotateCcw,
		Warehouse,
		User,
		FileText,
		ExternalLink
	} from 'lucide-svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { fade, fly } from 'svelte/transition';
	import { formatDateTimeInCn } from '$lib/datetime';

	let queryValue = $state('');
	let results = $state<any>(null);
	let loading = $state(false);
	let serialInfo = $state<any>(null);

	async function handleTrace() {
		if (!queryValue) return;
		loading = true;
		results = null;
		serialInfo = null;
		try {
			const data: any = await api.get(
				`/trace/material/serial?serial_code=${encodeURIComponent(queryValue.trim())}`
			);
			if (data.serial_info) {
				serialInfo = data.serial_info;
			}
			results = data.traces || [];
		} catch (err: any) {
			toast.error('查询失败: ' + err);
		} finally {
			loading = false;
		}
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

	function getActionColor(action: string) {
		const map: Record<string, string> = {
			stock_in: 'text-success bg-success/10',
			stock_out: 'text-warning bg-warning/10',
			return: 'text-info bg-info/10'
		};
		return map[action] || 'text-base-content/30 bg-base-200';
	}

	function getDocTypeLabel(refDocType: string) {
		const map: Record<string, string> = {
			stock_in: '入库单',
			stock_out: '出库单',
			purchase_order: '采购单',
			purchase_return: '采购退货单',
			consumption_order: '领料单',
			reversal_order: '退料单'
		};
		return map[refDocType] || refDocType || '-';
	}

	function docHref(trace: any) {
		if (!trace?.ref_doc_id || !trace?.ref_doc_type) return '';
		if (trace.ref_doc_type === 'stock_in') return `/stock/in/${trace.ref_doc_id}`;
		if (trace.ref_doc_type === 'stock_out') return `/stock/out/${trace.ref_doc_id}`;
		if (trace.ref_doc_type === 'purchase_order') return `/purchase/orders/${trace.ref_doc_id}`;
		if (trace.ref_doc_type === 'purchase_return') return `/purchase/return/${trace.ref_doc_id}`;
		if (trace.ref_doc_type === 'consumption_order')
			return `/consumption/orders/${trace.ref_doc_id}`;
		if (trace.ref_doc_type === 'reversal_order') return `/reversal/orders/${trace.ref_doc_id}`;
		return '';
	}

	function formatActionTime(v: string | Date) {
		if (!v) return '-';
		return formatDateTimeInCn(v);
	}
</script>

<div class="space-y-6">
	<div class="card bg-base-100 border-base-300 rounded-3xl border shadow-xl">
		<div class="card-body p-6">
			<div class="flex flex-col gap-3 md:flex-row">
				<div class="group relative flex-1">
					<Search
						size={20}
						class="text-base-content/30 group-focus-within:text-primary absolute top-1/2 left-4 -translate-y-1/2 transition-colors"
					/>
					<input
						type="text"
						placeholder="请输入物料具体编码（支持全模糊，如 MAT-240423）"
						bind:value={queryValue}
						onkeydown={(e) => e.key === 'Enter' && handleTrace()}
						class="input input-lg bg-base-200/50 focus:bg-base-100 focus:ring-primary/20 w-full rounded-2xl border-none pl-12 text-base transition-all focus:ring-2"
					/>
				</div>
				<button
					class="btn btn-primary btn-lg shadow-primary/30 gap-2 rounded-2xl px-10 shadow-lg"
					onclick={handleTrace}
					disabled={loading}
				>
					{#if loading}<span class="loading loading-spinner"></span>{/if}
					查询追踪
				</button>
			</div>
		</div>
	</div>

	{#if serialInfo}
		<div class="grid grid-cols-1 gap-4 md:grid-cols-4" in:fade>
			<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
				<div class="text-base-content/55 text-sm">具体编码</div>
				<div class="mt-1 font-mono text-base font-semibold">{serialInfo.serial_code}</div>
			</div>
			<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
				<div class="text-base-content/55 text-sm">物料</div>
				<div class="mt-1 text-base font-semibold">{serialInfo.material_name}</div>
				<div class="text-base-content/70 font-mono text-xs">{serialInfo.material_code}</div>
			</div>
			<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
				<div class="text-base-content/55 text-sm">当前状态</div>
				<div class="mt-1">
					<span class="badge {getStatusBadge(serialInfo.status)}"
						>{getStatusLabel(serialInfo.status)}</span
					>
				</div>
			</div>
			<div class="bg-base-100 border-base-300 rounded-2xl border p-4">
				<div class="text-base-content/55 text-sm">最近变动时间</div>
				<div class="mt-1 text-sm">
					{results?.length ? formatActionTime(results[results.length - 1].action_time) : '-'}
				</div>
			</div>
		</div>
	{/if}

	{#if results && results.length > 0}
		<div class="card bg-base-100 border-base-300 rounded-2xl border shadow-lg" in:fade>
			<div class="card-body p-6">
				<div class="mb-3 flex items-center justify-between">
					<div class="flex items-center gap-2">
						<ScanSearch size={18} class="text-primary" />
						<h2 class="text-lg font-bold">流转时间线</h2>
					</div>
					<div class="text-base-content/55 text-xs">共 {results.length} 条记录</div>
				</div>
				<div class="relative mt-1">
					<div class="bg-base-200 absolute top-0 bottom-0 left-8 w-0.5"></div>
					<div class="space-y-4">
						{#each results as trace, i}
							<div class="relative pl-16" in:fly={{ y: 20, delay: i * 80 }}>
								<div
									class="bg-base-100 border-primary absolute top-6 left-6 flex h-5 w-5 items-center justify-center rounded-full border-4"
								>
									<div class="bg-primary h-2 w-2 rounded-full"></div>
								</div>
								<div
									class="card bg-base-100 border-base-200 rounded-xl border transition-all hover:shadow-lg"
								>
									<div class="card-body p-5">
										<div class="mb-3 flex items-center justify-between">
											<div class="flex items-center gap-2">
												<div class="rounded-lg p-1.5 {getActionColor(trace.action)}">
													{#if trace.action === 'return'}
														<RotateCcw size={16} />
													{:else}
														<ArrowRightLeft size={16} />
													{/if}
												</div>
												<span class="text-sm font-bold">{trace.action_label || trace.action}</span>
												<span class="badge badge-ghost badge-sm"
													>{getDocTypeLabel(trace.ref_doc_type)}</span
												>
											</div>
											<div class="flex items-center gap-2 text-xs opacity-50">
												<Clock3 size={12} />
												{formatActionTime(trace.action_time)}
											</div>
										</div>
										<div class="grid grid-cols-2 gap-3 text-sm md:grid-cols-4">
											<div>
												<div class="flex items-center gap-1 text-[10px] opacity-50">
													<FileText size={10} /> 单据编号
												</div>
												{#if docHref(trace)}
													<a
														href={docHref(trace)}
														class="link link-primary inline-flex items-center gap-1 font-mono text-xs no-underline hover:underline"
													>
														{trace.ref_doc_no || '-'}
														<ExternalLink size={10} />
													</a>
												{:else}
													<div class="font-mono text-xs">{trace.ref_doc_no || '-'}</div>
												{/if}
											</div>
											<div>
												<div class="flex items-center gap-1 text-[10px] opacity-50">
													<Warehouse size={10} /> 来源仓库
												</div>
												<div>{trace.from_warehouse || '-'}</div>
											</div>
											<div>
												<div class="flex items-center gap-1 text-[10px] opacity-50">
													<Warehouse size={10} /> 目标仓库
												</div>
												<div>{trace.to_warehouse || '-'}</div>
											</div>
											<div>
												<div class="flex items-center gap-1 text-[10px] opacity-50">
													<User size={10} /> 操作人
												</div>
												<div>{trace.operator_name || '-'}</div>
											</div>
										</div>
										{#if trace.remark}
											<div class="text-base-content/50 mt-2 text-xs italic">
												{trace.remark}
											</div>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<div class="card bg-base-100 border-base-300 rounded-2xl border shadow-lg" in:fade>
			<div class="card-body p-0">
				<div class="border-base-200 flex items-center justify-between border-b px-5 py-4">
					<div class="flex items-center gap-2 font-semibold">
						<Package size={16} class="text-primary" /> 单据明细
					</div>
					<div class="text-base-content/55 text-xs">按时间升序展示</div>
				</div>
				<div class="overflow-x-auto">
					<table class="table-zebra table w-full text-sm">
						<thead>
							<tr>
								<th>时间</th>
								<th>动作</th>
								<th>业务类型</th>
								<th>业务单号</th>
								<th>来源仓库</th>
								<th>目标仓库</th>
								<th>操作人</th>
								<th>备注</th>
							</tr>
						</thead>
						<tbody>
							{#each results as trace}
								<tr>
									<td class="font-mono whitespace-nowrap">{formatActionTime(trace.action_time)}</td>
									<td>{trace.action_label || trace.action}</td>
									<td>{getDocTypeLabel(trace.ref_doc_type)}</td>
									<td class="font-mono">
										{#if docHref(trace)}
											<a
												href={docHref(trace)}
												class="link link-primary inline-flex items-center gap-1 no-underline hover:underline"
											>
												{trace.ref_doc_no || '-'}
												<ExternalLink size={10} />
											</a>
										{:else}
											{trace.ref_doc_no || '-'}
										{/if}
									</td>
									<td>{trace.from_warehouse || '-'}</td>
									<td>{trace.to_warehouse || '-'}</td>
									<td>{trace.operator_name || '-'}</td>
									<td>
										<span class="block max-w-[260px] truncate" title={trace.remark || '-'}>
											{trace.remark || '-'}
										</span>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		</div>
	{:else if results && results.length === 0 && !loading}
		<div
			class="bg-base-100 border-base-300 rounded-3xl border border-dashed p-20 text-center"
			in:fade
		>
			<Magnet size={48} class="text-base-content/10 mx-auto mb-4" />
			<p class="text-base-content/40 italic">未发现与 "{queryValue}" 相关的流转记录</p>
		</div>
	{:else if !loading}
		<div class="pointer-events-none p-20 text-center opacity-30 select-none">
			<ScanSearch size={64} class="mx-auto mb-4" />
			<p class="text-lg font-bold">输入物料具体编码后可查看完整出入库轨迹</p>
		</div>
	{/if}
</div>

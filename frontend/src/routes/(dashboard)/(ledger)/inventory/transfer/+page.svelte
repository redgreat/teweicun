<!--
功能：transfer页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';

	let transfers = $state([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	const columns = [
		{ key: 'transfer_no', label: '调拨单号', class: 'font-mono' },
		{ key: 'from_warehouse_name', label: '调出仓库' },
		{ key: 'to_warehouse_name', label: '调入仓库' },
		{ key: 'total_quantity', label: '总数量', class: 'text-right' },
		{ key: 'transfer_date', label: '调拨日期' },
		{ key: 'status', label: '状态', class: 'text-center' }
	];

	async function loadTransfers(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/stock-transfers?page=${page}&page_size=${pageSize}`);
			transfers = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	onMount(() => loadTransfers(1));
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={transfers}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadTransfers}
		onCreate={() => toast.info('新建调拨单功能开发中')}
	/>
</div>

<!--
功能：check页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';

	let checks = $state([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	const columns = [
		{ key: 'check_no', label: '盘点单号', class: 'font-mono' },
		{ key: 'warehouse_name', label: '盘点仓库' },
		{ key: 'check_date', label: '盘点日期' },
		{ key: 'check_type', label: '盘点类型' },
		{ key: 'total_difference', label: '差异总数', class: 'text-right' },
		{ key: 'status', label: '状态', class: 'text-center' }
	];

	async function loadChecks(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/inventory-checks?page=${page}&page_size=${pageSize}`);
			checks = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	onMount(() => loadChecks(1));
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={checks}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadChecks}
		onCreate={() => toast.info('新建盘点单功能开发中')}
	/>
</div>

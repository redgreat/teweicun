<!--
功能：布局组件
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import '../app.css';
	let { children } = $props();
</script>

<svelte:head>
	<title>特维存</title>
</svelte:head>

<div class="min-h-screen">
	{@render children()}
</div>

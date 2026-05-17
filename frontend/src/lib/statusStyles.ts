/**
 * 功能：统一的状态 Badge 样式映射
 * 创建时间：2026-04-26
 * 创建人：Antigravity
 */

export interface StatusStyle {
	class: string;
	label: string;
}

/** 通用状态样式映射函数 */
export function getStatusStyle(status: string, module: string): StatusStyle {
	const s = (status || '').toLowerCase();

	// 1. 采购订单 (Purchase Order)
	if (module === 'purchase_order') {
		const map: Record<string, StatusStyle> = {
			draft: {
				class: 'badge-ghost bg-base-200 text-base-content/70 border-base-300',
				label: '待提交'
			},
			ordered: { class: 'badge-info badge-outline', label: '已下单' },
			partial_received: { class: 'badge-primary badge-outline', label: '部分到货' },
			full_received: { class: 'badge-success text-white', label: '已完成' },
			closed: { class: 'badge-ghost text-base-content/40', label: '已关闭' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	// 2. 采购退货 (Purchase Return)
	if (module === 'purchase_return') {
		const map: Record<string, StatusStyle> = {
			draft: {
				class: 'badge-ghost bg-base-200 text-base-content/70 border-base-300',
				label: '待提交'
			},
			confirmed: { class: 'badge-info badge-outline', label: '待出库' },
			pending_out: { class: 'badge-info badge-outline', label: '待出库' },
			completed: { class: 'badge-success text-white', label: '已完成' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	// 3. 领料单 (Consumption)
	if (module === 'consumption_order') {
		const map: Record<string, StatusStyle> = {
			draft: {
				class: 'badge-ghost bg-base-200 text-base-content/70 border-base-300',
				label: '待提交'
			},
			pending: { class: 'badge-warning badge-outline', label: '待出库' },
			confirmed: { class: 'badge-warning badge-outline', label: '待出库' },
			completed: { class: 'badge-success text-white', label: '已完成' },
			cancelled: { class: 'badge-error badge-outline', label: '已取消' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	// 4. 退料单 (Reversal)
	if (module === 'reversal_order') {
		const map: Record<string, StatusStyle> = {
			draft: {
				class: 'badge-ghost bg-base-200 text-base-content/70 border-base-300',
				label: '待提交'
			},
			pending: { class: 'badge-info badge-outline', label: '待入库' },
			confirmed: { class: 'badge-info badge-outline', label: '待入库' },
			completed: { class: 'badge-success text-white', label: '已完成' },
			cancelled: { class: 'badge-error badge-outline', label: '已取消' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	// 5. 入库单 (Stock In)
	if (module === 'stock_in') {
		const map: Record<string, StatusStyle> = {
			preparing: { class: 'badge-warning badge-outline', label: '待入库' },
			pending: { class: 'badge-info badge-outline', label: '部分入库' },
			passed: { class: 'badge-success text-white', label: '已完成' },
			failed: { class: 'badge-error badge-outline', label: '已拒绝' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	// 6. 出库单 (Stock Out)
	if (module === 'stock_out') {
		const map: Record<string, StatusStyle> = {
			draft: { class: 'badge-warning badge-outline', label: '待出库' },
			pending: { class: 'badge-warning badge-outline', label: '待出库' },
			confirmed: { class: 'badge-success text-white', label: '已完成' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	// 7. 销售订单 (Sales Order)
	if (module === 'sales_order') {
		const map: Record<string, StatusStyle> = {
			draft: {
				class: 'badge-ghost bg-base-200 text-base-content/70 border-base-300',
				label: '待提交'
			},
			confirmed: { class: 'badge-info badge-outline', label: '待出库' },
			preparing: { class: 'badge-warning badge-outline', label: '出库中' },
			shipped: { class: 'badge-success text-white', label: '已完成' },
			cancelled: { class: 'badge-error badge-outline', label: '已取消' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	// 8. 销售退货 (Sales Return)
	if (module === 'sales_return') {
		const map: Record<string, StatusStyle> = {
			draft: {
				class: 'badge-ghost bg-base-200 text-base-content/70 border-base-300',
				label: '待提交'
			},
			confirmed: { class: 'badge-success text-white', label: '已完成' }
		};
		return map[s] || { class: 'badge-ghost', label: status || '-' };
	}

	return { class: 'badge-ghost', label: status || '-' };
}

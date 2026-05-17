/**
 * 功能：下拉浮层位置与宽度计算工具
 * 创建时间：2026-05-17
 * 创建人：GPT-5.4
 */

type FloatingDropdownPlacementOptions = {
	anchor: HTMLElement;
	contentTexts?: string[];
	minWidth?: number;
	maxWidth?: number;
	maxListHeight?: number;
	headerHeight?: number;
	padding?: number;
	gap?: number;
	preferBelowMinSpace?: number;
	extraWidth?: number;
	font?: string;
};

type FloatingDropdownPlacement = {
	left: number;
	top: number;
	width: number;
	listMaxHeight: number;
};

type FloatingDropdownGridLayoutOptions = {
	firstColumnTexts: string[];
	fixedColumnWidths: number[];
	minFirstColumnWidth?: number;
	firstColumnPadding?: number;
	gap?: number;
	panelPadding?: number;
	font?: string;
};

type FloatingDropdownGridLayout = {
	gridTemplate: string;
	firstColumnWidth: number;
	preferredPanelWidth: number;
};

let measureCanvas: HTMLCanvasElement | null = null;

function measureTextWidth(text: string, font: string) {
	if (typeof document === 'undefined') {
		return String(text || '').length * 14;
	}
	if (!measureCanvas) {
		measureCanvas = document.createElement('canvas');
	}
	const context = measureCanvas.getContext('2d');
	if (!context) {
		return String(text || '').length * 14;
	}
	context.font = font;
	return context.measureText(String(text || '')).width;
}

export function estimateDropdownTextWidth(text: string, font = '500 14px system-ui') {
	return measureTextWidth(text, font);
}

export function buildFloatingDropdownGridLayout(
	options: FloatingDropdownGridLayoutOptions
): FloatingDropdownGridLayout {
	const font = options.font ?? '500 14px system-ui';
	const firstColumnPadding = options.firstColumnPadding ?? 28;
	const gap = options.gap ?? 12;
	const panelPadding = options.panelPadding ?? 24;
	const firstColumnWidth = Math.max(
		options.minFirstColumnWidth ?? 260,
		...(options.firstColumnTexts || []).map(
			(text) => measureTextWidth(String(text || ''), font) + firstColumnPadding
		)
	);
	const fixedColumnsWidth = (options.fixedColumnWidths || []).reduce(
		(sum, width) => sum + Math.max(0, Number(width) || 0),
		0
	);
	const totalColumns = 1 + (options.fixedColumnWidths || []).length;
	return {
		gridTemplate: [
			`${Math.round(firstColumnWidth)}px`,
			...(options.fixedColumnWidths || []).map((width) => `${Math.round(width)}px`)
		].join(' '),
		firstColumnWidth: Math.round(firstColumnWidth),
		preferredPanelWidth: Math.round(
			firstColumnWidth + fixedColumnsWidth + gap * Math.max(0, totalColumns - 1) + panelPadding
		)
	};
}

export function calcFloatingDropdownPlacement(
	options: FloatingDropdownPlacementOptions
): FloatingDropdownPlacement {
	const padding = options.padding ?? 8;
	const gap = options.gap ?? 8;
	const headerHeight = options.headerHeight ?? 44;
	const maxListHeight = options.maxListHeight ?? 320;
	const preferBelowMinSpace = options.preferBelowMinSpace ?? headerHeight + 160;
	const font = options.font ?? '500 14px system-ui';
	const extraWidth = options.extraWidth ?? 96;
	const rect = options.anchor.getBoundingClientRect();
	const viewportWidth = typeof window === 'undefined' ? rect.width : window.innerWidth;
	const viewportHeight = typeof window === 'undefined' ? rect.height : window.innerHeight;
	const viewportMaxWidth = Math.max(rect.width, viewportWidth - padding * 2);
	const maxWidth = Math.max(
		options.minWidth ?? rect.width,
		Math.min(options.maxWidth ?? viewportMaxWidth, viewportMaxWidth)
	);
	const estimatedContentWidth = Math.max(
		0,
		...(options.contentTexts || []).map((text) => measureTextWidth(text, font))
	);
	const width = Math.min(
		Math.max(rect.width, options.minWidth ?? rect.width, estimatedContentWidth + extraWidth),
		maxWidth
	);
	const spaceBelow = viewportHeight - rect.bottom - padding;
	const spaceAbove = rect.top - padding;
	const openUp = spaceBelow < preferBelowMinSpace && spaceAbove > spaceBelow;
	let left = rect.left;
	if (left + width > viewportWidth - padding) {
		left = Math.max(padding, viewportWidth - padding - width);
	}
	if (openUp) {
		const listMax = Math.max(96, Math.min(maxListHeight, rect.top - padding - gap - headerHeight));
		return {
			left,
			top: Math.max(padding, rect.top - gap - headerHeight - listMax),
			width,
			listMaxHeight: listMax
		};
	}
	const listMax = Math.max(
		96,
		Math.min(maxListHeight, viewportHeight - padding - (rect.bottom + gap) - headerHeight)
	);
	return {
		left,
		top: Math.min(viewportHeight - padding - headerHeight - listMax, rect.bottom + gap),
		width,
		listMaxHeight: listMax
	};
}

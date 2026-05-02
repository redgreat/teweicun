/**
 * 功能：统一东八区日期时间格式化工具
 * 创建时间：2026-04-29
 * 创建人：Codex
 */

const TZ = 'Asia/Shanghai';

const dateFormatter = new Intl.DateTimeFormat('sv-SE', {
	timeZone: TZ,
	year: 'numeric',
	month: '2-digit',
	day: '2-digit'
});

const dateTimeFormatter = new Intl.DateTimeFormat('sv-SE', {
	timeZone: TZ,
	year: 'numeric',
	month: '2-digit',
	day: '2-digit',
	hour: '2-digit',
	minute: '2-digit',
	second: '2-digit',
	hour12: false
});

function toDate(value: string | Date | number): Date | null {
	if (value === null || value === undefined || value === '') return null;
	const d = value instanceof Date ? value : new Date(value);
	if (Number.isNaN(d.getTime())) return null;
	return d;
}

export function todayDateInCn(): string {
	return dateFormatter.format(new Date());
}

export function currentMonthInCn(): string {
	return todayDateInCn().slice(0, 7);
}

export function formatDateInCn(value: string | Date | number): string {
	const d = toDate(value);
	if (!d) return '-';
	return dateFormatter.format(d);
}

export function formatDateTimeInCn(value: string | Date | number): string {
	const d = toDate(value);
	if (!d) return '-';
	return dateTimeFormatter.format(d);
}

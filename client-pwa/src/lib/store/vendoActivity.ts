import { writable } from 'svelte/store';

/**
 * Live per-vendo data-pull progress, keyed by vendo id (0..100).
 *
 * The scheduler broadcasts `{type:"vendo_refresh", vendo_id, progress}` frames over
 * the WebSocket as it pulls each machine's logs + status. `setVendoProgress` records
 * the latest value so the status badge can animate its snake to match. Once a vendo
 * reaches 100% the entry is held briefly (so the ring visibly closes) then removed,
 * returning the badge to idle.
 */
export const refreshingVendos = writable<Map<number, number>>(new Map());

/**
 * Latest online state per vendo id, pushed by the scheduler's `vendo_status`
 * frames as each pull determines whether the device responded. The status badge
 * prefers this over the (load-time) row value so the Online/Offline text, colour
 * and snake gating reflect the most recent poll without a table reload.
 */
export const vendoOnline = writable<Map<number, boolean>>(new Map());

export function setVendoOnline(id: number, online: boolean): void {
	vendoOnline.update((m) => {
		const next = new Map(m);
		next.set(id, online);
		return next;
	});
}

// Pending "clear after completion" timers, keyed by vendo id.
const clearTimers = new Map<number, ReturnType<typeof setTimeout>>();

const COMPLETE_HOLD_MS = 500;

export function setVendoProgress(id: number, progress: number): void {
	refreshingVendos.update((m) => {
		const next = new Map(m);
		next.set(id, progress);
		return next;
	});

	const existing = clearTimers.get(id);
	if (existing) {
		clearTimeout(existing);
		clearTimers.delete(id);
	}

	if (progress >= 100) {
		clearTimers.set(
			id,
			setTimeout(() => {
				refreshingVendos.update((m) => {
					const next = new Map(m);
					next.delete(id);
					return next;
				});
				clearTimers.delete(id);
			}, COMPLETE_HOLD_MS)
		);
	}
}

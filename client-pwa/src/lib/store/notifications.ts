import { writable } from 'svelte/store';
import type { iNotification } from '$lib/types/models';

/**
 * Latest notification frame received over the WebSocket (already filtered for
 * the current user by Notifications.svelte). The notifications page listens to
 * this store to live-refresh its table when a new notification arrives.
 */
export const incomingNotification = writable<iNotification | null>(null);

export function pushIncoming(notification: iNotification): void {
	incomingNotification.set(notification);
}

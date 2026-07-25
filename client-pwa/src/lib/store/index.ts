import { writable, readable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

export const apiUrl = readable('/api');
export const count = writable(0);
export const currentUser: Writable<any> = writable(null);

// Sidebar collapsed state — persisted across sessions
const storedCollapsed = browser ? localStorage.getItem('nav-collapsed') === 'true' : false;
export const navCollapsed = writable<boolean>(storedCollapsed);
navCollapsed.subscribe((v) => {
	if (browser) localStorage.setItem('nav-collapsed', String(v));
});

// Notifications page type filter (info/alert/""=all) — persisted across visits
const storedNotifType = browser ? localStorage.getItem('notification-type-filter') || '' : '';
export const notificationTypeFilter = writable<string>(storedNotifType);
notificationTypeFilter.subscribe((v) => {
	if (browser) localStorage.setItem('notification-type-filter', v);
});

export interface iToast {
	message: string;
	type: 'success' | 'error';
}
export const toast = writable<iToast | null>(null);

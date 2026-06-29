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

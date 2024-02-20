import { writable,readable } from 'svelte/store';

export const apiUrl = readable('');
export const count = writable(0);

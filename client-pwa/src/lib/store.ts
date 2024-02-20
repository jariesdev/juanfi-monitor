import { writable,readable } from 'svelte/store';

export const apiUrl = readable('/api');
export const count = writable(0);

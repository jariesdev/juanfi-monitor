import { writable,readable } from 'svelte/store';

export const apiUrl = readable('http://localhost:8000');
export const count = writable(0);

import { writable,readable } from 'svelte/store';

export const apiUrl = readable('http://192.46.225.21:8001/');
export const count = writable(0);

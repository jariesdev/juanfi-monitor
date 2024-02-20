import { writable,readable } from 'svelte/store';

export const apiUrl = readable('https://pwifi.jaries.dev/api');
export const count = writable(0);

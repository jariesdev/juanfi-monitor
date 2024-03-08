import {writable, readable, type Writable} from 'svelte/store';

export const apiUrl = readable('/api');
export const count = writable(0);

export const currentUser: Writable<any> = writable(null)
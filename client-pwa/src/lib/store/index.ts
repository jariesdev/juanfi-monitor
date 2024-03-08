import {writable, readable, type Writable} from 'svelte/store';

export const apiUrl = readable('http://127.0.0.1:8000');
export const count = writable(0);

export const currentUser: Writable<any> = writable(null)
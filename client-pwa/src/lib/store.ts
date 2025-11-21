import {writable, readable, type Writable} from 'svelte/store';

// export const apiUrl = readable('http://host.docker.internal:8000');
export const apiUrl = readable('http://localhost');
// export const apiUrl = readable('https://pwifi.jaries.dev/api');
export const count = writable(0);

export const currentUser: Writable<any> = writable(null)

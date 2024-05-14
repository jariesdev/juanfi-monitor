import { type Writable, writable } from 'svelte/store';

export const useLocalTimeFormat: Writable<boolean> = writable(true);

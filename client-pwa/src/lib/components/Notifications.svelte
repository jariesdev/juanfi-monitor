<script lang="ts">
	import {onMount, onDestroy} from 'svelte';
	import type {iNotification} from "$lib/types/models";
	import Notification from "$lib/components/Notification.svelte";
	import { baseWsUrl } from '$lib/env';

	let messages: string[] = $state([]);
	let inputValue: string = $state('');
	let ws: WebSocket;
	let activeNotification: string = $state('')
	let pageVisible: DocumentVisibilityState|undefined|null = $state('visible')

	onMount(() => {
		ws = new WebSocket(`${baseWsUrl}/ws`); // Replace with your WebSocket server address

		ws.onopen = () => {
			console.log('WebSocket connected');
		};

		ws.onmessage = (event) => {
			const notification = JSON.parse(event.data)

			if (notification.type === 'notification') {
				messages.push(notification.message);

				if (! activeNotification) {
					nextNotification()
				}
			}

		};

		ws.onclose = () => {
			console.log('WebSocket disconnected');
		};

		ws.onerror = (error) => {
			console.error('WebSocket error:', error);
		};
	});

	onDestroy(() => {
		if (ws) {
			ws.close();
		}
	});

	// function sendMessage() {
	// 	if (ws && ws.readyState === WebSocket.OPEN) {
	// 		ws.send(JSON.stringify({message: inputValue}));
	// 		inputValue = ''; // Clear input after sending
	// 	}
	// }

	const nextNotification = (): void => {
		if (messages.length > 0 && pageVisible === 'visible') {
			activeNotification = messages.shift() || ''
		} else {
			activeNotification = ''
		}
	}

	const handleVisibilityChange = (): void => {
			if (pageVisible === 'visible') {
				nextNotification()
			}
	}
</script>

<svelte:document bind:visibilityState={pageVisible} onvisibilitychange={handleVisibilityChange} />

<div class="notifications uk-width-expand uk-position-absolute uk-position-bottom">
		<Notification message={activeNotification} onHide={nextNotification} />
</div>

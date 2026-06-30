<script lang="ts">
	import {onMount, onDestroy} from 'svelte';
	import type {iNotification} from "$lib/types/models";
	import Notification from "$lib/components/Notification.svelte";
	import { baseWsUrl } from '$lib/env';
	import { setVendoProgress, setVendoOnline } from '$lib/store/vendoActivity';

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

				if (pageVisible !== 'visible' && ! activeNotification) {
					showNotification()
				} else {
					pushNotification(notification.message)
				}
			} else if (notification.type === 'vendo_refresh') {
				setVendoProgress(notification.vendo_id, notification.progress);
			} else if (notification.type === 'vendo_status') {
				setVendoOnline(notification.vendo_id, notification.online);
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

	const showNotification = (): void => {
		if (pageVisible !== 'visible') return;

		if (messages.length > 0) {
			activeNotification = messages.shift() || ''
		} else {
			activeNotification = ''
		}
	}

	const getNotification = (): string => {
		if (messages.length > 0) {
			return messages.shift() || ''
		}

		return activeNotification = ''
	}

	const pushNotification = (message: string): void => {
		if (!("Notification" in window)) {
			// Check if the browser supports notifications
			console.log("This browser does not support desktop notification");
		} else if (window.Notification.permission === "granted") {
			// Check whether notification permissions have already been granted;
			// if so, create a notification
			const notification = new window.Notification(message);
		} else if (window.Notification.permission !== "denied") {
			// We need to ask the user for permission
			window.Notification.requestPermission().then((permission) => {
				// If the user accepts, let's create a notification
				if (permission === "granted") {
					const notification = new window.Notification(message);
				}
			});
		}
	}

	const handleVisibilityChange = (): void => {
		if (pageVisible === 'visible') {
			showNotification()
		}
	}
</script>

<svelte:document bind:visibilityState={pageVisible} onvisibilitychange={handleVisibilityChange} />

<div class="notifications uk-width-expand uk-position-absolute uk-position-bottom">
		<Notification message={activeNotification} onHide={showNotification} />
</div>

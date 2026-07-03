<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { iNotification } from '$lib/types/models';
	import Notification from '$lib/components/Notification.svelte';
	import { baseWsUrl } from '$lib/env';
	import { setVendoProgress, setVendoOnline } from '$lib/store/vendoActivity';
	import { pushIncoming } from '$lib/store/notifications';

	interface Props {
		// null when the user is an admin (sees every notification); otherwise
		// only global frames (user_id null) and the user's own are shown.
		currentUserId?: number | null;
	}

	const { currentUserId = null }: Props = $props();

	let messages: string[] = $state([]);
	let inputValue: string = $state('');
	let ws: WebSocket | undefined;
	let activeNotification: string = $state('');
	let pageVisible: DocumentVisibilityState | undefined | null = $state('visible');

	// Reconnect state: exponential backoff so a dropped socket recovers on its own.
	let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
	let reconnectDelay = 1000;
	const maxReconnectDelay = 30000;
	let closed = false; // set on component teardown to stop reconnecting

	const connect = (): void => {
		ws = new WebSocket(`${baseWsUrl}/ws`);

		ws.onopen = () => {
			console.log('WebSocket connected');
			reconnectDelay = 1000; // reset backoff after a successful connection
		};

		ws.onmessage = (event) => {
			const notification = JSON.parse(event.data);

			if (notification.type === 'notification') {
				// Skip frames targeted at another user.
				if (
					currentUserId !== null &&
					notification.user_id != null &&
					notification.user_id !== currentUserId
				) {
					return;
				}
				pushIncoming(notification);
				messages.push(notification.message);

				if (pageVisible !== 'visible' && !activeNotification) {
					showNotification();
				} else {
					pushNotification(notification.message);
				}
			} else if (notification.type === 'vendo_refresh') {
				setVendoProgress(notification.vendo_id, notification.progress);
			} else if (notification.type === 'vendo_status') {
				setVendoOnline(notification.vendo_id, notification.online);
			}
		};

		ws.onclose = () => {
			console.log('WebSocket disconnected');
			scheduleReconnect();
		};

		ws.onerror = (error) => {
			console.error('WebSocket error:', error);
			ws?.close(); // triggers onclose → reconnect
		};
	};

	const scheduleReconnect = (): void => {
		if (closed || reconnectTimer) return;
		reconnectTimer = setTimeout(() => {
			reconnectTimer = undefined;
			connect();
		}, reconnectDelay);
		reconnectDelay = Math.min(reconnectDelay * 2, maxReconnectDelay);
	};

	onMount(() => {
		connect();
	});

	onDestroy(() => {
		closed = true;
		if (reconnectTimer) clearTimeout(reconnectTimer);
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
			activeNotification = messages.shift() || '';
		} else {
			activeNotification = '';
		}
	};

	const getNotification = (): string => {
		if (messages.length > 0) {
			return messages.shift() || '';
		}

		return (activeNotification = '');
	};

	const pushNotification = (message: string): void => {
		if (!('Notification' in window)) {
			// Check if the browser supports notifications
			console.log('This browser does not support desktop notification');
		} else if (window.Notification.permission === 'granted') {
			// Check whether notification permissions have already been granted;
			// if so, create a notification
			const notification = new window.Notification(message);
		} else if (window.Notification.permission !== 'denied') {
			// We need to ask the user for permission
			window.Notification.requestPermission().then((permission) => {
				// If the user accepts, let's create a notification
				if (permission === 'granted') {
					const notification = new window.Notification(message);
				}
			});
		}
	};

	const handleVisibilityChange = (): void => {
		if (pageVisible === 'visible') {
			showNotification();
		}
	};
</script>

<svelte:document bind:visibilityState={pageVisible} onvisibilitychange={handleVisibilityChange} />

<div class="notifications uk-width-expand uk-position-absolute uk-position-bottom">
	<Notification message={activeNotification} onHide={showNotification} />
</div>

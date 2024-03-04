<script lang="ts">
    import {browser} from '$app/environment';
    import {onDestroy, onMount} from "svelte";

    let isOnline: boolean = false
    let intervalId: any = null

    function setOnline(): void {
        isOnline = true
    }

    function setOffline(): void {
        isOnline = false
    }

    onMount(() => {
        if (!browser) return;
        intervalId = setInterval(() => {
            isOnline = window.navigator.onLine
        }, 500)

        window.addEventListener('online', setOnline);
        window.addEventListener('offline', setOffline);
    })
    onDestroy(() => {
        if (intervalId) {
            clearInterval(intervalId)
        }
        if (!browser) return;
        window.removeEventListener('online', setOnline);
        window.removeEventListener('offline', setOffline);
    })
</script>

{#if (!isOnline)}
    <div class="uk-position-fixed uk-position-bottom-center uk-box-shadow-small uk-margin-small-bottom">
        <div class="uk-alert-danger uk-margin-remove-bottom uk-text-center" uk-alert>
            Not connected to the internet. Data may not accurate.
        </div>
    </div>
{/if}

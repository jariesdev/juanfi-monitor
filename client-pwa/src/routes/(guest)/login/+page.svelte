<script lang="ts">
    import {apiUrl} from "$lib/store";
    import LoginForm from "./LoginForm.svelte";
    import {goto} from '$app/navigation'

    let baseApiUrl: string = ''
    let isProcessing: boolean = false

    function refreshLogs(): void {
        isProcessing = true


        const request = new Request(`${baseApiUrl}/token`, {method: "POST"});
        fetch(request)
            .then((response) => {
                console.log(response)
            })
            .finally(() => {
                isProcessing = false
            })
    }

    function handleSuccess(): void {
        goto('/home')
    }

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })
</script>


<div class="uk-section">
    <div class="uk-container">
        <h3 class="uk-text-center">App Login</h3>
        <div class="uk-grid uk-grid-small uk-child-width-1-2@s uk-flex-center">
            <LoginForm on:success={handleSuccess}></LoginForm>
        </div>
    </div>
</div>


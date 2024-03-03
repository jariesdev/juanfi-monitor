<script lang="ts">
    import {apiUrl} from "$lib/store";
    import {createEventDispatcher} from "svelte";

    const dispatcher = createEventDispatcher()
    let isProcessing: boolean = false
    let form = {
        name: '',
        api_url: '',
        api_key: ''
    }
    let baseApiUrl: string = ''

    function submit(): void {
        const formData = new FormData
        Object.keys(form).forEach(key => formData.append(key, form[key]));
        isProcessing = true
        const request = new Request(`${baseApiUrl}/vendo-machines`, {
            method: "POST",
            body: JSON.stringify(form),
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
        })
        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((data) => {
                console.log(data)
                form = {
                    name: '',
                    api_url: '',
                    api_key: ''
                }
                dispatcher('success')
            })
            .catch((error) => {
                console.error(error);
            })
            .finally(() => {
                isProcessing = false
            })
    }

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })
</script>

<div>
    <form on:submit|preventDefault={submit}>
        <div class="uk-margin-small-bottom">
            <label class="uk-form-label" for="name">Name</label>
            <input bind:value={form.name} id="name" class="uk-input" type="text" placeholder="e.g. Juan Vendo">
        </div>
        <div class="uk-margin-small-bottom">
            <label class="uk-form-label" for="api_url">API URL</label>
            <input bind:value={form.api_url} id="api_url" class="uk-input" type="text" placeholder="e.g. http:10.0.12.2">
        </div>
        <div class="uk-margin-bottom">
            <label class="uk-form-label" for="api_key">API Key</label>
            <input bind:value={form.api_key} id="api_key" class="uk-input" type="text" placeholder="e.g. abcd1234">
        </div>
        <div class="uk-text-center">
            <button class="uk-modal-close uk-button" type="button">Cancel</button>
            <button class="uk-button uk-button-primary uk-margin-left"
                    type="submit"
                    disabled="{isProcessing}">
                Save
            </button>
        </div>
    </form>
</div>
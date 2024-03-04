<script lang="ts">
    import LoginForm from "./LoginForm.svelte";
    import {goto} from '$app/navigation'

    function handleSuccess(): void {
        goto('/home')
    }

</script>


<div class="uk-section">
    <div class="uk-container">
        <h3 class="uk-text-center">App Login</h3>
        <div class="uk-grid uk-grid-small uk-child-width-1-2@s uk-flex-center">
            <LoginForm on:success={handleSuccess}></LoginForm>
        </div>
    </div>
</div>


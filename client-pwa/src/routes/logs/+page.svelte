<script lang="ts">
    import LogTable from "./LogTable.svelte";

    let isReloading: boolean = false
    let reloadData: Function

    function refreshLogs(): void {
        isReloading = true

        const request = new Request("http://192.46.225.21:8000/log/refresh", {method: "POST"});
        fetch(request)
            .then(() => {
                reloadData()
            })
            .finally(() => {
                isReloading = false
            })
    }
</script>


<div class="uk-section">
    <div class="uk-container">
        <div class="uk-margin-bottom">
            <button class="uk-button uk-button-primary" disabled={isReloading} on:click={refreshLogs}>Refresh</button>
        </div>

        <LogTable bind:loadData={reloadData}></LogTable>
    </div>
</div>


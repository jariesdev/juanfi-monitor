<script lang="ts">
    import {apiUrl} from "$lib/store";
    import WithdrawalTable from "./WithdrawalTable.svelte";

    let isReloading: boolean = false
    let reloadData: Function
    let baseApiUrl: string = ''

    function refreshLogs(): void {
        isReloading = true

        const request = new Request(`${baseApiUrl}/withdrawals`, {method: "GET"});
        fetch(request)
            .then(() => {
                reloadData()
            })
            .finally(() => {
                isReloading = false
            })
    }

    apiUrl.subscribe(function(value) {
        baseApiUrl = value
    })
</script>


<div class="uk-section">
    <div class="uk-container">
        <WithdrawalTable bind:loadData={reloadData}></WithdrawalTable>
    </div>
</div>


<script lang="ts">
    import {apiUrl} from "$lib/store";
    import DataTable from "$lib/components/DataTable.svelte";
    import DateTime from "$lib/components/DateTime.svelte";
    import type {iVendo} from "$lib/interfaces";
    import {onMount} from "svelte";

    let baseApiUrl: string = ''

    let headers = [
        {label: 'Time', field: 'sale_time'},
        {label: 'MAC Address', field: 'mac_address'},
        {label: 'Vendo', field: 'vendo.name'},
        {label: 'Amount', field: 'amount'},
        {label: 'Voucher', field: 'voucher'}
    ];
    let vendoId: number
    let vendos: iVendo[] = []
    $: ({filters} = {
        filters: {
            vendo_id: vendoId
        }
    })

    function loadOptions(): void {
        let url = `${baseApiUrl}/vendo-machines`
        const request = new Request(url, {method: 'GET'});
        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((response) => {
                vendos = response.data
            })
            .catch(() => {
                vendos = []
            })
    }

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })

    onMount(() => {
        loadOptions()
    })
</script>

<DataTable url={`${baseApiUrl}/sales`} headers={headers} bind:filters={filters}>
    <div slot="before-table">
        <div class="uk-margin-small-top uk-grid uk-grid-small uk-child-width-1-2@s uk-child-width-1-3@m"
             style="row-gap: 15px">
            <div>
                <select bind:value={vendoId}
                        name="vendo_id"
                        id="vendo_id"
                        class="uk-select uk-form-small uk-child-width-1-1">
                    <option value="{undefined}">All</option>
                    {#each vendos as vendo}
                        <option value="{vendo.id}">{vendo.name}</option>
                    {/each}
                </select>
            </div>
        </div>
    </div>
    <span slot="cell" let:item let:header let:getCellValue>
        {#if (header.field === 'sale_time')}
            <DateTime humanized={true} date={item.sale_time}/>
        {:else }
            {getCellValue(item, header)}
        {/if}
    </span>
</DataTable>

<script lang="ts">
    import {onMount, onDestroy} from "svelte";
    import debounce from 'lodash/debounce'
    import {apiUrl} from "$lib/store";
    import VendoForm from "$lib/components/VendoForm.svelte";
    import type {iVendo} from "$lib/interfaces.js";

    let vendoMachines: iVendo[] = []
    let searchInput: string = ''
    let isLoading: boolean = false
    let baseApiUrl: string = ''
    let controller: AbortController | undefined = undefined

    export const loadData: Function = debounce(async (): Promise<void> => {
        isLoading = false

        let url = `${baseApiUrl}/vendo-machines`
        if (!!searchInput) {
            url = `${url}?q=${searchInput}`
        }
        controller = new AbortController()
        const signal = controller.signal
        const request = new Request(url, {method: "GET", signal: signal});

        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((response) => {
                vendoMachines = response.data
            })
            .catch((error) => {
                console.error(error);
                vendoMachines = []
            })
            .finally(() => {
                isLoading = false
            })
    }, 250, {maxWait: 1000})

    function withdrawCurrentSales(): string {
        alert('Feature not available yet.')
    }

    $: searchInput, loadData();

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })

    onMount(() => {
        loadData()
    })
    onDestroy(() => {
        controller && controller.abort('component destroyed')
    })
</script>

<div class="uk-card uk-card-default uk-card-body">
    <div class="uk-column-1-2@s">
        <h3 class="uk-card-title">Vendo Machines</h3>
        <div class="uk-text-right">
            <input bind:value={searchInput} class="uk-input uk-form-width-medium uk-form-small" type="search"
                   placeholder="Search" aria-label="Input">
        </div>
    </div>
    <div class="uk-overflow-auto">
        <table class="uk-table uk-table-divider">
            <thead>
            <tr>
                <th>Name</th>
                <th>API URL</th>
                <th>Current Sales</th>
                <th>Total Sales</th>
            </tr>
            </thead>
            <tbody>
            {#if vendoMachines.length === 0}
                <tr>
                    <td colspan="6" class="uk-text-center uk-text-italic uk-text-muted uk-text-small">
                        No record yet.
                    </td>
                </tr>
            {/if}
            {#each vendoMachines as vendo}
                <tr>
                    <td>{vendo.name}</td>
                    <td>{vendo.api_url}</td>
                    <td>{vendo.current_sales}</td>
                    <td>{vendo.total_sales}</td>
                    <td class="uk-text-nowrap">
                        <a href="{`/vendo/${vendo.id}/status`}" class="uk-margin-small-right">
                            <span class="uk-text-info" uk-icon="icon: info"></span>
                        </a>
                        <a href="#delete" on:click={withdrawCurrentSales} class="uk-margin-small-right">
                            <span class="uk-text-info" uk-icon="icon: pencil"></span>
                        </a>
                        <a href="#delete" on:click={withdrawCurrentSales}>
                            <span class="uk-text-danger" uk-icon="icon: trash"></span>
                        </a>
                    </td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>
    <div>
        <button uk-toggle="target: #add-modal" type="button" class="uk-button uk-button-primary uk-position-relative">
            Add vendo
        </button>
    </div>
    <!-- This is the modal -->
    <div id="add-modal" uk-modal class="uk-modal">
        <div class="uk-modal-dialog uk-modal-body">
            <h2 class="uk-modal-title">New Vendo</h2>
            <div class="uk-margin-small-bottom">
                <VendoForm on:success={loadData}/>
            </div>
        </div>
    </div>
</div>

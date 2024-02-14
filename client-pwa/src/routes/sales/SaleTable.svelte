<script lang="ts">
    import {onMount} from "svelte";
    import debounce from 'lodash/debounce'

    interface iSale {
        id: number
        sale_time: string
        mac_address: string
        voucher: string
        amount: number
        created_at: string
    }

    let sales: iSale[] = []
    let searchInput: string = ''
    let isLoading: boolean = false

    export const loadData: Function = debounce(async (): Promise<void> => {
        isLoading = false

        let url = "http://192.46.225.21:8000/sales"
        if (!!searchInput) {
            url = `${url}?q=${searchInput}`
        }
        const request = new Request(url, {method: "GET"});

        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((response) => {
                sales = response.data
            })
            .catch((error) => {
                console.error(error);
                sales = []
            })
            .finally(() => {
                isLoading = false
            })
    }, 250, {maxWait: 1000})

    function humanizeTime(time: string): string {
        return new Date(Date.parse(time)).toLocaleString()
    }

    $: searchInput, loadData();

    onMount(() => {
        loadData()
    })

</script>

<div class="uk-card uk-card-default uk-card-body">
    <div class="uk-flex uk-flex-between">
        <h3 class="uk-card-title">Sales</h3>
        <div>
            <input bind:value={searchInput} class="uk-input uk-form-width-medium uk-form-small" type="search"
                   placeholder="Search" aria-label="Input">
        </div>
    </div>
    <table class="uk-table uk-table-divider">
        <thead>
        <tr>
            <th>Time</th>
            <th>Mac Address</th>
            <th>Voucher</th>
            <th>Amount</th>
            <th>Created At</th>
        </tr>
        </thead>
        <tbody>
        {#each sales as log}
            <tr>
                <td title="{humanizeTime(log.sale_time)}">{log.sale_time}</td>
                <td>{log.mac_address}</td>
                <td>{log.voucher}</td>
                <td>{log.amount}</td>
                <td title="{humanizeTime(log.created_at)}">{log.created_at}</td>
            </tr>
        {/each}
        </tbody>
    </table>
</div>

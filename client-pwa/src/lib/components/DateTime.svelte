<script lang="ts">
    export let date: string = ''
    export let humanized: boolean = false

    function humanizeTime(): string {
        return date ? new Date(Date.parse(date)).toLocaleString() : ''
    }
</script>

<div on:click on:keydown on:keyup class="{$$restProps.class || ''}">
    {#if humanized}
        <span title="{date}">
            {humanizeTime()}
        </span>
    {:else}
        <span title="{humanizeTime()}">
            {date}
        </span>
    {/if}
</div>

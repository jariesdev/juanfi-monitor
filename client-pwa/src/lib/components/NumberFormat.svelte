<script lang="ts">
    export let value: number | null = null

    function getFormattedValue() {
        if(value === null) {
            return '';
        }
        return new Intl.NumberFormat('en-IN', {maximumSignificantDigits: 3}).format(value)
    }
</script>

<span>{getFormattedValue()}</span>
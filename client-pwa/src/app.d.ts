// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
// and what to do when importing types
declare namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface Platform {}
    interface Locals {
        user: {
            id: number
            username: string
            role: { id: number; name: string; permissions: string[] } | null
        } | null
        permissions: string[]
    }
}

import {apiUrl, currentUser} from "$lib/store";

let baseApiUrl: string

function getAuthToken(): string | null {
    return localStorage.getItem("auth_token")
}

export const get = (url: string): Promise<Response> => {
    const token: string | null = getAuthToken()
    const headers: Headers = new Headers()
    if (token) {
        headers.set("Authorization", `Bearer ${token}`)
    }
    const request: Request = new Request(`${baseApiUrl}${url}`, {method: "GET", headers});
    return fetch(request)
}

export const post = (url: string, formData: FormData | any): Promise<Response> => {
    const token: string | null = getAuthToken()
    const headers: Headers = new Headers()
    if (token) {
        headers.set("Authorization", `Bearer ${token}`)
    }
    const request: Request = new Request(`${baseApiUrl}${url}`, {method: "POST", body: formData});
    return fetch(request)
}


apiUrl.subscribe(function (value): void {
    baseApiUrl = value
})
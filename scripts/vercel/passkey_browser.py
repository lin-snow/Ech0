# SPDX-License-Identifier: AGPL-3.0-or-later
"""Exercise a real browser WebAuthn ceremony against the two smoke-test processes."""
import json
import sys
from urllib.parse import urlsplit

from playwright.sync_api import sync_playwright


def main():
    fixture = json.load(sys.stdin)
    origin = "http://localhost:" + str(fixture["first"])
    with sync_playwright() as driver:
        browser = driver.chromium.launch(executable_path=fixture.get("chrome"), headless=True)
        page = browser.new_page()
        session = page.context.new_cdp_session(page)
        session.send("WebAuthn.enable")
        session.send("WebAuthn.addVirtualAuthenticator", {"options": {
            "protocol": "ctap2", "transport": "internal", "hasResidentKey": True,
            "hasUserVerification": True, "isUserVerified": True,
            "automaticPresenceSimulation": True,
        }})

        # Both workers serve the same origin, as they would behind a load balancer.
        def finish(route):
            url = "http://localhost:" + str(fixture["second"]) + urlsplit(route.request.url).path
            response = route.fetch(url=url)
            route.fulfill(response=response)
            response.dispose()

        page.route("**/api/passkey/*/finish", finish)
        page.goto(origin, wait_until="domcontentloaded")
        invalid = page.request.post(origin + "/api/passkey/login/begin",
                                    data={}, headers={"Origin": "https://unrelated.invalid"})
        assert invalid.json()["code"] != 1
        invalid.dispose()
        result = page.evaluate("""async token => {
            async function post(path, data, authenticated = false) {
                const headers = { 'Content-Type': 'application/json' }
                if (authenticated) headers.Authorization = 'Bearer ' + token
                const response = await fetch('/api/passkey/' + path, {
                    method: 'POST', headers, body: JSON.stringify(data)
                })
                return response.json()
            }
            const reg = await post('register/begin', {device_name:'smoke-browser'}, true)
            if (reg.code !== 1) throw Error('registration begin failed')
            const options = PublicKeyCredential.parseCreationOptionsFromJSON(reg.data.publicKey)
            const credential = await navigator.credentials.create({publicKey:options})
            const attestation = {nonce:reg.data.nonce, credential:credential.toJSON()}
            const registered = await post('register/finish', attestation, true)
            if (registered.code !== 1) throw Error('registration finish failed')
            const login = await post('login/begin', {})
            if (login.code !== 1) throw Error('login begin failed')
            const assertion = await navigator.credentials.get({
                publicKey:PublicKeyCredential.parseRequestOptionsFromJSON(login.data.publicKey)
            })
            const payload = {nonce:login.data.nonce, credential:assertion.toJSON()}
            const loggedIn = await post('login/finish', payload)
            const replay = await post('login/finish', payload)
            return Boolean(loggedIn.code === 1 && loggedIn.data.access_token && replay.code !== 1)
        }""", fixture["token"])
        assert result, "failed to authenticate exactly once"
        browser.close()
    print("OK: browser passkey registration, discoverable login and single-use challenge across workers")


if __name__ == "__main__":
    main()

# End-to-end (e2e) tests for web

Browser tests written in plain English for [Kane CLI](https://www.testmuai.com/kane-cli/).

| Test | What it covers |
|---|---|
| `setup_test.md` | First setup (owner account), logging in and landing on the feed |
| `echo_test.md` | First setup, logging in, publishing a public echo and seeing it as a logged-out visitor |

## Run

Start a fresh instance on port 4877:

```bash
docker run -d --name ech0-e2e -p 4877:6277 -e JWT_SECRET="e2e-secret" sn0wl1n/ech0:latest
npm install -g @testmuai/kane-cli
kane-cli login
kane-cli testmd run e2e/kane/setup_test.md
kane-cli testmd run e2e/kane/echo_test.md
```

Both tests start from the first setup page, so recreate the container before each run.

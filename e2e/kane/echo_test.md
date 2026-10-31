---
mode: testing
url: http://localhost:4877/
max_steps: 40
---

# Publish a public echo

## Create the owner account and log in
On the first setup page, type "sam@example.com" in the "Owner email" field, "samrivera" in the "Owner username" field and "Ech0-Test-2026" in the "Owner password" field, then click "Initialize".
On the login page, type "samrivera" in the username field and "Ech0-Test-2026" in the password field, then click "Login".
Verify the home feed is shown with "Publish" in the menu on the right.

## Publish an echo
Click "Publish" in the menu on the right.
Type "Finished reading the Go memory model docs today" in the "Share what's on your mind~" box.
Click the "+" button below the box on the right, then click "Publish as public".
Verify the feed shows "Finished reading the Go memory model docs today".

## Check it as a visitor
Click the logout icon at the top of the page.
Go to http://localhost:4877/.
Verify the feed still shows "Finished reading the Go memory model docs today" and the menu no longer has "Settings".

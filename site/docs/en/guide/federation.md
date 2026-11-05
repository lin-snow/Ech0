---
title: Federation & aggregation
description: Connect federation and multi-instance timelines with the Hub
---

Ech0 can link **multiple instances** together: first register the other side's address in **Connect**, then merge the content of connected instances into a single timeline at **`/hub`**.

---

## Connect: instance federation

In **System settings → Connect**, enter **the root address of the other Ech0 instance** (for example `https://friend.example.com`) and save to add it.

**Make sure** to fill in the full URL of **your own site** under **System settings → Service address** (it must include `http://` or `https://`). Otherwise other instances may not see your avatar or site name when fetching your public information.

The “Connected Ech0” section in the home sidebar lists the instances you have added; an address that has been unreachable for a long time may not appear there (the backend only shows instances whose information was fetched successfully).

---

## Hub: federated aggregation timeline

Path: **`/hub`**.

Based on the instance list you saved in Connect, the Hub fetches public Echoes from each site and merges them in chronological order. Typical characteristics include:

- When one instance is temporarily unreachable, other instances' content remains browsable (behavior depends on your version).
- The merged result is deduplicated and sorted; scroll down to load more (per the UI).
- Instances unreachable for a long time may be dropped from the current Hub view.

**Order of use**: configure **Connect** first, then open **`/hub`**.

---

## Relationship with RSS and the single-site timeline

- Single-site home: only the current instance.
- Hub: multiple sites merged.
- RSS: usually subscribes to the **current instance's** public output; a different entry point from the Hub.

If you only run a single instance, you can skip Connect / Hub entirely.

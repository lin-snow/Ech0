---
title: Writing & editing
description: From new to published — modes, drafts, images, and extension cards
---

An **Echo** usually starts with "write the body text", then gains images, link cards, music, and so on as needed. UI wording may shift slightly between versions; the description below matches the current mainstream layout.

---

## Publishing your first Echo (flow)

1. Find the **Publish / New** entry in the UI (usually near the timeline) and open the editor.
2. Enter your Markdown body in the **main editor area**; common syntax is supported (headings, lists, quotes, etc., subject to the implementation).
3. When you need images, switch to the **Images** area and drag files in or pick them for upload (see "Images and uploads" below).
4. When you need "cards" such as music, video, or GitHub repositories, use **Panel → Extensions** and follow the prompts to enter a URL or metadata.
5. Once everything looks right, **publish**. If comments are enabled on the site, see [Comments](/docs/guide/comment) for how they appear to visitors.

If your post is not visible after publishing, first check that the current account has **posting permission** (see [FAQ](/docs/start/faq)).

---

## Modes

The editor supports several modes, reached through the **mode switch / panel** in the UI (names follow the current version):

| Mode        | Purpose                                                                        |
| ----------- | ------------------------------------------------------------------------------ |
| Main editor | Write the Markdown body                                                        |
| Panel       | Choose an extension type, open the inbox, and so on                            |
| Extensions  | Configure card data for music, video, GitHub repositories, website links, etc. |
| Images      | Upload attachments, choose a layout (such as masonry), toggle compression      |
| Tags        | Add or remove tags for the current Echo                                        |
| Inbox       | View system notifications and the like (if the version exposes this entry)     |

After entering Extensions from the panel, fill in the URL or metadata as prompted and the corresponding card is generated.

---

## Drafts

While you edit, **drafts are saved automatically** to your browser's local storage, so a page closed by accident does not lose your content. Drafts have an **expiry** (about one day, for example; subject to the implementation) and are cleaned up afterwards.

When you reopen the editor, if an unsent draft exists you are usually prompted to **restore or discard** it; editing an **already-published** Echo may behave slightly differently from "writing a new one", per the current version.

---

## Images and uploads

- Supports **drag and drop** and file selection for uploads.
- **Smart compression** can be enabled to reduce size: Safari commonly outputs JPEG while other browsers usually produce WebP (subject to the implementation).
- If the instance is configured with **S3**, large files may be uploaded directly or proxied through the backend, depending on storage settings and the network.
- Per-batch count and size limits are shown in the UI or in the **OpenAPI** spec.

---

## With comments, RSS, and the Hub

- **RSS**: once an Echo is published, readers can subscribe if the site offers RSS (the entry point follows the UI).
- **Comments**: the comment entry is visible to visitors only after it is enabled in [Comments](/docs/guide/comment).
- **Hub / Connect**: see [Federation & aggregation](/docs/guide/federation) for browsing across multiple instances; a single instance can ignore this.

When editing published content, use the "Edit" entry of the current version; when media URLs and storage backends are involved, mind the notes in [Object storage](/docs/guide/s3) and [Data control](/docs/guide/datacontrol).

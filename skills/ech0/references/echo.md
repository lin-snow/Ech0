# Publishing and editing Echoes

The Echo commands are `echo.create`, `echo.list`, `echo.get`, `echo.update`
and `echo.delete`, plus `file.upload` for attachments. **Create, update and
delete require an admin account.** Reads work for any logged-in account.

An Echo carries:

| Field | Meaning |
| --- | --- |
| `content` | Markdown body |
| `tags` | names; an existing tag with the same name is reused |
| `echo_files` | attachments: `{file_id, sort_order}` |
| `extension` | card payload: `{type, payload}` |
| `private` | visible only to the owner |
| `layout` | `waterfall` (default), `grid`, `horizontal`, `carousel`, `stack`, `none` |
| `created_at` | Unix seconds; defaults to now |

At least one of `content`, `echo_files` or `extension` must be non-empty.
Attachments inside one Echo must share a category (image, video, audio, file…),
and audio or video is limited to one per Echo.

## Create

```bash
ech0 echo create --content "今天的想法" --tag daily
ech0 echo create --content "只给自己看" --private
ech0 echo create --content "配图" --file-id <file_id>
```

- `--tag <name>` appends a tag and is repeatable; `--tags` takes a JSON array of
  objects instead: `--tags '[{"name":"daily"}]'`.
- `--file-id <id>` references an uploaded file and is repeatable; `--echo-files`
  takes JSON objects when you need `sort_order`.
- `--extension '{"type":"weather","payload":{"city":"Hangzhou"}}'` renders a card.
- `--layout`, `--private`, `--created-at` set the remaining fields.

Upload first; the upload prints a JSON object whose `id` is the file id:

```bash
ech0 file upload --file photo.png --json
cat photo.png | ech0 file upload --file - --file-name photo.png --json
```

`--storage-type local|s3|object` picks the target when the instance has more than
one. `external` files are registered by URL, not uploaded:

```bash
ech0 file external --url https://example.com/photo.png --category image --json
```

`upload` and `external` stay separate commands because they are separate
operations: `upload` sends `multipart/form-data`, `external` sends a JSON body
and never transfers bytes. Both still create the same kind of file record, so
`--file-id`, `file list` and `file get` treat them identically;
`file download` redirects external files to their original URL.

### Browse, inspect and download

```bash
ech0 file list --page 1 --page-size 20 --search photo --json
ech0 file get --id <file_id> --json            # metadata, including url
ech0 file tree --storage-type local            # browse by storage prefix
ech0 file download --id <file_id> --output ./photo.png
```

`file download` streams the content: `--output -` writes stdout, and a non-TTY
stdout is the default target. `file tree` needs the storage type the instance
uses (`local`, `s3`, `object`).

### Full payload

`--from <file>` reads a JSON object (`-` for stdin); `--edit` opens the same
template in `$EDITOR`. The template lists every field the server accepts.

```bash
ech0 echo create --from echo.json
ech0 echo get --id <echo_id> --json | ech0 echo update --id <echo_id> --from -
```

Merge rules: `--from`/`--edit` values override same-named flags, while `tags` and
`echo_files` from the payload, the flags and `--tag`/`--file-id` accumulate.

## Update

```bash
ech0 echo get --id <echo_id> --json
ech0 echo update --id <echo_id> --content "改过的正文"
ech0 echo update --id <echo_id> --content "改过的正文" --tag daily
```

Update is a **full replacement**: `content`, `private` and `layout` are written
as sent, and `echo_files`, `extension` and `tags` are replaced by what you pass.
Omitted values are cleared, so fetch the Echo first and send back the fields you
want to keep. `--id` is required.

## Delete

```bash
ech0 echo delete --id <echo_id> --yes
```

Destructive and admin-only. Interactive terminals print the exact `DELETE`
target and ask `[y/N]`; non-interactive runs (agents, pipes, CI) never prompt
and require `--yes`, so ask the human before passing it. Verify `--server` and
the id first. Deleting an Echo also deletes its attachments.

## Read

```bash
ech0 echo get --id <echo_id>
ech0 echo list --page 1 --page-size 20 --search daily
```

Responses look like `{code, msg, data}`. `ech0 echo list` returns `data.items`
plus `total`.

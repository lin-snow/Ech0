# Tags

Tags are created implicitly by `echo create` / `echo update --tag`. The CLI only
needs to look them up and clean them up; there is no `tag create`.

## List tags

```bash
ech0 tag list --json
```

Each entry has `id`, `name`, `usage_count` and `created_at`. Use the ids to
filter the timeline:

```bash
ech0 echo list --tag-ids <tag_id>,<tag_id> --json
```

`echo.by.tag` is deprecated; `echo list --tag-ids` is the supported path.

## Delete a tag

```bash
ech0 tag delete --id <tag_id> --yes
```

Admin only. Deletes the tag row and removes it from every Echo; the posts
themselves are kept. This cannot be undone.

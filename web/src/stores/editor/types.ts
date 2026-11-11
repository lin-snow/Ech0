// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export type LocationToAdd = {
  latitude: number | null
  longitude: number | null
  placeholder: string
}

export type ExtensionToAdd = {
  extension: string
  extension_type: string
}

export type WebsiteToAdd = {
  title: string
  site: string
}

export type TweetToAdd = {
  url: string
  username: string
  statusId: string
}

export type EditorSession = { kind: 'compose' } | { kind: 'update'; echo: App.Api.Ech0.Echo }

export type EditorDraft = {
  savedAt: number
  echoToAdd: Pick<App.Api.Ech0.EchoToAdd, 'content' | 'private' | 'layout' | 'extension'>
  filesToAdd: App.Api.Ech0.FileToAdd[]
  websiteToAdd: WebsiteToAdd
  videoURL: string
  musicURL: string
  githubRepo: string
  extensionToAdd: ExtensionToAdd
  locationToAdd: LocationToAdd
  tweetToAdd: TweetToAdd
  tagToAdd: string[]
}

export type Translate = (key: string, params?: Record<string, unknown>) => string

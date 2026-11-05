// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { DocLocale } from "./locale";

type DocsUi = {
  docsTitle: string;
  docsDescription: string;
  allDocs: string;
  home: string;
  documentation: string;
  intro: string;
  startHere: string;
  allGuides: string;
  browse: string;
  onThisPage: string;
  language: string;
};

export const DOCS_UI: Record<DocLocale, DocsUi> = {
  en: {
    docsTitle: "Documentation — Ech0",
    docsDescription: "Ech0 documentation: deployment, features, and policies.",
    allDocs: "All docs",
    home: "Home",
    documentation: "Documentation",
    intro: "Start with the essentials, then explore features and reference.",
    startHere: "Start here",
    allGuides: "All guides",
    browse: "Browse",
    onThisPage: "On this page",
    language: "Language",
  },
  zh: {
    docsTitle: "文档 — Ech0",
    docsDescription: "Ech0 文档：部署、功能与使用说明。",
    allDocs: "全部文档",
    home: "首页",
    documentation: "文档",
    intro: "先看基础，再浏览功能与参考。",
    startHere: "从这里开始",
    allGuides: "全部指南",
    browse: "浏览",
    onThisPage: "本页目录",
    language: "语言",
  },
};

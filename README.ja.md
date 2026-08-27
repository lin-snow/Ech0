<div align="center">

<img alt="Ech0" src="./docs/imgs/logo.svg" width="150">

# Ech0

[プレビュー](https://memo.vaaat.com/) · [公式サイト & ドキュメント](https://www.ech0.app/) · [リリース](https://lin-snow.github.io/Ech0/) · [Ech0 Hub](https://hub.ech0.app/)

<a title="en-US" href="./README.md"><img src="https://img.shields.io/badge/-English-545759?style=for-the-badge" alt="English"></a> <a title="zh" href="./README.zh.md"><img src="https://img.shields.io/badge/-简体中文-545759?style=for-the-badge" alt="简体中文"></a> <a title="de" href="./README.de.md"><img src="https://img.shields.io/badge/-Deutsch-545759?style=for-the-badge" alt="Deutsch"></a> <img src="https://img.shields.io/badge/-日本語-F54A00?style=for-the-badge" alt="日本語">

[![GitHub release](https://img.shields.io/github/v/release/lin-snow/Ech0?logo=github&logoColor=white&labelColor=%20%23528BFF&color=%20%23155EEF)](https://github.com/lin-snow/Ech0/releases)
[![License](https://img.shields.io/github/license/lin-snow/Ech0?labelColor=%20%237D89B0&color=%20%235D6B98)](./LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/lin-snow/Ech0?style=flat)](https://goreportcard.com/report/github.com/lin-snow/Ech0)
[![Go Version](https://img.shields.io/github/go-mod/go-version/lin-snow/Ech0?logo=go&logoColor=white&labelColor=%20%235FD4EC&color=%20%2300ADD8)](./go.mod)
[![Release Build](https://img.shields.io/github/actions/workflow/status/lin-snow/Ech0/release.yml?logo=github&logoColor=white&label=build&labelColor=%20%2332B583&color=%20%2312B76A)](https://github.com/lin-snow/Ech0/actions/workflows/release.yml)
[![i18n](https://img.shields.io/badge/i18n-4_locales-7A5AF8?logo=googletranslate&logoColor=white&labelColor=%20%239B8AFB)](./web/src/locales/messages)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lin-snow/Ech0)
[![Docker Pulls](https://img.shields.io/docker/pulls/sn0wl1n/ech0?logo=docker&logoColor=white&labelColor=%20%235CADF6&color=%20%232496ED)](https://hub.docker.com/r/sn0wl1n/ech0)
[![Docker Image Size](https://img.shields.io/docker/image-size/sn0wl1n/ech0/latest?logo=docker&logoColor=white&labelColor=%20%235CADF6&color=%20%231570EF)](https://hub.docker.com/r/sn0wl1n/ech0)
[![Stars](https://img.shields.io/github/stars/lin-snow/Ech0?logo=github&logoColor=white&labelColor=%20%23FDB022&color=%20%23DC6803)](https://github.com/lin-snow/Ech0/stargazers)
[![Forks](https://img.shields.io/github/forks/lin-snow/Ech0?logo=github&logoColor=white&labelColor=%20%232ED3B7&color=%20%2315B79E)](https://github.com/lin-snow/Ech0/network/members)
[![Discussions](https://img.shields.io/github/discussions/lin-snow/Ech0?logo=github&logoColor=white&labelColor=%20%23B692F6&color=%20%239E77ED)](https://github.com/lin-snow/Ech0/discussions)
[![Last Commit](https://img.shields.io/github/last-commit/lin-snow/Ech0?logo=github&logoColor=white&labelColor=%20%238098F9&color=%20%23444CE7)](https://github.com/lin-snow/Ech0/commits/main)
[![Contributors](https://img.shields.io/github/contributors/lin-snow/Ech0?logo=github&logoColor=white&labelColor=%20%23FDA29B&color=%20%23F04438)](https://github.com/lin-snow/Ech0/graphs/contributors)
[![Sponsor](https://img.shields.io/badge/sponsor-Afdian-EE46BC?logo=githubsponsors&logoColor=white&labelColor=%20%23FD8FB8)](https://afdian.com/a/l1nsn0w)

<br />

<a href="https://hellogithub.com/repository/lin-snow/Ech0" target="_blank"><img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=neutral" alt="Featured｜HelloGitHub" width="250" height="54" /></a>

</div>

> セルフホスト型のパーソナル・マイクロブログ。あなたのタイムラインを共有・議論しつつ、データの所有権はあなたの手に。

Memos のようなツールは思いつきを素早く記録するのに最適です。Ech0 は「その先」のためのツールです — 個人のタイムラインに公開し、他の人がフォローしたり交流したりできるようにします。
自分のサーバーで運用してコンテンツの完全なコントロールを保ちながら、任意のコメントや共有によってつながりを残せます。
軽量で、デプロイが簡単、そして完全にオープンソースです。

**こんな人にぴったり：**
- 自分のドメインで公開・準公開のタイムラインを運用したい
- 短い投稿、リンク、メディアをひとつのシンプルな UI から発信したい
- データの所有権を保ちつつ、RSS や任意のコメントも欲しい
- 重厚な SNS ではなく、軽い社会的つながりのある個人スペースが欲しい

**おそらく合わない：**
- 双方向リンクのナレッジベース型ワークフロー（Obsidian 風 PKM など）
- チーム前提の共同編集ドキュメント環境（Notion 風など）
- タイムラインや公開機能を一切必要としない、純粋なメモ/ノート専用ツール

![インターフェイス・プレビュー](./docs/imgs/screenshot.png)

---

<details>
   <summary><strong>目次</strong></summary>

- [60 秒で試す](#60-秒で試す)
- [機能一覧](#機能一覧)
- [デプロイとアップグレード](#デプロイとアップグレード)
- [FAQ](#faq)
- [フィードバックとコミュニティ](#フィードバックとコミュニティ)
- [オープンソース運営と開発](#オープンソース運営と開発)
- [スポンサーと謝辞](#スポンサーと謝辞)
- [Star ヒストリー](#star-ヒストリー)

</details>

---

## 60 秒で試す

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

その後 `http://ip:6277` を開きます：

1. 最初のアカウントを登録します。
2. 最初のアカウントが Owner（管理者権限）になります。
3. デフォルトでは、投稿は権限のあるアカウントに制限されます。

Docker Compose や Helm を使った手順は [かんたんデプロイ](#かんたんデプロイ) を参照してください。

## 機能一覧

<details>
<summary><strong>クリックして全機能を表示</strong></summary>

### ハイライト

- ☁️ **軽量で効率的なアーキテクチャ**：低リソース・小サイズのイメージで、個人サーバーから ARM デバイスまで対応。
- 🚀 **すばやいデプロイ**：すぐ使える Docker デプロイ。インストールから初回起動までコマンド 1 つ。
- 📦 **自己完結したディストリビューション**：完全なバイナリとコンテナイメージ。追加のランタイム依存なし。
- 💻 **クロスプラットフォーム**：Linux、Windows、ARM デバイス（例：Raspberry Pi）に対応。
- 🧳 **持ち運べるコンテンツカプセル**：書いた内容を自己完結した[カプセル](./docs/usage/capsule.md)としてエクスポート。別のインスタンスへインポートしたり、コマンド 1 つで静的サイトにビルドして任意のホスティングへ配置できます。

### ストレージとデータ

- 🗂️ **VireFS 統合ストレージ層**：**VireFS** によりローカルストレージと S3 互換オブジェクトストレージを統一管理。
- ☁️ **S3 オブジェクトストレージ対応**：S3 互換オブジェクトストレージをネイティブサポート。
- 📦 **データ主権**：コンテンツとメタデータはユーザーの手元に。RSS 出力にも対応。
- 🔄 **データ移行ワークフロー**：履歴データの移行インポートに加え、移行とアーカイブ向けのスナップショット/カプセルエクスポート。ダッシュボードと CLI の両方から利用可能。
- 🔐 **自動バックアップ**：Web、CLI、TUI からのエクスポート/バックアップに加え、バックグラウンド自動バックアップ。

### 執筆とコンテンツ

- ✍️ **Markdown 執筆体験**：**markdown-it** ベースの編集/レンダリングエンジン。プラグイン拡張とライブプレビュー対応。
- 🏷️ **タグ管理**：タグによる分類、絞り込み、検索。
- 🃏 **リッチメディアカード**：ウェブサイトリンクや GitHub プロジェクトなどをカード表示。
- 🎥 **動画コンテンツ解析**：Bilibili と YouTube の埋め込み再生。

### メディアとアセット

- 🎬 **マルチフォーマット・メディア添付**：各 Echo に画像・音声・動画を直接添付でき、ギャラリーと組み込みの音声/動画プレーヤーでインライン表示。
- 📁 **ビジュアル・ファイルマネージャー**：ファイルアップロード、閲覧、アセット管理機能を内蔵。

### ソーシャルと交流

- 💬 **組み込みコメントシステム**：コメントとモデレーション設定。
- 🃏 **コンテンツ・インタラクション**：いいねや共有などの社会的アクション。

### 認証とセキュリティ

- 🔑 **OAuth2 / OIDC 認証**：OAuth2 と OIDC による外部 ID 連携。
- 🙈 **Passkey によるパスワードレスログイン**：生体認証またはハードウェアセキュリティキー。
- 🔑 **アクセストークン管理**：スコープ付きトークンの発行と取り消し。API 呼び出しや外部連携に。
- 👤 **マルチアカウント権限管理**：マルチユーザー協業と権限制御。

### システムと開発

- 🧱 **Busen データバスアーキテクチャ**：自社開発 Busen による疎結合なモジュール通信と確実なメッセージ配送。
- 📊 **構造化ロギング**：システムログを統一的な構造化フォーマットで出力。
- 🖥️ **リアルタイム・ログコンソール**：Web 上でログをリアルタイム表示し、デバッグ・障害調査に活用可能。
- 📟 **TUI 管理画面**：サーバー運用に最適なターミナル UI。
- 🧰 **CLI ツールチェーン**：自動化やスクリプト連携用の CLI。
- 🔗 **オープン API と Webhook**：外部連携・自動化ワークフロー向けの完全な API と Webhook。
- 🤖 **MCP（Model Context Protocol）**：内蔵 [MCP Server](./docs/usage/mcp-usage.md) が中核機能（投稿、ファイル、統計など）を **ほぼ全カバー**で AI レイヤーに公開。**Streamable HTTP**、**Tools & Resources**、**スコープ付き JWT** を採用。
- 🧠 **Ech0 Copilot**：タイムラインに組み込まれた AI アシスタント——**近況まとめ**がワンクリックで最近の投稿を簡潔なダイジェストに凝縮し、**Chat** では自分のコンテンツに直接質問できます。**ベクトル RAG** 検索と**ストリーミング（SSE）**応答を採用し、すべての回答が*あなた自身の*データに基づきます。

### エクスペリエンス

- 🌍 **クロスデバイス対応**：レスポンシブデザインでデスクトップ、タブレット、モバイルに対応。
- 🌐 **i18n 多言語対応**：UI の多言語切り替えに対応。
- 👾 **PWA 対応**：Web アプリとしてインストールでき、ネイティブアプリ風の体験。
- 🌗 **テーマとダークモード**：ダークモードとテーマ拡張に対応。

### ライセンス

- 🎉 **完全オープンソース**：**AGPL-3.0** で公開。トラッキング、サブスク、SaaS 依存はありません。

</details>

---

## デプロイとアップグレード

**Docker Compose**、**スクリプトインストール**、**Kubernetes (Helm)** の詳細な手順、そして既存インスタンスの**アップグレード**手順は **[DEPLOYMENT.md](./DEPLOYMENT.md)** にまとめています。

最短で起動したい場合は、上の [60 秒で試す](#60-秒で試す) のコマンド一つで十分です。

---

## FAQ

<details>
<summary><strong>FAQ を展開</strong></summary>

1. **Ech0 とは？**
   Ech0 は、個人の思考、文章、リンクを素早く公開・共有するための軽量なオープンソース・セルフホスト・プラットフォームです。シンプルな UI と気が散らない体験を提供し、データは常にあなたの管理下に置かれます。

2. **Ech0 は何ではない？**
   Ech0 は伝統的なプロのノートアプリ（Obsidian や Notion など）ではありません。中心的なユースケースは、ソーシャルフィードやマイクロブログ・ストリームに近いものです。

3. **Ech0 は無料？**
   はい。Ech0 は AGPL-3.0 のもと完全に無料・オープンソースです。広告、トラッキング、サブスクリプション、サービスロックインはありません。

4. **データのバックアップとインポートは？**
   Ech0 は「スナップショット・エクスポート」と「マイグレーション・インポート」によるリストア／移行をサポートしています。デプロイ層では、マウントしたデータディレクトリ（例：`/opt/ech0/data`）を定期的にバックアップしてください。デフォルトではコアデータはローカル DB に保存されます。オブジェクトストレージを有効化している場合、メディアアセットは設定したストレージバックエンドに書き込まれます。

5. **Ech0 は RSS をサポート？**
   はい。Ech0 は RSS 配信に対応しており、RSS リーダーから更新を購読できます。

6. **「管理者に連絡してください」と表示されて投稿できないのは？**
   投稿はデフォルトで権限のあるアカウントに制限されています。初期化時、最初のアカウントが Owner（管理権限あり）になります。一般ユーザーは権限のあるアカウントから明示的に許可されるまで投稿できません。初回セットアップであれば [60 秒で試す](#60-秒で試す) を参照し、Owner となっているアカウントを確認してください。

7. **詳細な権限マトリクスがないのはなぜ？**
   Ech0 は現在、運用をシンプルかつ予測可能に保つため、軽量なロールモデル（Owner / Admin / 一般ユーザー）を採用しています。権限モデルはコミュニティのフィードバックに基づき今後も進化します。

8. **他の人の Connect アバターが表示されないのは？**
   `システム設定 - サービス URL` に現在のインスタンス URL を設定してください。例：`https://memo.vaaat.com`（`http://` または `https://` を含めること）。

9. **設定の MetingAPI とは？**
   音楽カードが再生用ストリーム情報を解決するために使う API エンドポイントです。自前または信頼できるエンドポイントを指定できます。空欄の場合は、Ech0 のデフォルト・リゾルバが使われます。本番環境では自前のエンドポイントを推奨します。

10. **新たに追加した Connect の一部しか表示されないのは？**
    バックエンドはすべての Connect エントリのインスタンス情報を取得しようとします。インスタンスがダウンしていたり到達不能だったりする場合は破棄され、有効でアクセス可能な Connect データのみがフロントエンドに返されます。

11. **コメントを有効化するには？**
    パネルのコメント管理でコメントを有効にし、必要に応じてモデレーションと CAPTCHA を設定してください。Ech0 は CAPTCHA 検証用に `gocap` を内蔵しているため、独立した CAPTCHA サービスのデプロイは不要です。

12. **S3 ストレージの設定方法は？**
    ストレージ設定でプロバイダー、Endpoint、Bucket、Access Key、Secret Key などを入力します。endpoint は `http://` や `https://` なしの形式を推奨します。ブラウザがメディアを直接取得する場合、選択したポリシー（public-read 相当、または同等の CDN/Gateway 構成）でオブジェクトが読み取り可能であることを確認してください。

13. **Passkey ログインを有効化するには？**
    `SSO - Passkey` で `WebAuthn RP ID` と `WebAuthn Origins` を設定し、保存して「Passkey 準備完了」と表示された後、ブラウザのプロンプトに従って生体認証またはセキュリティキーをバインドします。

14. **サードパーティ統合に関する公式声明**
    Ech0 が公式に承認していないサードパーティ統合プラットフォームやサービスは、公式サポートの対象外です。これらの利用に起因するセキュリティインシデント、データ損失、アカウントの問題、その他のリスクは、利用者およびサードパーティの責任となります。

15. **サードパーティ統合（AI／自動化）からコメントを投稿するには？**
    Ech0 は CAPTCHA とフォームトークン検証をバイパスする専用の統合用コメントエンドポイント `POST /api/comments/integration` を提供しています。「アクセストークン」管理から、scope `comment:write` と audience `integration` を持つアクセストークンを作成し、`Authorization: Bearer <token>` ヘッダーに付与してください。リクエストボディとレスポンスについては、インスタンスが提供する OpenAPI ドキュメント（`/swagger/index.html`、ローカル開発では通常 `http://localhost:6277/swagger/index.html`）を参照してください。このエンドポイントには独立したレートリミットが設定されており、コメントは `source=integration` でタグ付けされ、管理画面で識別できます。

16. **ローカル／S3 ストレージのデータ配置、`key` のマッピング規則、S3 やローカル ⇄ オブジェクト間の移行についての詳細は？**
    リポジトリ内ドキュメント [ストレージ移行ガイド](./docs/usage/storage-migration.md) を参照してください。フラットな `key` の値が、ディスク上のパスや S3 オブジェクトキーにどうマップされるか（`schema.Resolve` と `PathPrefix` を含む）、保存される `File.url` のスナップショットが UI とどう対応するか、静的な `/api/files` アクセスと認証付き `stream` ルートの違い、S3 プロバイダーの切り替えやローカル⇄オブジェクトストレージ間のデータ移行に関する実務的なガイダンスを記載しています。

</details>

---

## フィードバックとコミュニティ

- バグを見つけた場合は [Issues](https://github.com/lin-snow/Ech0/issues) で報告してください。
- 新機能や改善のアイデアは [Discussions](https://github.com/lin-snow/Ech0/discussions) で歓迎します。
- 公式 QQ グループ：`1065435773`

### Ech0 Hub に参加する

[Ech0 Hub](https://hub.ech0.app/) は、登録された公開 Ech0 インスタンスのタイムラインを集約する公開ディレクトリです。**あなたの**公開インスタンスを登録する手順については、[`hub/README.md`](./hub/README.md) を参照してください。

| 公式 QQ コミュニティ                                              | その他のグループ |
| ----------------------------------------------------------------- | ---------------- |
| <img src="./docs/imgs/qq.png" alt="QQ グループ" style="height:250px;"> | なし             |

---

## オープンソース運営と開発

**運営**

- [コントリビューションガイド](./CONTRIBUTING.md)
- [行動規範](./CODE_OF_CONDUCT.md)
- [セキュリティポリシー](./SECURITY.md)
- [ライセンス](./LICENSE)

**開発**

ローカル開発のセットアップ、環境要件、フロントエンド／バックエンド連携については **[docs/dev/development.md](./docs/dev/development.md)** を参照してください。より高レベルなアーキテクチャと規約については [`CLAUDE.md`](./CLAUDE.md) と [`CONTRIBUTING.md`](./CONTRIBUTING.md) をご覧ください。

---

## スポンサーと謝辞

🌟 **Ech0** が気に入ったら、ぜひプロジェクトに Star を付けてください！🚀

**Ech0** は完全にオープンソースかつ無料です。継続的なメンテナンスと改善はコミュニティの支援に支えられています。プロジェクトがあなたのお役に立っているなら、ご支援いただけると嬉しいです。下の QR コードから寄付し、メモに GitHub のユーザー名を残していただければ、[スポンサーウォール](./SPONSOR.md) に追加されます。

|                  プラットフォーム            | QR コード                                              |
| :----------------------------------------: | :----------------------------------------------------- |
| [**Afdian**](https://afdian.com/a/l1nsn0w) | <img src="./docs/imgs/pay.jpeg" alt="Pay" width="200"> |

スポンサー、コントリビューター、ユーザーの皆さんに心より感謝します — 完全なスポンサーリストは [SPONSOR.md](./SPONSOR.md) を参照してください。

[LinuxDO](https://linux.do/) コミュニティのご支持と評価に特別な感謝を申し上げます。

[![Contributors](https://contrib.rocks/image?repo=lin-snow/Ech0)](https://contrib.rocks/image?repo=lin-snow/Ech0)

![Repobeats analytics image](https://repobeats.axiom.co/api/embed/d69b9177e4a121e31aaed95354ff862c928ca22d.svg "Repobeats analytics image")

---

```cpp

███████╗     ██████╗    ██╗  ██╗     ██████╗
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝

```

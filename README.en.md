<p align="left">
  <a href="https://hellogithub.com/repository/lin-snow/Ech0" target="_blank">
    <img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=neutral"
         alt="Featured｜HelloGitHub"
         width="250"
         height="54" />
  </a>
</p>

<p align="right">
  <a title="zh" href="./README.md">
    <img src="https://img.shields.io/badge/-简体中文-545759?style=for-the-badge" alt="简体中文">
  </a>
  <img src="https://img.shields.io/badge/-English-F54A00?style=for-the-badge" alt="English">
</p>



<div align="center">
  <img alt="Ech0" src="./docs/imgs/logo.svg" width="150">

  [Preview](https://memo.vaaat.com/) | [Official Site & Doc](https://www.ech0.app/) | [Ech0 Hub](https://hub.ech0.app/)

  # Ech0
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/lin-snow/Ech0)](https://github.com/lin-snow/Ech0/releases) ![License](https://img.shields.io/github/license/lin-snow/Ech0) [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lin-snow/Ech0) [![Hello Github](https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=small)](https://hellogithub.com/repository/lin-snow/Ech0)

</div>

> A next-generation open-source, self-hosted, lightweight federated publishing platform focused on personal idea sharing.

Ech0 is a new-generation open-source self-hosted platform designed for individual users. It is ultra-lightweight and low-cost, supporting the ActivityPub protocol to let you easily publish and share ideas, writings, and links. With a clean, intuitive interface and powerful command-line tools, content management becomes simple and flexible. Your data is fully owned and controlled by you, always connected to the world, building your own network of thoughts.

![Interface Preview](./docs/imgs/screenshot.png)

---

<details>
   <summary><strong>Table of Contents</strong></summary>

- [Ech0](#ech0)
  - [Highlights](#highlights)
  - [Quick Deployment](#quick-deployment)
    - [🐳 Docker (Recommended)](#-docker-recommended)
    - [🐋 Docker Compose](#-docker-compose)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm)
  - [Upgrading](#upgrading)
    - [🔄 Docker](#-docker)
    - [💎 Docker Compose](#-docker-compose-1)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm-1)
  - [FAQ](#faq)
  - [Feedback \& Community](#feedback--community)
  - [Architecture](#architecture)
  - [Development Guide](#development-guide)
    - [Backend Requirements](#backend-requirements)
    - [Frontend Requirements](#frontend-requirements)
    - [Start Backend \& Frontend](#start-backend--frontend)
  - [Thanks for Your Support!](#thanks-for-your-support)
  - [Star History](#star-history)
  - [Acknowledgements](#acknowledgements)
  - [Support](#support)
</details>

---

## Highlights

☁️ **Atomically Lightweight**: Consumes less than **15MB** of memory with an image size under **50MB**, powered by a single-file SQLite architecture  
🚀 **Instant Deployment**: Zero configuration required — from installation to operation in just one command  
✍️ **Distraction-Free Writing**: A clean, online Markdown editor with rich plugin support and real-time preview  
📦 **Data Sovereignty**: All content is stored locally in SQLite, with full RSS feed support  
🔐 **Secure Backup Mechanism**: One-click export and full data backup across Web, TUI, and CLI modes, with automatic background backup support  
♻️ **Seamless Recovery**: Supports TUI/CLI snapshot restoration and Web-based zero-downtime recovery, ensuring data safety with ease  
🎉 **Forever Free**: Open-sourced under the AGPL-3.0 license — no tracking, no subscriptions, no external dependencies  
🌍 **Cross-Platform Adaptation**: Fully responsive design optimized for desktop, tablet, and mobile browsers  
👾 **PWA Ready**: Installable as a web application, offering a near-native experience  
🏷️ **Elegant Tag Management & Filtering**: Intelligent tagging system with fast filtering and precise search for effortless organization  
☁️ **S3 Storage Integration** — Native support for S3-compatible object storage enables efficient cloud synchronization  
🌐 **ActivityPub Federation** — Seamlessly federates with Mastodon, Misskey, and other decentralized platforms  
🔑 **OAuth2 & OIDC Authentication** — Native support for OAuth2 and OIDC protocols, enabling seamless third-party login and API authorization  
🙈 **Passkey Passwordless Login**: Supports passkey login based on biometrics or hardware keys, greatly enhancing security and login experience  
🪶 **Highly Available Webhook**: Enables real-time integration and collaboration with external systems, supporting event-driven automated workflows  
📝 **Built-in Todo Management**: Easily capture and manage daily tasks to stay organized and productive  
🧘 **Quiet Inbox Mode**: Minimizes system-level interruptions by default—messages are surfaced only as needed, letting the tool assist without intruding.
🌗 **Dark Mode & Theme Extensions**: Supports adaptive system dark mode or manual switching, with future extensibility for custom color schemes  
🤖 **Quick Agent AI Setup**: Easily configure multiple large language models for instant AI experience, no manual setup required  
🧰 **Command-Line Powerhouse**: A built-in high-availability CLI that empowers developers and advanced users with precision control and seamless automation  
🔑 **Quick Access Token Management**: Generate and revoke access tokens with one click for secure and efficient API calls and third-party integrations  
📊 **Real-Time System Resource Monitoring**: High-performance WebSocket-based monitoring dashboard for instant visibility into runtime status  
📟 **Refined TUI Experience**: A beautifully designed terminal interface offering intuitive management of Ech0  
🔗 **Ech0 Connect**: A multi-instance connectivity feature that enables real-time status sharing and synchronization between Ech0 nodes  
🎵 **Seamless Music Integration**: Lightweight embedded music player providing immersive soundscapes and focus modes  
🎥 **Instant Video Sharing**: Natively supports intelligent parsing of Bilibili and YouTube videos  
🃏 **Rich Smart Cards**: Instantly share websites, GitHub projects, and other media in visually engaging cards  
⚙️ **Advanced Customization**: Easily personalize styles and scripts for expressive, unique content presentation  
💬 **Comment System**: Quick Twikoo integration for lightweight, instant, and non-intrusive interactions  
💻 **Cross-Platform Compatibility**: Runs natively on Windows, Linux, and ARM devices like Raspberry Pi for stable deployment anywhere  
🔗 **Ech0 Hub Square**: Built-in Ech0 Hub Square for easily discovering, subscribing to, and sharing high-quality content  
📦 **Self-Contained Binary**: Includes all required resources — no extra dependencies, no setup hassle  
🔗 **Rich API Support**: Open APIs for seamless integration with external systems and workflows  
🃏 **Dynamic Content Display**: Supports Twitter-like card layouts with likes and social interactions  
👤 **Multi-Account & Permission Management**: Flexible user and role-based access control ensuring privacy and security  


---

## Quick Deployment
<!-- 
### 🧙 One-Click Script Deployment (Recommended, make sure your network can access GitHub Release)
```shell
curl -fsSL "https://sh.soopy.cn/ech0.sh" -o ech0.sh && bash ech0.sh
``` -->

### 🐳 Docker (Recommended)

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

> 💡 After deployment, access `ip:6277` to use  
> 🚷 It is recommended to change `JWT_SECRET="Hello Echos"` to a secure secret  
> 📍 The first registered user will be set as administrator  
> 🎈 Data stored under `/opt/ech0/data`

### 🐋 Docker Compose

1. Create a new directory and place `docker-compose.yml` inside.  
2. Run:

```shell
docker-compose up -d
```

### ☸️ Kubernetes (Helm)

If you want to deploy Ech0 in a Kubernetes cluster, you can use the Helm Chart provided in this project.

Since this project does not provide an online Helm repository, you need to clone the repository to your local machine first, and then install from the local directory.

1.  **Clone the repository:**
    ```shell
    git clone https://github.com/lin-snow/Ech0.git
    cd Ech0
    ```

2.  **Install with Helm:**
    ```shell
    # helm install <release-name> <chart-directory>
    helm install ech0 ./charts/ech0
    ```

    You can also customize the release name and namespace:
    ```shell
    helm install my-ech0 ./charts/ech0 --namespace my-namespace --create-namespace
    ```

---

## Upgrading

### 🔄 Docker

```shell
docker stop ech0
docker rm ech0
docker pull sn0wl1n/ech0:latest
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

### 💎 Docker Compose

```shell
cd /path/to/compose
docker-compose pull && \
docker-compose up -d --force-recreate
docker image prune -f
```

### ☸️ Kubernetes (Helm)

1.  **Update the repository:**
    Navigate to your local Ech0 repository directory and pull the latest changes.
    ```shell
    cd Ech0
    git pull
    ```

2.  **Upgrade the Helm Release:**
    Use the `helm upgrade` command to update your release.
    ```shell
    # helm upgrade <release-name> <chart-directory>
    helm upgrade ech0 ./charts/ech0
    ```
    If you used a custom release name and namespace, use the corresponding names:
    ```shell
    helm upgrade my-ech0 ./charts/ech0 --namespace my-namespace
    ```

<!-- ---

## Access Modes

### 🖥️ TUI Mode

![TUI Mode](./docs/imgs/tui.png)

Run the binary directly (for example, on Windows double-click `Ech0.exe`).

### 🔐 SSH Mode

Connect to the instance via port 6278:

```shell
ssh -p 6278 ssh.vaaat.com
``` -->

---

## FAQ

1. **What is Ech0?**  
   A lightweight, open-source self-hosted platform for quickly sharing thoughts, writings, and links. All content is locally stored.  

2. **What Ech0 is NOT?**  
   Not a professional note-taking app like Obsidian or Notion; its core function is similar to social feed/microblog.  

3. **Is Ech0 free?**  
   Yes, fully free and open-source under AGPL-3.0, no ads, tracking, subscription, or service dependency.  

4. **How do I back up and restore data?**  
  Since all content is stored in a local SQLite file, you only need to back up the files in the `/opt/ech0/data` directory (or the mapped path you chose during deployment). To restore, simply replace the data files with your backup. You can also use the online data management features in the settings under "Data Management" to quickly create, export, or restore snapshots. If the latest content does not appear after restoring, try manually restarting the Docker container.

5. **Does Ech0 support RSS?**  
   Yes, content updates can be subscribed via RSS.  

6. **Why can't I publish content?**  
   Only administrators can publish. First registered user is admin.  

7. **Why no detailed permission system?**  
   Ech0 emphasizes simplicity: admin vs non-admin only, for smooth experience.  

8. **Why Connect avatars may not show?**  
   Set your instance URL in `System Settings - Service URL` (with `http://` or `https://`).  

9. **What is MetingAPI?**  
   Used to parse music streaming URLs for music cards. If empty, default API provided by Ech0 is used.  

10. **Why not all Connect items show?**  
    Instances that are offline or unreachable are ignored; only valid instances are displayed.  

11. **What content is not recommended?**  
    Avoid publishing dense content mixing text + images + extension cards. Long posts or extension cards alone are okay.  

12. **How to enable comments?**  
    Set up Twikoo backend URL in settings. Only Twikoo is supported.  

13. **How to configure S3?**  
    Fill in endpoint (without http/https) and bucket with public access.

14. **How to join the Fediverse?**  
  You need to bind Ech0 to a domain name and fill in the domain in the server address field in the settings page. Once set, Ech0 will automatically join the Fediverse. Example: `https://memo.vaaat.com`

---

## Feedback & Community

- Report bugs via [Issues](https://github.com/lin-snow/Ech0/issues).
- Propose features or share ideas in [Discussions](https://github.com/lin-snow/Ech0/discussions).

---

## Architecture

![Architecture Diagram](./docs/imgs/Ech0技术架构图.svg)  
> by ExcaliDraw

---

## Development Guide

### Backend Requirements
- Go 1.25.3+  
- C Compiler for CGO (`go-sqlite3`):
  - Windows: [MinGW-w64](https://winlibs.com/)  
  - macOS: `brew install gcc`  
  - Linux: `sudo apt install build-essential`  
- Google Wire: `go install github.com/google/wire/cmd/wire@latest`  
- Golangci-Lint: `golangci-lint run` / `golangci-lint fmt`  
- Air (optional, backend hot reload): `make air-install` or `go install github.com/air-verse/air@latest`  
- Swagger: `swag init -g internal/server/server.go -o internal/swagger`  

### Frontend Requirements
- NodeJS v25.5.0+, PNPM v10.30.0+  
- Use [fnm](https://github.com/Schniz/fnm) if multiple Node versions needed

### Start Backend & Frontend
```shell
# Backend
make run # normal backend start (equivalent to go run main.go web)
make dev # backend hot reload with Air

# Frontend
cd web
pnpm install
pnpm dev
# or from project root: make web-dev
```

Preview: Backend `http://localhost:6277`, Frontend `http://localhost:5173`

> When importing layered packages, prefer consistent aliases such as `xxxModel`, `xxxService`, `xxxRepository`, and so on.


---

## Thanks for Your Support!

Thank you to all the friends who have supported this project! Your contributions keep it thriving 💡✨

|                        ⚙️ User                        |   🔋 Date   | 💬 Message                                       |
| :--------------------------------------------------: | :--------: | :---------------------------------------------- |
|                  🧑‍💻 Anonymous Friend                  | 2025-5-19  | Silly programmer, buy yourself some sweet drink |
|        🧑‍💻 [@sseaan](https://github.com/sseaan)        | 2025-7-27  | Ech0 is a great thing🥳                          |
| 🧑‍💻 [@QYG2297248353](https://github.com/QYG2297248353) | 2025-10-10 | None                                            |
|    🧑‍💻 [@continue33](https://github.com/continue33)    | 2025-10-23 | Thanks for fixing R2                            |
|    🧑‍💻 [@hoochanlon](https://github.com/hoochanlon)      | 2025-10-28 | None             |
|       🧑‍💻 [@Rvn0xsy](https://github.com/Rvn0xsy)       | 2025-11-12 | Great project, I will keep following! |
|                     🧑‍💻 王贼臣                     | 2025-11-20 | Thanks www.cardopt.cn             |
|       🧑‍💻 [@ljxme](https://github.com/ljxme)    | 2025-11-30 | Doing my humble part 😋             |
|       🧑‍💻 [@he9ab2l](https://github.com/he9ab2l)    | 2025-12-23 | None            |
|       🧑‍💻 鸿运当头(windfore)    | 2026-1-6 | Thank you for creating ech0           |
|       🧑‍💻 Anonymous User    | 2026-01-23  | None           |

---

## Star History

<a href="https://www.star-history.com/#lin-snow/Ech0&Timeline">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
 </picture>
</a>


---

## Acknowledgements

- Thanks to all users for their valuable suggestions and feedback.
- Thanks to all contributors and supporters from the open-source community.


![Alt](https://repobeats.axiom.co/api/embed/d69b9177e4a121e31aaed95354ff862c928ca22d.svg "Repobeats analytics image")

---

## Support

🌟 If you like **Ech0**, please give it a Star! 🚀  
Ech0 is completely free and open-source. Support helps the project continue improving.  

|                  Platform                  | QR Code                                                |
| :----------------------------------------: | :----------------------------------------------------- |
| [**Afdian**](https://afdian.com/a/l1nsn0w) | <img src="./docs/imgs/pay.jpeg" alt="Pay" width="200"> |

---

```cpp

███████╗     ██████╗    ██╗  ██╗     ██████╗ 
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝ 

``` 

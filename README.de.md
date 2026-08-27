<div align="center">

<img alt="Ech0" src="./docs/imgs/logo.svg" width="150">

# Ech0

[Vorschau](https://memo.vaaat.com/) · [Offizielle Seite & Dokumentation](https://www.ech0.app/) · [Releases](https://lin-snow.github.io/Ech0/) · [Ech0 Hub](https://hub.ech0.app/)

<a title="en-US" href="./README.md"><img src="https://img.shields.io/badge/-English-545759?style=for-the-badge" alt="English"></a> <a title="zh" href="./README.zh.md"><img src="https://img.shields.io/badge/-简体中文-545759?style=for-the-badge" alt="简体中文"></a> <img src="https://img.shields.io/badge/-Deutsch-F54A00?style=for-the-badge" alt="Deutsch"> <a title="ja" href="./README.ja.md"><img src="https://img.shields.io/badge/-日本語-545759?style=for-the-badge" alt="日本語"></a>

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

> Ein selbstgehosteter persönlicher Microblog, dessen Timeline geteilt, diskutiert und vollständig von dir kontrolliert werden kann.

Tools wie Memos eignen sich hervorragend zum schnellen Festhalten von Gedanken. Ech0 ist für das gebaut, was danach kommt: diese Ideen auf einer persönlichen Timeline zu veröffentlichen, der andere folgen und mit der sie interagieren können.
Hoste es auf deinem eigenen Server, behalte die volle Kontrolle über deine Inhalte und bewahre dir einen persönlichen Raum, der durch optionale Kommentare und Sharing trotzdem verbunden bleibt.
Es bleibt dabei leichtgewichtig, einfach zu deployen und vollständig Open Source.

**Gut geeignet, wenn du:**
- eine persönliche öffentliche oder halböffentliche Timeline auf deiner eigenen Domain betreiben willst
- kurze Beiträge, Links und Medien über eine schlanke Oberfläche veröffentlichen möchtest
- Datenhoheit behalten und gleichzeitig RSS sowie optionale Kommentare nutzen willst
- einen persönlichen Raum mit leichter sozialer Interaktion willst, ohne ein vollwertiges soziales Netzwerk zu betreiben

**Eher nicht geeignet, wenn du brauchst:**
- einen bidirektionalen Knowledge-Base-Workflow (z. B. Obsidian-artiges PKM)
- einen Team-orientierten kollaborativen Docs-Workspace (z. B. Notion-artige Dokumente)
- ein reines Notiz-/Memo-Tool ganz ohne Timeline oder Veröffentlichung

![Oberflächen-Vorschau](./docs/imgs/screenshot.png)

---

<details>
   <summary><strong>Inhaltsverzeichnis</strong></summary>

- [In 60 Sekunden ausprobieren](#in-60-sekunden-ausprobieren)
- [Vollständige Funktionsliste](#vollständige-funktionsliste)
- [Deployment & Aktualisierung](#deployment--aktualisierung)
- [FAQ](#faq)
- [Feedback & Community](#feedback--community)
- [Open Source & Entwicklung](#open-source--entwicklung)
- [Sponsoren & Danksagungen](#sponsoren--danksagungen)
- [Star-Verlauf](#star-verlauf)

</details>

---

## In 60 Sekunden ausprobieren

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

Öffne anschließend `http://ip:6277`:

1. Registriere deinen ersten Account.
2. Der erste Account wird automatisch zum Owner (Admin-Rechte).
3. Standardmäßig dürfen nur privilegierte Accounts veröffentlichen.

Weitere Optionen wie Docker Compose und Helm findest du unter [Schnelles Deployment](./DEPLOYMENT.md#quick-deployment).

## Vollständige Funktionsliste

<details>
<summary><strong>Klicken, um die vollständige Funktionsliste anzuzeigen</strong></summary>

### Highlights

- ☁️ **Leichtgewichtige, effiziente Architektur**: Geringer Ressourcenverbrauch und kompakte Images, geeignet von persönlichen Servern bis zu ARM-Geräten.
- 🚀 **Schnelles Deployment**: Out-of-the-box-Docker-Deployment, von der Installation bis zum ersten Start mit einem einzigen Befehl.
- 📦 **Eigenständige Distribution**: Vollständige Binaries und Container-Images ohne zusätzliche Laufzeitabhängigkeiten.
- 💻 **Plattformübergreifend**: Unterstützt Linux, Windows und ARM-Geräte (z. B. Raspberry Pi).
- 🧳 **Portable Inhaltskapseln**: Exportieren Sie alles Geschriebene als eigenständige [Kapsel](./docs/usage/capsule.md), importieren Sie sie in eine andere Instanz oder bauen Sie daraus eine statische Seite für beliebiges Hosting.

### Speicher & Daten

- 🗂️ **VireFS Unified Storage Layer**: **VireFS** vereinheitlicht das Mounten und die Verwaltung von lokalem Speicher und S3-kompatiblem Object Storage.
- ☁️ **S3-Object-Storage-Unterstützung**: Native Unterstützung für S3-kompatiblen Object Storage zur Erweiterung in die Cloud.
- 📦 **Datensouveränität**: Inhalte und Metadaten bleiben in Nutzerhand und unter Nutzerkontrolle, inklusive RSS-Ausgabe.
- 🔄 **Datenmigrations-Workflow**: Migrationsimport für historische Daten sowie Snapshot- und Kapsel-Export für Migration und Archivierung, aus dem Dashboard und über die CLI.
- 🔐 **Automatisiertes Backup-System**: Export/Backup über Web, CLI und TUI sowie automatische Hintergrund-Backups.

### Schreiben & Inhalt

- ✍️ **Markdown-Schreiberlebnis**: Auf **markdown-it** basierende Editor-/Rendering-Engine mit Plugin-Erweiterungen und Live-Vorschau.
- 🏷️ **Tag-Verwaltung**: Tag-Organisation, schnelles Filtern und präzise Suche.
- 🃏 **Rich-Media-Karten**: Karten-Rendering für Website-Links, GitHub-Projekte und mehr.
- 🎥 **Video-Parsing**: Eingebettete Wiedergabe für Bilibili- und YouTube-Videos.

### Medien & Assets

- 🎬 **Multi-Format-Medienanhänge**: Bilder, Audio und Video direkt an jede Echo anhängen – inline dargestellt mit Galerie und eingebauten Audio-/Video-Playern.
- 📁 **Visueller Datei-Manager**: Eingebaute Funktionen für Datei-Upload, Browsing und Asset-Verwaltung.

### Soziales & Interaktion

- 💬 **Eingebautes Kommentarsystem**: Kommentare und konfigurierbare Moderation.
- 🃏 **Inhalts-Interaktion**: Soziale Interaktionen wie Likes und Sharing.

### Authentifizierung & Sicherheit

- 🔑 **OAuth2 / OIDC**: Anbindung an Drittanbieter-Logins über OAuth2 und OIDC.
- 🙈 **Passkey ohne Passwort**: Anmeldung per Biometrie oder Hardware-Sicherheitsschlüssel.
- 🔑 **Access-Token-Verwaltung**: Erzeugen und Widerrufen von scopebasierten Tokens für API-Aufrufe und Drittanbieter-Integrationen.
- 👤 **Mehrbenutzer-Rechteverwaltung**: Mehrbenutzer-Kollaboration und Rechtekontrolle.

### System & Entwicklung

- 🧱 **Busen-Datenbus-Architektur**: Das hauseigene Busen sorgt für entkoppelte Modulkommunikation und zuverlässige Nachrichtenzustellung.
- 📊 **Strukturiertes Logging**: System-Logs in einheitlichem strukturiertem Format für Lesbarkeit und Analyse.
- 🖥️ **Echtzeit-Log-Konsole**: Eingebaute Web-Konsole für Live-Log-Streams, Debugging und Troubleshooting.
- 📟 **TUI-Verwaltung**: Terminal-UI, ideal für die Verwaltung auf Servern.
- 🧰 **CLI-Toolchain**: CLI-Tools für Automatisierung und Skript-Integration.
- 🔗 **Open API & Webhook**: Vollständige API- und Webhook-Unterstützung für externe Integrationen und Automation.
- 🤖 **MCP (Model Context Protocol)**: Eingebauter [MCP Server](./docs/usage/mcp-usage.md) deckt **nahezu vollständig** die Kernfunktionen für die KI-Schicht ab (Beiträge, Dateien, Statistiken usw.) — **Streamable HTTP**, **Tools & Resources**, **scoped JWT**.
- 🧠 **Ech0 Copilot**: Ein KI-Begleiter direkt in deiner Timeline — **Rückblick** verdichtet deine jüngsten Beiträge per Klick zu einer prägnanten Zusammenfassung, und **Chat** lässt dich direkt mit deinen eigenen Inhalten sprechen und alles fragen. Angetrieben von **Vektor-RAG**-Retrieval mit **Streaming-(SSE)**-Antworten, jede Antwort basiert auf *deinen* Daten.

### Erlebnis

- 🌍 **Geräteübergreifende Anpassung**: Responsive Design für Desktop, Tablet und mobile Browser.
- 🌐 **i18n-Mehrsprachigkeit**: Mehrsprachige Oberfläche für unterschiedliche Einsatzszenarien.
- 👾 **PWA-Unterstützung**: Als Web-App installierbar — fast wie eine native App.
- 🌗 **Themes & Dark Mode**: Dark Mode und Theme-Erweiterungen.

### Lizenz

- 🎉 **Vollständig Open Source**: Veröffentlicht unter **AGPL-3.0**, ohne Tracking, ohne Abo, ohne SaaS-Abhängigkeit.

</details>

---

## Deployment & Aktualisierung

Ausführliche Anleitungen für **Docker Compose**, **Skript-Installation**, **Kubernetes (Helm)** sowie das **Upgrade** einer bestehenden Instanz findest du in **[DEPLOYMENT.md](./DEPLOYMENT.md)**.

Für den schnellsten Einstieg reicht der oben gezeigte [In 60 Sekunden ausprobieren](#in-60-sekunden-ausprobieren)-Befehl bereits aus.

---

## FAQ

<details>
<summary><strong>FAQ ausklappen</strong></summary>

1. **Was ist Ech0?**
   Ech0 ist eine leichtgewichtige Open-Source-Self-Hosting-Plattform zum schnellen Veröffentlichen und Teilen persönlicher Gedanken, Texte und Links. Sie bietet eine schlanke Oberfläche und ein ablenkungsfreies Erlebnis — und deine Daten bleiben in deiner Hand.

2. **Was ist Ech0 nicht?**
   Ech0 ist keine klassische professionelle Notiz-App (wie Obsidian oder Notion). Der Kern-Use-Case ähnelt eher einem Social-Feed bzw. Microblog-Stream.

3. **Ist Ech0 kostenlos?**
   Ja. Ech0 ist vollständig kostenlos und Open Source unter AGPL-3.0 — ohne Werbung, Tracking, Abo oder Service-Lock-in.

4. **Wie sichere und importiere ich Daten?**
   Ech0 unterstützt Datenwiederherstellung/-migration via „Snapshot-Export" und „Migrations-Import". Auf Deployment-Ebene sollte das gemappte Datenverzeichnis (z. B. `/opt/ech0/data`) regelmäßig gesichert werden. Standardmäßig liegen Kerndaten in der lokalen Datenbank; bei aktiviertem Object Storage werden Medien-Assets in das konfigurierte Backend geschrieben.

5. **Unterstützt Ech0 RSS?**
   Ja. Ech0 unterstützt RSS-Abonnements, sodass du Updates in RSS-Readern verfolgen kannst.

6. **Warum schlägt das Veröffentlichen mit der Meldung „Administrator kontaktieren" fehl?**
   Das Veröffentlichen ist standardmäßig auf privilegierte Accounts beschränkt. Bei der Initialisierung wird der erste Account zum Owner (mit Verwaltungsrechten). Reguläre Nutzer dürfen erst veröffentlichen, wenn ein privilegierter Account dies explizit erlaubt. Wenn das deine erste Einrichtung ist, prüfe unter [In 60 Sekunden ausprobieren](#in-60-sekunden-ausprobieren), welcher Account Owner ist.

7. **Warum gibt es keine detaillierte Rechtematrix?**
   Ech0 verwendet derzeit ein leichtgewichtiges Rollenmodell (Owner / Admin / regulärer Nutzer), um den Betrieb einfach und vorhersehbar zu halten. Das Rechtemodell wird auf Basis von Community-Feedback weiterentwickelt.

8. **Warum sehen andere mein Connect-Avatar nicht?**
   Setze die URL deiner aktuellen Instanz in `Systemeinstellungen — Service-URL`, z. B. `https://memo.vaaat.com` (mit `http://` oder `https://`).

9. **Was ist die Option MetingAPI in den Einstellungen?**
   Das ist der API-Endpunkt, den Music-Cards verwenden, um abspielbare Stream-Metadaten aufzulösen. Du kannst einen eigenen, vertrauenswürdigen Endpunkt angeben; bleibt das Feld leer, fällt Ech0 auf einen Standard-Resolver zurück. Für den produktiven Einsatz empfiehlt sich ein selbstgehosteter Endpunkt.

10. **Warum zeigt eine neu hinzugefügte Connect nur Teilergebnisse?**
    Das Backend versucht, Instanz-Informationen für alle Connect-Einträge abzurufen. Ist eine Instanz offline oder nicht erreichbar, wird sie verworfen — nur gültige/erreichbare Connect-Daten werden ans Frontend zurückgegeben.

11. **Wie aktiviere ich Kommentare?**
    Aktiviere Kommentare in der Kommentar-Verwaltung im Panel und konfiguriere bei Bedarf Moderation und Captcha. Ech0 enthält bereits `gocap` zur Captcha-Verifikation — ein eigener Captcha-Service ist nicht nötig.

12. **Wie konfiguriere ich S3-Speicher?**
    Trage in den Speichereinstellungen Provider, Endpoint, Bucket, Access Key, Secret Key und weitere Felder ein. Endpoint vorzugsweise ohne `http://` oder `https://`. Wenn Medien direkt vom Browser abgerufen werden, müssen die Objekte über die gewählte Policy lesbar sein (z. B. public-read oder ein gleichwertiges CDN-/Gateway-Setup).

13. **Wie aktiviere ich Passkey-Login?**
    Konfiguriere unter `SSO — Passkey` `WebAuthn RP ID` und `WebAuthn Origins`. Nach dem Speichern und der Anzeige „Passkey bereit" folgt man den Browser-Prompts, um Biometrie oder Sicherheitsschlüssel zu binden.

14. **Offizielle Erklärung zu Drittanbieter-Integrationen**
    Drittanbieter-Integrationsplattformen oder -Dienste, die nicht offiziell von Ech0 autorisiert sind, liegen außerhalb des offiziellen Supports. Sicherheitsvorfälle, Datenverluste, Account-Probleme oder andere Risiken durch deren Nutzung liegen in der Verantwortung des Nutzers und des Drittanbieters.

15. **Wie poste ich Kommentare über eine Drittanbieter-Integration (KI / Automation)?**
    Ech0 stellt einen dedizierten Integration-Endpunkt unter `POST /api/comments/integration` bereit, der Captcha- und Form-Token-Prüfung umgeht. Erstelle in der „Access Token"-Verwaltung ein Token mit Scope `comment:write` und Audience `integration` und sende es im Header `Authorization: Bearer <token>`. Request-Body und Antworten siehe OpenAPI-Doku deiner Instanz unter `/swagger/index.html` (lokal typischerweise `http://localhost:6277/swagger/index.html`). Dieser Endpunkt hat eigene Rate-Limits, und die Kommentare werden mit `source=integration` markiert, sodass sie im Admin-Panel erkennbar sind.

16. **Wo finde ich detaillierte Doku zu Storage-Regeln (lokal vs. S3), Object-Keys und Migration?**
    Siehe den [Storage-Migration-Guide](./docs/usage/storage-migration.md) im Repo. Er erklärt, wie flache `key`-Werte auf Pfade auf der Festplatte und S3-Objekt-Keys gemappt werden (inkl. `schema.Resolve` und `PathPrefix`), wie gespeicherte `File.url`-Snapshots zur UI passen, den Unterschied zwischen statischem `/api/files`-Zugriff und authentifizierten `stream`-Routen sowie praktische Hinweise zum Wechsel des S3-Anbieters und zur Migration zwischen lokalem Storage und Object Storage.

</details>

---

## Feedback & Community

- Bei Bugs bitte ein [Issue](https://github.com/lin-snow/Ech0/issues) öffnen.
- Für Feature-Ideen oder Verbesserungen ist [Discussions](https://github.com/lin-snow/Ech0/discussions) der richtige Ort.
- Offizielle QQ-Gruppe: `1065435773`

### Ech0 Hub beitreten

[Ech0 Hub](https://hub.ech0.app/) ist ein öffentliches Verzeichnis, das die Timelines gelisteter Ech0-Instanzen zusammenführt. Eine Schritt-für-Schritt-Anleitung zum Registrieren **deiner** öffentlichen Instanz findest du in [`hub/README.md`](./hub/README.md).

| Offizielle QQ-Community                                          | Weitere Gruppen |
| ---------------------------------------------------------------- | --------------- |
| <img src="./docs/imgs/qq.png" alt="QQ-Gruppe" style="height:250px;"> | —               |

---

## Open Source & Entwicklung

**Governance**

- [Beitragsleitfaden](./CONTRIBUTING.md)
- [Verhaltenskodex](./CODE_OF_CONDUCT.md)
- [Sicherheitsrichtlinie](./SECURITY.md)
- [Lizenz](./LICENSE)

**Entwicklung**

Das lokale Setup, Umgebungsanforderungen und das Front-/Backend-Zusammenspiel sind in **[docs/dev/development.md](./docs/dev/development.md)** dokumentiert. Architektur und Konventionen findest du in [`CLAUDE.md`](./CLAUDE.md) und [`CONTRIBUTING.md`](./CONTRIBUTING.md).

---

## Sponsoren & Danksagungen

🌟 Wenn dir **Ech0** gefällt, gib dem Projekt gerne einen Star! 🚀

**Ech0** ist vollständig Open Source und kostenlos. Wartung und Weiterentwicklung leben von der Unterstützung der Community. Wenn dir das Projekt hilft, freuen wir uns über jede Spende. Scanne den QR-Code unten und hinterlasse deinen GitHub-Namen in der Notiz — du wirst dann auf der [Sponsorenliste](./SPONSOR.md) eingetragen.

|                  Plattform                 | QR-Code                                                |
| :----------------------------------------: | :----------------------------------------------------- |
| [**Afdian**](https://afdian.com/a/l1nsn0w) | <img src="./docs/imgs/pay.jpeg" alt="Pay" width="200"> |

Ein großes Dankeschön an alle Sponsoren, Beitragenden und Nutzer — die vollständige Sponsorenliste findest du in [SPONSOR.md](./SPONSOR.md).

Besonderen Dank an die [LinuxDO](https://linux.do/) Community für ihre Anerkennung und Unterstützung.

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

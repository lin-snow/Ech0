<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0"
  xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
  xmlns:atom="http://www.w3.org/2005/Atom">

  <xsl:output method="html" encoding="UTF-8" indent="yes"
    doctype-system="about:legacy-compat"/>

  <xsl:template match="/atom:feed">
    <html lang="zh-CN">
      <head>
        <meta charset="UTF-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1"/>
        <meta name="robots" content="noindex,follow"/>
        <title>
          <xsl:value-of select="atom:title"/> · RSS
        </title>
        <link rel="icon" href="/favicon.svg" type="image/svg+xml"/>
        <style>
          :root {
            --bg-canvas: #f4f1ec;
            --bg-surface: #fff;
            --bg-muted: #fdf9f4;
            --text-primary: #33251d;
            --text-secondary: #5b4f46;
            --text-muted: #8e847d;
            --border-subtle: #e4dfdb;
            --border-strong: #c2b9b2;
            --accent: #b84200;
            --accent-soft: #f4e3cd;
            --shadow: 0 1px 2px rgb(51 37 29 / 4%), 0 8px 24px rgb(51 37 29 / 6%);
          }
          @media (prefers-color-scheme: dark) {
            :root {
              --bg-canvas: #1a1816;
              --bg-surface: #201d1b;
              --bg-muted: #292623;
              --text-primary: #f0e9e0;
              --text-secondary: #cfcac4;
              --text-muted: #a29e99;
              --border-subtle: #3e3c3a;
              --border-strong: #5e5a57;
              --accent: #e7b063;
              --accent-soft: #493e30;
              --shadow: 0 1px 2px rgb(0 0 0 / 24%), 0 8px 24px rgb(0 0 0 / 32%);
            }
          }
          * { box-sizing: border-box; }
          html, body {
            margin: 0;
            padding: 0;
            background: var(--bg-canvas);
            color: var(--text-primary);
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC",
                         "Hiragino Sans GB", "Microsoft YaHei", Roboto, sans-serif;
            font-size: 16px;
            line-height: 1.65;
            -webkit-font-smoothing: antialiased;
          }
          ::selection { background: var(--accent-soft); color: var(--text-primary); }
          a { color: var(--accent); text-decoration: none; }
          a:hover { text-decoration: underline; }
          .wrap { max-width: 560px; margin: 0 auto; padding: 28px 16px 80px; }
          .masthead {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            padding-bottom: 20px;
            border-bottom: 1px solid var(--border-subtle);
            margin-bottom: 24px;
          }
          .masthead-brand {
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 0;
          }
          .logo {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            object-fit: cover;
            border: 2px solid var(--bg-surface);
            box-shadow:
              0 0 0 1px var(--border-subtle),
              0 1px 2px rgb(0 0 0 / 6%);
            flex-shrink: 0;
          }
          .masthead-brand h1 {
            margin: 0;
            font-size: 20px;
            font-weight: 700;
            letter-spacing: -0.01em;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }
          .masthead-brand h1 .rss-suffix {
            font-weight: 500;
            color: var(--text-muted);
            margin-left: 6px;
          }
          .feed-badge {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 4px 12px;
            border-radius: 999px;
            font-size: 12px;
            font-weight: 600;
            letter-spacing: 0.06em;
            background: var(--accent-soft);
            color: var(--accent);
            flex-shrink: 0;
          }
          .feed-badge::before {
            content: "";
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: var(--accent);
            animation: pulse 2s ease-in-out infinite;
          }
          @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.4; }
          }
          .hint {
            background: var(--bg-muted);
            border: 1px solid var(--border-subtle);
            border-radius: 12px;
            padding: 14px 16px;
            font-size: 13.5px;
            color: var(--text-secondary);
            margin-bottom: 28px;
            display: flex;
            gap: 12px;
            align-items: flex-start;
          }
          .hint svg { flex-shrink: 0; margin-top: 2px; color: var(--accent); }
          .hint code {
            background: var(--bg-surface);
            border: 1px solid var(--border-subtle);
            padding: 1px 6px;
            border-radius: 4px;
            font-size: 12.5px;
            color: var(--text-primary);
            word-break: break-all;
          }
          .entries { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 16px; }
          .entry {
            background: var(--bg-surface);
            border: 1px solid var(--border-subtle);
            border-radius: 14px;
            padding: 20px 22px;
            box-shadow: var(--shadow);
            transition: transform .18s ease, border-color .18s ease;
          }
          .entry:hover { transform: translateY(-1px); border-color: var(--border-strong); }
          .entry-meta {
            display: flex;
            align-items: center;
            gap: 10px;
            font-size: 13px;
            color: var(--text-muted);
            margin-bottom: 10px;
          }
          .entry-author {
            font-weight: 600;
            color: var(--text-secondary);
          }
          .entry-meta .dot {
            width: 3px;
            height: 3px;
            border-radius: 50%;
            background: var(--border-strong);
          }
          .entry-meta time {
            font-variant-numeric: tabular-nums;
            font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
            font-size: 12.5px;
          }
          .entry-content {
            position: relative;
            color: var(--text-primary);
            font-size: 15px;
            line-height: 1.7;
            overflow-wrap: break-word;
            max-height: 16em;
            overflow: hidden;
            -webkit-mask-image: linear-gradient(to bottom, #000 70%, transparent 100%);
                    mask-image: linear-gradient(to bottom, #000 70%, transparent 100%);
          }
          .entry-content p { margin: 0 0 0.6em; }
          .entry-content p:last-child { margin-bottom: 0; }
          .entry-content img {
            display: inline-block !important;
            width: auto !important;
            max-width: calc(50% - 6px) !important;
            height: 120px !important;
            object-fit: cover;
            border-radius: 8px;
            border: 1px solid var(--border-subtle);
            margin: 0 4px 4px 0;
            vertical-align: top;
          }
          .entry-content img:only-of-type {
            max-width: 100% !important;
            height: auto !important;
            max-height: 240px;
            display: block !important;
          }
          .entry-content blockquote {
            margin: 12px 0;
            padding: 4px 14px;
            border-left: 3px solid var(--accent);
            background: var(--bg-muted);
            border-radius: 0 8px 8px 0;
            color: var(--text-secondary);
          }
          .entry-content code {
            background: var(--bg-muted);
            border: 1px solid var(--border-subtle);
            padding: 1px 6px;
            border-radius: 4px;
            font-size: 0.9em;
          }
          .entry-content pre {
            background: var(--bg-muted);
            border: 1px solid var(--border-subtle);
            border-radius: 8px;
            padding: 12px 14px;
            overflow-x: auto;
            font-size: 13.5px;
          }
          .entry-content pre code { background: none; border: none; padding: 0; }
          .entry-content .tag {
            display: inline-block;
            background: var(--accent-soft);
            color: var(--accent);
            padding: 1px 8px;
            border-radius: 999px;
            font-size: 12px;
            font-weight: 500;
            margin-right: 4px;
          }
          .entry-footer {
            margin-top: 14px;
            padding-top: 12px;
            border-top: 1px dashed var(--border-subtle);
            font-size: 13px;
          }
          .entry-footer a {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            color: var(--text-muted);
            transition: color .15s ease;
          }
          .entry-footer a:hover { color: var(--accent); text-decoration: none; }
          .empty {
            text-align: center;
            padding: 64px 20px;
            color: var(--text-muted);
            background: var(--bg-surface);
            border: 1px dashed var(--border-strong);
            border-radius: 14px;
          }
          .site-footer {
            margin-top: 40px;
            text-align: center;
            font-size: 12.5px;
            color: var(--text-muted);
          }
          .site-footer a { color: var(--text-secondary); }
          @media (max-width: 520px) {
            .wrap { padding: 20px 14px 60px; }
            .entry { padding: 16px 16px; border-radius: 12px; }
            .masthead-meta h1 { font-size: 19px; }
          }
        </style>
      </head>
      <body>
        <main class="wrap">
          <header class="masthead">
            <div class="masthead-brand">
              <img class="logo" alt="">
                <xsl:attribute name="src">
                  <xsl:choose>
                    <xsl:when test="atom:logo"><xsl:value-of select="atom:logo"/></xsl:when>
                    <xsl:otherwise>/favicon.svg</xsl:otherwise>
                  </xsl:choose>
                </xsl:attribute>
              </img>
              <h1>
                <xsl:value-of select="atom:title"/>
                <span class="rss-suffix">RSS</span>
              </h1>
            </div>
            <span class="feed-badge">Feed</span>
          </header>

          <div class="hint">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 11a9 9 0 0 1 9 9"/>
              <path d="M4 4a16 16 0 0 1 16 16"/>
              <circle cx="5" cy="19" r="1.5" fill="currentColor" stroke="none"/>
            </svg>
            <div>
              This is an RSS feed. Copy this page's URL into your favorite reader (Feedly, Inoreader, Reeder…) to subscribe.
              <br/>
              <span style="display:inline-block; margin-top:6px;">
                <code id="feed-url"></code>
              </span>
            </div>
          </div>

          <xsl:choose>
            <xsl:when test="atom:entry">
              <ul class="entries">
                <xsl:for-each select="atom:entry">
                  <li class="entry">
                    <div class="entry-meta">
                      <span class="entry-author">
                        <xsl:value-of select="atom:author/atom:name"/>
                      </span>
                      <span class="dot"></span>
                      <time>
                        <xsl:attribute name="datetime">
                          <xsl:value-of select="atom:updated"/>
                        </xsl:attribute>
                        <xsl:value-of select="substring(atom:updated, 1, 10)"/>
                      </time>
                    </div>
                    <div class="entry-content">
                      <xsl:value-of select="atom:summary" disable-output-escaping="yes"/>
                    </div>
                    <xsl:if test="atom:link/@href">
                      <div class="entry-footer">
                        <a target="_blank" rel="noopener">
                          <xsl:attribute name="href">
                            <xsl:value-of select="atom:link/@href"/>
                          </xsl:attribute>
                          查看原文
                          <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                               stroke="currentColor" stroke-width="2.2"
                               stroke-linecap="round" stroke-linejoin="round">
                            <path d="M7 17 17 7"/>
                            <path d="M8 7h9v9"/>
                          </svg>
                        </a>
                      </div>
                    </xsl:if>
                  </li>
                </xsl:for-each>
              </ul>
            </xsl:when>
            <xsl:otherwise>
              <div class="empty">暂无内容</div>
            </xsl:otherwise>
          </xsl:choose>

          <footer class="site-footer">
            <a target="_blank" rel="noopener">
              <xsl:attribute name="href">
                <xsl:value-of select="atom:link/@href"/>
              </xsl:attribute>
              访问站点 →
            </a>
            <span style="margin: 0 10px; opacity: 0.4;">|</span>
            Powered by Ech0
          </footer>
        </main>
        <script>
          document.getElementById('feed-url').textContent = window.location.href;
        </script>
      </body>
    </html>
  </xsl:template>
</xsl:stylesheet>

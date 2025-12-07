import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import http from 'http'
import { EventEmitter } from 'events'
import { parse } from 'url'
import { z } from 'zod'

const host = String(process.env.NOTE_HOST || 'http://localhost:1314').trim().replace(/^`+|`+$/g, '')
const token = process.env.NOTE_TOKEN || ''
let session = process.env.NOTE_SESSION || ''

const s = new McpServer({ name: 'ech0-mcp', version: '0.1.0' })
const bus = new EventEmitter()

const authHeaders = () => {
  const h = { 'Content-Type': 'application/json' }
  if (token) h['Authorization'] = `Bearer ${token}`
  if (session) h['Cookie'] = session
  return h
}

async function searchTool(args) {
  try {
    const q = (args.keyword || '').trim()
    const page = Number(args.page || 1)
    const pageSize = Number(args.pageSize || 10)
    const fmt = String(args.format || '').toLowerCase()
    
    if (!q) {
      // 获取分页内容
      const url = `${host}/api/echo/page?page=${page}&pageSize=${pageSize}`
      const r = await fetch(url, { headers: authHeaders() })
      if (!r.ok) throw new Error(`HTTP ${r.status}`)
      const j = await r.json()
      const arr = (j && j.data && Array.isArray(j.data.items)) ? j.data.items : (Array.isArray(j.items) ? j.items : [])
      const lines = Array.isArray(arr) ? arr.map((it) => {
        const content = String(it.content || '')
        const imgs = Array.isArray(it.images) ? it.images.map(img => img.image_url || img.imageURL) : []
        const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(content)
        const imgMd = !hasMdImg && Array.isArray(imgs) && imgs.length ? imgs.map((u) => `![image](${u})`).join('\n') : ''
        const body = imgMd ? `${imgMd}\n${content}` : content
        const t = formatTimeShort(it.created_at || it.createdAt)
        return `[${it.id}] ${t}\n\n${body}`
      }) : []
      const header = '以下是最新内容：'
      const imgCount = Array.isArray(arr) ? arr.filter((it) => {
        const c = String(it.content || '')
        const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(c)
        const imgs = Array.isArray(it.images) ? it.images : []
        return hasMdImg || (Array.isArray(imgs) && imgs.length)
      }).length : 0
      const sample = Array.isArray(arr) ? arr.slice(0,3).map((it) => String(it.content || '').trim()).filter(Boolean) : []
      const summary = [`共${Array.isArray(arr) ? arr.length : 0}条`, imgCount ? `包含图片${imgCount}条` : ''].concat(sample.map((s) => `- ${s.slice(0,60)}`)).filter(Boolean).join('\n')
      const text = lines.length ? `${header}\n\n${lines.join('\n\n')}\n\n摘要：\n${summary}` : '无匹配结果'
      if (fmt === 'json') return { content: [{ type: 'text', text: JSON.stringify(j) }] }
      return { content: [{ type: 'text', text }] }
    }
    
    if (q.startsWith('#')) {
      const tag = encodeURIComponent(q.slice(1))
      const url = `${host}/api/echo/tag/${tag}?page=${page}&pageSize=${pageSize}`
      const r = await fetch(url, { headers: authHeaders() })
      if (!r.ok) throw new Error(`HTTP ${r.status}`)
      const j = await r.json()
      const arr = (j && j.data && Array.isArray(j.data.items)) ? j.data.items : (Array.isArray(j.items) ? j.items : [])
      const lines = Array.isArray(arr) ? arr.map((it) => {
        const content = String(it.content || '')
        const imgs = Array.isArray(it.images) ? it.images.map(img => img.image_url || img.imageURL) : []
        const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(content)
        const imgMd = !hasMdImg && Array.isArray(imgs) && imgs.length ? imgs.map((u) => `![image](${u})`).join('\n') : ''
        const body = imgMd ? `${imgMd}\n${content}` : content
        const t = formatTimeShort(it.created_at || it.createdAt)
        return `[${it.id}] ${t}\n\n${body}`
      }) : []
      const header = `关于"${q}"的相关内容：`
      const imgCount = Array.isArray(arr) ? arr.filter((it) => {
        const c = String(it.content || '')
        const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(c)
        const imgs = Array.isArray(it.images) ? it.images : []
        return hasMdImg || (Array.isArray(imgs) && imgs.length)
      }).length : 0
      const sample = Array.isArray(arr) ? arr.slice(0,3).map((it) => String(it.content || '').trim()).filter(Boolean) : []
      const summary = [`共${Array.isArray(arr) ? arr.length : 0}条`, imgCount ? `包含图片${imgCount}条` : ''].concat(sample.map((s) => `- ${s.slice(0,60)}`)).filter(Boolean).join('\n')
      const text = lines.length ? `${header}\n\n${lines.join('\n\n')}\n\n摘要：\n${summary}` : '无匹配结果'
      if (fmt === 'json') return { content: [{ type: 'text', text: JSON.stringify(j) }] }
      return { content: [{ type: 'text', text }] }
    }
    
    // 对于其他搜索，目前API没有直接的搜索端点，可以使用分页获取并过滤
    // 这里简单返回分页内容作为替代方案
    const url = `${host}/api/echo/page?page=${page}&pageSize=${pageSize}`
    const r = await fetch(url, { headers: authHeaders() })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    const arr = (j && j.data && Array.isArray(j.data.items)) ? j.data.items : (Array.isArray(j.items) ? j.items : [])
    
    // 简单的客户端过滤
    const filtered = Array.isArray(arr) ? arr.filter((it) => {
      const content = String(it.content || '').toLowerCase()
      return content.includes(q.toLowerCase())
    }) : []
    
    const lines = filtered.map((it) => {
      const content = String(it.content || '')
      const imgs = Array.isArray(it.images) ? it.images.map(img => img.image_url || img.imageURL) : []
      const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(content)
      const imgMd = !hasMdImg && Array.isArray(imgs) && imgs.length ? imgs.map((u) => `![image](${u})`).join('\n') : ''
      const body = imgMd ? `${imgMd}\n${content}` : content
      const t = formatTimeShort(it.created_at || it.createdAt)
      return `[${it.id}] ${t}\n\n${body}`
    })
    const header = `关于"${q}"的搜索结果：`
    const imgCount = filtered.filter((it) => {
      const c = String(it.content || '')
      const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(c)
      const imgs = Array.isArray(it.images) ? it.images : []
      return hasMdImg || (Array.isArray(imgs) && imgs.length)
    }).length
    const sample = filtered.slice(0,3).map((it) => String(it.content || '').trim()).filter(Boolean)
    const summary = [`共${filtered.length}条`, imgCount ? `包含图片${imgCount}条` : ''].concat(sample.map((s) => `- ${s.slice(0,60)}`)).filter(Boolean).join('\n')
    const text = lines.length ? `${header}\n\n${lines.join('\n\n')}\n\n摘要：\n${summary}` : '无匹配结果'
    if (fmt === 'json') return { content: [{ type: 'text', text: JSON.stringify({ ...j, data: { ...j.data, items: filtered } }) }] }
    return { content: [{ type: 'text', text }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function pageTool(args) {
  try {
    const page = Number(args.page || args.page_number || 1)
    const pageSize = Number(args.pageSize || args.page_size || 10)
    const fmt = String(args.format || '').toLowerCase()
    const url = `${host}/api/echo/page?page=${page}&pageSize=${pageSize}`
    const r = await fetch(url, { headers: authHeaders() })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    const arr = (j && j.data && Array.isArray(j.data.items)) ? j.data.items : (Array.isArray(j.items) ? j.items : [])
    const lines = Array.isArray(arr) ? arr.map((it) => {
      const content = String(it.content || '')
      const imgs = Array.isArray(it.images) ? it.images.map(img => img.image_url || img.imageURL) : []
      const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(content)
      const imgMd = !hasMdImg && Array.isArray(imgs) && imgs.length ? imgs.map((u) => `![image](${u})`).join('\n') : ''
      const body = imgMd ? `${imgMd}\n${content}` : content
      const t = formatTimeShort(it.created_at || it.createdAt)
      return `${body}\n\n[${it.id}] ${t}`
    }) : []
    const header = '以下是分页内容：'
    const imgCount = Array.isArray(arr) ? arr.filter((it) => {
      const c = String(it.content || '')
      const hasMdImg = /!\[[^\]]*\]\([^\)]+\)/.test(c)
      const imgs = Array.isArray(it.images) ? it.images : []
      return hasMdImg || (Array.isArray(imgs) && imgs.length)
    }).length : 0
    const sample = Array.isArray(arr) ? arr.slice(0,3).map((it) => String(it.content || '').trim()).filter(Boolean) : []
    const summary = [`共${Array.isArray(arr) ? arr.length : 0}条`, imgCount ? `包含图片${imgCount}条` : ''].concat(sample.map((s) => `- ${s.slice(0,60)}`)).filter(Boolean).join('\n')
    const text = lines.length ? `${header}\n\n${lines.join('\n\n')}\n\n摘要：\n${summary}` : '无匹配结果'
    if (fmt === 'json') return { content: [{ type: 'text', text: JSON.stringify(j) }] }
    return { content: [{ type: 'text', text }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function getTool(args) {
  try {
    const id = String(args.id)
    const r = await fetch(`${host}/api/echo/${id}`, { headers: authHeaders() })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    return { content: [{ type: 'text', text: JSON.stringify(j) }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function publishTool(args) {
  try {
    const priv = Boolean(args.private || false)
    const typeRaw = String(args.type || 'text').toLowerCase()
    const type = ['text','markdown','image','multipart','md'].includes(typeRaw) ? (typeRaw === 'md' ? 'markdown' : typeRaw) : 'text'
    const contentRaw = typeof args.content === 'string' ? args.content : (args.content ? JSON.stringify(args.content) : '')
    const content = String(contentRaw || '').trim()
    const images = Array.isArray(args.images) ? args.images.map(String) : []
    const image = args.image ? String(args.image) : (args.imageURL ? String(args.imageURL) : (images[0] || ''))
    
    const body = { content, private: priv }
    if (type === 'image' || type === 'multipart') {
      if (images.length) {
        // 将images数组转换为符合API期望的格式
        // 注意：API可能需要单独上传图片，这里简化处理
        body.content = content + (images.length ? '\n\n' + images.map(url => `![image](${url})`).join('\n') : '')
      } else if (image) {
        body.content = content + '\n\n' + `![image](${image})`
      }
    }
    
    const endpoint = `${host}/api/echo`
    const r = await fetch(endpoint, { method: 'POST', headers: authHeaders(), body: JSON.stringify(body) })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    return { content: [{ type: 'text', text: JSON.stringify(j) }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function deleteTool(args) {
  try {
    const idRaw = args && args.id
    const id = String(typeof idRaw === 'number' ? idRaw : (idRaw || '')).trim()
    
    if (!id || id === 'undefined' || id === 'null' || !/^[0-9]+$/.test(id)) {
      const kw = String(args && (args.keyword || args.content) || '').trim()
      if (!kw) {
        const text = '使用 MCP 删除失败：参数错误。请提供有效的 id（数字字符串）'
        return { content: [{ type: 'text', text }] }
      }
      // 尝试通过内容搜索获取ID
      try {
        const url = `${host}/api/echo/page?page=1&pageSize=10`
        const r = await fetch(url, { headers: authHeaders() })
        if (!r.ok) throw new Error(`HTTP ${r.status}`)
        const j = await r.json()
        const arr = (j && j.data && Array.isArray(j.data.items)) ? j.data.items : (Array.isArray(j.items) ? j.items : [])
        const found = Array.isArray(arr) ? arr.find((it) => {
          const content = String(it.content || '').toLowerCase()
          return content.includes(kw.toLowerCase())
        }) : null
        if (!found) {
          const text = '使用 MCP 删除失败：未找到匹配信息，无法删除'
          return { content: [{ type: 'text', text }] }
        }
        args.id = String(found.id)
      } catch (e) {
        const text = `使用 MCP 删除失败：搜索错误 ${String(e && e.message || e)}`
        return { content: [{ type: 'text', text }] }
      }
    }
    
    const endpoint = `${host}/api/echo/${args.id}`
    const r = await fetch(endpoint, { method: 'DELETE', headers: authHeaders() })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    const text = `使用 MCP 删除成功：id=${args.id}`
    return { content: [{ type: 'text', text }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function updateTool(args) {
  try {
    const id = String(args.id)
    const content = String(args.content || '')
    const body = { content }
    
    const endpoint = `${host}/api/echo`
    const r = await fetch(endpoint, { method: 'PUT', headers: authHeaders(), body: JSON.stringify(body) })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    return { content: [{ type: 'text', text: JSON.stringify(j) }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function statusTool() {
  try {
    const r = await fetch(`${host}/api/status`)
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    return { content: [{ type: 'text', text: JSON.stringify(j) }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function todayTool() {
  try {
    const r = await fetch(`${host}/api/echo/today`, { headers: authHeaders() })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    return { content: [{ type: 'text', text: JSON.stringify(j) }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function tagsTool() {
  try {
    const r = await fetch(`${host}/api/tags`)
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    return { content: [{ type: 'text', text: JSON.stringify(j) }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

async function loginTool(args) {
  try {
    const username = String(args.username || '')
    const password = String(args.password || '')
    const r = await fetch(`${host}/api/login`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username, password }) })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const j = await r.json()
    const sc = r.headers.get('set-cookie') || ''
    let ck = ''
    if (sc) {
      ck = String(sc).split(',')[0].split(';')[0].trim()
      session = ck
    }
    return { content: [{ type: 'text', text: JSON.stringify({ cookie: ck || sc, response: j }) }] }
  } catch (e) {
    const msg = String(e && e.message || e)
    return { content: [{ type: 'text', text: `error=${msg}` }] }
  }
}

function wrap(name, fn) {
  return async (args) => {
    bus.emit('tool_start', { name, args })
    const res = await fn(args)
    bus.emit('tool_end', { name, res })
    return res
  }
}

function formatTime(v) {
  const d = new Date(String(v || ''))
  if (isNaN(d.getTime())) {
    const s = String(v || '').replace('T', ' ').replace('Z', '')
    return s.split('.')[0]
  }
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

function formatTimeShort(v) {
  const d = new Date(String(v || ''))
  if (!isNaN(d.getTime())) {
    const p = (n) => String(n).padStart(2, '0')
    return `${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
  }
  const s = String(v || '').replace('T', ' ').replace('Z', '')
  const m = s.match(/(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2})/)
  if (m) return `${m[2]}-${m[3]} ${m[4]}:${m[5]}`
  return s.slice(5, 16)
}

// 定义 schemas
const searchSchema = z.object({
  keyword: z.string().optional(),
  query: z.string().optional(),
  page: z.number().optional(),
  pageSize: z.number().optional(),
  format: z.string().optional()
})

const pageSchema = z.object({ 
  page: z.number().optional(), 
  pageSize: z.number().optional(), 
  format: z.string().optional() 
})

const deleteSchema = z.object({ 
  id: z.string().optional(), 
  keyword: z.string().optional(), 
  content: z.string().optional() 
})

const idSchema = z.object({ id: z.string() })
const anyJson = z.any()

const publishSchema = z.object({
  content: anyJson.optional(),
  type: z.string().optional(),
  private: z.boolean().optional(),
  image: z.string().optional(),
  images: z.array(z.string()).optional(),
  imageURL: z.string().optional()
})

const updateSchema = z.object({ 
  id: z.string(), 
  content: z.string() 
})

const loginSchema = z.object({ 
  username: z.string(), 
  password: z.string() 
})

// 注册工具
s.registerTool('search', { 
  description: '搜索Echo内容：默认输出为人类可读文本；若需原始 JSON 传入 format=json', 
  inputSchema: searchSchema 
}, wrap('search', (args) => searchTool({ ...args, keyword: args.keyword || args.query || '' })))

s.registerTool('页面', { 
  description: '分页列表：触发 page/pageSize 参数时必须调用本工具', 
  inputSchema: pageSchema 
}, wrap('页面', pageTool))

s.registerTool('echo', { 
  description: '获取Echo详情：传入 id 时必须调用本工具', 
  inputSchema: idSchema 
}, wrap('echo', getTool))

s.registerTool('publish', { 
  description: '发布Echo内容：触发发布参数时必须调用本工具；支持令牌或会话', 
  inputSchema: publishSchema 
}, wrap('publish', publishTool))

s.registerTool('delete', { 
  description: '删除Echo：需要认证；支持 id 或 keyword/content', 
  inputSchema: deleteSchema 
}, wrap('delete', deleteTool))

s.registerTool('搜索', { 
  description: '搜索Echo内容：默认输出为人类可读文本；若需原始 JSON 传入 format=json', 
  inputSchema: searchSchema 
}, wrap('搜索', (args) => searchTool({ ...args, keyword: args.keyword || args.query || '' })))

s.registerTool('发布', { 
  description: '发布Echo内容：触发发布参数时必须调用本工具；支持令牌或会话', 
  inputSchema: publishSchema 
}, wrap('发布', publishTool))

s.registerTool('笔记', { 
  description: '发布Echo内容：触发发布参数时必须调用本工具；支持令牌或会话', 
  inputSchema: publishSchema 
}, wrap('笔记', publishTool))

s.registerTool('说说', { 
  description: '发布Echo内容：触发发布参数时必须调用本工具；支持令牌或会话', 
  inputSchema: publishSchema 
}, wrap('说说', publishTool))

s.registerTool('删除', { 
  description: '删除Echo：需要认证；支持 id 或 keyword/content', 
  inputSchema: deleteSchema 
}, wrap('删除', deleteTool))

s.registerTool('更新', { 
  description: '更新Echo：需要认证；传入 id/content 时必须调用本工具', 
  inputSchema: updateSchema 
}, wrap('更新', updateTool))

s.registerTool('状态', { 
  description: '系统状态：无入参；需要获取状态时调用本工具', 
  inputSchema: z.object({}) 
}, statusTool)

s.registerTool('今日', { 
  description: '今日Echo：获取今天的Echo列表；需要认证', 
  inputSchema: z.object({}) 
}, todayTool)

s.registerTool('标签', { 
  description: '获取所有标签：无入参；需要获取标签时调用本工具', 
  inputSchema: z.object({}) 
}, tagsTool)

s.registerTool('登录', { 
  description: '会话登录：传入用户名与密码时必须调用本工具', 
  inputSchema: loginSchema 
}, loginTool)

// 启动服务器
const t = new StdioServerTransport()
await s.connect(t)

// HTTP 服务器部分 (可选，用于Web界面)
const httpPort = Number(process.env.NOTE_HTTP_PORT || 0)
if (httpPort) {
  const tools = new Map([
    ['search', searchTool], ['搜索', searchTool],
    ['publish', publishTool], ['发布', publishTool],
    ['delete', deleteTool], ['删除', deleteTool],
    ['update', updateTool], ['更新', updateTool],
    ['echo', getTool], ['消息', getTool],
    ['page', pageTool], ['页面', pageTool],
    ['status', statusTool], ['状态', statusTool],
    ['today', todayTool], ['今日', todayTool],
    ['tags', tagsTool], ['标签', tagsTool],
    ['login', loginTool], ['登录', loginTool]
  ])

  const clients = new Set()

  const srv = http.createServer(async (req, res) => {
    const u = parse(req.url || '', true)
    const p = u.pathname || ''
    const cors = {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization'
    }
    if (req.method === 'OPTIONS') {
      res.writeHead(204, cors)
      res.end()
      return
    }
    if (req.method === 'GET' && p === '/mcp/tools') {
      res.writeHead(200, { 'Content-Type': 'application/json', ...cors })
      res.end(JSON.stringify(Array.from(tools.keys())))
      return
    }
    if (req.method === 'GET' && p === '/mcp/sse') {
      res.writeHead(200, {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        Connection: 'keep-alive',
        ...cors
      })
      res.write('retry: 5000\n\n')
      res.write(`event: mcp_hello\ndata: ${JSON.stringify({ name: 'ech0-mcp', version: '0.1.0' })}\n\n`)
      res.write(`event: mcp_tools\ndata: ${JSON.stringify(Array.from(tools.keys()))}\n\n`)
      const onStart = (e) => res.write(`event: tool_start\ndata: ${JSON.stringify(e)}\n\n`)
      const onEnd = (e) => res.write(`event: tool_end\ndata: ${JSON.stringify(e)}\n\n`)
      const iv = setInterval(() => { try { res.write('event: keepalive\ndata: {}\n\n') } catch {} }, 30000)
      bus.on('tool_start', onStart)
      bus.on('tool_end', onEnd)
      clients.add(res)
      req.on('close', () => {
        bus.off('tool_start', onStart)
        bus.off('tool_end', onEnd)
        clients.delete(res)
        clearInterval(iv)
        res.end()
      })
      return
    }
    if (req.method === 'GET' && p.startsWith('/mcp/sse/tool/')) {
      const name = decodeURIComponent(p.replace('/mcp/sse/tool/', ''))
      const fn = tools.get(name)
      if (!fn) {
        res.writeHead(404, { 'Content-Type': 'text/event-stream', ...cors })
        res.write('retry: 5000\n\n')
        res.write(`event: error\ndata: ${JSON.stringify({ error: 'tool_not_found' })}\n\n`)
        res.end()
        return
      }
      res.writeHead(200, {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        Connection: 'keep-alive',
        ...cors
      })
      res.write('retry: 5000\n\n')
      let args = {}
      try {
        const q = u && u.query ? u.query : {}
        if (q && typeof q.input === 'string' && q.input.length) {
          try { args = JSON.parse(q.input) } catch {}
        } else if (q && typeof q.args === 'string' && q.args.length) {
          try { args = JSON.parse(q.args) } catch {}
        }
      } catch {}
      bus.emit('tool_start', { name, args })
      res.write(`event: tool_start\ndata: ${JSON.stringify({ name, args })}\n\n`)
      ;(async () => {
        try {
          const result = await fn(args)
          res.write(`event: tool_end\ndata: ${JSON.stringify({ name, result })}\n\n`)
        } catch (e) {
          res.write(`event: error\ndata: ${JSON.stringify({ error: String(e && e.message || e) })}\n\n`)
        }
        setTimeout(() => res.end(), 100)
      })()
      return
    }
    res.writeHead(404, cors)
    res.end('Not Found')
  })

  srv.listen(httpPort, () => {
    console.log(`MCP HTTP Server running on port ${httpPort}`)
  })
}
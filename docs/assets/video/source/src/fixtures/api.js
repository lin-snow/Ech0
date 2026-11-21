// Fixture responses for Ech0's real API routes (paths/shapes from web/src/service/api/* and
// web/src/typings/*). Demo persona "Mira" is fictional; no user data is copied from any instance.
import {PANEL_FIXTURES} from './panel.js';
export const API_PREFIXES = ['/api/'];

const ok = data => ({code: 1, msg: '', data});
const DAY = 86400;
// Film clock is fixed at 2026-09-18 09:30 (+08:00): see src/fake-api.js.
const NOW = Math.floor(new Date('2026-09-18T09:30:00+08:00').getTime() / 1000);
const at = (daysAgo, hh = 9, mm = 0) => NOW - daysAgo * DAY - (9 * 3600 + 1800) + hh * 3600 + mm * 60;

export const USER = {id: 'u-mira', username: 'Mira', email: 'mira@example.com', is_admin: true, avatar: '', locale: 'en-US'};
const tag = (id, name, n) => ({id, name, usage_count: n, created_at: at(90)});
export const TAGS = [tag('t-build', 'build', 24), tag('t-notes', 'notes', 41), tag('t-photo', 'photo', 18), tag('t-reading', 'reading', 12)];
const T = n => TAGS.find(t => t.name === n);
const img = (id, echo, name, w, h, sort = 0) => ({
  id: 'ef-' + id, echo_id: echo, file_id: 'f-' + id, sort_order: sort,
  file: {id: 'f-' + id, key: `demo/${name}.jpg`, storage_type: 'local', url: `/demo/${name}.jpg`, name: `${name}.jpg`, content_type: 'image/jpeg', category: 'image', width: w, height: h},
});
const echo = (id, content, created_at, extra = {}) => ({id, content, username: 'Mira', user_id: 'u-mira', private: false, fav_count: 3, layout: 'waterfall', echo_files: [], tags: [], extension: null, created_at, ...extra});

// Newest first. The editor shot publishes NEW_ECHO on top of these.
export const ECHOS = [
  echo('e-harbor', 'Up before the city. The harbor is all gold for about four minutes, then it is just water again.', at(0, 6, 42),
    {tags: [T('photo')], echo_files: [img('harbor', 'e-harbor', 'harbor', 1600, 1066)], fav_count: 12}),
  echo('e-quiet', '### Notes to self\n\nWrite for yourself first.\nPublish if you like.\nOwn it either way.', at(1, 23, 5), {tags: [T('notes')], fav_count: 27}),
  echo('e-ridge', 'Last light over the ridge. Three layers of rose I could never mix on purpose.', at(3, 18, 30),
    {tags: [T('photo')], echo_files: [img('ridge', 'e-ridge', 'ridge', 1600, 1200)]}),
  echo('e-link', 'Saving this for the next time someone asks why I stopped posting on platforms.', at(5, 12, 15),
    {tags: [T('reading')], extension: {type: 'WEBSITE', payload: {title: 'The quiet web is still here', site: 'https://example.com/quiet-web'}}}),
  echo('e-build', 'Shipped reading mode for my little side project. Quiet pages, bigger type, nothing else.', at(8, 22, 40), {tags: [T('build')], fav_count: 9}),
];
const vid = (id, echo, name, w, h) => ({
  id: 'ef-' + id, echo_id: echo, file_id: 'f-' + id, sort_order: 0,
  file: {id: 'f-' + id, key: `videos/${name}`, storage_type: 'local', url: `/demo/${name}`, name, content_type: 'video/mp4', category: 'video', width: w, height: h, size: 1245184},
});
// Published in the write shot, with the clip uploaded there (real upload flow, fixture answer).
export const NEW_ECHO = echo('e-new', 'Moved my notes off the feed and onto my own server tonight.\n\nSame words. **My address.**', at(0, 9, 30),
  {tags: [T('notes')], fav_count: 0, echo_files: [vid('cats', 'e-new', 'cats-at-dusk.mp4', 1916, 1080)]});

const heatmap = (() => {
  const out = []; let s = 11;
  for (let i = 364; i >= 0; i--) {
    s = (s * 16807) % 2147483647;
    const r = (s % 1000) / 1000, d = new Date((NOW - i * DAY) * 1000);
    const count = r < 0.28 ? 0 : r < 0.62 ? 1 : r < 0.85 ? 2 : r < 0.95 ? 3 : 5;
    out.push({date: d.toISOString().slice(0, 10), count: d.getDay() === 0 && r < 0.5 ? 0 : count});
  }
  return out;
})();

const cm = (id, echo, nickname, content, hoursAgo) => ({id, echo_id: echo, nickname, email: '', website: '', content, status: 'approved', hot: false, source: 'guest', created_at: NOW - hoursAgo * 3600, updated_at: NOW - hoursAgo * 3600});
const COMMENTS = [
  cm('c1', 'e-harbor', 'Jonas', 'Four minutes of gold is worth the alarm. Great shot.', 1),
  cm('c2', 'e-quiet', 'Aiko', 'Printing "own it either way" and taping it to my monitor.', 5),
  cm('c3', 'e-link', 'Sam', 'Found you through Ech0 Hub. Subscribed via RSS!', 20),
];
// Other (fictional) Ech0 instances this one is connected to.
const inst = (id, name, user, total, today) => ({id, server_name: name, server_url: `https://${id}.example`, logo: `/demo/av-${id}.png`, total_echos: total, today_echos: today, sys_username: user, version: '5.7.0'});
const CONNECTS = [inst('jonas', 'Jonas', 'jonas', 812, 2), inst('aiko', 'Aiko Notes', 'aiko', 356, 1), inst('sam', 'sam.log', 'sam', 1204, 3), inst('lena', 'Lena', 'lena', 97, 0)];

const settings = {
  site_title: 'Mira — notes & light', server_logo: '', server_name: 'Mira', server_url: 'https://mira.example',
  allow_register: false, default_locale: 'en-US', ICP_number: '', footer_content: '', footer_link: '', meting_api: '', custom_css: '', custom_js: '',
};
const state = {published: false, liked: false, readerComment: false};
// Detail chapter: a reader's reply to the new Echo (posted through the real comment form).
export const READER = {nickname: 'Jonas', email: 'jonas@example.com', content: 'Same words, my address. Love this. Following along via RSS.'};
const detailComments = () => [
  ...(state.readerComment ? [cm('c-reader', 'e-new', READER.nickname, READER.content, 0)] : []),
  cm('c-lena', 'e-new', 'Lena', 'Stealing "same words, my address" for my own about page.', 0.2),
];
export const resetDetail = () => { state.liked = false; state.readerComment = false; };
export const setPublished = v => { state.published = v; };
export const isPublished = () => state.published;

export const API_FIXTURES = {
  '/api/init/status': ok({initialized: true, owner_exists: true}),
  '/api/auth/refresh': ok({access_token: 'film-token', refresh_token: 'film-refresh'}),
  '/api/user': ok(USER),
  '/api/settings': ok(settings),
  '/api/agent/info': ok({enable: true, model: 'claude-sonnet-5', protocol: 'anthropic'}),
  '/api/hello': ok({hello: 'Hello, Ech0!', copyright: '', version: '5.7.0', commit: '', build_time: '', license: 'AGPL-3.0-or-later', author: 'lin-snow', repo_url: 'https://github.com/lin-snow/Ech0'}),
  '/api/oauth2/status': ok({enabled: false, provider: '', oauth_ready: false}),
  '/api/passkey/status': ok({passkey_ready: false}),
  '/api/tags': ok(TAGS),
  '/api/heatmap': ok(heatmap),
  '/api/echo/query': () => {
    const items = state.published ? [NEW_ECHO, ...ECHOS] : ECHOS;
    return ok({items, total: items.length});
  },
  '/api/echo/today': () => ok(state.published ? [NEW_ECHO] : []),
  '/api/echo/onthisday': ok([]),
  '/api/echo/hot': ok(ECHOS.slice(0, 3)),
  '/api/echo': (url, init) => {
    if ((init?.method || 'GET').toUpperCase() === 'POST') { state.published = true; return ok(NEW_ECHO); }
    return ok(null);
  },
  '/api/chat/session': ok([]),
  '/api/echo/e-new': () => ok({...NEW_ECHO, fav_count: NEW_ECHO.fav_count + (state.liked ? 1 : 0)}),
  '/api/echo/like/e-new': (url, init) => { if ((init?.method || 'GET').toUpperCase() !== 'GET') state.liked = true; return ok(null); },
  '/api/comments': (url, init) => {
    if ((init?.method || 'GET').toUpperCase() === 'POST') { state.readerComment = true; return ok({status: 'approved'}); }
    return ok(url.searchParams.get('echo_id') === 'e-new' ? detailComments() : []);
  },
  '/api/comments/public': ok(COMMENTS),
  '/api/comments/form': ok({form_token: 'film', min_submit_ms: 0, captcha_enabled: false, captcha_api_endpoint: '', enable_comment: true}),
  '/api/connect/list': ok(CONNECTS.map(c => ({id: c.id, connect_url: c.server_url}))),
  '/api/connects/info': ok(CONNECTS),
  ...PANEL_FIXTURES,
  '/api/agent/recent': ok('This week Mira moved her notes onto her own server, photographed the harbor at dawn, and kept coming back to one idea: **write for yourself first.**'),
};

// Title chapters. open/echo are carried by the persistent rail (src/shots/rail.jsx); this file holds
// the "own" command card and the closing logo assembly.
import React from 'react';
import {Split, reveal} from '../kit/text.jsx';
import {shot, onRender} from '../engine.js';
import {showShot, q, Typed, typeInto, BEAT} from './common.jsx';
import {DOT} from './rail.jsx';

export const Empty = () => null;

// README "Try in 60 Seconds" (README.md#L74), verbatim.
const CMD = ['docker run -d \\', '  --name ech0 \\', '  -p 6277:6277 \\', '  -v /opt/ech0/data:/app/data \\', '  sn0wl1n/ech0:latest'];
export function OwnView() {
  return <div className="cmd-card">
    <div className="cmd-head"><i /><i /><i /><span>server</span></div>
    <pre className="cmd-body">{CMD.map((l, i) => <div key={i} className="cmd-line" data-line={i}><span className="cmd-typed" /></div>)}</pre>
  </div>;
}
export function buildOwn(tl) {
  const s = shot('own');
  showShot(tl, 'own');
  const card = q('own', '.cmd-card');
  tl.fromTo(card, {opacity: 0, y: 40}, {opacity: 1, y: 0, duration: 0.7, ease: 'expo.out', immediateRender: false}, s.start + 0.8);
  tl.set(card, {opacity: 0}, 0);
  const TYPE_AT = 1.25, CPS = 52, MARK_AT = 3.75;
  const total = CMD.join('\n');
  onRender('own', local => {
    const n = local < TYPE_AT ? 0 : Math.min(total.length, Math.floor((local - TYPE_AT) * CPS));
    let left = n;
    CMD.forEach((line, i) => {
      const el = card.querySelector(`[data-line="${i}"] .cmd-typed`);
      const shown = line.slice(0, Math.max(0, Math.min(line.length, left)));
      left -= line.length + 1;
      const marked = local >= MARK_AT && line.includes('/app/data');
      const html = marked ? shown.replace('/app/data', '<mark>/app/data</mark>') : shown.replace(/</g, '&lt;');
      if (el.innerHTML !== html) el.innerHTML = html;
    });
  });
}

// Ech0.svg geometry (web/public/Ech0.svg, 408×408): the four strokes of the face.
export function CloseView() {
  return <div className="close-stage">
    <svg className="logo" viewBox="0 0 408 408" width="340" height="340">
      <rect className="logo-plate" width="408" height="408" rx="31" fill="#FFF4E4" stroke="#FFF3E8" />
      <g className="logo-dash"><path d="M261.448 114.512C257.029 114.513 253.447 110.931 253.448 106.512L253.45 98.7982C253.451 94.3806 257.033 90.799 261.451 90.798L301.505 90.7884C305.924 90.7873 309.505 94.3692 309.504 98.7882L309.502 106.502C309.501 110.919 305.919 114.501 301.502 114.502L261.448 114.512Z" fill="#000" /></g>
      <g className="logo-slash"><path d="M199.965 160.5L215 181.845L215 234.5L199.965 209L199.965 160.5Z" fill="#000" /></g>
      <g className="logo-bar"><rect x="150" y="291" width="124" height="20" fill="#000" /></g>
      <circle className="logo-eye" cx="132" cy="102" r="14" fill="#F54A00" />
    </svg>
    <div className="close-copy">
      <div className="close-word"><Typed className="close-typed" /></div>
      <div className="close-line"><Split text="A timeline you own." /></div>
      <div className="close-meta">Open source · Self-hosted · ech0.app</div>
    </div>
  </div>;
}
export function buildClose(tl) {
  const s = shot('close');
  showShot(tl, 'close');
  const svg = q('close', '.logo');
  const LOGO = {x: 560, y: 370, size: 340};   // svg top-left on screen
  const eye = {x: LOGO.x + 132 / 408 * LOGO.size, y: LOGO.y + 102 / 408 * LOGO.size};
  // The film dot travels from the rail to the logo's eye; then the logo takes over.
  tl.to('#rail .film-dot', {x: eye.x, y: eye.y, duration: 0.9, ease: 'power3.inOut'}, s.start + 0.1);
  tl.set('#rail .film-dot', {opacity: 0}, s.start + 1.0);
  tl.set('#rail .film-dot', {opacity: 1}, 0);
  tl.set(svg, {opacity: 1}, s.start);
  tl.fromTo(q('close', '.logo-eye'), {opacity: 0}, {opacity: 1, duration: 0.01, immediateRender: false}, s.start + 0.99);
  tl.set(q('close', '.logo-eye'), {opacity: 0}, 0);
  const pieces = [['.logo-dash', {x: 180, y: -120}], ['.logo-slash', {x: 0, y: -220}], ['.logo-bar', {x: -200, y: 140}]];
  pieces.forEach(([sel, from], i) => {
    tl.fromTo(q('close', sel), {...from, opacity: 0}, {x: 0, y: 0, opacity: 1, duration: 0.8, ease: 'expo.out', immediateRender: false}, s.start + 0.5 + i * BEAT * 0.5);
    tl.set(q('close', sel), {opacity: 0}, 0);
  });
  tl.fromTo(q('close', '.logo-plate'), {opacity: 0, scale: 0.92, transformOrigin: '50% 50%'}, {opacity: 1, scale: 1, duration: 0.7, ease: 'expo.out', immediateRender: false}, s.start + 1.5);
  tl.set(q('close', '.logo-plate'), {opacity: 0}, 0);
  const word = q('close', '.close-typed');
  onRender('close', local => typeInto(word, local, 2.0, 'Ech0', 7, {cursorFrom: 1.7}));
  reveal(tl, q('close', '.close-line'), s.start + 2.9, {stagger: 0.03});
  tl.fromTo(q('close', '.close-meta'), {opacity: 0, y: 14}, {opacity: 1, y: 0, duration: 0.6, ease: 'expo.out', immediateRender: false}, s.start + 3.6);
  tl.set(q('close', '.close-meta'), {opacity: 0}, 0);
  void DOT;
}

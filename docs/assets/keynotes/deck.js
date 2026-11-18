/*
 * Ech0 keynote runtime — shared by every deck under docs/assets/keynotes/.
 * Classic script (no ES modules) so the deck also runs from file:// without a server.
 *
 * Keys: → ↓ Space PageDown / ← ↑ PageUp · Home End · digits+Enter · f · n · t · Esc
 * URL:  #<n> addresses a slide, so a link can point at one.
 */
(function () {
  'use strict';

  var slides = Array.prototype.slice.call(document.querySelectorAll('.slide'));
  if (!slides.length) return;

  var panel = document.querySelector('.notes-panel');
  var progress = document.querySelector('.progress');
  var counter = document.querySelector('.counter');
  var index = 0;
  var jump = '';
  var jumpTimer = 0;

  function clamp(n) {
    return Math.min(Math.max(n, 0), slides.length - 1);
  }

  function fromHash() {
    var n = parseInt(String(location.hash).slice(1), 10);
    return isNaN(n) ? 0 : clamp(n - 1);
  }

  function paint() {
    slides.forEach(function (slide, i) {
      slide.classList.toggle('is-active', i === index);
    });
    if (progress) progress.style.width = ((index + 1) / slides.length) * 100 + '%';
    if (counter) counter.textContent = index + 1 + ' / ' + slides.length;
    if (panel && !panel.hidden) panel.textContent = notesOf(index);
    history.replaceState(null, '', '#' + (index + 1));
  }

  function notesOf(i) {
    var note = slides[i] && slides[i].querySelector('.notes');
    return note ? note.textContent.trim() : '';
  }

  function go(n) {
    index = clamp(n);
    paint();
  }

  function fit() {
    // Derive from the logical 1600×900 box, then center it by offset: the stage
    // is transform-origin 0 0, so no layout centering is involved.
    var scale = Math.min(window.innerWidth / 1600, window.innerHeight / 900);
    var ox = (window.innerWidth - 1600 * scale) / 2;
    var oy = (window.innerHeight - 900 * scale) / 2;
    var style = document.documentElement.style;
    style.setProperty('--scale', String(scale));
    style.setProperty('--ox', ox + 'px');
    style.setProperty('--oy', oy + 'px');
  }

  function toggleNotes() {
    if (!panel) return;
    panel.hidden = !panel.hidden;
    if (!panel.hidden) panel.textContent = notesOf(index);
  }

  function isDark() {
    var theme = document.documentElement.dataset.theme;
    if (theme) return theme === 'dark';
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  }

  function toggleTheme() {
    document.documentElement.dataset.theme = isDark() ? 'light' : 'dark';
  }

  function toggleFullscreen() {
    if (document.fullscreenElement) document.exitFullscreen();
    else document.documentElement.requestFullscreen();
  }

  document.addEventListener('keydown', function (e) {
    if (e.metaKey || e.ctrlKey || e.altKey) return;

    var key = e.key;
    if (/^[0-9]$/.test(key)) {
      jump += key;
      clearTimeout(jumpTimer);
      jumpTimer = setTimeout(function () {
        jump = '';
      }, 900);
      return;
    }

    switch (key) {
      case 'ArrowRight':
      case 'ArrowDown':
      case 'PageDown':
      case ' ':
      case 'Enter':
        if (jump) {
          go(parseInt(jump, 10) - 1);
          jump = '';
        } else {
          go(index + 1);
        }
        break;
      case 'ArrowLeft':
      case 'ArrowUp':
      case 'PageUp':
        go(index - 1);
        break;
      case 'Home':
        go(0);
        break;
      case 'End':
        go(slides.length - 1);
        break;
      case 'f':
        toggleFullscreen();
        break;
      case 'n':
        toggleNotes();
        break;
      case 't':
        toggleTheme();
        break;
      case 'Escape':
        if (panel) panel.hidden = true;
        break;
      default:
        return;
    }
    e.preventDefault();
  });

  // Touch / trackpad swipe.
  var startX = 0;
  var startY = 0;
  document.addEventListener(
    'touchstart',
    function (e) {
      startX = e.touches[0].clientX;
      startY = e.touches[0].clientY;
    },
    { passive: true }
  );
  document.addEventListener(
    'touchend',
    function (e) {
      var dx = e.changedTouches[0].clientX - startX;
      var dy = e.changedTouches[0].clientY - startY;
      if (Math.abs(dx) < 60 || Math.abs(dx) < Math.abs(dy)) return;
      go(index + (dx < 0 ? 1 : -1));
    },
    { passive: true }
  );

  window.addEventListener('hashchange', function () {
    go(fromHash());
  });
  window.addEventListener('resize', fit);

  fit();
  go(fromHash());
  window.addEventListener('load', fit);
})();

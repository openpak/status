(function () {
  'use strict';
  document.documentElement.classList.add('js');

  // Deep links: #switch opens Switch, #switch/game-servers opens the category too.
  function openPath(id) {
    var el = id && document.getElementById(id);
    if (!el) return;
    for (var d = el; d; d = d.parentElement) if (d.tagName === 'DETAILS') d.open = true;
    el.scrollIntoView({ block: 'start' });
  }
  function fromHash() { try { openPath(decodeURIComponent(location.hash.slice(1))); } catch (e) {} }
  fromHash();
  window.addEventListener('hashchange', fromHash);

  // Search: a match opens its path; everything else hides. Clearing restores what the page opened.
  var q = document.getElementById('q'), none = document.getElementById('none');
  var details = [].slice.call(document.querySelectorAll('#nets details'));
  var initial = details.map(function (d) { return d.open; });
  var rows = [].slice.call(document.querySelectorAll('#nets .row'));
  if (q) q.addEventListener('input', function () {
    var s = q.value.trim().toLowerCase(), shown = 0;
    [].forEach.call(document.querySelectorAll('#nets > .fam'), function (f) { f.hidden = !!s; });
    if (!s) {
      rows.forEach(function (r) { r.hidden = false; });
      details.forEach(function (d, i) { d.open = initial[i]; d.parentElement.hidden = false; });
      none.style.display = 'none';
      return;
    }
    rows.forEach(function (r) { r.hidden = r.getAttribute('data-text').toLowerCase().indexOf(s) < 0; });
    // Innermost first, so a category is decided before its network.
    details.slice().reverse().forEach(function (d) {
      var any = d.querySelector('.row:not([hidden])');
      var own = d.querySelector(':scope > summary .t').textContent.toLowerCase().indexOf(s) >= 0;
      if (own && !any) { // the category or network itself matches: show all beneath
        [].forEach.call(d.querySelectorAll('.row'), function (r) { r.hidden = false; });
        any = true;
      }
      d.parentElement.hidden = !any;
      d.open = !!any && !own || (own && d.classList.contains('cat') ? false : !!any);
      if (any && d.classList.contains('net')) shown++;
    });
    none.style.display = shown ? 'none' : 'block';
  });

  // The uptime strip's tooltip: time and answer for the check under the pointer.
  var tip = document.createElement('div');
  tip.className = 'sv-tip'; tip.hidden = true; document.body.appendChild(tip);
  function show(h, x, y, i) {
    var n = h.getAttribute('data-h').length, end = +h.getAttribute('data-end'), step = +h.getAttribute('data-step');
    var ok = h.getAttribute('data-h').charAt(i) === '1';
    var t = new Date((end - (n - 1 - i) * step) * 1000);
    tip.textContent = '';
    var b = document.createElement('b'); b.textContent = ok ? 'Working' : 'Not responding';
    tip.appendChild(b);
    tip.appendChild(document.createTextNode(' · ' + t.toLocaleString([], { hour: '2-digit', minute: '2-digit', day: 'numeric', month: 'short' })));
    tip.hidden = false;
    tip.style.left = Math.max(6, Math.min(innerWidth - tip.offsetWidth - 6, x - tip.offsetWidth / 2)) + 'px';
    tip.style.top = (y - tip.offsetHeight - 8) + 'px';
  }
  document.addEventListener('pointermove', function (e) {
    var h = e.target.closest && e.target.closest('.sv-hist');
    if (!h) { tip.hidden = true; return; }
    var r = h.getBoundingClientRect(), n = h.getAttribute('data-h').length;
    var i = Math.max(0, Math.min(n - 1, Math.floor((e.clientX - r.left) / r.width * n)));
    show(h, e.clientX, r.top, i);
  });
  document.addEventListener('focusin', function (e) {
    var h = e.target.closest && e.target.closest('.sv-hist');
    if (!h) return;
    var r = h.getBoundingClientRect();
    show(h, r.right - 6, r.top, h.getAttribute('data-h').length - 1);
  });
  document.addEventListener('focusout', function () { tip.hidden = true; });
})();

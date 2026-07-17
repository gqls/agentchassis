/* Tiny deterministic "hand-drawn" SVG helper.
   Seeded so every render of the PDF is byte-identical. */

function rng(seed) {
  let s = seed >>> 0 || 1;
  return function () {
    s ^= s << 13; s >>>= 0;
    s ^= s >> 17;
    s ^= s << 5; s >>>= 0;
    return s / 4294967296;
  };
}

const SVGNS = 'http://www.w3.org/2000/svg';

function el(name, attrs) {
  const n = document.createElementNS(SVGNS, name);
  for (const k in attrs) n.setAttribute(k, attrs[k]);
  return n;
}

/* A wobbly line: two passes, each a quadratic bezier bowed by a random amount. */
function roughLine(g, x1, y1, x2, y2, o) {
  const r = o.r, passes = o.passes === undefined ? 2 : o.passes;
  const len = Math.hypot(x2 - x1, y2 - y1);
  const wob = Math.min(o.wobble === undefined ? 2.6 : o.wobble, len / 12);
  for (let p = 0; p < passes; p++) {
    const jx1 = x1 + (r() - 0.5) * wob, jy1 = y1 + (r() - 0.5) * wob;
    const jx2 = x2 + (r() - 0.5) * wob, jy2 = y2 + (r() - 0.5) * wob;
    const mx = (x1 + x2) / 2 + (r() - 0.5) * wob * 2.1;
    const my = (y1 + y2) / 2 + (r() - 0.5) * wob * 2.1;
    g.appendChild(el('path', {
      d: `M ${jx1.toFixed(2)} ${jy1.toFixed(2)} Q ${mx.toFixed(2)} ${my.toFixed(2)} ${jx2.toFixed(2)} ${jy2.toFixed(2)}`,
      fill: 'none',
      stroke: o.stroke,
      'stroke-width': o.width * (p ? 0.8 : 1),
      'stroke-linecap': 'round',
      opacity: p ? 0.55 : 1
    }));
  }
}

function roughRect(g, x, y, w, h, o) {
  if (o.fill) {
    g.appendChild(el('path', {
      d: `M ${x + 3} ${y + 1} L ${x + w - 2} ${y + 2} L ${x + w - 1} ${y + h - 2} L ${x + 2} ${y + h - 1} Z`,
      fill: o.fill, stroke: 'none', opacity: o.fillOpacity === undefined ? 1 : o.fillOpacity
    }));
  }
  roughLine(g, x, y, x + w, y, o);
  roughLine(g, x + w, y, x + w, y + h, o);
  roughLine(g, x + w, y + h, x, y + h, o);
  roughLine(g, x, y + h, x, y, o);
}

function roughEllipse(g, cx, cy, rx, ry, o) {
  const r = o.r;
  if (o.fill) {
    g.appendChild(el('ellipse', { cx, cy, rx, ry, fill: o.fill, stroke: 'none', opacity: o.fillOpacity === undefined ? 1 : o.fillOpacity }));
  }
  for (let p = 0; p < 2; p++) {
    const pts = [];
    const start = r() * 0.5;
    for (let i = 0; i <= 16; i++) {
      const a = start + (i / 16) * Math.PI * 2;
      const k = 1 + (r() - 0.5) * 0.035;
      pts.push([cx + Math.cos(a) * rx * k, cy + Math.sin(a) * ry * k]);
    }
    let d = `M ${pts[0][0].toFixed(2)} ${pts[0][1].toFixed(2)}`;
    for (let i = 1; i < pts.length; i++) d += ` L ${pts[i][0].toFixed(2)} ${pts[i][1].toFixed(2)}`;
    g.appendChild(el('path', {
      d: d + ' Z', fill: 'none', stroke: o.stroke,
      'stroke-width': o.width * (p ? 0.75 : 1), 'stroke-linejoin': 'round',
      'stroke-linecap': 'round', opacity: p ? 0.5 : 1
    }));
  }
}

/* Arrow along a slightly bowed path, with a scratchy two-stroke head. */
function roughArrow(g, x1, y1, x2, y2, o) {
  const r = o.r;
  const bow = o.bow || 0;
  const dx = x2 - x1, dy = y2 - y1;
  const len = Math.hypot(dx, dy) || 1;
  const nx = -dy / len, ny = dx / len;
  const mx = (x1 + x2) / 2 + nx * bow + (r() - 0.5) * 2;
  const my = (y1 + y2) / 2 + ny * bow + (r() - 0.5) * 2;
  for (let p = 0; p < 2; p++) {
    g.appendChild(el('path', {
      d: `M ${x1 + (r() - 0.5) * 1.4} ${y1 + (r() - 0.5) * 1.4} Q ${mx} ${my} ${x2 + (r() - 0.5) * 1.4} ${y2 + (r() - 0.5) * 1.4}`,
      fill: 'none', stroke: o.stroke, 'stroke-width': o.width * (p ? 0.7 : 1),
      'stroke-linecap': 'round', opacity: p ? 0.5 : 1,
      'stroke-dasharray': o.dash || 'none'
    }));
  }
  // head: angle taken from the curve's final tangent (control point -> tip)
  const ang = Math.atan2(y2 - my, x2 - mx);
  const hl = o.head || 9;
  for (const s of [1, -1]) {
    const a = ang + s * 0.42 + Math.PI;
    roughLine(g, x2, y2, x2 + Math.cos(a) * hl, y2 + Math.sin(a) * hl,
      { r, stroke: o.stroke, width: o.width, wobble: 0.8, passes: 1 });
  }
}

function label(g, x, y, text, o) {
  o = o || {};
  /* Hand-written text never sits perfectly level: nudge each line a fraction
     of a degree when a generator is available and no angle was asked for. */
  let rot = o.rotate;
  if (rot === undefined && o.r) rot = (o.r() - 0.5) * 2.6;
  const t = el('text', {
    x, y,
    'font-family': o.family || "'Caveat', cursive",
    'font-size': o.size || 19,
    'font-weight': o.weight || 600,
    fill: o.fill || '#1d2b45',
    'text-anchor': o.anchor || 'middle'
  });
  if (rot) t.setAttribute('transform', `rotate(${rot.toFixed(2)} ${x} ${y})`);
  t.textContent = text;
  g.appendChild(t);
  return t;
}

/* Multi-line label helper */
function labelLines(g, x, y, lines, o) {
  o = o || {};
  const lh = o.lineHeight || (o.size || 19) * 0.92;
  lines.forEach((ln, i) => label(g, x, y + i * lh, ln, o));
}

function box(svg, x, y, w, h, lines, o) {
  o = o || {};
  const g = el('g', {});
  const rot = o.rotate || 0;
  if (rot) g.setAttribute('transform', `rotate(${rot} ${x + w / 2} ${y + h / 2})`);
  roughRect(g, x, y, w, h, {
    r: o.r, stroke: o.stroke || '#1d2b45', width: o.width || 1.6,
    fill: o.fill, fillOpacity: o.fillOpacity
  });
  const size = o.size || 17;
  const lh = o.lineHeight || size * 0.95;
  const startY = y + h / 2 - ((lines.length - 1) * lh) / 2 + size * 0.34;
  lines.forEach((ln, i) => {
    label(g, x + w / 2, startY + i * lh, ln, {
      size, family: o.family, weight: o.weight, fill: o.textFill || '#16233c', r: o.r
    });
  });
  svg.appendChild(g);
  return g;
}

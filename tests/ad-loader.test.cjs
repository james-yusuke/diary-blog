const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const vm = require('node:vm');

const source = readFileSync('assets/site.js', 'utf8');
const loader = source.slice(source.lastIndexOf('(() => {'));
const flush = () => new Promise(resolve => setImmediate(resolve));

function setup() {
  const inserted = [];
  const resize = [];
  const intersections = [];
  const window = {};
  const slots = ['banner', 'sidebar'].map(kind => {
    const classes = new Set();
    const mount = {
      style: {}, children: [],
      replaceChildren() { this.children = []; },
      appendChild(script) {
        this.children.push(script);
        inserted.push({ script, options: { ...window.atOptions }, mount });
      },
    };
    return {
      dataset: { adSlot: kind }, width: 800, visible: true, mount,
      querySelector: () => mount,
      getBoundingClientRect() { return { width: this.width }; },
      getClientRects() { return this.visible ? [{}] : []; },
      classList: { add: c => classes.add(c), remove: c => classes.delete(c) },
    };
  });
  vm.runInNewContext(loader, {
    window, document: {
      querySelectorAll: () => slots,
      createElement(tag) { assert.equal(tag, 'script'); return {}; },
    },
    ResizeObserver: class {
      constructor(callback) { resize.push(callback); }
      observe() {}
    },
    IntersectionObserver: class {
      constructor(callback) { intersections.push(callback); }
      observe() {}
      disconnect() {}
    },
  });
  return { slots, inserted, resize, intersections };
}

test('loads nearby ads in the publisher document with serialized options', async () => {
  const state = setup();
  assert.equal(state.inserted.length, 0);
  state.intersections.forEach(callback => callback([{ isIntersecting: true }]));
  await flush();
  assert.equal(state.inserted.length, 1);
  const first = state.inserted[0];
  assert.equal(first.options.width, 728);
  assert.equal(first.script.src, `https://www.highrevenueformat.com/${first.options.key}/invoke.js`);
  first.script.onload();
  await flush();
  assert.equal(state.inserted.length, 2);
  assert.equal(state.inserted[1].options.width, 160);
});

test('a blocked ad releases the queue and collapses its reserved space', async () => {
  const state = setup();
  state.intersections.forEach(callback => callback([{ isIntersecting: true }]));
  await flush();
  state.inserted[0].script.onerror();
  await flush();
  assert.equal(state.slots[0].mount.style.height, '0px');
  assert.equal(state.slots[0].mount.children.length, 0);
  assert.equal(state.inserted.length, 2);
});

test('does not load a queued sidebar after switching to a mobile layout', async () => {
  const state = setup();
  state.intersections.forEach(callback => callback([{ isIntersecting: true }]));
  await flush();
  state.slots[1].visible = false;
  state.resize[1]();
  state.inserted[0].script.onload();
  await flush();
  assert.equal(state.inserted.length, 1);
});

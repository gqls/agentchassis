/* ===========================================================================
   state.js — Vibe Equalizer
   ---------------------------------------------------------------------------
   WHY THIS FILE EXISTS. The tool has been broken since it shipped on
   websitedesign.com: its index.html loads `../../js/state.js`, and that file
   does not exist anywhere in the source repo. `StateManager` was therefore
   undefined, `loadState()` threw on page load, and no slider ever worked. This
   is the missing half, written from the calls script.js actually makes.

   THE CONTRACT, read off script.js:
     StateManager.save({m, d, dp, r})   called on every slider input
     StateManager.load() -> object|null called once at startup by loadState()

   STATE LIVES IN THE URL, not just localStorage. That is not a preference: the
   Share button does `navigator.clipboard.writeText(window.location.href)` and
   promises "you can share this exact vibe". A localStorage-only implementation
   would copy a URL carrying none of the configuration, so the button would
   appear to work and silently share nothing. localStorage is kept as a
   secondary so a plain revisit restores your last position.
   =========================================================================== */

(function (global) {
    'use strict';

    var KEY = 'wd-vibe-equalizer';
    var FIELDS = ['m', 'd', 'dp', 'r'];

    function save(state) {
        if (!state) return;

        var params = new URLSearchParams();
        FIELDS.forEach(function (f) {
            if (state[f] !== undefined && state[f] !== null) {
                params.set(f, String(state[f]));
            }
        });

        // replaceState, not pushState: the sliders fire on every input event and
        // pushState would bury the back button under hundreds of entries.
        try {
            history.replaceState(null, '', location.pathname + location.search + '#' + params.toString());
        } catch (e) {
            location.hash = params.toString();
        }

        try {
            localStorage.setItem(KEY, JSON.stringify(state));
        } catch (e) {
            /* private mode, or storage full — the URL still carries the state */
        }
    }

    function load() {
        var hash = (location.hash || '').replace(/^#/, '');
        if (hash) {
            var params = new URLSearchParams(hash);
            var fromUrl = {};
            var found = false;
            FIELDS.forEach(function (f) {
                var v = params.get(f);
                if (v !== null) { fromUrl[f] = v; found = true; }
            });
            if (found) return complete(fromUrl);
        }

        try {
            var raw = localStorage.getItem(KEY);
            if (raw) return complete(JSON.parse(raw));
        } catch (e) {
            /* unreadable or not JSON — fall through to defaults */
        }
        return null;
    }

    // A partially-saved object would assign `undefined` to a range input's
    // value, which the browser coerces to the string "undefined" and the slider
    // then snaps to its minimum. Fill any missing field with the control's own
    // current value so a half-written URL degrades to a sensible position
    // instead of a broken one.
    function complete(state) {
        FIELDS.forEach(function (f) {
            if (state[f] === undefined || state[f] === null || state[f] === '') {
                delete state[f];
            }
        });
        var ids = { m: 'mood', d: 'density', dp: 'depth', r: 'radius' };
        FIELDS.forEach(function (f) {
            if (state[f] === undefined) {
                var el = document.getElementById(ids[f]);
                if (el) state[f] = el.value;
            }
        });
        return state;
    }

    global.StateManager = { save: save, load: load };

    // The Copy button in the markup has no handler in script.js — it has never
    // done anything. Wired here rather than by patching script.js, so the port
    // keeps that file byte-identical to its source.
    document.addEventListener('DOMContentLoaded', function () {
        var copy = document.getElementById('btn-copy');
        var output = document.getElementById('prompt-output');
        if (!copy || !output) return;
        copy.addEventListener('click', function () {
            navigator.clipboard.writeText(output.value).then(function () {
                var original = copy.textContent;
                copy.textContent = 'Copied';
                setTimeout(function () { copy.textContent = original; }, 1500);
            });
        });
    });
})(window);

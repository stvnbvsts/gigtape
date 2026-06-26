# Handoff: Gigtape "Mixtape" Redesign

## Overview
A full visual redesign of the Gigtape web app (`apps/web`, Vue 3 + Vite SPA) in a
**90s mixtape / zine** aesthetic — cream paper, masking tape, marker handwriting,
spinning cassette reels, hard-offset "stamped" buttons. It covers the entire existing
flow: connect/landing → artist search → setlist preview & edit → playlist result →
festival mode → festival result.

The redesign is intentionally nostalgic: the app should feel like making a tape for a
friend before a show, not like a generic web form.

## About the Design Files
The file in this bundle — **`Gigtape.prototype.html`** — is a **design reference**, not
production code to copy. It is a single self-contained HTML/JS prototype that demonstrates
the intended look, layout, copy, and interactions with **mock data**.

The task is to **recreate this look in the existing `apps/web` Vue codebase**, reusing its
current components, router, and API client. Do **not** introduce React, do **not** ship the
HTML file, and do **not** change any networking or business logic. This is a **restyle +
light restructure** of existing Vue views — the data layer (`src/api/client.ts`), session
handling, and OAuth flow stay exactly as they are.

> The prototype's JS class (`renderVals`, mock `setlistsFor`, `defaultFest`, etc.) exists
> only to make the prototype interactive. Ignore it for data — your real data already comes
> from `src/api/client.ts`. Use it only to understand **interaction logic and derived UI
> state** (e.g. how removed tracks renumber, how the festival CTA label changes).

## Fidelity
**High-fidelity (hifi).** Colors, typography, spacing, copy, and interactions are final.
Match them precisely. The one flexible area is the **landing/connect split** (see Routing).

---

## Target codebase map

Existing files in `apps/web/src` and how each maps to a redesign screen:

| Prototype screen | Existing Vue file | What changes |
|---|---|---|
| Landing / Connect | `views/ArtistSearchView.vue` (top half) | New hero with cassette + "Connect Spotify". See Routing note. |
| Artist search + results | `views/ArtistSearchView.vue` (search + list) | Restyle search input, GO button, result cards. |
| Setlist preview & edit | `views/SetlistPreviewView.vue` | Restyle: show selector → "pick a show" tabs; track list → lined-paper list; manual add; short-warning; "Make the tape" CTA. |
| Track row | `components/TrackList.vue` | Rebuild row as the lined-paper handwritten row (number, Caveat title, duration, ✗/↩ toggle). |
| Playlist result | `views/PlaylistResultView.vue` | Finished-cassette result, "Open in Spotify", unmatched-tracks panel. |
| Festival mode (search + lineup + options) | `views/FestivalSearchView.vue` / `views/FestivalModeView.vue` | Restyle search, lineup checklist with include toggles, add-artist, merged vs per-artist, CTA. |
| Festival result | `views/FestivalResultView.vue` | Merged cassette OR per-artist tape list, skipped-artist panel. |

Keep all existing imports from `src/api/client.ts` (`searchArtists`, `getSetlists`,
`createArtistPlaylist`, `searchEvents`, `createFestivalPlaylist`, `toSpotifyAppURI`, session
helpers) and the existing `vue-router` navigation. Only the templates and `<style>` change.

### Routing note (one real decision)
Today `ArtistSearchView.vue` combines **connect** and **search** on `/`. The prototype
splits them into a **landing** screen (hero + Connect Spotify + "festival mode" link) and a
**search** screen. Two options — pick one:
- **A (recommended):** add a `/` landing route and move search to `/search`. Cleaner, matches
  the prototype's `01 / SEARCH` step tag and back-button flow.
- **B (smaller diff):** keep `/` combined and just restyle, dropping the explicit landing.
Either is fine; the visuals are identical per section.

---

## Design Tokens

### Color
```
--desk:        #211c17   /* app backdrop behind the paper; striped w/ #241f19 / #1f1a15 */
--paper:       #f4ecd8   /* main sheet */
--paper-inset: #fbf6e9   /* inputs, result cards, list items */
--ink:         #1c1813   /* near-black: cassette body, buttons, headlines */
--ink-text:    #231f1a   /* body text on paper */
--label-cream: #efe7d5   /* cassette label face */
--accent:      #b5402a   /* MARKER RED — primary accent, themeable */
--hand-brown:  #7a4a3a   /* handwritten sub-headlines */
--muted:       #9a8f78   /* meta / attribution */
--muted-2:     #8a8073   /* secondary meta */
--muted-3:     #6f665b   /* footer counts */
--tape:        rgba(214,196,140,.6)   /* masking-tape strips + title highlight */
--line:        #e7d9b4   /* notebook rule line */
/* borders */  #cdbb8f #d8c79f #ddcca3 #d3c29a #cabd9e
/* warning */  bg #f3e2c2  border #c79a4a (dashed)  text #8a5a1a
/* spotify  */ green #1DB954  on-green text #06371a  focus ring #14401f
/* halftone */ radial-gradient(#0000000d 1px, transparent 1.4px); background-size:5px 5px;
```
`--accent` is the only themeable token. Default `#b5402a`. Everything that reads
`var(--accent, #b5402a)` (button shadows, track numbers, links, "SIDE A") should pull from it.

### Typography (Google Fonts)
```html
<link href="https://fonts.googleapis.com/css2?family=Archivo:wght@600;700;800;900&family=Archivo+Black&family=Permanent+Marker&family=Caveat:wght@500;600;700&family=Courier+Prime:wght@400;700&display=swap" rel="stylesheet">
```
| Role | Family | Notes |
|---|---|---|
| Display headlines, all buttons, stamped labels | **Archivo Black** | letter-spacing −.02em on big headers |
| Artist names in lists, structured labels | **Archivo** 700–800 | |
| Marker titles, cassette labels, track numbers | **Permanent Marker** | |
| Handwritten sub-copy, tracklist titles | **Caveat** 600–700 | |
| Typewriter meta, attribution, inputs, step tags | **Courier Prime** 400/700 | |

Type scale (px): H1 38 / marker section title 26–28 / tracklist title 22 (Caveat) /
sub-headline 20–22 (Caveat) / button 14–19 / meta + attribution 9–11 (Courier).

### Form & shape language
- **Stamped button:** `background:#1c1813; color:#f4ecd8; font-family:'Archivo Black'; box-shadow:5px 5px 0 var(--accent); transform:rotate(-1deg);` Active state: `transform: rotate(-1deg) translate(2px,2px); box-shadow:3px 3px 0 var(--accent);`
- **Spotify button:** same shape but `background:#1DB954; color:#06371a; box-shadow:5px 5px 0 #06371a;`
- **Inputs:** `background:#fbf6e9; border:1.5px solid #cdbb8f; border-radius:2px; font-family:'Courier Prime';` focus → `border-color:var(--accent)`.
- **Hand-placed feel:** small rotations (−0.4° to −1.6°) on tape strips, labels, buttons, cassettes. Keep subtle.
- Border-radius is near-0 everywhere EXCEPT the cassette body (11px) and its label (3px).

### Key CSS recipes (lift these verbatim)
**Cassette reel** (two per cassette, spin at different speeds):
```css
/* outer reel */
width:44px; height:44px; border-radius:50%;
background:radial-gradient(circle,#efe7d5 0 7px,#bcae8e 7px 18px,#efe7d5 18px 20px,transparent 20px);
display:flex; align-items:center; justify-content:center;
animation:spin 7s linear infinite;   /* second reel: 4.5s for parallax */
/* inner hub (spokes) */
width:15px; height:15px; border-radius:50%;
background:conic-gradient(#1c1813 0 30deg,#efe7d5 30deg 60deg,#1c1813 60deg 90deg,#efe7d5 90deg 120deg,#1c1813 120deg 150deg,#efe7d5 150deg 180deg,#1c1813 180deg 210deg,#efe7d5 210deg 240deg,#1c1813 240deg 270deg,#efe7d5 270deg 300deg,#1c1813 300deg 330deg,#efe7d5 330deg 360deg);
@keyframes spin{to{transform:rotate(360deg)}}
```
**Lined notebook paper** (tracklist background):
```css
background:repeating-linear-gradient(#fbf6e9 0 33px,#e7d9b4 33px 34px);
border:1px solid #d8c79f; box-shadow:inset 0 1px 3px rgba(0,0,0,.05);
/* each track row is height:34px to sit on a rule line */
```
**Halftone overlay** (absolutely positioned, `pointer-events:none`, over the paper):
```css
position:absolute; inset:0; opacity:.5;
background-image:radial-gradient(#0000000d 1px, transparent 1.4px); background-size:5px 5px;
```
**Masking-tape strip / title highlight:** `background:rgba(214,196,140,.6); padding:8px 16px; transform:rotate(-1deg); width:fit-content;`

### Layout
- App is a **single centered column**, `max-width:468px`, full-height `#f4ecd8` sheet on the
  `#211c17` desk, `box-shadow:0 0 60px -10px rgba(0,0,0,.6)`. On mobile it fills the width.
  This is already responsive — no breakpoints needed beyond the max-width clamp.
- Content padding: `18px 22px 40px`. Top bar padding `16px 22px 10px`.
- Persistent **top bar**: left `← back` (Caveat, `#7a4a3a`, hidden on landing), center
  `GIGTAPE` wordmark (Archivo Black), right step tag (Courier, e.g. `02 / EDIT`).
  Under it a dashed rule: `repeating-linear-gradient(90deg,#c9bd9f 0 7px,transparent 7px 12px)`.

---

## Screens / Views (detail)

### 1. Landing / Connect
- Tilted cassette hero (284px) with marker "Gigtape" + "SIDE A" and two spinning reels.
- H1 (Archivo Black 38px): "Make the tape\nbefore the show."
- Sub (Caveat 22px, `#7a4a3a`): "Turn any band's live setlist into a Spotify playlist — a mixtape for the gig you're about to see. ♪"
- Stamped "Connect Spotify" button with a green dot → triggers existing OAuth (`getAuthUrl` → redirect).
- Text link (Caveat, underlined accent): "Going to a festival? Make the whole lineup →" → festival route.
- Footer (Courier 10px): "NO ACCOUNT STORED · NOTHING SAVED · IN & OUT" / "SETLIST DATA PROVIDED BY SETLIST.FM".

### 2. Artist search
- Marker title on tape: "Find a band". Sub (Caveat): "Who are you about to go see?"
- Row: Courier input (`placeholder="artist name…"`) + stamped "GO" button. Submit on Enter or GO → `searchArtists(query)`.
- Results: `{n} RESULTS` (Courier) then cards. Each card: `#fbf6e9`, `border-left:5px solid var(--accent)`, Archivo 800 name, Courier disambiguation, accent "→" on the right. Hover: `#fff9ec` + nudge right 2px. Click → setlist route with the artist ref/name (existing `pick()` logic).

### 3. Setlist preview & edit  *(the core screen)*
- Marker title on tape = artist name. Sub (Caveat): "songs they've been playing live ♪".
- "PICK A SHOW" (Courier) then a wrap of **show tabs** — one per recent setlist. Selected tab: `#1c1813` bg, cream text, `3px 3px 0 var(--accent)` shadow. Unselected: `#fbf6e9`. Each tab shows event name (Archivo 700, 12px) + `date · location` (Courier 10px). Selecting a tab loads that setlist's tracks. Caption: "via setlist.fm · {n} recent shows".
- **Short-setlist warning** (only when the API's `short_warning` is true): dashed `#c79a4a` box, ⚠ + Caveat text "Only {n} songs logged for this show — the setlist may be incomplete. Add any you remember."
- **Lined tracklist** (see CSS recipe). Each row (height 34px): Permanent-Marker number in accent (`01.`), Caveat 22px title (ellipsis on overflow), Courier duration, and a toggle glyph. Active track shows red **✗** (muted color) to remove; a removed track renders with `—` number, strike-through, `#b6a98a` title, and an accent **↩** to restore. **Numbering counts only kept tracks** and re-flows when you remove/restore (see prototype `visibleTracks`).
- **Manual add** row inside the list: borderless Caveat input "+ add a song you remember…" + accent "ADD". Adds a track (duration shows "—"). Existing `addManualTrack` already does this — keep it.
- Footer line (Courier): "{keptCount} SONGS · {mm:ss total}" left, "SIDE A" right. Duration total = sum of kept track durations.
- Stamped "MAKE THE TAPE →" CTA → existing `createArtistPlaylist(...)` → result route. Attribution caption below.

### 4. Playlist result
- Marker eyebrow "★ your tape is ready ★".
- Finished **cassette** (tilted +1.3°) with the artist name + "live · setlist mix" and reels.
- Caveat 24px: "{matchedCount} songs added to Spotify."
- Spotify-green "▸ Open in Spotify" button → `toSpotifyAppURI(playlist_url)` with browser-link fallback (existing logic).
- **Unmatched panel** (only when `unmatched_tracks.length`): dashed warning box, Courier eyebrow "COULDN'T FIND THESE ON SPOTIFY —" then each title as a Caveat "· {title}" line. This honesty is required — never hide unmatched tracks.
- Caveat 21px: "Now go. Pass it on. ♪" then "make another tape" link → home.

### 5. Festival mode
- Inverted marker title (cream text on `#1c1813` block): "Festival mode". Sub: "One big tape for the whole lineup."
- Courier input "festival or venue…" + GO → existing `searchEvents(query)`.
- On result: event name (Archivo 800) + "{date} · {location} · {included} of {total} artists in" (Courier).
- **Lineup checklist** — one row per artist: a 22px square checkbox (filled `#1c1813` with cream ✓ when included, empty when not), Archivo 700 name, and right-aligned Courier meta = "{count} songs" / "added" / accent "no setlist yet" for artists lacking setlist data. Click toggles include; excluded rows dim to opacity .5. (Mirrors existing `FestivalArtistEntry.include`.)
- **Add artist** row: dashed Caveat input "+ add an artist they missed…" + accent "ADD".
- **"HOW DO YOU WANT IT?"** — two selectable cards: "ONE BIG TAPE / everyone, merged" (mode `merged`) and "A TAPE EACH / one per artist" (mode `per_artist`). Selected card: `#1c1813` + `4px 4px 0 var(--accent)`.
- Stamped CTA whose label depends on mode: `MAKE ONE BIG TAPE →` or `MAKE {n} TAPES →` → existing `createFestivalPlaylist({mode, artists})`. Caption: "LINEUPS VIA SETLIST.FM · MAY BE PARTIAL".

### 6. Festival result
- Eyebrow "★ your festival tape is ready ★" (merged) or "★ your tapes are ready ★" (per-artist).
- **Merged:** one cassette card with event name + "{tracks} songs · {artists} artists" and a green "Open in Spotify".
- **Per-artist:** a list of tapes, one per artist — `#fbf6e9` row, `border-left:5px solid #1DB954`, accent "▸", Archivo 800 name, Courier "{count} songs", each links to its playlist.
- **Skipped panel** (when any included artist had no setlist data): dashed box "SKIPPED — NO SETLIST DATA YET —" + Caveat "· {name}" lines. (Maps to the API's `skipped_artists`.) Required honesty, same as unmatched tracks.
- "start over" link → home.

---

## Interactions & Behavior
- **Navigation:** screen-to-screen via `vue-router` (existing). Back button targets:
  search→landing, setlist→search, result→landing, festival→landing, festResult→festival.
- **Reels** spin continuously (`@keyframes spin`, two speeds for parallax). Cosmetic only.
- **Buttons** press in on `:active` (translate 2px into the shadow). No other hover choreography
  except result/search cards' subtle background lift.
- **Inputs** submit on Enter as well as button click.
- **NOTE:** do not re-add a CSS entrance/`fadeup` animation gated on `opacity:0` — it caused
  stuck-invisible content in the prototype runtime. If you want an intro fade, drive it from
  mounted state, not a fill-mode animation.

## State Management
All real state already lives in the existing Vue views and `src/api/client.ts` (session id,
search results, setlists, selected setlist index, editable tracks, playlist result, festival
event + entries + mode). **Reuse it.** The only *derived* UI state to add:
- kept-track renumbering + total duration (compute from the existing tracks array),
- festival CTA label from selected mode + included count,
- which "show tab" is active (selected setlist index — already exists).

## Assets
- **Fonts:** Google Fonts (link above). If you prefer to self-host for the nginx build, add the
  five families to the build and drop the remote `<link>`.
- **No images.** The cassette, reels, halftone, tape, and lined paper are all pure CSS.
- **Icons/glyphs** are Unicode (✗ ↩ ✓ ▸ → ♪ ★ ⚠) — no icon library needed.
- setlist.fm attribution text must remain wherever setlist data is shown (already a project rule).

## Files
- `Gigtape.prototype.html` — the interactive design reference (open in a browser to click through
  all six screens; mock data only).
- `gigtape.css` — **production-ready** design system: all tokens as CSS variables + reusable
  `.gt-*` component classes (shell, top bar, buttons, inputs, cassette + spinning reel, show
  tabs, lined tracklist, festival lineup rows, mode cards, result cards, honesty panels).
  Drop it at `src/styles/gigtape.css` and `import './styles/gigtape.css'` once in `main.ts`.
  Start every view from these classes instead of re-deriving inline styles; override the accent
  per-element with `style="--gt-accent:#…"`. See the comment header for markup examples
  (e.g. the reel is `<div class="gt-reel"><div class="gt-reel__hub"></div></div>`).

### Suggested build order for Claude Code
1. Add `gigtape.css`, import in `main.ts`, wrap the app in `.gt-desk > .gt-sheet` (+ `.gt-grain`).
2. Restyle `SetlistPreviewView.vue` + `TrackList.vue` (the core screen) — verify, then continue.
3. `ArtistSearchView.vue` (+ landing split if using option A), `PlaylistResultView.vue`.
4. Festival views, then the top bar / back-button wiring across all routes.

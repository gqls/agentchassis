// FILE: main.go
//
// HTTP shell for the paired-provocation prototype. Run it, open the page,
// feel the shape. All the interesting rules live in paired.go; this file is
// deliberately dumb, because a handler that has to remember not to leak is a
// handler that eventually forgets.
//
//	go run .            # then open http://localhost:8099
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var store = NewStore()

func main() {
	addr := flag.String("addr", "localhost:8099", "listen address")
	flag.Parse()

	base := "http://" + *addr
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handleNew)
	mux.HandleFunc("POST /create", func(w http.ResponseWriter, r *http.Request) { handleCreate(w, r) })
	mux.HandleFunc("GET /o/{id}", func(w http.ResponseWriter, r *http.Request) { handleOrganiser(w, r, base) })
	mux.HandleFunc("POST /o/{id}/reveal", handleForceReveal)
	mux.HandleFunc("GET /p/{token}", handleParticipant)
	mux.HandleFunc("POST /p/{token}", handleCommit)

	log.Printf("paired-provocation prototype on %s", base)
	log.Printf("IN-MEMORY ONLY — restarting the process destroys every session")
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	render(w, tplNew, nil)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	names := strings.FieldsFunc(r.FormValue("participants"), func(c rune) bool {
		return c == ',' || c == '\n' || c == '\r'
	})
	rule := RevealRule(r.FormValue("rule"))
	quorum, _ := strconv.Atoi(r.FormValue("quorum"))

	deadline := time.Time{}
	if mins, _ := strconv.Atoi(r.FormValue("deadline_minutes")); mins > 0 {
		deadline = time.Now().Add(time.Duration(mins) * time.Minute)
	}

	s, err := NewSession(r.FormValue("organiser"), r.FormValue("provocation"), names, rule, quorum, deadline, time.Now())
	if err != nil {
		render(w, tplNew, map[string]any{"Error": err.Error()})
		return
	}
	store.Put(s)
	http.Redirect(w, r, "/o/"+s.ID, http.StatusSeeOther)
}

func handleOrganiser(w http.ResponseWriter, r *http.Request, base string) {
	s, ok := store.BySession(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	render(w, tplOrganiser, s.OrganiserView(base, time.Now()))
}

func handleForceReveal(w http.ResponseWriter, r *http.Request) {
	s, ok := store.BySession(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.ForceReveal(time.Now())
	http.Redirect(w, r, "/o/"+s.ID, http.StatusSeeOther)
}

func handleParticipant(w http.ResponseWriter, r *http.Request) {
	tok := r.PathValue("token")
	s, ok := store.ByToken(tok)
	if !ok {
		http.NotFound(w, r)
		return
	}
	sealed, revealed, err := s.ViewFor(tok, time.Now())
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	// Two templates, because there are two view types. The sealed template
	// could not render a peer's position even if it tried — SealedView has
	// no field holding one.
	if revealed != nil {
		render(w, tplRevealed, revealed)
		return
	}
	render(w, tplSealed, map[string]any{"V": sealed, "Token": tok})
}

func handleCommit(w http.ResponseWriter, r *http.Request) {
	tok := r.PathValue("token")
	s, ok := store.ByToken(tok)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := s.Commit(tok, r.FormValue("position"), time.Now()); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/p/"+tok, http.StatusSeeOther)
}

func render(w http.ResponseWriter, t *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		log.Printf("template: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

const css = `<style>
:root{color-scheme:dark}
body{background:#0e1116;color:#e6e6e6;font:16px/1.6 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;
     max-width:44rem;margin:0 auto;padding:2.5rem 1.25rem}
h1{font-size:1.6rem;line-height:1.25;margin:0 0 .3rem}
h2{font-size:1.05rem;letter-spacing:.04em;text-transform:uppercase;color:#8b95a5;margin:2.2rem 0 .8rem}
.eyebrow{letter-spacing:.12em;text-transform:uppercase;font-size:.72rem;color:#7c8899}
.prov{background:#161b22;border-left:3px solid #d24b3e;padding:1rem 1.2rem;margin:1.2rem 0;font-size:1.15rem}
label{display:block;margin:1rem 0 .3rem;font-size:.85rem;color:#9aa5b4}
input,textarea,select{width:100%;padding:.6rem .7rem;background:#161b22;color:#e6e6e6;
     border:1px solid #2b323c;border-radius:4px;font:inherit}
textarea{min-height:8rem;resize:vertical}
button{margin-top:1.2rem;padding:.7rem 1.6rem;background:#d24b3e;color:#fff;border:0;border-radius:4px;
     font:inherit;font-weight:600;cursor:pointer}
button.ghost{background:#2b323c}
table{width:100%;border-collapse:collapse;margin-top:.5rem}
td,th{text-align:left;padding:.55rem .4rem;border-bottom:1px solid #232a33;font-size:.92rem}
th{color:#7c8899;font-weight:500;font-size:.75rem;text-transform:uppercase;letter-spacing:.06em}
a{color:#7fb2ff;word-break:break-all}
.pill{display:inline-block;padding:.12rem .55rem;border-radius:999px;font-size:.72rem;font-weight:600}
.yes{background:#1d3b2a;color:#6ee7a0}.no{background:#3a2320;color:#f0a08e}
.note{color:#7c8899;font-size:.85rem;margin-top:2rem;border-top:1px solid #232a33;padding-top:1rem}
.err{background:#3a2320;color:#f0a08e;padding:.7rem 1rem;border-radius:4px;margin:1rem 0}
.pos{background:#161b22;border:1px solid #232a33;border-radius:6px;padding:1rem 1.2rem;margin:.8rem 0}
.pos .who{font-weight:700;color:#fff;margin-bottom:.35rem}
</style>`

var tplNew = template.Must(template.New("new").Parse(css + `
<h1>Set a paired provocation</h1>
<p class="eyebrow">Prototype · in memory · nothing is saved</p>
{{with .Error}}<div class="err">{{.}}</div>{{end}}
<form method="post" action="/create">
  <label>Your name (the organiser)</label>
  <input name="organiser" value="Fran" required>
  <label>The provocation — you write this one, not the machine</label>
  <textarea name="provocation" required>Remote work killed mentorship. You cannot absorb judgement over a video call.</textarea>
  <label>Participants (comma or newline separated)</label>
  <textarea name="participants" style="min-height:5rem" required>Alice, Bob, Carol</textarea>
  <label>When do the positions open?</label>
  <select name="rule">
    <option value="all_committed">When everyone has committed</option>
    <option value="quorum">As soon as N have committed</option>
    <option value="deadline">At a deadline, whoever has answered</option>
  </select>
  <label>N (quorum rule only)</label>
  <input name="quorum" type="number" value="2" min="1">
  <label>Deadline in minutes from now (deadline rule only)</label>
  <input name="deadline_minutes" type="number" value="5" min="1">
  <button type="submit">Create and get the links</button>
</form>
<p class="note">Nobody — including you — can read a position before the reveal.
That is enforced by the type system, not by a check: see <code>paired.go</code>.</p>
`))

var tplOrganiser = template.Must(template.New("org").Parse(css + `
<p class="eyebrow">Organiser view · session {{.ID}}</p>
<h1>{{.Organiser}}'s paired provocation</h1>
<div class="prov">{{.Provocation}}</div>
<p>{{.Committed}} of {{.Total}} committed ·
{{if .Revealed}}<span class="pill yes">OPEN</span>{{else}}<span class="pill no">SEALED</span>{{end}}
· rule: <code>{{.Rule}}</code>{{if eq (printf "%s" .Rule) "quorum"}} (N={{.Quorum}}){{end}}</p>
<h2>Send one link to each person</h2>
<table>
<tr><th>Who</th><th>Answered</th><th>Their private link</th></tr>
{{range .Rows}}<tr>
  <td>{{.Name}}</td>
  <td>{{if .Committed}}<span class="pill yes">yes</span>{{else}}<span class="pill no">not yet</span>{{end}}</td>
  <td><a href="{{.Link}}">{{.Link}}</a></td>
</tr>{{end}}
</table>
{{if not .Revealed}}
<form method="post" action="/o/{{.ID}}/reveal">
  <button class="ghost" type="submit">Force the reveal now</button>
</form>
<p class="note"><strong>You cannot see anyone's position on this page, and that is deliberate.</strong>
A facilitator who has read the answers cannot run the session neutrally, and participants who
suspect you have will hedge. You get who has answered, so you know who to chase — nothing more.</p>
{{else}}
<p class="note">Positions are open. They are visible to the participants who committed —
not to you on this page, and not publicly anywhere.</p>
{{end}}
`))

var tplSealed = template.Must(template.New("sealed").Parse(css + `
<p class="eyebrow">{{.V.Organiser}} asked you · sealed</p>
<h1>{{if .V.YouCommitted}}Locked in. Waiting for the others.{{else}}What's your position?{{end}}</h1>
<div class="prov">{{.V.Provocation}}</div>
{{if .V.YouCommitted}}
  <h2>Your position</h2>
  <div class="pos"><div class="who">{{.V.YourName}} (you)</div>{{.V.YourPosition}}</div>
{{else}}
  <form method="post" action="/p/{{.Token}}">
    <label>Commit your position — you cannot change it afterwards, and nobody sees it until the reveal</label>
    <textarea name="position" required></textarea>
    <button type="submit">Commit</button>
  </form>
{{end}}
<h2>The room · {{.V.Committed}} of {{.V.Total}} committed</h2>
<table>
<tr><th>Who</th><th>Answered</th></tr>
{{range .V.Peers}}<tr><td>{{.Name}}</td>
<td>{{if .Committed}}<span class="pill yes">yes</span>{{else}}<span class="pill no">not yet</span>{{end}}</td></tr>{{end}}
</table>
<p class="note">You can see <em>that</em> they have answered and not <em>what</em> they answered.
This page is built from a <code>SealedView</code>, a type with nowhere to put another person's words —
so it could not leak one even if this template asked it to.</p>
`))

var tplRevealed = template.Must(template.New("revealed").Parse(css + `
<p class="eyebrow">{{.Organiser}} asked you · open {{.RevealedAt.Format "15:04"}}</p>
<h1>Everyone's position, all at once</h1>
<div class="prov">{{.Provocation}}</div>
{{range .Positions}}
<div class="pos"><div class="who">{{.Name}}</div>{{.Position}}</div>
{{end}}
{{with .DidNotCommit}}
<h2>Did not answer</h2>
<p>{{range $i, $n := .}}{{if $i}}, {{end}}{{$n}}{{end}}</p>
{{end}}
<p class="note">Private to the people who committed. There is no public listing of this
session anywhere, and anyone who stayed silent does not get to read the room.</p>
`))

var _ = fmt.Sprintf

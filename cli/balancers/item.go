package balancers

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

// Item is a balancer card with the organization and datacenter names
// resolved.
type Item struct {
	balnc      *balancer.Balancer
	names      *names
	org        string
	datacenter string
}

func (i *Item) Balancer() *balancer.Balancer {
	return i.balnc
}

func (i *Item) Id() string {
	return i.balnc.Id.Hex()
}

func (i *Item) Name() string {
	return i.balnc.Name
}

func (i *Item) Tag() string {
	return i.balnc.Id.Hex()
}

func (i *Item) domains() string {
	domains := []string{}
	for _, domn := range i.balnc.Domains {
		domains = append(domains, domn.Domain)
	}
	return strings.Join(domains, ", ")
}

// stateColor colors the active state green and inactive red.
func stateColor(active bool) color.Color {
	if active {
		return widget.ColorGreen
	}
	return widget.ColorRed
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Organization",
			Value: widget.Default(i.org, resource.IdHex(i.balnc.Organization)),
		},
		{
			Label: "Datacenter",
			Value: widget.Default(i.datacenter,
				resource.IdHex(i.balnc.Datacenter)),
		},
		{
			Label: "State",
			Value: widget.Bool(i.balnc.State, "Active", "Inactive"),
			Color: stateColor(i.balnc.State),
		},
		{
			Label: "Domains",
			Value: i.domains(),
		},
	}
}

// backendStates merges the backend states reported by every node the
// same way as the detailed balancer view, the worst state of a backend
// wins, and returns the totals with one line per backend.
func (i *Item) backendStates() (requests, retries, websockets int,
	lines []string) {

	if !i.balnc.State || i.balnc.States == nil {
		return
	}

	const (
		offline = iota + 1
		unknownLow
		unknownMid
		unknownHigh
		online
	)
	labels := map[int]string{
		online:      "Online",
		unknownHigh: "Unknown High",
		unknownMid:  "Unknown Mid",
		unknownLow:  "Unknown Low",
		offline:     "Offline",
	}

	states := map[string]int{}
	mark := func(backends []string, state int) {
		for _, backend := range backends {
			cur, ok := states[backend]
			if !ok || cur > state {
				states[backend] = state
			}
		}
	}

	for _, state := range i.balnc.States {
		if state == nil {
			continue
		}
		requests += state.Requests
		retries += state.Retries
		websockets += state.WebSockets
		mark(state.Offline, offline)
		mark(state.UnknownLow, unknownLow)
		mark(state.UnknownMid, unknownMid)
		mark(state.UnknownHigh, unknownHigh)
		mark(state.Online, online)
	}

	for _, state := range []int{online, unknownHigh, unknownMid,
		unknownLow, offline} {

		backends := []string{}
		for backend, cur := range states {
			if cur == state {
				backends = append(backends, backend)
			}
		}
		sort.Strings(backends)

		for _, backend := range backends {
			lines = append(lines, backend+" - "+labels[state])
		}
	}

	return
}

// Info mirrors the fields of the detailed balancer view.
func (i *Item) Info() []widget.InfoField {
	requests, retries, websockets, states := i.backendStates()

	return []widget.InfoField{
		{"ID", i.balnc.Id.Hex()},
		{"Requests", fmt.Sprintf("%d/min", requests)},
		{"Retries", fmt.Sprintf("%d/min", retries)},
		{"WebSockets", fmt.Sprintf("%d", websockets)},
		{"Backends", strings.Join(states, ", ")},
	}
}

// Actions mirrors the delete button of the detailed view.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("balancer", i.balnc.Name,
			resource.Delete("balancer", "balancer.change", i.balnc.Id, balancer.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.balnc, i.names)
}

// Package todo keeps track of a two layer todo list
package todo

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/leep-frog/cli/color"
	"github.com/leep-frog/cli/commands"
)

const (
	cacheKey = "todo-list.json"
)

type List struct {
	Items map[string]map[string]bool

	PrimaryFormats map[string]*color.Format

	changed bool
}

func (tl *List) Load(jsn string) error {
	if jsn == "" {
		tl = &List{}
		return nil
	}

	if err := json.Unmarshal([]byte(jsn), tl); err != nil {
		return fmt.Errorf("failed to unmarshal todo list json: %v", err)
	}
	return nil
}

func (tl *List) ListItems(cos commands.CommandOS, _, _ map[string]*commands.Value) (*commands.ExecutorResponse, bool) {
	ps := make([]string, 0, len(tl.Items))
	count := 0
	for k, v := range tl.Items {
		ps = append(ps, k)
		count += len(v)
	}
	sort.Strings(ps)

	for _, p := range ps {
		f := tl.PrimaryFormats[p]
		cos.Stdout(f.Format(p))
		ss := make([]string, 0, len(tl.Items[p]))
		for s := range tl.Items[p] {
			ss = append(ss, s)
		}
		sort.Strings(ss)
		for _, s := range ss {
			cos.Stdout(fmt.Sprintf("  %s", s))
		}
	}

	return &commands.ExecutorResponse{}, true
}

// TODO: can this just be a generic feature in color package?
func (tl *List) FormatPrimary(cos commands.CommandOS, args, flags map[string]*commands.Value) (*commands.ExecutorResponse, bool) {
	primary := *args["primary"].String()
	codes := *args["format"].StringList()

	if tl.PrimaryFormats == nil {
		tl.PrimaryFormats = map[string]*color.Format{}
	}
	f, ok := tl.PrimaryFormats[primary]
	if !ok {
		f = &color.Format{}
		tl.PrimaryFormats[primary] = f
	}
	for _, c := range codes {
		if err := f.AddAttribute(c); err != nil {
			cos.Stderr("error adding todo list attribute: %v", err)
			return nil, false
		}
	}
	tl.changed = true

	return &commands.ExecutorResponse{}, true
}

func (tl *List) Changed() bool {
	return tl.changed
}

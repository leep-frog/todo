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

func (tl *List) ListItems(_, _ map[string]*commands.Value) (*commands.ExecutorResponse, error) {
	ps := make([]string, 0, len(tl.Items))
	count := 0
	for k, v := range tl.Items {
		ps = append(ps, k)
		count += len(v)
	}
	sort.Strings(ps)
	r := make([]string, 0, count)
	for _, p := range ps {
		f := tl.PrimaryFormats[p]
		r = append(r, f.Format(p))
		ss := make([]string, 0, len(tl.Items[p]))
		for s := range tl.Items[p] {
			ss = append(ss, s)
		}
		sort.Strings(ss)
		for _, s := range ss {
			r = append(r, fmt.Sprintf("  %s", s))
		}
	}

	resp := &commands.ExecutorResponse{}
	if len(r) > 0 {
		resp.Stdout = r
	}
	return resp, nil
}

// TODO: can this just be a generic feature in color package?
func (tl *List) FormatPrimary(args, flags map[string]*commands.Value) (*commands.ExecutorResponse, error) {
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
			return &commands.ExecutorResponse{
				Stderr: []string{
					fmt.Sprintf("error adding todo list attribute: %v", err),
				},
			}, nil
		}
	}
	tl.changed = true

	return &commands.ExecutorResponse{}, nil
}

func (tl *List) Changed() bool {
	return tl.changed
}

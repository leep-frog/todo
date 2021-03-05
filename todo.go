// Package todo keeps track of a two layer todo list
package todo

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/leep-frog/commands/color"
	"github.com/leep-frog/commands/commands"
)

const (
	cacheKey = "todo-list.json"
)

type List struct {
	Items map[string]map[string]bool

	PrimaryFormats map[string]*color.Format

	changed bool
}

func (tl *List) Option() *commands.Option { return nil }

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

func (tl *List) ListItems(cos commands.CommandOS, _, _ map[string]*commands.Value, _ *commands.OptionInfo) (*commands.ExecutorResponse, bool) {
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

	return nil, true
}

func (tl *List) FormatPrimary(cos commands.CommandOS, args, flags map[string]*commands.Value, _ *commands.OptionInfo) (*commands.ExecutorResponse, bool) {
	primary := args[primaryArg].GetString_()

	if tl.PrimaryFormats == nil {
		tl.PrimaryFormats = map[string]*color.Format{}
	}

	tl.PrimaryFormats[primary], tl.changed = color.ApplyCodes(tl.PrimaryFormats[primary], args)
	return nil, true
}

func (tl *List) Changed() bool {
	return tl.changed
}

// Package todo keeps track of a two layer todo list
package todo

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/leep-frog/command"
	"github.com/leep-frog/command/color"
)

func CLI() *List {
	return &List{}
}

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

func (tl *List) ListItems(output command.Output, data *command.Data) error {
	ps := make([]string, 0, len(tl.Items))
	count := 0
	for k, v := range tl.Items {
		ps = append(ps, k)
		count += len(v)
	}
	sort.Strings(ps)

	for _, p := range ps {
		f := tl.PrimaryFormats[p]
		output.Stdout(f.Format(p))
		ss := make([]string, 0, len(tl.Items[p]))
		for s := range tl.Items[p] {
			ss = append(ss, s)
		}
		sort.Strings(ss)
		for _, s := range ss {
			output.Stdout(fmt.Sprintf("  %s", s))
		}
	}

	return nil
}

func (tl *List) FormatPrimary(output command.Output, data *command.Data) error {
	primary := data.Values[primaryArg].String()

	if tl.PrimaryFormats == nil {
		tl.PrimaryFormats = map[string]*color.Format{}
	}

	var err error
	if tl.PrimaryFormats[primary], err = color.ApplyCodes(tl.PrimaryFormats[primary], output, data); err != nil {
		return err
	}
	tl.changed = true
	return nil
}

func (tl *List) Setup() []string { return nil }

func (tl *List) Changed() bool {
	return tl.changed
}

package todo

import (
	"github.com/leep-frog/command"
	"github.com/leep-frog/command/color"
)

const (
	primaryArg   = "primary"
	secondaryArg = "secondary"
)

func (tl *List) AddItem(output command.Output, data *command.Data) error {
	if tl.Items == nil {
		tl.Items = map[string]map[string]bool{}
		tl.changed = true
	}

	p := data.Values[primaryArg].String()
	if _, ok := tl.Items[p]; !ok {
		tl.Items[p] = map[string]bool{}
		tl.changed = true
	}

	if data.Values[secondaryArg].Provided() {
		s := data.Values[secondaryArg].String()
		if tl.Items[p][s] {
			return output.Stderr("item %q, %q already exists", p, s)
		}
		tl.Items[p][s] = true
		tl.changed = true
	} else if !tl.changed {
		return output.Stderr("primary item %q already exists", p)
	}
	return nil
}

func (tl *List) DeleteItem(output command.Output, data *command.Data) error {
	if tl.Items == nil {
		return output.Stderr("can't delete from empty list")
	}

	p := data.Values[primaryArg].String()
	if _, ok := tl.Items[p]; !ok {
		return output.Stderr("Primary item %q does not exist", p)
	}

	// Delete secondary if provided
	if data.Values[secondaryArg].Provided() {
		s := data.Values[secondaryArg].String()
		if tl.Items[p][s] {
			delete(tl.Items[p], s)
			tl.changed = true
			return nil
		} else {
			return output.Stderr("Secondary item %q does not exist", s)
		}
	}

	if len(tl.Items[p]) != 0 {
		return output.Stderr("Can't delete primary item that still has secondary items")
	}

	delete(tl.Items, p)
	tl.changed = true
	return nil
}

// Name returns the name of the CLI.
func (tl *List) Name() string {
	return "td"
}

type fetcher struct {
	List *List
	// Primary is whether or not to complete primary or secondary result.
	Primary bool
}

func (f *fetcher) Fetch(value *command.Value, data *command.Data) *command.Completion {
	if f.Primary {
		primaries := make([]string, 0, len(f.List.Items))
		for p := range f.List.Items {
			primaries = append(primaries, p)
		}
		return &command.Completion{
			Suggestions: primaries,
		}
	}

	p := data.Values[primaryArg].String()
	sMap := f.List.Items[p]
	secondaries := make([]string, 0, len(sMap))
	for s := range sMap {
		secondaries = append(secondaries, s)
	}
	return &command.Completion{
		Suggestions: secondaries,
	}
}

func (tl *List) Node() *command.Node {
	pf := &command.ArgOpt{
		Completor: &command.Completor{
			SuggestionFetcher: &fetcher{
				List:    tl,
				Primary: true,
			},
		},
	}
	sf := &command.ArgOpt{
		Completor: &command.Completor{
			SuggestionFetcher: &fetcher{List: tl},
		},
	}
	return command.BranchNode(
		map[string]*command.Node{
			"a": command.SerialNodes(
				command.StringNode(primaryArg, pf),
				command.OptionalStringNode(secondaryArg, nil),
				command.ExecutorNode(tl.AddItem),
			),
			"d": command.SerialNodes(
				command.StringNode(primaryArg, pf),
				command.OptionalStringNode(secondaryArg, sf),
				command.ExecutorNode(tl.DeleteItem),
			),
			"f": command.SerialNodes(
				command.StringNode(primaryArg, pf),
				color.Arg,
				command.ExecutorNode(tl.FormatPrimary),
			),
		},
		command.SerialNodes(command.ExecutorNode(tl.ListItems)),
		true,
	)
}

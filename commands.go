package todo

import (
	"github.com/leep-frog/command"
	"github.com/leep-frog/command/color"
)

const (
	primaryArg    = "primary"
	primaryDesc   = "Primary list"
	secondaryArg  = "secondary"
	secondaryDesc = "Secondary list"
)

func (tl *List) AddItem(output command.Output, data *command.Data) error {
	if tl.Items == nil {
		tl.Items = map[string]map[string]bool{}
		tl.changed = true
	}

	p := data.String(primaryArg)
	if _, ok := tl.Items[p]; !ok {
		tl.Items[p] = map[string]bool{}
		tl.changed = true
	}

	if data.Has(secondaryArg) {
		s := data.String(secondaryArg)
		if tl.Items[p][s] {
			return output.Stderrf("item %q, %q already exists", p, s)
		}
		tl.Items[p][s] = true
		tl.changed = true
	} else if !tl.changed {
		return output.Stderrf("primary item %q already exists", p)
	}
	return nil
}

func (tl *List) DeleteItem(output command.Output, data *command.Data) error {
	if tl.Items == nil {
		return output.Stderr("can't delete from empty list")
	}

	p := data.String(primaryArg)
	if _, ok := tl.Items[p]; !ok {
		return output.Stderrf("Primary item %q does not exist", p)
	}

	// Delete secondary if provided
	if data.Has(secondaryArg) {
		s := data.String(secondaryArg)
		if tl.Items[p][s] {
			delete(tl.Items[p], s)
			tl.changed = true
			return nil
		} else {
			return output.Stderrf("Secondary item %q does not exist", s)
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

func (f *fetcher) Fetch(value string, data *command.Data) (*command.Completion, error) {
	if f.Primary {
		primaries := make([]string, 0, len(f.List.Items))
		for p := range f.List.Items {
			primaries = append(primaries, p)
		}
		return &command.Completion{
			Suggestions: primaries,
		}, nil
	}

	p := data.String(primaryArg)
	sMap := f.List.Items[p]
	secondaries := make([]string, 0, len(sMap))
	for s := range sMap {
		secondaries = append(secondaries, s)
	}
	return &command.Completion{
		Suggestions: secondaries,
	}, nil
}

func (tl *List) Node() *command.Node {
	pf := &command.Completor[string]{
		SuggestionFetcher: &fetcher{
			List:    tl,
			Primary: true,
		},
	}
	sf := &command.Completor[string]{
		SuggestionFetcher: &fetcher{List: tl},
	}
	return command.BranchNode(
		map[string]*command.Node{
			"a": command.SerialNodes(
				command.Arg[string](primaryArg, primaryDesc, pf),
				command.OptionalArg[string](secondaryArg, secondaryDesc),
				command.ExecuteErrNode(tl.AddItem),
			),
			"d": command.SerialNodes(
				command.Arg[string](primaryArg, primaryDesc, pf),
				command.OptionalArg[string](secondaryArg, secondaryDesc, sf),
				command.ExecuteErrNode(tl.DeleteItem),
			),
			"f": command.SerialNodes(
				command.Arg[string](primaryArg, primaryDesc, pf),
				color.Arg,
				command.ExecuteErrNode(tl.FormatPrimary),
			),
		},
		command.SerialNodes(command.ExecutorNode(tl.ListItems)),
	)
}

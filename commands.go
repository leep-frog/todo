package todo

import (
	"github.com/leep-frog/commands/color"
	"github.com/leep-frog/commands/commands"
)

var (
	primaryArg   = "primary"
	secondaryArg = "secondary"
)

func (tl *List) AddItem(cos commands.CommandOS, args, flag map[string]*commands.Value, _ *commands.OptionInfo) (*commands.ExecutorResponse, bool) {
	if tl.Items == nil {
		tl.Items = map[string]map[string]bool{}
		tl.changed = true
	}

	p := args[primaryArg].String()
	if _, ok := tl.Items[p]; !ok {
		tl.Items[p] = map[string]bool{}
		tl.changed = true
	}

	if args[secondaryArg].Provided() {
		s := args[secondaryArg].String()
		if tl.Items[p][s] {
			cos.Stderr("item %q, %q already exists", p, s)
			return nil, false
		}
		tl.Items[p][s] = true
		tl.changed = true
	} else if !tl.changed {
		cos.Stderr("primary item %q already exists", p)
		return nil, false
	}
	return nil, true
}

func (tl *List) DeleteItem(cos commands.CommandOS, args, flag map[string]*commands.Value, _ *commands.OptionInfo) (*commands.ExecutorResponse, bool) {
	if tl.Items == nil {
		cos.Stderr("can't delete from empty list")
		return nil, false
	}

	p := args[primaryArg].String()
	if _, ok := tl.Items[p]; !ok {
		cos.Stderr("Primary item %q does not exist", p)
		return nil, false
	}

	// Delete secondary if provided
	if args[secondaryArg].Provided() {
		s := args[secondaryArg].String()
		if tl.Items[p][s] {
			delete(tl.Items[p], s)
			tl.changed = true
			return nil, true
		} else {
			cos.Stderr("Secondary item %q does not exist", s)
			return nil, false
		}
	}

	if len(tl.Items[p]) != 0 {
		cos.Stderr("Can't delete primary item that still has secondary items")
		return nil, false
	}

	delete(tl.Items, p)
	tl.changed = true
	return nil, true
}

// Name returns the name of the CLI.
func (tl *List) Name() string {
	return "todo-list"
}

// Alias returns the CLI alias.
func (tl *List) Alias() string {
	return "td"
}

type fetcher struct {
	List *List
	// Primary is whether or not to complete primary or secondary result.
	Primary bool
}

func (f *fetcher) Fetch(_ *commands.Value, args, _ map[string]*commands.Value) *commands.Completion {
	if f.Primary {
		primaries := make([]string, 0, len(f.List.Items))
		for p := range f.List.Items {
			primaries = append(primaries, p)
		}
		return &commands.Completion{
			Suggestions: primaries,
		}
	}

	p := args[primaryArg].String()
	sMap := f.List.Items[p]
	secondaries := make([]string, 0, len(sMap))
	for s := range sMap {
		secondaries = append(secondaries, s)
	}
	return &commands.Completion{
		Suggestions: secondaries,
	}
}

func (tl *List) Command() commands.Command {
	pf := &commands.Completor{
		SuggestionFetcher: &fetcher{
			List:    tl,
			Primary: true,
		},
	}
	sf := &commands.Completor{
		SuggestionFetcher: &fetcher{List: tl},
	}
	return &commands.CommandBranch{
		TerminusCommand: &commands.TerminusCommand{
			Executor: tl.ListItems,
		},
		Subcommands: map[string]commands.Command{
			// Add item
			"a": &commands.TerminusCommand{
				Args: []commands.Arg{
					commands.StringArg(primaryArg, true, pf),
					commands.StringArg(secondaryArg, false, nil),
				},
				Executor: tl.AddItem,
			},
			// Delete item
			"d": &commands.TerminusCommand{
				Args: []commands.Arg{
					commands.StringArg(primaryArg, true, pf),
					commands.StringArg(secondaryArg, false, sf),
				},
				Executor: tl.DeleteItem,
			},
			// Format items
			"f": &commands.TerminusCommand{
				Executor: tl.FormatPrimary,
				Args: []commands.Arg{
					commands.StringArg(primaryArg, true, pf),
					color.Arg,
				},
			},
		},
	}
}

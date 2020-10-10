package todo

import (
	"github.com/leep-frog/cli/commands"
)

func (tl *List) AddItem(cos commands.CommandOS, args, flag map[string]*commands.Value) (*commands.ExecutorResponse, bool) {
	if tl.Items == nil {
		tl.Items = map[string]map[string]bool{}
		tl.changed = true
	}

	p := *args["primary"].String()
	if _, ok := tl.Items[p]; !ok {
		tl.Items[p] = map[string]bool{}
		tl.changed = true
	}

	if s, ok := args["secondary"]; ok {
		if tl.Items[p][*s.String()] {
			cos.Stderr("item %q, %q already exists", p, *s.String())
			return nil, false
		}
		tl.Items[p][*s.String()] = true
		tl.changed = true
	} else if !tl.changed {
		cos.Stderr("primary item %q already exists", p)
		return nil, false
	}
	return &commands.ExecutorResponse{}, true
}

func (tl *List) DeleteItem(cos commands.CommandOS, args, flag map[string]*commands.Value) (*commands.ExecutorResponse, bool) {
	if tl.Items == nil {
		cos.Stderr("can't delete from empty list")
		return nil, false
	}

	p := *args["primary"].String()
	if _, ok := tl.Items[p]; !ok {
		cos.Stderr("Primary item %q does not exist", p)
		return nil, false
	}

	// Delete secondary if provided
	if s, ok := args["secondary"]; ok {
		if tl.Items[p][*s.String()] {
			delete(tl.Items[p], *s.String())
			tl.changed = true
			return &commands.ExecutorResponse{}, true
		} else {
			cos.Stderr("Secondary item %q does not exist", *s.String())
			return nil, false
		}
	}

	if len(tl.Items[p]) != 0 {
		cos.Stderr("Can't delete primary item that still has secondary items")
		return nil, false
	}

	delete(tl.Items, p)
	tl.changed = true
	return &commands.ExecutorResponse{}, true
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

	// TODO: make this string a constant throughout the package (same with secondary)
	p := *args["primary"].String()
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
					commands.StringArg("primary", true, pf),
					commands.StringArg("secondary", false, nil),
				},
				Executor: tl.AddItem,
			},
			// Delete item
			"d": &commands.TerminusCommand{
				Args: []commands.Arg{
					// TODO: make these constants (primaryArgGroup, secondaryArgGroup)
					commands.StringArg("primary", true, pf),
					commands.StringArg("secondary", false, sf),
				},
				Executor: tl.DeleteItem,
			},
			// Format items
			"f": &commands.TerminusCommand{
				Executor: tl.FormatPrimary,
				Args: []commands.Arg{
					commands.StringArg("primary", true, pf),
					// TODO: make color completor.
					commands.StringListArg("format", 1, -1, nil),
				},
			},
		},
	}
}

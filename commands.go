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
			return output.Stderrf("item %q, %q already exists\n", p, s)
		}
		tl.Items[p][s] = true
		tl.changed = true
	} else if !tl.changed {
		return output.Stderrf("primary item %q already exists\n", p)
	}
	return nil
}

func (tl *List) DeleteItem(output command.Output, data *command.Data) error {
	if tl.Items == nil {
		return output.Stderr("can't delete from empty list\n")
	}

	p := data.String(primaryArg)
	if _, ok := tl.Items[p]; !ok {
		return output.Stderrf("Primary item %q does not exist\n", p)
	}

	// Delete secondary if provided
	if data.Has(secondaryArg) {
		s := data.String(secondaryArg)
		if tl.Items[p][s] {
			delete(tl.Items[p], s)
			tl.changed = true
			return nil
		} else {
			return output.Stderrf("Secondary item %q does not exist\n", s)
		}
	}

	if len(tl.Items[p]) != 0 {
		return output.Stderr("Can't delete primary item that still has secondary items\n")
	}

	delete(tl.Items, p)
	tl.changed = true
	return nil
}

// Name returns the name of the CLI.
func (tl *List) Name() string {
	return "td"
}

func completer(l *List, primary bool) command.Completer[string] {
	return command.CompleterFromFunc(func(value string, data *command.Data) (*command.Completion, error) {
		if primary {
			primaries := make([]string, 0, len(l.Items))
			for p := range l.Items {
				primaries = append(primaries, p)
			}
			return &command.Completion{
				Suggestions: primaries,
			}, nil
		}

		p := data.String(primaryArg)
		sMap := l.Items[p]
		secondaries := make([]string, 0, len(sMap))
		for s := range sMap {
			secondaries = append(secondaries, s)
		}
		return &command.Completion{
			Suggestions: secondaries,
		}, nil
	})
}

func (tl *List) Node() command.Node {
	pf := completer(tl, true)
	sf := completer(tl, false)
	return &command.BranchNode{
		Branches: map[string]command.Node{
			"a": command.SerialNodes(
				command.Arg[string](primaryArg, primaryDesc, pf),
				command.OptionalArg[string](secondaryArg, secondaryDesc),
				&command.ExecutorProcessor{F: tl.AddItem},
			),
			"d": command.SerialNodes(
				command.Arg[string](primaryArg, primaryDesc, pf),
				command.OptionalArg[string](secondaryArg, secondaryDesc, sf),
				&command.ExecutorProcessor{F: tl.DeleteItem},
			),
			"f": command.SerialNodes(
				command.Arg[string](primaryArg, primaryDesc, pf),
				color.Arg,
				&command.ExecutorProcessor{F: tl.FormatPrimary},
			),
		},
		Default: command.SerialNodes(&command.ExecutorProcessor{F: tl.ListItems}),
	}
}

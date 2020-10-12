package todo

import (
	"strings"
	"testing"

	"github.com/leep-frog/commands/color"
	"github.com/leep-frog/commands/commands"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLoad(t *testing.T) {
	for _, test := range []struct {
		name    string
		json    string
		want    *List
		wantErr string
	}{
		{
			name: "handles empty string",
			want: &List{},
		},
		{
			name:    "errors on invalid json",
			json:    "}",
			want:    &List{},
			wantErr: "failed to unmarshal todo list json",
		},
		{
			name: "properly unmarshals",
			json: `{"Items": {"write": {"tests": true, "code": false}}, "PrimaryFormats": {"write": {"Color": "red", "Thickness": true }}}`,
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"tests": true,
						"code":  false,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": &color.Format{
						Color:     color.Red,
						Thickness: color.Bold,
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			l := &List{}

			err := l.Load(test.json)
			if err != nil && test.wantErr == "" {
				t.Fatalf("Load(%v) returned error (%v); want nil", test.json, err)
			} else if err == nil && test.wantErr != "" {
				t.Fatalf("Load(%v) returned nil; want error (%v)", test.json, test.wantErr)
			} else if err != nil && test.wantErr != "" && !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Load(%v) returned error (%v); want (%v)", test.json, err, test.wantErr)
			}

			if diff := cmp.Diff(test.want, l, cmpopts.IgnoreUnexported(List{})); diff != "" {
				t.Errorf("Load(%v) produced todo list diff (-want, +got):\n%s", test.json, diff)
			}
		})
	}
}

func TestExecution(t *testing.T) {
	for _, test := range []struct {
		name        string
		l           *List
		args        []string
		wantOK      bool
		want        *List
		wantResp    *commands.ExecutorResponse
		wantChanged bool
		wantStderr  []string
		wantStdout  []string
	}{
		{
			name:       "errors on unknown arg",
			l:          &List{},
			args:       []string{"uhh"},
			want:       &List{},
			wantStderr: []string{"extra unknown args ([uhh])"},
		},
		// ListItems
		{
			name: "lists on nil args",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  false,
						"tests": true,
					},
					"sleep": map[string]bool{},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": &color.Format{
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
			wantOK: true,
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  false,
						"tests": true,
					},
					"sleep": map[string]bool{},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": &color.Format{
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
			wantResp: &commands.ExecutorResponse{},
			wantStdout: []string{
				color.Blue.Format(color.Bold.Format("sleep")),
				"write",
				"  code",
				"  tests",
			},
		},
		{
			name: "lists on empty args",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  false,
						"tests": true,
					},
					"sleep": map[string]bool{},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": &color.Format{
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
			args:   []string{},
			wantOK: true,
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  false,
						"tests": true,
					},
					"sleep": map[string]bool{},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": &color.Format{
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
			wantResp: &commands.ExecutorResponse{},
			wantStdout: []string{
				color.Blue.Format(color.Bold.Format("sleep")),
				"write",
				"  code",
				"  tests",
			},
		},
		// AddItem
		{
			name:       "errors if no arguments",
			l:          &List{},
			args:       []string{"a"},
			want:       &List{},
			wantStderr: []string{`no argument provided for "primary"`},
		},
		{
			name:       "errors if too many arguments",
			l:          &List{},
			args:       []string{"a", "write", "tests", "exclusively"},
			want:       &List{},
			wantStderr: []string{"extra unknown args ([exclusively])"},
		},
		{
			name:   "adds primary to empty list",
			l:      &List{},
			args:   []string{"a", "sleep"},
			wantOK: true,
			want: &List{
				Items: map[string]map[string]bool{
					"sleep": map[string]bool{},
				},
			},
			wantChanged: true,
			wantResp:    &commands.ExecutorResponse{},
		},
		{
			name:   "adds primary and secondary to empty list",
			l:      &List{},
			args:   []string{"a", "write", "tests"},
			wantOK: true,
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"tests": true,
					},
				},
			},
			wantChanged: true,
			wantResp:    &commands.ExecutorResponse{},
		},
		{
			name: "adds just secondary to empty list",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code": true,
					},
				},
			},
			args:   []string{"a", "write", "tests"},
			wantOK: true,
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
			},
			wantChanged: true,
			wantResp:    &commands.ExecutorResponse{},
		},
		{
			name: "error if primary already exists",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{},
				},
			},
			args: []string{"a", "write"},
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{},
				},
			},
			wantStderr: []string{
				`primary item "write" already exists`,
			},
		},
		{
			name: "error if secondary already exists",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code": true,
					},
				},
			},
			args: []string{"a", "write", "code"},
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code": true,
					},
				},
			},
			wantStderr: []string{
				`item "write", "code" already exists`,
			},
		},
		// DeleteItem
		{
			name:       "errors if no arguments",
			l:          &List{},
			args:       []string{"d"},
			want:       &List{},
			wantStderr: []string{`no argument provided for "primary"`},
		},
		{
			name:       "errors if too many arguments",
			l:          &List{},
			args:       []string{"d", "write", "tests", "exclusively"},
			want:       &List{},
			wantStderr: []string{"extra unknown args ([exclusively])"},
		},
		{
			name:       "error if empty items and deleting primary",
			l:          &List{},
			args:       []string{"d", "write"},
			want:       &List{},
			wantStderr: []string{"can't delete from empty list"},
		},
		{
			name:       "error if empty items and deleting secondary",
			l:          &List{},
			args:       []string{"d", "write", "code"},
			want:       &List{},
			wantStderr: []string{"can't delete from empty list"},
		},
		{
			name: "error if unknown primary when deleting primary",
			l: &List{
				Items: map[string]map[string]bool{},
			},
			args:       []string{"d", "write"},
			wantStderr: []string{`Primary item "write" does not exist`},
			want: &List{
				Items: map[string]map[string]bool{},
			},
		},
		{
			name: "error if unknown primary when deleting secondary",
			l: &List{
				Items: map[string]map[string]bool{},
			},
			args:       []string{"d", "write", "code"},
			wantStderr: []string{`Primary item "write" does not exist`},
			want: &List{
				Items: map[string]map[string]bool{},
			},
		},
		{
			name: "error if unknown secondary",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{},
				},
			},
			args:       []string{"d", "write", "code"},
			wantStderr: []string{`Secondary item "code" does not exist`},
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{},
				},
			},
		},
		{
			name: "error if deleting primary that has secondaries",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  false,
						"tests": true,
					},
				},
			},
			args:       []string{"d", "write"},
			wantStderr: []string{"Can't delete primary item that still has secondary items"},
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  false,
						"tests": true,
					},
				},
			},
		},
		{
			name: "successfully deletes primary",
			l: &List{
				Items: map[string]map[string]bool{
					"design": map[string]bool{
						"solutions": true,
					},
					"write": map[string]bool{},
				},
			},
			args:     []string{"d", "write"},
			wantOK:   true,
			wantResp: &commands.ExecutorResponse{},
			want: &List{
				Items: map[string]map[string]bool{
					"design": map[string]bool{
						"solutions": true,
					},
				},
			},
			wantChanged: true,
		},
		{
			name: "successfully deletes secondary",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
			},
			args:     []string{"d", "write", "code"},
			wantOK:   true,
			wantResp: &commands.ExecutorResponse{},
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"tests": true,
					},
				},
			},
			wantChanged: true,
		},
		// FormatPrimary
		{
			name: "successfully adds format",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
			},
			args:     []string{"f", "write", "bold", string(color.Red)},
			wantOK:   true,
			wantResp: &commands.ExecutorResponse{},
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": &color.Format{
						Thickness: color.Bold,
						Color:     color.Red,
					},
				},
			},
			wantChanged: true,
		},
		{
			name: "successfully updates format",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": &color.Format{
						Thickness: color.Bold,
						Color:     color.Red,
					},
				},
			},
			args:     []string{"f", "write", "shy", string(color.Green)},
			wantOK:   true,
			wantResp: &commands.ExecutorResponse{},
			want: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": &color.Format{
						Color: color.Green,
					},
				},
			},
			wantChanged: true,
		},
		{
			name: "error with format",
			l: &List{
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
			},
			args: []string{"f", "write", "crazy"},
			wantStderr: []string{
				"error adding todo list attribute: invalid attribute! crazy",
			},
			want: &List{
				PrimaryFormats: map[string]*color.Format{
					"write": &color.Format{},
				},
				Items: map[string]map[string]bool{
					"write": map[string]bool{
						"code":  true,
						"tests": true,
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			tcos := &commands.TestCommandOS{}
			got, ok := commands.Execute(tcos, test.l.Command(), test.args, nil)
			if ok != test.wantOK {
				t.Fatalf("commands.Execute(%v) returned %v for ok; want %v", test.args, ok, test.wantOK)
			}
			if diff := cmp.Diff(test.wantResp, got); diff != "" {
				t.Fatalf("Execute(%v) produced response diff (-want, +got):\n%s", test.args, diff)
			}

			if diff := cmp.Diff(test.wantStdout, tcos.GetStdout()); diff != "" {
				t.Errorf("command.Execute(%v) produced stdout diff (-want, +got):\n%s", test.args, diff)
			}
			if diff := cmp.Diff(test.wantStderr, tcos.GetStderr()); diff != "" {
				t.Errorf("command.Execute(%v) produced stderr diff (-want, +got):\n%s", test.args, diff)
			}

			if diff := cmp.Diff(test.want, test.l, cmpopts.IgnoreUnexported(List{})); diff != "" {
				t.Fatalf("Execute(%v) produced todo list diff (-want, +got):\n%s", test.args, diff)
			}

			changed := test.l != nil && test.l.Changed()
			if changed != test.wantChanged {
				t.Fatalf("Execute(%v) marked Changed as %v; want %v", test.args, changed, test.wantChanged)
			}
		})
	}
}

func TestAutocomplete(t *testing.T) {
	l := &List{
		Items: map[string]map[string]bool{
			"design": map[string]bool{
				"solutions": true,
			},
			"write": map[string]bool{
				"code":   false,
				"tests":  true,
				"things": false,
			},
		},
		PrimaryFormats: map[string]*color.Format{
			"write": &color.Format{
				Color:     color.Red,
				Thickness: color.Bold,
			},
		},
	}

	for _, test := range []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "first arg suggests all options",
			want: []string{
				"a",
				"d",
				"f",
			},
		},
		{
			name: "suggests nothing if unknown subcommand",
			args: []string{"z", ""},
		},
		// AddItems
		{
			name: "add suggests all primaries",
			args: []string{"a", ""},
			want: []string{
				"design",
				"write",
			},
		},
		{
			name: "add suggests no secondaries",
			args: []string{"a", "write", ""},
		},
		{
			name: "add handles unknown primary",
			args: []string{"a", "huh", ""},
		},
		// DeleteItem
		{
			name: "delete suggests all primaries",
			args: []string{"d", ""},
			want: []string{
				"design",
				"write",
			},
		},
		{
			name: "delete suggests secondaries",
			args: []string{"d", "write", ""},
			want: []string{
				"code",
				"tests",
				"things",
			},
		},
		{
			name: "delete handles unknown primary",
			args: []string{"a", "huh", ""},
		},
		// FormatPrimary
		{
			name: "format suggests all primaries",
			args: []string{"f", ""},
			want: []string{
				"design",
				"write",
			},
		},
		{
			// TODO: should suggest formats
			name: "format suggests no secondaries",
			args: []string{"f", "write", ""},
		},
		{
			name: "format handles unknown primary",
			args: []string{"f", "uhh", ""},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			suggestions := commands.Autocomplete(l.Command(), test.args, 0)
			// Empty list is equivalent to nil.
			if len(suggestions) == 0 {
				suggestions = nil
			}
			if diff := cmp.Diff(test.want, suggestions); diff != "" {
				t.Errorf("Complete(%v) produced diff (-want, +got):\n%s", test.args, diff)
			}
		})
	}
}

func TestUsage(t *testing.T) {
	l := &List{}
	wantUsage := []string{
		"a", "PRIMARY", "[", "SECONDARY", "]", "\n",
		"d", "PRIMARY", "[", "SECONDARY", "]", "\n",
		"f", "PRIMARY", "FORMAT", "[FORMAT ...]", "\n",
	}
	usage := l.Command().Usage()
	if diff := cmp.Diff(wantUsage, usage); diff != "" {
		t.Errorf("Usage() produced diff:\n%s", diff)
	}
}

func TestMetadata(t *testing.T) {
	l := &List{}
	want := "todo-list"
	if l.Name() != want {
		t.Errorf("Incorrect todo list name: got %s; want %s", l.Name(), want)
	}

	want = "td"
	if l.Alias() != want {
		t.Errorf("Incorrect todo list alias: got %s; want %s", l.Alias(), want)
	}
}

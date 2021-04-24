package todo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/leep-frog/command"
	"github.com/leep-frog/command/color"

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
					"write": {
						"tests": true,
						"code":  false,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": {
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
		name       string
		l          *List
		args       []string
		want       *List
		wantErr    error
		wantResp   *command.ExecuteData
		wantStderr []string
		wantStdout []string
		wantData   *command.Data
	}{
		{
			name:       "errors on unknown arg",
			l:          &List{},
			args:       []string{"uhh"},
			want:       &List{},
			wantStderr: []string{"Unprocessed extra args: [uhh]"},
			wantErr:    fmt.Errorf("Unprocessed extra args: [uhh]"),
		},
		// ListItems
		{
			name: "lists on nil args",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  false,
						"tests": true,
					},
					"sleep": {},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": {
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
			want: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  false,
						"tests": true,
					},
					"sleep": {},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": {
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
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
					"write": {
						"code":  false,
						"tests": true,
					},
					"sleep": {},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": {
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
			args: []string{},
			want: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  false,
						"tests": true,
					},
					"sleep": {},
				},
				PrimaryFormats: map[string]*color.Format{
					"sleep": {
						Color:     color.Blue,
						Thickness: color.Bold,
					},
				},
			},
			wantStdout: []string{
				color.Blue.Format(color.Bold.Format("sleep")),
				"write",
				"  code",
				"  tests",
			},
		},
		// AddItem
		{
			name: "errors if no arguments",
			l:    &List{},
			args: []string{"a"},
			want: &List{},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg: command.StringValue(""),
				},
			},
			wantStderr: []string{"not enough arguments"},
			wantErr:    fmt.Errorf("not enough arguments"),
		},
		{
			name: "errors if too many arguments",
			l:    &List{},
			args: []string{"a", "write", "tests", "exclusively"},
			want: &List{},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("tests"),
				},
			},
			wantStderr: []string{"Unprocessed extra args: [exclusively]"},
			wantErr:    fmt.Errorf("Unprocessed extra args: [exclusively]"),
		},
		{
			name: "adds primary to empty list",
			l:    &List{},
			args: []string{"a", "sleep"},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"sleep": {},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("sleep"),
					secondaryArg: command.StringValue(""),
				},
			},
		},
		{
			name: "adds primary and secondary to empty list",
			l:    &List{},
			args: []string{"a", "write", "tests"},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"tests": true,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("tests"),
				},
			},
		},
		{
			name: "adds just secondary to empty list",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code": true,
					},
				},
			},
			args: []string{"a", "write", "tests"},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("tests"),
				},
			},
		},
		{
			name: "error if primary already exists",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {},
				},
			},
			args: []string{"a", "write"},
			want: &List{
				Items: map[string]map[string]bool{
					"write": {},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue(""),
				},
			},
			wantStderr: []string{
				`primary item "write" already exists`,
			},
			wantErr: fmt.Errorf(`primary item "write" already exists`),
		},
		{
			name: "error if secondary already exists",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code": true,
					},
				},
			},
			args: []string{"a", "write", "code"},
			want: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code": true,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("code"),
				},
			},
			wantStderr: []string{
				`item "write", "code" already exists`,
			},
			wantErr: fmt.Errorf(`item "write", "code" already exists`),
		},
		// DeleteItem
		{
			name:       "errors if no arguments",
			l:          &List{},
			args:       []string{"d"},
			want:       &List{},
			wantStderr: []string{"not enough arguments"},
			wantErr:    fmt.Errorf("not enough arguments"),
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg: command.StringValue(""),
				},
			},
		},
		{
			name:       "errors if too many arguments",
			l:          &List{},
			args:       []string{"d", "write", "tests", "exclusively"},
			wantStderr: []string{"Unprocessed extra args: [exclusively]"},
			wantErr:    fmt.Errorf("Unprocessed extra args: [exclusively]"),
			want:       &List{},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("tests"),
				},
			},
		},
		{
			name: "error if empty items and deleting primary",
			l:    &List{},
			args: []string{"d", "write"},
			want: &List{},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue(""),
				},
			},
			wantStderr: []string{"can't delete from empty list"},
			wantErr:    fmt.Errorf("can't delete from empty list"),
		},
		{
			name: "error if empty items and deleting secondary",
			l:    &List{},
			args: []string{"d", "write", "code"},
			want: &List{},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("code"),
				},
			},
			wantStderr: []string{"can't delete from empty list"},
			wantErr:    fmt.Errorf("can't delete from empty list"),
		},
		{
			name: "error if unknown primary when deleting primary",
			l: &List{
				Items: map[string]map[string]bool{},
			},
			args:       []string{"d", "write"},
			wantStderr: []string{`Primary item "write" does not exist`},
			wantErr:    fmt.Errorf(`Primary item "write" does not exist`),
			want: &List{
				Items: map[string]map[string]bool{},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue(""),
				},
			},
		},
		{
			name: "error if unknown primary when deleting secondary",
			l: &List{
				Items: map[string]map[string]bool{},
			},
			args:       []string{"d", "write", "code"},
			wantStderr: []string{`Primary item "write" does not exist`},
			wantErr:    fmt.Errorf(`Primary item "write" does not exist`),
			want: &List{
				Items: map[string]map[string]bool{},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("code"),
				},
			},
		},
		{
			name: "error if unknown secondary",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {},
				},
			},
			args:       []string{"d", "write", "code"},
			wantStderr: []string{`Secondary item "code" does not exist`},
			wantErr:    fmt.Errorf(`Secondary item "code" does not exist`),
			want: &List{
				Items: map[string]map[string]bool{
					"write": {},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("code"),
				},
			},
		},
		{
			name: "error if deleting primary that has secondaries",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  false,
						"tests": true,
					},
				},
			},
			args:       []string{"d", "write"},
			wantStderr: []string{"Can't delete primary item that still has secondary items"},
			wantErr:    fmt.Errorf("Can't delete primary item that still has secondary items"),
			want: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  false,
						"tests": true,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue(""),
				},
			},
		},
		{
			name: "successfully deletes primary",
			l: &List{
				Items: map[string]map[string]bool{
					"design": {
						"solutions": true,
					},
					"write": {},
				},
			},
			args: []string{"d", "write"},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"design": {
						"solutions": true,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue(""),
				},
			},
		},
		{
			name: "successfully deletes secondary",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
			},
			args: []string{"d", "write", "code"},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"tests": true,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:   command.StringValue("write"),
					secondaryArg: command.StringValue("code"),
				},
			},
		},
		// FormatPrimary
		{
			name: "successfully adds format",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
			},
			args: []string{"f", "write", "bold", string(color.Red)},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": {
						Thickness: color.Bold,
						Color:     color.Red,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:    command.StringValue("write"),
					color.ArgName: command.StringListValue("bold", string(color.Red)),
				},
			},
		},
		{
			name: "successfully updates format",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": {
						Thickness: color.Bold,
						Color:     color.Red,
					},
				},
			},
			args: []string{"f", "write", "shy", string(color.Green)},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
				PrimaryFormats: map[string]*color.Format{
					"write": {
						Color: color.Green,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:    command.StringValue("write"),
					color.ArgName: command.StringListValue("shy", string(color.Green)),
				},
			},
		},
		{
			name: "error with format",
			l: &List{
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
			},
			args: []string{"f", "write", "crazy"},
			want: &List{
				PrimaryFormats: map[string]*color.Format{
					"write": nil,
				},
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
				},
			},
			wantData: &command.Data{
				Values: map[string]*command.Value{
					primaryArg:    command.StringValue("write"),
					color.ArgName: command.StringListValue("crazy"),
				},
			},
			wantStderr: []string{"invalid attribute: crazy"},
			wantErr:    fmt.Errorf("invalid attribute: crazy"),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			command.ExecuteTest(t, test.l.Node(), test.args, test.wantErr, nil, test.wantData, test.wantStdout, test.wantStderr)
			if diff := cmp.Diff(test.want, test.l, cmp.AllowUnexported(List{})); diff != "" {
				t.Fatalf("Execute(%v) produced todo list diff (-want, +got):\n%s", test.args, diff)
			}
		})
	}
}

func TestAutocomplete(t *testing.T) {
	l := &List{
		Items: map[string]map[string]bool{
			"design": {
				"solutions": true,
			},
			"write": {
				"code":   false,
				"tests":  true,
				"things": false,
			},
		},
		PrimaryFormats: map[string]*color.Format{
			"write": {
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
			name: "format suggests no secondaries",
			args: []string{"f", "write", ""},
			want: color.Attributes(),
		},
		{
			name: "format handles unknown primary",
			args: []string{"f", "uhh", ""},
			want: color.Attributes(),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			suggestions := command.Autocomplete(l.Node(), test.args)
			if diff := cmp.Diff(test.want, suggestions, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Complete(%v) produced diff (-want, +got):\n%s", test.args, diff)
			}
		})
	}
}

func TestMetadata(t *testing.T) {
	l := &List{}
	want := "td"
	if l.Name() != want {
		t.Errorf("Incorrect todo list name: got %s; want %s", l.Name(), want)
	}
}

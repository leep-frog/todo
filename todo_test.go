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
		WantErr string
	}{
		{
			name: "handles empty string",
		},
		{
			name:    "errors on invalid json",
			json:    "}",
			WantErr: "failed to unmarshal todo list json",
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

			want := test.want
			if want == nil {
				want = &List{}
			}

			err := l.Load(test.json)
			if err != nil && test.WantErr == "" {
				t.Fatalf("Load(%v) returned error (%v); want nil", test.json, err)
			} else if err == nil && test.WantErr != "" {
				t.Fatalf("Load(%v) returned nil; want error (%v)", test.json, test.WantErr)
			} else if err != nil && test.WantErr != "" && !strings.Contains(err.Error(), test.WantErr) {
				t.Fatalf("Load(%v) returned error (%v); want (%v)", test.json, err, test.WantErr)
			}

			if diff := cmp.Diff(want, l, cmpopts.IgnoreUnexported(List{})); diff != "" {
				t.Errorf("Load(%v) produced todo list diff (-want, +got):\n%s", test.json, diff)
			}
		})
	}
}

func TestExecution(t *testing.T) {
	for _, test := range []struct {
		name string
		l    *List
		etc  *command.ExecuteTestCase
		want *List
	}{
		{
			name: "errors on unknown arg",
			etc: &command.ExecuteTestCase{
				Args:       []string{"uhh"},
				WantStderr: "Unprocessed extra args: [uhh]\n",
				WantErr:    fmt.Errorf("Unprocessed extra args: [uhh]"),
			},
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
			etc: &command.ExecuteTestCase{
				WantStdout: strings.Join([]string{
					color.Blue.Format(color.Bold.Format("sleep")),
					"write",
					"  code",
					"  tests",
					"",
				}, "\n"),
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
			etc: &command.ExecuteTestCase{
				Args: []string{},
				WantStdout: strings.Join([]string{
					color.Blue.Format(color.Bold.Format("sleep")),
					"write",
					"  code",
					"  tests",
					"",
				}, "\n"),
			},
		},
		// AddItem
		{
			name: "errors if no arguments",
			etc: &command.ExecuteTestCase{
				Args:       []string{"a"},
				WantStderr: "Argument \"primary\" requires at least 1 argument, got 0\n",
				WantErr:    fmt.Errorf(`Argument "primary" requires at least 1 argument, got 0`),
			},
		},
		{
			name: "errors if too many arguments",
			etc: &command.ExecuteTestCase{
				Args: []string{"a", "write", "tests", "exclusively"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "tests",
					},
				},
				WantStderr: "Unprocessed extra args: [exclusively]\n",
				WantErr:    fmt.Errorf("Unprocessed extra args: [exclusively]"),
			},
		},
		{
			name: "adds primary to empty list",
			etc: &command.ExecuteTestCase{
				Args: []string{"a", "sleep"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "sleep",
					},
				},
			},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"sleep": {},
				},
			},
		},
		{
			name: "adds primary and secondary to empty list",
			etc: &command.ExecuteTestCase{
				Args: []string{"a", "write", "tests"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "tests",
					},
				},
			},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"tests": true,
					},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"a", "write", "tests"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "tests",
					},
				},
			},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"code":  true,
						"tests": true,
					},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"a", "write"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "write",
					},
				},
				WantStderr: "primary item \"write\" already exists\n",
				WantErr:    fmt.Errorf(`primary item "write" already exists`),
			},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"a", "write", "code"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "code",
					},
				},
				WantStderr: "item \"write\", \"code\" already exists\n",
				WantErr:    fmt.Errorf(`item "write", "code" already exists`),
			},
		},
		// DeleteItem
		{
			name: "errors if no arguments",
			etc: &command.ExecuteTestCase{
				Args:       []string{"d"},
				WantStderr: "Argument \"primary\" requires at least 1 argument, got 0\n",
				WantErr:    fmt.Errorf(`Argument "primary" requires at least 1 argument, got 0`),
			},
		},
		{
			name: "errors if too many arguments",
			etc: &command.ExecuteTestCase{
				Args:       []string{"d", "write", "tests", "exclusively"},
				WantStderr: "Unprocessed extra args: [exclusively]\n",
				WantErr:    fmt.Errorf("Unprocessed extra args: [exclusively]"),
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "tests",
					},
				},
			},
		},
		{
			name: "error if empty items and deleting primary",
			etc: &command.ExecuteTestCase{
				Args: []string{"d", "write"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "write",
					},
				},
				WantStderr: "can't delete from empty list\n",
				WantErr:    fmt.Errorf("can't delete from empty list"),
			},
		},
		{
			name: "error if empty items and deleting secondary",
			etc: &command.ExecuteTestCase{
				Args: []string{"d", "write", "code"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "code",
					},
				},
				WantStderr: "can't delete from empty list\n",
				WantErr:    fmt.Errorf("can't delete from empty list"),
			},
		},
		{
			name: "error if unknown primary when deleting primary",
			l: &List{
				Items: map[string]map[string]bool{},
			},
			etc: &command.ExecuteTestCase{
				Args:       []string{"d", "write"},
				WantStderr: "Primary item \"write\" does not exist\n",
				WantErr:    fmt.Errorf(`Primary item "write" does not exist`),
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "write",
					},
				},
			},
		},
		{
			name: "error if unknown primary when deleting secondary",
			l: &List{
				Items: map[string]map[string]bool{},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"d", "write", "code"},
				WantStderr: strings.Join([]string{
					`Primary item "write" does not exist`,
					"",
				}, "\n"),
				WantErr: fmt.Errorf(`Primary item "write" does not exist`),
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "code",
					},
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
			etc: &command.ExecuteTestCase{
				Args:       []string{"d", "write", "code"},
				WantStderr: "Secondary item \"code\" does not exist\n",
				WantErr:    fmt.Errorf(`Secondary item "code" does not exist`),
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "code",
					},
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
			etc: &command.ExecuteTestCase{
				Args:       []string{"d", "write"},
				WantStderr: "Can't delete primary item that still has secondary items\n",
				WantErr:    fmt.Errorf("Can't delete primary item that still has secondary items"),
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "write",
					},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"d", "write"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "write",
					},
				},
			},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"design": {
						"solutions": true,
					},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"d", "write", "code"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "code",
					},
				},
			},
			want: &List{
				changed: true,
				Items: map[string]map[string]bool{
					"write": {
						"tests": true,
					},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"f", "write", "bold", string(color.Red)},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:    "write",
						color.ArgName: []string{"bold", string(color.Red)},
					},
				},
			},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"f", "write", "shy", string(color.Green)},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:    "write",
						color.ArgName: []string{"shy", string(color.Green)},
					},
				},
			},
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
			etc: &command.ExecuteTestCase{
				Args: []string{"f", "write", "crazy"},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:    "write",
						color.ArgName: []string{"crazy"},
					},
				},
				WantStderr: "invalid attribute: crazy\n",
				WantErr:    fmt.Errorf("invalid attribute: crazy"),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.l == nil {
				test.l = &List{}
			}
			test.etc.Node = test.l.Node()
			command.ExecuteTest(t, test.etc)
			command.ChangeTest(t, test.want, test.l, cmp.AllowUnexported(List{}))
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
		ctc  *command.CompleteTestCase
	}{
		{
			name: "first arg suggests all options",
			ctc: &command.CompleteTestCase{
				Want: []string{
					"a",
					"d",
					"f",
				},
			},
		},
		{
			name: "suggests nothing if unknown subcommand",
			ctc: &command.CompleteTestCase{
				Args:    "td z ",
				WantErr: fmt.Errorf("Unprocessed extra args: [z ]"),
			},
		},
		// AddItems
		{
			name: "add suggests all primaries",
			ctc: &command.CompleteTestCase{
				Args: "td a ",
				Want: []string{
					"design",
					"write",
				},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "",
					},
				},
			},
		},
		{
			name: "add suggests no secondaries",
			ctc: &command.CompleteTestCase{
				Args: "td a write ",
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "",
					},
				},
			},
		},
		{
			name: "add handles unknown primary",
			ctc: &command.CompleteTestCase{
				Args: "td a huh ",
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "huh",
						secondaryArg: "",
					},
				},
			},
		},
		// DeleteItem
		{
			name: "delete suggests all primaries",
			ctc: &command.CompleteTestCase{
				Args: "td d ",
				Want: []string{
					"design",
					"write",
				},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "",
					},
				},
			},
		},
		{
			name: "delete suggests secondaries",
			ctc: &command.CompleteTestCase{
				Args: "td d write ",
				Want: []string{
					"code",
					"tests",
					"things",
				},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "write",
						secondaryArg: "",
					},
				},
			},
		},
		{
			name: "delete handles unknown primary",
			ctc: &command.CompleteTestCase{
				Args: "td a huh ",
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg:   "huh",
						secondaryArg: "",
					},
				},
			},
		},
		// FormatPrimary
		{
			name: "format suggests all primaries",
			ctc: &command.CompleteTestCase{
				Args: "td f ",
				Want: []string{
					"design",
					"write",
				},
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "",
					},
				},
			},
		},
		{
			name: "format suggests no secondaries",
			ctc: &command.CompleteTestCase{
				Args: "td f write ",
				Want: color.Attributes(),
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "write",
						"format":   []string{""},
					},
				},
			},
		},
		{
			name: "format handles unknown primary",
			ctc: &command.CompleteTestCase{
				Args: "td f uhh ",
				Want: color.Attributes(),
				WantData: &command.Data{
					Values: map[string]interface{}{
						primaryArg: "uhh",
						"format":   []string{""},
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.ctc.Node = l.Node()
			command.CompleteTest(t, test.ctc)
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

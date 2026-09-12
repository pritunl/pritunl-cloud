package resource

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/view"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	filterClear = widget.DialogCustom
)

// filterFieldsMsg carries the filter dialog fields loaded in the
// background.
type filterFieldsMsg struct {
	provider Provider
	fields   []FilterField
	err      error
}

// applyFilterMsg replaces the list filter and reloads the first page.
type applyFilterMsg struct {
	provider Provider
	filter   Filter
	summary  string
}

// filterBinding links a filter field to its dialog option.
type filterBinding struct {
	field  FilterField
	text   *widget.OptionText
	choice *widget.OptionSelect
}

func (b filterBinding) value() (value, label string) {
	if b.choice != nil {
		return b.choice.GetValue(), b.choice.GetLabel()
	}
	value = b.text.GetValue()
	return value, value
}

// filterFieldsCmd loads the filter fields of the provider.
func filterFieldsCmd(provider Provider) tea.Cmd {
	return func() tea.Msg {
		db := database.GetDatabase()
		if db == nil {
			return filterFieldsMsg{
				provider: provider,
				err: &errortypes.DatabaseError{
					errors.New("resource: Database not connected"),
				},
			}
		}
		defer db.Close()

		fields, err := provider.FilterFields(db)
		return filterFieldsMsg{
			provider: provider,
			fields:   fields,
			err:      err,
		}
	}
}

// filterDialog opens the filter dialog with the current filter values,
// applying returns an applyFilterMsg to the list.
func (l *List) filterDialog(fields []FilterField) tea.Cmd {
	opts := []widget.Option{}
	bindings := []filterBinding{}

	for _, field := range fields {
		binding := filterBinding{
			field: field,
		}

		if field.Options != nil {
			choices := make([]widget.SelectOption, 0, len(field.Options)+1)
			choices = append(choices, widget.SelectOption{
				Label: "Any " + field.Label,
				Value: "",
			})
			choices = append(choices, field.Options...)

			choice := &widget.OptionSelect{
				Label:   field.Label,
				Options: choices,
			}
			choice.SetValue(l.filter[field.Key])

			binding.choice = choice
			opts = append(opts, choice)
		} else {
			text := &widget.OptionText{
				Label:       field.Label,
				Placeholder: field.Placeholder,
				Value:       l.filter[field.Key],
			}

			binding.text = text
			opts = append(opts, text)
		}

		bindings = append(bindings, binding)
	}

	opts = append(opts,
		&widget.OptionButton{
			Label:  "Clear",
			Return: filterClear,
		},
		&widget.OptionButton{
			Label:  "Cancel",
			Return: widget.DialogCancel,
		},
		&widget.OptionButton{
			Label:  "Apply",
			Return: widget.DialogOk,
		},
	)

	provider := l.provider
	dialog := widget.NewDialog(
		"Filter "+provider.Title(),
		"",
		opts...,
	)

	return view.Dialog(dialog, func(ret int) tea.Cmd {
		switch ret {
		case filterClear:
			return func() tea.Msg {
				return applyFilterMsg{
					provider: provider,
					filter:   Filter{},
				}
			}
		case widget.DialogOk:
			filter := Filter{}
			parts := []string{}

			for _, binding := range bindings {
				value, label := binding.value()
				if value == "" {
					continue
				}

				filter[binding.field.Key] = value
				parts = append(parts, fmt.Sprintf(
					"%s=%s", binding.field.Label, label))
			}

			return func() tea.Msg {
				return applyFilterMsg{
					provider: provider,
					filter:   filter,
					summary:  strings.Join(parts, ", "),
				}
			}
		}

		return nil
	})
}

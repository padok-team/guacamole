package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// noArgs is a cobra.Args validator for commands that don't take positional
// arguments (only flags and, where relevant, subcommands).
//
// Cobra's default suggestion mechanism ("Did you mean this?") only kicks in
// when a command has no Run/RunE of its own, because in that case an unknown
// word is treated as a genuinely unknown command. Commands like `static`
// have both subcommands (layer, module) *and* their own default Run (used
// when called with no subcommand), so Cobra instead treats an unrecognized
// word (e.g. "layers") as a positional argument and silently runs the
// default behavior instead of failing.
//
// Using this validator as a command's Args rejects any stray positional
// argument and, if it looks like a typo of a sibling/child command,
// suggests the closest match.
func noArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	msg := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())

	// cmd.SuggestionsMinimumDistance defaults to 0, which would reject every
	// suggestion (Cobra only sets it to its real default of 2 internally,
	// inside the unexported findSuggestions helper). Mirror that default here.
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}

	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += "\n\nDid you mean this?\n"
		for _, s := range suggestions {
			msg += fmt.Sprintf("\t%s\n", s)
		}
	}

	return errors.New(msg)
}

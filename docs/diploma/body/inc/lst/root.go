package command

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "backupcli",
	Short: "A CLI utility for working with files",
}

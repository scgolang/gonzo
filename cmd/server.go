// Copyright © 2017 Brian Sorahan <bsorahan@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a gonzo server.",
	Long:  "Run a gonzo server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := NewConfig(args, cmd.Flags())
		if err != nil {
			return err
		}
		app, err := NewApp(context.Background(), config)
		if err != nil {
			return err
		}
		return app.Wait()
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}

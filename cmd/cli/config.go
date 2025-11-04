package cli

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/utils"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration operations",
	Long:  `Commands for managing server configuration`,
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		cfg, err := c.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		if RawFlag {
			tomlData, err := toml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			fmt.Print(string(tomlData))
			return nil
		}

		return outputJSON(cfg)
	},
}

var configUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update configuration (interactive editor)",
	Long:  `Fetches the current config, opens it in an editor, then sends the updated config back to the server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient()
		if err != nil {
			return err
		}

		cfg, err := c.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		tomlData, err := toml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		editedData, err := utils.EditTextWithEditor(string(tomlData), "toml")
		if err != nil {
			return fmt.Errorf("failed to edit config: %w", err)
		}

		var updatedConfig config.Config
		if err := toml.Unmarshal([]byte(editedData), &updatedConfig); err != nil {
			return fmt.Errorf("invalid TOML: %w", err)
		}

		if err := c.UpdateConfig(&updatedConfig); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		fmt.Println("Configuration updated successfully")
		return nil
	},
}

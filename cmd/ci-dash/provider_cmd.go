package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joelklabo/ceye/internal/providers"
	"github.com/joelklabo/ceye/internal/providers/manager"
)

func providerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage runtime provider entries",
	}
	cmd.AddCommand(providerListCmd())
	cmd.AddCommand(providerAddCmd())
	cmd.AddCommand(providerUpdateCmd())
	cmd.AddCommand(providerEnableCmd(true))
	cmd.AddCommand(providerEnableCmd(false))
	cmd.AddCommand(providerRemoveCmd())
	return cmd
}

func providerListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List dynamic providers stored for the dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := manager.New(resolveProviderStorePath(providerStoreFlag))
			if err != nil {
				return err
			}
			entries := store.List()
			if len(entries) == 0 {
				fmt.Println("no stored providers")
				return nil
			}
			fmt.Printf("%-36s  %-8s  %-12s  %s\n", "ID", "Enabled", "Type", "Name")
			for _, e := range entries {
				fmt.Printf("%-36s  %-8t  %-12s  %s\n", e.ID, e.Enabled, e.Config.Type, providers.DisplayName(e.Config))
			}
			return nil
		},
	}
}

func providerAddCmd() *cobra.Command {
	var filePath string
	var jsonSpec string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a runtime provider without editing the primary config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := readProviderConfig(filePath, jsonSpec)
			if err != nil {
				return err
			}
			store, err := manager.New(resolveProviderStorePath(providerStoreFlag))
			if err != nil {
				return err
			}
			record, err := store.Add(cfg)
			if err != nil {
				return err
			}
			fmt.Printf("added provider %s (%s)\n", record.ID, providers.DisplayName(record.Config))
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "config", "", "Path to YAML/JSON snippet specifying the provider")
	cmd.Flags().StringVar(&jsonSpec, "json", "", "Inline JSON object describing the provider")
	cmd.MarkFlagsMutuallyExclusive("config", "json")
	return cmd
}

func providerUpdateCmd() *cobra.Command {
	var filePath string
	var jsonSpec string
	var id string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a stored provider configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}
			cfg, err := readProviderConfig(filePath, jsonSpec)
			if err != nil {
				return err
			}
			store, err := manager.New(resolveProviderStorePath(providerStoreFlag))
			if err != nil {
				return err
			}
			if err := store.Update(id, cfg); err != nil {
				return err
			}
			fmt.Printf("updated provider %s (%s)\n", id, providers.DisplayName(cfg))
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Identifier of the stored provider")
	cmd.Flags().StringVar(&filePath, "config", "", "Path to YAML/JSON snippet specifying the provider")
	cmd.Flags().StringVar(&jsonSpec, "json", "", "Inline JSON object describing the provider")
	cmd.MarkFlagsMutuallyExclusive("config", "json")
	cmd.MarkFlagRequired("id")
	return cmd
}

func providerEnableCmd(enable bool) *cobra.Command {
	var id string
	use := "enable"
	if !enable {
		use = "disable"
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("%s a stored provider", use),
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}
			store, err := manager.New(resolveProviderStorePath(providerStoreFlag))
			if err != nil {
				return err
			}
			if err := store.SetEnabled(id, enable); err != nil {
				return err
			}
			fmt.Printf("%sd provider %s\n", use, id)
			return nil
		},
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringVar(&id, "id", "", "Identifier of the stored provider")
	return cmd
}

func providerRemoveCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete a stored provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}
			store, err := manager.New(resolveProviderStorePath(providerStoreFlag))
			if err != nil {
				return err
			}
			if err := store.Remove(id); err != nil {
				return err
			}
			fmt.Printf("removed provider %s\n", id)
			return nil
		},
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringVar(&id, "id", "", "Identifier of the stored provider")
	return cmd
}

func readProviderConfig(path, inline string) (providers.ProviderConfig, error) {
	var cfg providers.ProviderConfig
	switch {
	case path != "":
		v := viper.New()
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return cfg, err
		}
		if err := v.Unmarshal(&cfg); err != nil {
			return cfg, err
		}
	case inline != "":
		if err := json.Unmarshal([]byte(inline), &cfg); err != nil {
			return cfg, err
		}
	default:
		return cfg, fmt.Errorf("either --config or --json must be provided")
	}
	if cfg.Type == "" {
		return cfg, fmt.Errorf("provider type is required")
	}
	return cfg, nil
}

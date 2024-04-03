package cmd

import (
	"github.com/spf13/cobra"
	"media-search/migration"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate data of books and movies into database",
	RunE:  migrate,
}

func migrate(cmd *cobra.Command, args []string) error {
	config := initViper()
	db := initDBConnection(config)
	esClient := initElasticSearchClient()
	migration.Migrate(db, esClient)
	return nil
}

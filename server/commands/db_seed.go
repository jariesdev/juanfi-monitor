package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/spf13/cobra"
)

var dbSeedCmd = &cobra.Command{
	Use:   "db-seed",
	Short: "Seed the database with default roles",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		roleRepo := repository.NewRoleRepository(db)

		roles, err := roleRepo.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		// Check if Admin role already exists
		for _, r := range roles {
			if strings.EqualFold(r.Name, "Admin") {
				fmt.Println("Admin role already exists — skipping.")
				return
			}
		}

		admin := &models.Role{Name: "Admin"}
		admin.SetPermissions(models.AllPermissions())

		if err := roleRepo.Create(admin); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to create Admin role: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Admin role created with permissions: %s\n", strings.Join(models.AllPermissions(), ", "))
	},
}

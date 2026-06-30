package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/spf13/cobra"
)

var userAssignRoleCmd = &cobra.Command{
	Use:   "user-assign-role",
	Short: "Assign one or more roles to an existing user (replaces current roles)",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		roleRepo := repository.NewRoleRepository(db)
		userRepo := repository.NewUserRepository(db)

		roles, err := roleRepo.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(roles) == 0 {
			fmt.Fprintln(os.Stderr, "error: no roles found — create a role first")
			os.Exit(1)
		}

		fmt.Println("Available roles:")
		for _, r := range roles {
			fmt.Printf("  %s  (%s)\n", r.Name, strings.Join(r.GetPermissions(), ", "))
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("\nEnter username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		user, err := userRepo.GetByUsername(username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: user %q not found\n", username)
			os.Exit(1)
		}

		if len(user.Roles) > 0 {
			var current []string
			for _, r := range user.Roles {
				current = append(current, r.Name)
			}
			fmt.Printf("Current roles: %s\n", strings.Join(current, ", "))
		} else {
			fmt.Println("Current roles: (none)")
		}

		fmt.Print("Enter role names to assign (comma-separated, leave empty to clear): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var roleIDs []uint
		if input != "" {
			for _, name := range strings.Split(input, ",") {
				name = strings.TrimSpace(name)
				found := false
				for _, r := range roles {
					if strings.EqualFold(r.Name, name) {
						roleIDs = append(roleIDs, r.ID)
						found = true
						break
					}
				}
				if !found {
					fmt.Fprintf(os.Stderr, "error: role %q not found\n", name)
					os.Exit(1)
				}
			}
		}

		if err := userRepo.AssignRoles(user.ID, roleIDs); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if len(roleIDs) == 0 {
			fmt.Printf("All roles cleared for user %q.\n", username)
		} else {
			fmt.Printf("Roles assigned to user %q.\n", username)
		}
	},
}

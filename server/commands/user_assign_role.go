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
	Short: "Assign a role to an existing user",
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
			fmt.Printf("  [%d] %s  (%s)\n", r.ID, r.Name, strings.Join(r.GetPermissions(), ", "))
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

		fmt.Print("Enter role name: ")
		roleName, _ := reader.ReadString('\n')
		roleName = strings.TrimSpace(roleName)

		var matchedRoleID *uint
		for _, r := range roles {
			if strings.EqualFold(r.Name, roleName) {
				id := r.ID
				matchedRoleID = &id
				break
			}
		}
		if matchedRoleID == nil {
			fmt.Fprintf(os.Stderr, "error: role %q not found\n", roleName)
			os.Exit(1)
		}

		user.RoleID = matchedRoleID
		if err := userRepo.Update(user); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Role %q assigned to user %q.\n", roleName, username)
	},
}

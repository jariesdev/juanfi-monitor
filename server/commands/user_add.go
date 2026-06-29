package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/spf13/cobra"
)

var userAddCmd = &cobra.Command{
	Use:   "user-add",
	Short: "Interactively add a new user account",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Print("Enter password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		repo := repository.NewUserRepository(db)
		user := &models.User{
			Username: username,
			Password: password,
			IsActive: true,
		}
		if err := repo.Create(user); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User %q was added.\n", username)
	},
}

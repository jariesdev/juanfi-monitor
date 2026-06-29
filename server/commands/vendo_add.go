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

var vendoAddCmd = &cobra.Command{
	Use:   "vendo-add",
	Short: "Interactively register a new vendo machine",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter vendo name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("Enter API URL (e.g. http://192.168.42.10:8081): ")
		apiURL, _ := reader.ReadString('\n')
		apiURL = strings.TrimSpace(apiURL)

		fmt.Print("Enter API key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)

		repo := repository.NewVendoRepository(db)
		vendo := &models.Vendo{
			Name:   name,
			APIURL: &apiURL,
			APIKey: &apiKey,
		}
		if err := repo.Create(vendo); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Vendo %q added (id=%d).\n", name, vendo.ID)
	},
}

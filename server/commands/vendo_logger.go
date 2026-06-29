package commands

import (
	"fmt"
	"os"

	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
	"github.com/spf13/cobra"
)

var vendoLoggerCmd = &cobra.Command{
	Use:   "vendo-logger",
	Short: "Pull the latest logs and sales from all active vendo machines",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		repo := repository.NewVendoRepository(db)
		vendos, err := repo.AllActive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching vendos: %v\n", err)
			os.Exit(1)
		}

		if len(vendos) == 0 {
			fmt.Println("No active vendo machines registered. Add one with vendo-add first.")
			return
		}

		for i := range vendos {
			v := &vendos[i]
			fmt.Printf("Pulling logs from %s...\n", v.Name)
			logger := services.NewJuanfiLogger(v, db)
			if err := logger.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "  [%s] error: %v\n", v.Name, err)
			}
		}
		fmt.Println("Done.")
	},
}

package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/spf13/cobra"
)

// notificationsClearCmd permanently deletes notifications created before the
// given date. Deletion reuses NotificationRepository.DeleteOlderThan so the
// row-removal logic lives in the repository layer, not the command.
var notificationsClearCmd = &cobra.Command{
	Use:   "notifications-clear <date>",
	Short: "Delete notifications created before the given date (YYYY-MM-DD, PHT)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		loc := time.FixedZone("PHT", 8*60*60)
		cutoff, err := time.ParseInLocation(time.DateOnly, args[0], loc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid date %q (want YYYY-MM-DD): %v\n", args[0], err)
			os.Exit(1)
		}

		repo := repository.NewNotificationRepository(db)
		deleted, err := repo.DeleteOlderThan(cutoff)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: delete notifications: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted %d notification(s) created before %s.\n", deleted, cutoff.Format(time.DateOnly))
	},
}

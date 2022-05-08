package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/koltyakov/gosip"
	"github.com/koltyakov/gosip/api"
	strategy "github.com/koltyakov/gosip/auth/saml"
	"github.com/koltyakov/gosip/cpass"
	"github.com/koltyakov/spsync"
)

func main() {
	// Load environment variables needed for sync processing
	_ = godotenv.Load()

	ctx := context.Background()

	// Setup sync configuration per a specific entity.
	// Usually you need to sync multiple Lists so most of a logic from this sample
	// iterates for synced entities.
	// In a serverless scenarios, sync session timeout is controlled and terminated
	// by the caller after timeout. It's OK, the resumed session is designed
	// to continue since a previous successful incremental state.

	opts := &spsync.Options{
		// Create and bind SharePoint API client, see more https://go.spflow.com
		// For multiple entities sync initiate the client externally
		SP: NewSP(),
		// Pass current state from a persistent storage (e.g. database or SharePoint list)
		State: &spsync.State{
			EntID:    "Lists/MyList",
			SyncMode: spsync.Full,
		},
		// Pass configuration for a specific entity, usually stored in config file
		Ent: &spsync.Ent{
			Select: []string{"Id", "Title"},
		},
		// Provide a handler to deal with created and updated items
		Upsert: func(ctx context.Context, items []spsync.Item) error {
			// Implement your logic here for your target system:
			// Bulk create or update if a target system supports batch processing
			// alternatively, check if an item's Id exists and update otherwise create
			fmt.Printf("Upsert %d items\n", len(items))
			return nil
		},
		// Provide a handler to deal with deleted items
		Delete: func(ctx context.Context, ids []int) error {
			// Implement your logic here for your target system:
			// Bulk delete if a target system supports batch processing
			// alternatively, delete item by Id and ignore NotFound error types
			fmt.Printf("Delete %d items\n", len(ids))
			return nil
		},
		// Save the updated entity sync state to a persistent storage
		Persist: func(s *spsync.State) error {
			// The state is used to resume sync session from a previous state.
			// Even when a sync session ended up with errors we need to save it.
			// Some scenarious, e.g. sync in serverless jobs might be designed
			// to terminate in 10-15 minutes and resume on a next run.
			return nil
		},
		// Hook to sync events and use logging middleware of your choice
		Events: &spsync.Events{},
	}

	// Run sync session
	state, err := spsync.Run(ctx, opts)
	if err != nil {
		state.Fails += 1
		// Persist state here as well
		fmt.Println(err)
	}
}

// NewSP constructs SharePoint authenticated API client instance
func NewSP() *api.SP {
	// Simplest strategy for testing
	// for prod the AzureAD auth is recommended:
	// https://go.spflow.com/auth/custom-auth/azure-certificate-auth

	// Create .env with the following variables: SP_SITE_URL, SP_USERNAME, SP_PASSWORD
	// or export the variables in your shell

	c := cpass.Cpass("")
	password, _ := c.Decode(os.Getenv("SP_PASSWORD"))

	auth := &strategy.AuthCnfg{
		SiteURL:  os.Getenv("SP_SITE_URL"),
		Username: os.Getenv("SP_USERNAME"),
		Password: password,
	}

	client := &gosip.SPClient{AuthCnfg: auth}
	return api.NewSP(client)
}

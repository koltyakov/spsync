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
	godotenv.Load()

	ctx := context.Background()

	// Setup sync configuration per a specific entity.
	// Usually you need to sync multiple Lists so most of a logic from this sample
	// iterates for synced entities.
	// In a serverless scenarios, sync session timeout is controlled and terminated
	// by the caller after timeout. It's OK, the resumed session is designed
	// to continue since a previous successful incremental state.

	opts := &spsync.SyncOptns{
		// Create and bind SharePoint API client, see more https://go.spflow.com
		SP: NewSP(),
		// Pass current state from a persistent storage (e.g. database or SharePoint list)
		State: &spsync.SyncState{
			EntID:     "Lists/SPFTSheetsTimeEntries",
			ChngToken: "1;3;b9904727-dd73-4459-8149...37874554928300000;347735073",
			SyncMode:  spsync.IncrSync,
		},
		// Pass configuration for a specific entity, usually stored in config file
		EntConf: &spsync.EntConf{
			Select: []string{"Id", "Title"},
		},
		// Provide a handler to deal with created and updated items
		Upsert: func(ctx context.Context, items []spsync.AbstItem) error {
			// Implement your logic here for your target system:
			// Bulk create or update if a target system supports batch processing
			// alternatively, check if an item's Id exists and update otherwise create
			for _, item := range items {
				fmt.Printf("Upsert %+v\n", item.Data)
			}
			return nil
		},
		// Provide a handler to deal with deleted items
		Delete: func(ctx context.Context, ids []int) error {
			// Implement your logic here for your target system:
			// Bulk delete if a target system supports batch processing
			// alternatively, delete item by Id and ignore NotFound error types
			fmt.Printf("Deletes %+v\n", ids)
			return nil
		},
	}

	// Run sync session
	state, err := spsync.Run(ctx, opts)
	if err != nil {
		state.Fails += 1
		fmt.Println(err)
	}

	// Save the updated entity sync state to a persistent storage.
	// The state is used to resume sync session from a previous state.
	// Even when a sync session ended up with errors we need to save it.
	// Some scenarious, e.g. sync in serverless jobs might be designed
	// to terminate in 10-15 minutes and resume on a next run.

	fmt.Printf("State %+v\n", state)
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
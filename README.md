# spsync

> Go library for robust cloud native SharePoint Lists synchronization or backup

<!--suppress HtmlDeprecatedAttribute -->
<div align="center">
  <img alt="Gosip" src="https://raw.githubusercontent.com/koltyakov/gosip-docs/master/.gitbook/assets/gosip.png" />
</div>

## Entity sync metadata persistent state

| Entity     | Change Token  | Page Token           | Mode | Stage  | Sync Date | Failed Sessions |
| ---------- | ------------- | -------------------- | ---- | ------ | --------- | --------------- |
| List/ListA | 1;3;b9...6db9 | Paged=TRUE&p_ID=6387 | Full | Upsert |           | 0               |

### Synchronization Modes

- `Full`
- `Incr`

### Synchronization Stages

- Upsert
- Delete

## Full synchronization mode

### Full synchronization mode condition(s)

`Mode` == `Full` or a `Change token` is not set.

### Full synchronization mode logic

- On a blank session:
  - save current change token and timestamp (sync start not completion)
  - change mode to `Full`
- On a continue session:
  - start from a previous page skip token (if any)
- Default sort (by ID ascending)
- Paginate by N items
- Save a previous page skip token (on page successfully processed)
- Bulk processing upserts (in a custom hook handler)
- Control time (for serverless scenarios)
- Fail aware and continue logic (continue from page token)
- On completion:

  - update sync date after a successful session
  - clear session page skip token
  - change mode to `Incr`

- Tracking deletions
  - Paginate through all items with only ID selected
  - Using max page size of 5000 items
  - Detect gaps in ID sequence
  - Bulk process deletions (in a custom hook handler)

## Incremental synchronization mode

### Incremental synchronization mode condition(s)

`Mode` == `Incr` and a `Change token` is not empty.

### Incremental synchronization mode logic

- Use Change Token and change API
- Track: Adds, Updates, Deletes, Restores
- Paginate through by N items
- Save a previous page skip token (on page successfully processed)
- Process changes list
  - For added/updated/restored items: request by IDs, and process (in a custom hook handler)
  - Bulk process deletions (in a custom hook handler)
- Save the latest change token from a previously processed page
- Control time (for serverless scenarios)
- By design, a failure is amended on a next incremental run
  - however, failed sessions should be tracked
  - with too many fails - create an incident
- On completion:
  - update sync date after a successful session
  - reset failed sessions counter

## Usage sample

```golang
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

	opts := &spsync.Options{
		// Create and bind SharePoint API client, see more https://go.spflow.com
		SP: NewSP(),
		// Pass current state from a persistent storage (e.g. database or SharePoint list)
		State: &spsync.State{
			EntID:       "Lists/SPFTSheetsTimeEntries",
			ChangeToken: "1;3;b9904727-dd73-4459-8149...37874554928300000;347735073",
			SyncMode:    spsync.Incr,
		},
		// Pass configuration for a specific entity, usually stored in config file
		EntConf: &spsync.EntConf{
			Select: []string{"Id", "Title"},
		},
		// Provide a handler to deal with created and updated items
		Upsert: func(ctx context.Context, items []spsync.ListItem) error {
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
```

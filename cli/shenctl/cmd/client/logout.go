package client

import (
	"fmt"
	"net/http"

	clientutils "github.com/elmq0022/shen/cli/shenctl/cmd/client/utils"
	cmdutils "github.com/elmq0022/shen/cli/shenctl/cmd/utils"
	"github.com/elmq0022/shen/cli/shenctl/utils"
)

func Logout() error {
	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return err
	}

	req, err := utils.NewRequestBuilder(http.MethodPost, "/api/v1/auth/logout").
		WithAuthHeader(authHeader).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build logout request: %w", err)
	}

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("logout failed: server returned status %d", resp.StatusCode)
	}

	if err := cmdutils.ClearSession(); err != nil {
		return fmt.Errorf("logout successful but failed to remove local session file: %w", err)
	}

	return nil
}

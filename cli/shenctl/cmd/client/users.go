package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	cmdutils "github.com/elmq0022/shen/cli/shenctl/cmd/utils"
	"github.com/elmq0022/shen/cli/shenctl/utils"
	db "github.com/elmq0022/shen/db/sqlc"
)

func ListUsers(cursor string, limit int) ([]db.ListUsersRow, error) {
	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	req, err := utils.NewRequestBuilder(http.MethodGet, "/api/v1/users").
		WithHeader("Authorization", authHeader).Build()
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("cursor", cursor)
	query.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = query.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var users []db.ListUsersRow
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users response: %w", err)
	}

	return users, nil
}

func CreateUser() error {
	return nil
}

func UpdateUser() error {
	return nil
}

func DeleteUser() error {
	return nil
}

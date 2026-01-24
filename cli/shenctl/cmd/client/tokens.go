package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	clientutils "github.com/elmq0022/shen/cli/shenctl/cmd/client/utils"
	cmdutils "github.com/elmq0022/shen/cli/shenctl/cmd/utils"
	"github.com/elmq0022/shen/cli/shenctl/utils"
	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers/tokens"
)

func ListTokens(user string, cursor int32, limit int32) ([]db.ListTokensByUserRow, error) {
	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	req, err := utils.NewRequestBuilder(http.MethodGet, "/api/v1/tokens").
		WithAuthHeader(authHeader).Build()
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if cursor > 0 {
		query.Add("cursor", strconv.Itoa(int(cursor)))
	}
	query.Add("limit", strconv.Itoa(int(limit)))
	if user != "" {
		query.Add("user", user)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return nil, fmt.Errorf("forbidden: insufficient permissions to list tokens")
		case http.StatusNotFound:
			return nil, fmt.Errorf("user not found")
		default:
			if errMsg != "" {
				return nil, fmt.Errorf("failed to list tokens (%s): %s", resp.Status, errMsg)
			}
			return nil, fmt.Errorf("unexpected status: %s", resp.Status)
		}
	}

	var tokens []db.ListTokensByUserRow
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode tokens response: %w", err)
	}

	return tokens, nil
}

func CreateToken(name, application, expiration string) (*tokens.CreatePATResponse, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("token name is required")
	}

	application = strings.TrimSpace(application)
	if application == "" {
		return nil, fmt.Errorf("application name is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	endpoint := fmt.Sprintf("/api/v1/tokens/%s/%s", name, application)
	req, err := utils.NewRequestBuilder(http.MethodPost, endpoint).
		WithAuthHeader(authHeader).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	if expiration != "" {
		query := req.URL.Query()
		query.Add("exp", expiration)
		req.URL.RawQuery = query.Encode()
	}

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return nil, fmt.Errorf("forbidden: user not authorized for this application")
		case http.StatusNotFound:
			return nil, fmt.Errorf("application %q not found", application)
		case http.StatusConflict:
			return nil, fmt.Errorf("token %q already exists for application %q", name, application)
		case http.StatusBadRequest:
			if errMsg != "" {
				return nil, fmt.Errorf("invalid request: %s", errMsg)
			}
			return nil, fmt.Errorf("invalid request: check token name and expiration format")
		default:
			if errMsg != "" {
				return nil, fmt.Errorf("failed to create token (%s): %s", resp.Status, errMsg)
			}
			return nil, fmt.Errorf("failed to create token: server returned %s", resp.Status)
		}
	}

	var patResp tokens.CreatePATResponse
	if err := json.NewDecoder(resp.Body).Decode(&patResp); err != nil {
		return nil, fmt.Errorf("failed to parse server response: %w", err)
	}

	return &patResp, nil
}

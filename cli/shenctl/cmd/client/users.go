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
	"github.com/elmq0022/shen/internal/handlers/users"
)

func ListActiveUsers(cursor string, limit int) ([]db.ListActiveUsersRow, error) {
	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	req, err := utils.NewRequestBuilder(http.MethodGet, "/api/v1/users/").
		WithAuthHeader(authHeader).Build()
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("cursor", cursor)
	query.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = query.Encode()

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var users []db.ListActiveUsersRow
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users response: %w", err)
	}

	return users, nil
}

func CreateUser(username, password, role string) (db.CreateUserRow, error) {
	// Input validation
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	role = strings.TrimSpace(role)

	if username == "" {
		return db.CreateUserRow{}, fmt.Errorf("username is required")
	}
	if password == "" && role != "service" {
		return db.CreateUserRow{}, fmt.Errorf("password is required")
	}
	if role == "" {
		return db.CreateUserRow{}, fmt.Errorf("role is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return db.CreateUserRow{}, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodPost, "/api/v1/users/").
		WithAuthHeader(authHeader).
		WithJSON(users.CreateUserRequest{
			UserName: username,
			Password: password,
			Role:     role,
		}).
		Build()
	if err != nil {
		return db.CreateUserRow{}, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return db.CreateUserRow{}, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return db.CreateUserRow{}, fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return db.CreateUserRow{}, fmt.Errorf("forbidden: insufficient permissions to create users")
		case http.StatusConflict:
			return db.CreateUserRow{}, fmt.Errorf("user %q already exists", username)
		case http.StatusBadRequest:
			if errMsg != "" {
				return db.CreateUserRow{}, fmt.Errorf("invalid request: %s", errMsg)
			}
			return db.CreateUserRow{}, fmt.Errorf("invalid request: check username, password, and role values")
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			if errMsg != "" {
				return db.CreateUserRow{}, fmt.Errorf("server error (%s): %s", resp.Status, errMsg)
			}
			return db.CreateUserRow{}, fmt.Errorf("server error (%s): please try again later", resp.Status)
		default:
			if errMsg != "" {
				return db.CreateUserRow{}, fmt.Errorf("failed to create user (%s): %s", resp.Status, errMsg)
			}
			return db.CreateUserRow{}, fmt.Errorf("failed to create user: server returned %s", resp.Status)
		}
	}

	var user db.CreateUserRow
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return db.CreateUserRow{}, fmt.Errorf("failed to parse server response: %w", err)
	}

	return user, nil
}

func UpdateUser(username string, role *string, password *string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if role == nil && password == nil {
		return fmt.Errorf("at least one of role or password must be provided")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodPatch, "/api/v1/users/"+username).
		WithAuthHeader(authHeader).
		WithJSON(users.UpdateUserRequest{
			Role:     role,
			Password: password,
		}).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return fmt.Errorf("forbidden: %s", errMsg)
		case http.StatusNotFound:
			return fmt.Errorf("user %q not found", username)
		case http.StatusBadRequest:
			if errMsg != "" {
				return fmt.Errorf("invalid request: %s", errMsg)
			}
			return fmt.Errorf("invalid request: check role value")
		default:
			if errMsg != "" {
				return fmt.Errorf("failed to update user (%s): %s", resp.Status, errMsg)
			}
			return fmt.Errorf("failed to update user: server returned %s", resp.Status)
		}
	}

	return nil
}

func DeleteUser(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodDelete, "/api/v1/users/"+username).
		WithAuthHeader(authHeader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return fmt.Errorf("forbidden: insufficient permissions to delete users")
		case http.StatusNotFound:
			return fmt.Errorf("user %q not found", username)
		case http.StatusBadRequest:
			if errMsg != "" {
				return fmt.Errorf("invalid request: %s", errMsg)
			}
			return fmt.Errorf("invalid request: username is required")
		default:
			if errMsg != "" {
				return fmt.Errorf("failed to delete user (%s): %s", resp.Status, errMsg)
			}
			return fmt.Errorf("failed to delete user: server returned %s", resp.Status)
		}
	}

	return nil
}

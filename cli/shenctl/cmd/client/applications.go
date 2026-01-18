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
	"github.com/elmq0022/shen/internal/handlers/applications"
)

func ListActiveApplications(cursor string, limit int) ([]db.ShenApplication, error) {
	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	req, err := utils.NewRequestBuilder(http.MethodGet, "/api/v1/applications/").
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
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return nil, fmt.Errorf("forbidden: insufficient permissions to list applications")
		default:
			if errMsg != "" {
				return nil, fmt.Errorf("failed to list applications (%s): %s", resp.Status, errMsg)
			}
			return nil, fmt.Errorf("unexpected status: %s", resp.Status)
		}
	}

	var apps []db.ShenApplication
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, fmt.Errorf("failed to decode applications response: %w", err)
	}

	return apps, nil
}

func CreateApplication(name string) (db.ShenApplication, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.ShenApplication{}, fmt.Errorf("application name is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return db.ShenApplication{}, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodPost, "/api/v1/applications/").
		WithAuthHeader(authHeader).
		WithJSON(applications.CreateApplicationRequest{
			Name: name,
		}).
		Build()
	if err != nil {
		return db.ShenApplication{}, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return db.ShenApplication{}, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return db.ShenApplication{}, fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return db.ShenApplication{}, fmt.Errorf("forbidden: insufficient permissions to create applications")
		case http.StatusConflict:
			return db.ShenApplication{}, fmt.Errorf("application %q already exists", name)
		case http.StatusBadRequest:
			if errMsg != "" {
				return db.ShenApplication{}, fmt.Errorf("invalid request: %s", errMsg)
			}
			return db.ShenApplication{}, fmt.Errorf("invalid request: check application name")
		default:
			if errMsg != "" {
				return db.ShenApplication{}, fmt.Errorf("failed to create application (%s): %s", resp.Status, errMsg)
			}
			return db.ShenApplication{}, fmt.Errorf("failed to create application: server returned %s", resp.Status)
		}
	}

	var app db.ShenApplication
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return db.ShenApplication{}, fmt.Errorf("failed to parse server response: %w", err)
	}

	return app, nil
}

func DeleteApplication(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("application name is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodDelete, "/api/v1/applications/"+name).
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
			return fmt.Errorf("forbidden: insufficient permissions to delete applications")
		case http.StatusNotFound:
			return fmt.Errorf("application %q not found", name)
		case http.StatusBadRequest:
			if errMsg != "" {
				return fmt.Errorf("invalid request: %s", errMsg)
			}
			return fmt.Errorf("invalid request: application name is required")
		default:
			if errMsg != "" {
				return fmt.Errorf("failed to delete application (%s): %s", resp.Status, errMsg)
			}
			return fmt.Errorf("failed to delete application: server returned %s", resp.Status)
		}
	}

	return nil
}

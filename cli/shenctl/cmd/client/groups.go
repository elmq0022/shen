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
	"github.com/elmq0022/shen/internal/handlers/groups"
)

func ListActiveGroups(cursor string, limit int) ([]db.ShenGroup, error) {
	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	req, err := utils.NewRequestBuilder(http.MethodGet, "/api/v1/groups/").
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
			return nil, fmt.Errorf("forbidden: insufficient permissions to list groups")
		default:
			if errMsg != "" {
				return nil, fmt.Errorf("failed to list groups (%s): %s", resp.Status, errMsg)
			}
			return nil, fmt.Errorf("unexpected status: %s", resp.Status)
		}
	}

	var groupList []db.ShenGroup
	if err := json.NewDecoder(resp.Body).Decode(&groupList); err != nil {
		return nil, fmt.Errorf("failed to decode groups response: %w", err)
	}

	return groupList, nil
}

func CreateGroup(name string) (db.ShenGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.ShenGroup{}, fmt.Errorf("group name is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return db.ShenGroup{}, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodPost, "/api/v1/groups/").
		WithAuthHeader(authHeader).
		WithJSON(groups.CreateGroupRequest{
			Name: name,
		}).
		Build()
	if err != nil {
		return db.ShenGroup{}, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := clientutils.DefaultClient.Do(req)
	if err != nil {
		return db.ShenGroup{}, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errMsg := clientutils.ReadErrorBody(resp)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return db.ShenGroup{}, fmt.Errorf("unauthorized: invalid or expired credentials")
		case http.StatusForbidden:
			return db.ShenGroup{}, fmt.Errorf("forbidden: insufficient permissions to create groups")
		case http.StatusConflict:
			return db.ShenGroup{}, fmt.Errorf("group %q already exists", name)
		case http.StatusBadRequest:
			if errMsg != "" {
				return db.ShenGroup{}, fmt.Errorf("invalid request: %s", errMsg)
			}
			return db.ShenGroup{}, fmt.Errorf("invalid request: check group name")
		default:
			if errMsg != "" {
				return db.ShenGroup{}, fmt.Errorf("failed to create group (%s): %s", resp.Status, errMsg)
			}
			return db.ShenGroup{}, fmt.Errorf("failed to create group: server returned %s", resp.Status)
		}
	}

	var group db.ShenGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return db.ShenGroup{}, fmt.Errorf("failed to parse server response: %w", err)
	}

	return group, nil
}

func DeleteGroup(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("group name is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodDelete, "/api/v1/groups/"+name).
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
			return fmt.Errorf("forbidden: insufficient permissions to delete groups")
		case http.StatusNotFound:
			return fmt.Errorf("group %q not found", name)
		case http.StatusBadRequest:
			if errMsg != "" {
				return fmt.Errorf("invalid request: %s", errMsg)
			}
			return fmt.Errorf("invalid request: group name is required")
		default:
			if errMsg != "" {
				return fmt.Errorf("failed to delete group (%s): %s", resp.Status, errMsg)
			}
			return fmt.Errorf("failed to delete group: server returned %s", resp.Status)
		}
	}

	return nil
}

// GroupMember represents a member in a group
type GroupMember struct {
	Username string `json:"username"`
}

func ListGroupMembers(groupName string, cursor string, limit int) ([]GroupMember, error) {
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		return nil, fmt.Errorf("group name is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	req, err := utils.NewRequestBuilder(http.MethodGet, "/api/v1/groups/"+groupName+"/members").
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
			return nil, fmt.Errorf("forbidden: insufficient permissions to list group members")
		case http.StatusNotFound:
			return nil, fmt.Errorf("group %q not found", groupName)
		default:
			if errMsg != "" {
				return nil, fmt.Errorf("failed to list group members (%s): %s", resp.Status, errMsg)
			}
			return nil, fmt.Errorf("unexpected status: %s", resp.Status)
		}
	}

	var members []GroupMember
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode group members response: %w", err)
	}

	return members, nil
}

// AddUsersResponse represents the response from adding users to a group
type AddUsersResponse struct {
	Added    []string `json:"added"`
	NotFound []string `json:"not_found"`
}

func AddUsersToGroup(groupName string, usernames []string) (*AddUsersResponse, error) {
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if len(usernames) == 0 {
		return nil, fmt.Errorf("at least one username is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodPost, "/api/v1/groups/"+groupName+"/members").
		WithAuthHeader(authHeader).
		WithJSON(groups.MembersRequest{
			Usernames: usernames,
		}).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
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
			return nil, fmt.Errorf("forbidden: insufficient permissions to add users to group")
		case http.StatusNotFound:
			return nil, fmt.Errorf("group %q not found or no users found", groupName)
		case http.StatusBadRequest:
			if errMsg != "" {
				return nil, fmt.Errorf("invalid request: %s", errMsg)
			}
			return nil, fmt.Errorf("invalid request")
		default:
			if errMsg != "" {
				return nil, fmt.Errorf("failed to add users to group (%s): %s", resp.Status, errMsg)
			}
			return nil, fmt.Errorf("failed to add users to group: server returned %s", resp.Status)
		}
	}

	var result AddUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse server response: %w", err)
	}

	return &result, nil
}

// RemoveUsersResponse represents the response from removing users from a group
type RemoveUsersResponse struct {
	Removed  []string `json:"removed"`
	NotFound []string `json:"not_found"`
}

func RemoveUsersFromGroup(groupName string, usernames []string) (*RemoveUsersResponse, error) {
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if len(usernames) == 0 {
		return nil, fmt.Errorf("at least one username is required")
	}

	authHeader, err := cmdutils.GetAuthHeader()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodDelete, "/api/v1/groups/"+groupName+"/members").
		WithAuthHeader(authHeader).
		WithJSON(groups.MembersRequest{
			Usernames: usernames,
		}).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
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
			return nil, fmt.Errorf("forbidden: insufficient permissions to remove users from group")
		case http.StatusNotFound:
			return nil, fmt.Errorf("group %q not found or no users found", groupName)
		case http.StatusBadRequest:
			if errMsg != "" {
				return nil, fmt.Errorf("invalid request: %s", errMsg)
			}
			return nil, fmt.Errorf("invalid request")
		default:
			if errMsg != "" {
				return nil, fmt.Errorf("failed to remove users from group (%s): %s", resp.Status, errMsg)
			}
			return nil, fmt.Errorf("failed to remove users from group: server returned %s", resp.Status)
		}
	}

	var result RemoveUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse server response: %w", err)
	}

	return &result, nil
}

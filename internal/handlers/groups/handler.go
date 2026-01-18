package groups

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewHandler(pool *pgxpool.Pool, queries *db.Queries) Handler {
	return Handler{pool: pool, queries: queries}
}

func (h *Handler) CreateGroup(c echo.Context) error {
	var req CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	// Normalize name to lowercase
	name := strings.ToLower(strings.TrimSpace(req.Name))
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("group name is required"))
	}

	group, err := h.queries.CreateGroup(c.Request().Context(), name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.JSON(http.StatusConflict, handlers.NewErrorResponse("group already exists"))
		}
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to create group"))
	}

	return c.JSON(http.StatusCreated, group)
}

func (h *Handler) ListGroups(c echo.Context) error {
	cursor := c.QueryParam("cursor")
	limitStr := c.QueryParam("limit")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid limit parameter"))
	}

	groups, err := h.queries.ListActiveGroups(c.Request().Context(), db.ListActiveGroupsParams{
		Column1: cursor,
		Limit:   int32(limit),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to list groups"))
	}

	return c.JSON(http.StatusOK, groups)
}

func (h *Handler) DeleteGroup(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no group name provided"))
	}

	group, err := h.queries.GetGroupByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group does not exist"))
	}

	if err := h.queries.DeactivateGroup(c.Request().Context(), group.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to deactivate group"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListGroupMembers(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no group name provided"))
	}

	group, err := h.queries.GetGroupByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group does not exist"))
	}

	cursor := c.QueryParam("cursor")
	limitStr := c.QueryParam("limit")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid limit parameter"))
	}

	users, err := h.queries.ListUsersByGroup(c.Request().Context(), db.ListUsersByGroupParams{
		GroupID: group.ID,
		Column2: cursor,
		Limit:   int32(limit),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to list group members"))
	}

	// Convert to response format with only username
	members := make([]GroupMemberResponse, len(users))
	for i, user := range users {
		members[i] = GroupMemberResponse{Username: user.Username}
	}

	return c.JSON(http.StatusOK, members)
}

func (h *Handler) AddGroupMembers(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no group name provided"))
	}

	group, err := h.queries.GetGroupByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group does not exist"))
	}

	var req MembersRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	if len(req.Usernames) == 0 {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("at least one username is required"))
	}

	var addedUsers []string
	var notFoundUsers []string

	for _, username := range req.Usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}

		user, err := h.queries.GetUserByUsername(c.Request().Context(), username)
		if err != nil {
			notFoundUsers = append(notFoundUsers, username)
			continue
		}

		_, err = h.queries.AddUserToGroup(c.Request().Context(), db.AddUserToGroupParams{
			UserID:  user.ID,
			GroupID: group.ID,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			// Ignore duplicate membership errors (unique constraint violation)
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				addedUsers = append(addedUsers, username)
				continue
			}
			return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to add user to group"))
		}
		addedUsers = append(addedUsers, username)
	}

	if len(notFoundUsers) > 0 && len(addedUsers) == 0 {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("none of the specified users were found"))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"added":     addedUsers,
		"not_found": notFoundUsers,
	})
}

func (h *Handler) RemoveGroupMembers(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no group name provided"))
	}

	group, err := h.queries.GetGroupByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group does not exist"))
	}

	var req MembersRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	if len(req.Usernames) == 0 {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("at least one username is required"))
	}

	var removedUsers []string
	var notFoundUsers []string

	for _, username := range req.Usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}

		user, err := h.queries.GetUserByUsername(c.Request().Context(), username)
		if err != nil {
			notFoundUsers = append(notFoundUsers, username)
			continue
		}

		err = h.queries.RemoveUserFromGroup(c.Request().Context(), db.RemoveUserFromGroupParams{
			UserID:  user.ID,
			GroupID: group.ID,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to remove user from group"))
		}
		removedUsers = append(removedUsers, username)
	}

	if len(notFoundUsers) > 0 && len(removedUsers) == 0 {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("none of the specified users were found"))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"removed":   removedUsers,
		"not_found": notFoundUsers,
	})
}

func (h *Handler) AddGroupRole(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no group name provided"))
	}

	group, err := h.queries.GetGroupByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group not found"))
	}

	if !group.Active {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group not found"))
	}

	var req GroupRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	appName := strings.ToLower(strings.TrimSpace(req.Application))
	roleName := strings.ToLower(strings.TrimSpace(req.Role))

	if appName == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("application name is required"))
	}
	if roleName == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("role name is required"))
	}

	// Validate application exists and is active
	app, err := h.queries.GetApplicationByName(c.Request().Context(), appName)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("application not found"))
	}
	if !app.Active {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("application not found"))
	}

	// Validate role exists (from seeded values)
	role, err := h.queries.GetApplicationRoleByName(c.Request().Context(), roleName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid role: must be one of authenticated, viewer, auditor, operator, admin"))
	}

	// Add the role assignment
	_, err = h.queries.AddGroupApplicationRole(c.Request().Context(), db.AddGroupApplicationRoleParams{
		GroupID:       group.ID,
		ApplicationID: app.ID,
		RoleID:        role.ID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		// Handle duplicate - return success (idempotent)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.JSON(http.StatusCreated, GroupRoleResponse{
				Application: appName,
				Role:        roleName,
			})
		}
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to add role to group"))
	}

	return c.JSON(http.StatusCreated, GroupRoleResponse{
		Application: appName,
		Role:        roleName,
	})
}

func (h *Handler) RemoveGroupRole(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no group name provided"))
	}

	group, err := h.queries.GetGroupByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group not found"))
	}

	var req GroupRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	appName := strings.ToLower(strings.TrimSpace(req.Application))
	roleName := strings.ToLower(strings.TrimSpace(req.Role))

	if appName == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("application name is required"))
	}
	if roleName == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("role name is required"))
	}

	// Validate application exists
	app, err := h.queries.GetApplicationByName(c.Request().Context(), appName)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("application not found"))
	}

	// Validate role exists
	role, err := h.queries.GetApplicationRoleByName(c.Request().Context(), roleName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid role: must be one of authenticated, viewer, auditor, operator, admin"))
	}

	// Delete the role assignment
	err = h.queries.DeleteGroupApplicationRole(c.Request().Context(), db.DeleteGroupApplicationRoleParams{
		GroupID:       group.ID,
		ApplicationID: app.ID,
		RoleID:        role.ID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to remove role from group"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListGroupRoles(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no group name provided"))
	}

	group, err := h.queries.GetGroupByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group not found"))
	}

	if !group.Active {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("group not found"))
	}

	appFilter := strings.ToLower(strings.TrimSpace(c.QueryParam("application")))
	cursorApp := c.QueryParam("cursor_app")
	cursorRole := c.QueryParam("cursor_role")
	limitStr := c.QueryParam("limit")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid limit parameter"))
	}

	// If application filter is provided, use the filtered query
	if appFilter != "" {
		app, err := h.queries.GetApplicationByName(c.Request().Context(), appFilter)
		if err != nil {
			return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("application not found"))
		}

		roles, err := h.queries.ListGroupApplicationRolesByGroupAndApplication(c.Request().Context(), db.ListGroupApplicationRolesByGroupAndApplicationParams{
			GroupID:        group.ID,
			ApplicationID:  app.ID,
			CursorRoleName: cursorRole,
			Limit:          int32(limit),
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to list group roles"))
		}

		// Convert to response format
		response := make([]GroupRoleResponse, len(roles))
		for i, role := range roles {
			response[i] = GroupRoleResponse{
				Application: appFilter,
				Role:        role.RoleName,
			}
		}
		return c.JSON(http.StatusOK, response)
	}

	// No application filter - list all roles for the group
	roles, err := h.queries.ListGroupApplicationRolesByGroup(c.Request().Context(), db.ListGroupApplicationRolesByGroupParams{
		GroupID:               group.ID,
		CursorApplicationName: cursorApp,
		CursorRoleName:        cursorRole,
		Limit:                 int32(limit),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to list group roles"))
	}

	// Convert to response format
	response := make([]GroupRoleResponse, len(roles))
	for i, role := range roles {
		response[i] = GroupRoleResponse{
			Application: role.ApplicationName,
			Role:        role.RoleName,
		}
	}
	return c.JSON(http.StatusOK, response)
}

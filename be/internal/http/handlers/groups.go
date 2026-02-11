package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

// Request structs

type CreateGroupRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	Description  string `json:"description" validate:"max=500"`
	CurrencyCode string `json:"currency_code" validate:"omitempty,len=3"`
}

// Response structs

type CreateGroupResponse struct {
	ID           pgtype.UUID `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	CurrencyCode string      `json:"currency_code"`
	CreatedAt    string      `json:"created_at"`
	Role         string      `json:"role"`
}

type GroupMemberResponse struct {
	ID        pgtype.UUID `json:"id"`
	GroupID   pgtype.UUID `json:"group_id"`
	UserID    pgtype.UUID `json:"user_id"`
	Role      string      `json:"role"`
	Status    string      `json:"status"`
	InvitedAt string      `json:"invited_at,omitempty"`
	JoinedAt  string      `json:"joined_at,omitempty"`
}

type GroupMemberWithUserResponse struct {
	ID              pgtype.UUID `json:"id"`
	GroupID         pgtype.UUID `json:"group_id"`
	UserID          pgtype.UUID `json:"user_id"`
	InvitationToken string      `json:"invitation_token,omitempty"`
	Role            string      `json:"role"`
	Status          string      `json:"status"`
	InvitedAt       string      `json:"invited_at,omitempty"`
	JoinedAt        string      `json:"joined_at,omitempty"`
	User            UserInfo    `json:"user"`
}

type UserInfo struct {
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UserGroupResponse struct {
	ID             pgtype.UUID `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description,omitempty"`
	CurrencyCode   string      `json:"currency_code"`
	CreatedAt      string      `json:"created_at"`
	MembershipID   pgtype.UUID `json:"membership_id"`
	MemberRole     string      `json:"member_role"`
	MemberStatus   string      `json:"member_status"`
	MemberJoinedAt string      `json:"member_joined_at,omitempty"`
}

// Handlers

func CreateGroupHandler(groupService service.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateGroupRequest](r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.request.context_invalid", "Invalid request context.")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		result, err := groupService.CreateGroup(r.Context(), service.CreateGroupInput{
			Name:         req.Name,
			Description:  req.Description,
			CurrencyCode: req.CurrencyCode,
			CreatedBy:    userID,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			code := "system.group.create_failed"
			message := "Unable to create group."
			switch err {
			case service.ErrInvalidGroupName:
				statusCode = http.StatusUnprocessableEntity
				code = "validation.group.name.invalid"
				message = "Group name is required."
			}
			response.SendErrorWithCode(w, statusCode, code, message)
			return
		}

		resp := CreateGroupResponse{
			ID:           result.Group.ID,
			Name:         result.Group.Name,
			Description:  result.Group.Description.String,
			CurrencyCode: result.Group.CurrencyCode,
			CreatedAt:    result.Group.CreatedAt.Time.String(),
			Role:         result.Membership.Role,
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func ListGroupMembersHandler(groupService service.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.group.group_id.invalid", "Invalid group id.")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		members, err := groupService.ListGroupMembers(r.Context(), groupID, userID)
		if err != nil {
			statusCode := http.StatusBadRequest
			code := "system.group.members.list_failed"
			message := "Unable to load group members."
			switch err {
			case service.ErrGroupNotFound:
				statusCode = http.StatusNotFound
				code = "resource.group.not_found"
				message = "Group not found."
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
				code = "permission.group.member_required"
				message = "You are not a member of this group."
			}
			response.SendErrorWithCode(w, statusCode, code, message)
			return
		}

		// Merge results
		resp := make([]GroupMemberWithUserResponse, len(members))
		for i, m := range members {
			// Map user details
			user := UserInfo{
				Email:     m.UserEmail,
				Name:      m.UserName.String,
				AvatarURL: m.UserAvatarUrl.String,
			}

			if m.IsPending {
				// If we have a pending user ID, merge it with UserId
				if m.PendingUserID.Valid {
					m.UserID = m.PendingUserID
				}

				// If we have a name for the pending user (from invitation input), use it
				if m.UserName.Valid {
					user.Name = m.UserName.String
				}
			}

			resp[i] = GroupMemberWithUserResponse{
				ID:              m.ID,
				GroupID:         m.GroupID,
				UserID:          m.UserID,
				InvitationToken: m.InvitationToken,
				Role:            m.Role,
				Status:          m.Status,
				InvitedAt:       formatTimestamp(m.InvitedAt),
				JoinedAt:        formatTimestamp(m.JoinedAt),
				User:            user,
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func ListUserGroupsHandler(groupService service.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		groups, err := groupService.ListUserGroups(r.Context(), userID)
		if err != nil {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.group.list_failed", "Unable to load groups.")
			return
		}

		resp := make([]UserGroupResponse, len(groups))
		for i, g := range groups {
			resp[i] = UserGroupResponse{
				ID:             g.ID,
				Name:           g.Name,
				Description:    g.Description.String,
				CurrencyCode:   g.CurrencyCode,
				CreatedAt:      formatTimestamp(g.CreatedAt),
				MembershipID:   g.MembershipID,
				MemberRole:     g.MemberRole,
				MemberStatus:   g.MemberStatus,
				MemberJoinedAt: formatTimestamp(g.MemberJoinedAt),
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

// Helper to convert UUID to string
func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return formatUUID(uuid.Bytes)
}

func formatUUID(b [16]byte) string {
	return string([]byte{
		hexDigit(b[0] >> 4), hexDigit(b[0]),
		hexDigit(b[1] >> 4), hexDigit(b[1]),
		hexDigit(b[2] >> 4), hexDigit(b[2]),
		hexDigit(b[3] >> 4), hexDigit(b[3]),
		'-',
		hexDigit(b[4] >> 4), hexDigit(b[4]),
		hexDigit(b[5] >> 4), hexDigit(b[5]),
		'-',
		hexDigit(b[6] >> 4), hexDigit(b[6]),
		hexDigit(b[7] >> 4), hexDigit(b[7]),
		'-',
		hexDigit(b[8] >> 4), hexDigit(b[8]),
		hexDigit(b[9] >> 4), hexDigit(b[9]),
		'-',
		hexDigit(b[10] >> 4), hexDigit(b[10]),
		hexDigit(b[11] >> 4), hexDigit(b[11]),
		hexDigit(b[12] >> 4), hexDigit(b[12]),
		hexDigit(b[13] >> 4), hexDigit(b[13]),
		hexDigit(b[14] >> 4), hexDigit(b[14]),
		hexDigit(b[15] >> 4), hexDigit(b[15]),
	})
}

func hexDigit(b byte) byte {
	b = b & 0x0f
	if b < 10 {
		return '0' + b
	}
	return 'a' + b - 10
}

// Helper functions

func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}

func formatTimestamp(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.String()
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type ActivityResponse struct {
	ID         pgtype.UUID            `json:"id"`
	GroupID    pgtype.UUID            `json:"group_id"`
	UserID     pgtype.UUID            `json:"user_id"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   pgtype.UUID            `json:"entity_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  string                 `json:"created_at"`
	User       UserInfo               `json:"user,omitempty"`
}

func ListGroupActivitiesHandler(activityService service.GroupActivityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		// Pagination
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := int32(20)
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = int32(l)
			}
		}

		offset := int32(0)
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = int32(o)
			}
		}

		activities, err := activityService.ListGroupActivities(r.Context(), groupID, limit, offset)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := make([]ActivityResponse, len(activities))
		for i, a := range activities {
			var metadata map[string]interface{}
			if len(a.Metadata) > 0 {
				_ = json.Unmarshal(a.Metadata, &metadata)
			}

			resp[i] = ActivityResponse{
				ID:         a.ID,
				GroupID:    a.GroupID,
				UserID:     a.UserID,
				Action:     a.Action,
				EntityType: a.EntityType,
				EntityID:   a.EntityID,
				Metadata:   metadata,
				CreatedAt:  formatTimestamp(a.CreatedAt),
				User: UserInfo{
					Email:     a.UserEmail,
					Name:      a.UserName.String,
					AvatarURL: a.UserAvatarUrl.String,
				},
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func GetExpenseHistoryHandler(activityService service.GroupActivityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expenseID, err := parseUUID(chi.URLParam(r, "expense_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid expense_id")
			return
		}

		activities, err := activityService.GetExpenseHistory(r.Context(), expenseID)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := make([]ActivityResponse, len(activities))
		for i, a := range activities {
			var metadata map[string]interface{}
			if len(a.Metadata) > 0 {
				_ = json.Unmarshal(a.Metadata, &metadata)
			}

			resp[i] = ActivityResponse{
				ID:         a.ID,
				GroupID:    a.GroupID,
				UserID:     a.UserID,
				Action:     a.Action,
				EntityType: a.EntityType,
				EntityID:   a.EntityID,
				Metadata:   metadata,
				CreatedAt:  formatTimestamp(a.CreatedAt),
				User: UserInfo{
					Email:     a.UserEmail,
					Name:      a.UserName.String,
					AvatarURL: a.UserAvatarUrl.String,
				},
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

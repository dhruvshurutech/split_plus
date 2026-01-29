package handlers

import (
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateCategoryRequest struct {
	Name  string `json:"name" validate:"required"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name" validate:"required"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type CreateFromPresetsRequest struct {
	PresetSlugs []string `json:"preset_slugs" validate:"required,min=1"`
}

type CategoryResponse struct {
	ID        string `json:"id"`
	GroupID   string `json:"group_id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Icon      string `json:"icon,omitempty"`
	Color     string `json:"color,omitempty"`
	CreatedBy string `json:"created_by"`
	UpdatedBy string `json:"updated_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CategoryPresetResponse struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

func GetCategoryPresetsHandler(categoryService service.ExpenseCategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		presets := categoryService.GetCategoryPresets()

		resp := make([]CategoryPresetResponse, len(presets))
		for i, p := range presets {
			resp[i] = CategoryPresetResponse{
				Slug:  p.Slug,
				Name:  p.Name,
				Icon:  p.Icon,
				Color: p.Color,
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func ListGroupCategoriesHandler(categoryService service.ExpenseCategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		categories, err := categoryService.ListCategoriesForGroup(r.Context(), groupID, userID)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if err == service.ErrNotGroupMember {
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := make([]CategoryResponse, len(categories))
		for i, cat := range categories {
			resp[i] = CategoryResponse{
				ID:        uuid.UUID(cat.ID.Bytes).String(),
				GroupID:   uuid.UUID(cat.GroupID.Bytes).String(),
				Slug:      cat.Slug,
				Name:      cat.Name,
				Icon:      cat.Icon.String,
				Color:     cat.Color.String,
				CreatedBy: uuid.UUID(cat.CreatedBy.Bytes).String(),
				UpdatedBy: uuid.UUID(cat.UpdatedBy.Bytes).String(),
				CreatedAt: cat.CreatedAt.Time.String(),
				UpdatedAt: cat.UpdatedAt.Time.String(),
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func CreateGroupCategoryHandler(categoryService service.ExpenseCategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateCategoryRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		input := service.CreateGroupCategoryInput{
			GroupID:   groupID,
			Name:      req.Name,
			Icon:      req.Icon,
			Color:     req.Color,
			CreatedBy: userID,
		}

		category, err := categoryService.CreateCategoryForGroup(r.Context(), input)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrInvalidCategoryName:
				statusCode = http.StatusBadRequest
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrCategoryAlreadyExists:
				statusCode = http.StatusConflict
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := CategoryResponse{
			ID:        uuid.UUID(category.ID.Bytes).String(),
			GroupID:   uuid.UUID(category.GroupID.Bytes).String(),
			Slug:      category.Slug,
			Name:      category.Name,
			Icon:      category.Icon.String,
			Color:     category.Color.String,
			CreatedBy: uuid.UUID(category.CreatedBy.Bytes).String(),
			UpdatedBy: uuid.UUID(category.UpdatedBy.Bytes).String(),
			CreatedAt: category.CreatedAt.Time.String(),
			UpdatedAt: category.UpdatedAt.Time.String(),
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func CreateCategoriesFromPresetsHandler(categoryService service.ExpenseCategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateFromPresetsRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		categories, err := categoryService.CreateCategoriesFromPresets(r.Context(), groupID, userID, req.PresetSlugs)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if err == service.ErrNotGroupMember {
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := make([]CategoryResponse, len(categories))
		for i, cat := range categories {
			resp[i] = CategoryResponse{
				ID:        uuid.UUID(cat.ID.Bytes).String(),
				GroupID:   uuid.UUID(cat.GroupID.Bytes).String(),
				Slug:      cat.Slug,
				Name:      cat.Name,
				Icon:      cat.Icon.String,
				Color:     cat.Color.String,
				CreatedBy: uuid.UUID(cat.CreatedBy.Bytes).String(),
				UpdatedBy: uuid.UUID(cat.UpdatedBy.Bytes).String(),
				CreatedAt: cat.CreatedAt.Time.String(),
				UpdatedAt: cat.UpdatedAt.Time.String(),
			}
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func UpdateGroupCategoryHandler(categoryService service.ExpenseCategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[UpdateCategoryRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		categoryID, err := parseUUID(chi.URLParam(r, "id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid category ID")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		input := service.UpdateCategoryInput{
			CategoryID: categoryID,
			GroupID:    groupID,
			Name:       req.Name,
			Icon:       req.Icon,
			Color:      req.Color,
			UpdatedBy:  userID,
		}

		category, err := categoryService.UpdateCategory(r.Context(), input)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrCategoryNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrInvalidCategoryName:
				statusCode = http.StatusBadRequest
			case service.ErrCategoryAlreadyExists:
				statusCode = http.StatusConflict
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := CategoryResponse{
			ID:        uuid.UUID(category.ID.Bytes).String(),
			GroupID:   uuid.UUID(category.GroupID.Bytes).String(),
			Slug:      category.Slug,
			Name:      category.Name,
			Icon:      category.Icon.String,
			Color:     category.Color.String,
			CreatedBy: uuid.UUID(category.CreatedBy.Bytes).String(),
			UpdatedBy: uuid.UUID(category.UpdatedBy.Bytes).String(),
			CreatedAt: category.CreatedAt.Time.String(),
			UpdatedAt: category.UpdatedAt.Time.String(),
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func DeleteGroupCategoryHandler(categoryService service.ExpenseCategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		categoryID, err := parseUUID(chi.URLParam(r, "id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid category ID")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		err = categoryService.DeleteCategory(r.Context(), categoryID, groupID, userID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrCategoryNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		response.SendSuccess(w, http.StatusNoContent, struct{}{})
	}
}

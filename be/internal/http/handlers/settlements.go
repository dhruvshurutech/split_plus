package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

// Request structs

type CreateSettlementRequest struct {
	PayerID              string `json:"payer_id" validate:"omitempty,uuid"`
	PayerPendingUserID   string `json:"payer_pending_user_id" validate:"omitempty,uuid"`
	PayeeID              string `json:"payee_id" validate:"omitempty,uuid"`
	PayeePendingUserID   string `json:"payee_pending_user_id" validate:"omitempty,uuid"`
	Amount               string `json:"amount" validate:"required"`
	CurrencyCode         string `json:"currency_code" validate:"omitempty,len=3"`
	Status               string `json:"status" validate:"omitempty,oneof=pending completed cancelled"`
	PaymentMethod        string `json:"payment_method" validate:"max=50"`
	TransactionReference string `json:"transaction_reference" validate:"max=100"`
	Notes                string `json:"notes" validate:"max=500"`
}

type UpdateSettlementRequest struct {
	Amount               string `json:"amount" validate:"required"`
	CurrencyCode         string `json:"currency_code" validate:"omitempty,len=3"`
	Status               string `json:"status" validate:"required,oneof=pending completed cancelled"`
	PaymentMethod        string `json:"payment_method" validate:"max=50"`
	TransactionReference string `json:"transaction_reference" validate:"max=100"`
	Notes                string `json:"notes" validate:"max=500"`
}

type UpdateSettlementStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending completed cancelled"`
}

// Response structs

type SettlementResponse struct {
	ID                   pgtype.UUID `json:"id"`
	GroupID              pgtype.UUID `json:"group_id"`
	PayerID              pgtype.UUID `json:"payer_id"`
	PayerPendingUserID   pgtype.UUID `json:"payer_pending_user_id"`
	PayeeID              pgtype.UUID `json:"payee_id"`
	PayeePendingUserID   pgtype.UUID `json:"payee_pending_user_id"`
	Amount               string      `json:"amount"`
	CurrencyCode         string      `json:"currency_code"`
	Status               string      `json:"status"`
	PaymentMethod        string      `json:"payment_method,omitempty"`
	TransactionReference string      `json:"transaction_reference,omitempty"`
	PaidAt               string      `json:"paid_at,omitempty"`
	Notes                string      `json:"notes,omitempty"`
	CreatedAt            string      `json:"created_at"`
	CreatedBy            pgtype.UUID `json:"created_by"`
	UpdatedAt            string      `json:"updated_at"`
	UpdatedBy            pgtype.UUID `json:"updated_by"`
}

type SettlementWithUsersResponse struct {
	ID                   pgtype.UUID               `json:"id"`
	GroupID              pgtype.UUID               `json:"group_id"`
	PayerID              pgtype.UUID               `json:"payer_id"`
	PayerPendingUserID   pgtype.UUID               `json:"payer_pending_user_id"`
	PayeeID              pgtype.UUID               `json:"payee_id"`
	PayeePendingUserID   pgtype.UUID               `json:"payee_pending_user_id"`
	Amount               string                    `json:"amount"`
	CurrencyCode         string                    `json:"currency_code"`
	Status               string                    `json:"status"`
	PaymentMethod        string                    `json:"payment_method,omitempty"`
	TransactionReference string                    `json:"transaction_reference,omitempty"`
	PaidAt               string                    `json:"paid_at,omitempty"`
	Notes                string                    `json:"notes,omitempty"`
	CreatedAt            string                    `json:"created_at"`
	CreatedBy            pgtype.UUID               `json:"created_by"`
	UpdatedAt            string                    `json:"updated_at"`
	UpdatedBy            pgtype.UUID               `json:"updated_by"`
	Payer                SettlementParticipantInfo `json:"payer"`
	Payee                SettlementParticipantInfo `json:"payee"`
}

type SettlementParticipantInfo struct {
	UserID        pgtype.UUID `json:"user_id"`
	PendingUserID pgtype.UUID `json:"pending_user_id"`
	Email         string      `json:"email"`
	Name          string      `json:"name,omitempty"`
	AvatarURL     string      `json:"avatar_url,omitempty"`
	IsPending     bool        `json:"is_pending"`
}

// Helper to convert settlement to response
func settlementToResponse(s sqlc.Settlement) SettlementResponse {
	amount := "0"
	if s.Amount.Valid {
		if val, err := s.Amount.Value(); err == nil {
			if str, ok := val.(string); ok {
				amount = str
			}
		}
	}

	paymentMethod := ""
	if s.PaymentMethod.Valid {
		paymentMethod = s.PaymentMethod.String
	}

	transactionReference := ""
	if s.TransactionReference.Valid {
		transactionReference = s.TransactionReference.String
	}

	notes := ""
	if s.Notes.Valid {
		notes = s.Notes.String
	}

	paidAt := ""
	if s.PaidAt.Valid {
		paidAt = s.PaidAt.Time.Format(time.RFC3339)
	}

	return SettlementResponse{
		ID:                   s.ID,
		GroupID:              s.GroupID,
		PayerID:              s.PayerID,
		PayerPendingUserID:   s.PayerPendingUserID,
		PayeeID:              s.PayeeID,
		PayeePendingUserID:   s.PayeePendingUserID,
		Amount:               amount,
		CurrencyCode:         s.CurrencyCode,
		Status:               s.Status,
		PaymentMethod:        paymentMethod,
		TransactionReference: transactionReference,
		PaidAt:               paidAt,
		Notes:                notes,
		CreatedAt:            s.CreatedAt.Time.Format(time.RFC3339),
		CreatedBy:            s.CreatedBy,
		UpdatedAt:            s.UpdatedAt.Time.Format(time.RFC3339),
		UpdatedBy:            s.UpdatedBy,
	}
}

// Handlers

func CreateSettlementHandler(settlementService service.SettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		req, ok := middleware.GetBody[CreateSettlementRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		groupIDStr := chi.URLParam(r, "group_id")
		var groupID pgtype.UUID
		if err := groupID.Scan(groupIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group id")
			return
		}

		var payerID pgtype.UUID
		if req.PayerID != "" {
			if err := payerID.Scan(req.PayerID); err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payer id")
				return
			}
		}

		var payerPendingUserID pgtype.UUID
		if req.PayerPendingUserID != "" {
			if err := payerPendingUserID.Scan(req.PayerPendingUserID); err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payer pending user id")
				return
			}
		}

		var payeeID pgtype.UUID
		if req.PayeeID != "" {
			if err := payeeID.Scan(req.PayeeID); err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payee id")
				return
			}
		}

		var payeePendingUserID pgtype.UUID
		if req.PayeePendingUserID != "" {
			if err := payeePendingUserID.Scan(req.PayeePendingUserID); err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payee pending user id")
				return
			}
		}

		status := req.Status
		if status == "" {
			status = "pending"
		}

		settlement, err := settlementService.CreateSettlement(r.Context(), service.CreateSettlementInput{
			GroupID:              groupID,
			PayerID:              payerID,
			PayerPendingUserID:   payerPendingUserID,
			PayeeID:              payeeID,
			PayeePendingUserID:   payeePendingUserID,
			Amount:               req.Amount,
			CurrencyCode:         req.CurrencyCode,
			Status:               status,
			PaymentMethod:        req.PaymentMethod,
			TransactionReference: req.TransactionReference,
			Notes:                req.Notes,
			CreatedBy:            requesterID,
		})
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrGroupNotFound, service.ErrSettlementNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrInvalidAmount, service.ErrInvalidStatus, service.ErrInvalidSettlement:
				statusCode = http.StatusBadRequest
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := settlementToResponse(settlement)
		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func GetSettlementHandler(settlementService service.SettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		settlementIDStr := chi.URLParam(r, "settlement_id")
		var settlementID pgtype.UUID
		if err := settlementID.Scan(settlementIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid settlement id")
			return
		}

		settlement, err := settlementService.GetSettlementByID(r.Context(), settlementID, requesterID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrSettlementNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := settlementToResponse(settlement)
		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func ListSettlementsByGroupHandler(settlementService service.SettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		groupIDStr := chi.URLParam(r, "group_id")
		var groupID pgtype.UUID
		if err := groupID.Scan(groupIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group id")
			return
		}

		rows, err := settlementService.ListSettlementsByGroup(r.Context(), groupID, requesterID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrGroupNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		// Convert rows to response with user info
		settlements := make([]SettlementWithUsersResponse, 0, len(rows))
		for _, row := range rows {
			amount := "0"
			if val, err := row.Amount.Value(); err == nil {
				if str, ok := val.(string); ok {
					amount = str
				}
			}

			paymentMethod := ""
			if row.PaymentMethod.Valid {
				paymentMethod = row.PaymentMethod.String
			}

			transactionReference := ""
			if row.TransactionReference.Valid {
				transactionReference = row.TransactionReference.String
			}

			notes := ""
			if row.Notes.Valid {
				notes = row.Notes.String
			}

			paidAt := ""
			if row.PaidAt.Valid {
				paidAt = row.PaidAt.Time.Format(time.RFC3339)
			}

			settlements = append(settlements, SettlementWithUsersResponse{
				ID:                   row.ID,
				GroupID:              row.GroupID,
				PayerID:              row.PayerID,
				PayerPendingUserID:   row.PayerPendingUserID,
				PayeeID:              row.PayeeID,
				PayeePendingUserID:   row.PayeePendingUserID,
				Amount:               amount,
				CurrencyCode:         row.CurrencyCode,
				Status:               row.Status,
				PaymentMethod:        paymentMethod,
				TransactionReference: transactionReference,
				PaidAt:               paidAt,
				Notes:                notes,
				CreatedAt:            row.CreatedAt.Time.Format(time.RFC3339),
				CreatedBy:            row.CreatedBy,
				UpdatedAt:            row.UpdatedAt.Time.Format(time.RFC3339),
				UpdatedBy:            row.UpdatedBy,
				Payer: SettlementParticipantInfo{
					UserID:        row.PayerID,
					PendingUserID: row.PayerPendingUserID,
					Email:         row.PayerEmail,
					Name:          row.PayerName.String,
					AvatarURL:     row.PayerAvatarUrl.String,
					IsPending:     row.PayerPendingUserID.Valid,
				},
				Payee: SettlementParticipantInfo{
					UserID:        row.PayeeID,
					PendingUserID: row.PayeePendingUserID,
					Email:         row.PayeeEmail,
					Name:          row.PayeeName.String,
					AvatarURL:     row.PayeeAvatarUrl.String,
					IsPending:     row.PayeePendingUserID.Valid,
				},
			})
		}

		response.SendSuccess(w, http.StatusOK, settlements)
	}
}

func UpdateSettlementHandler(settlementService service.SettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		req, ok := middleware.GetBody[UpdateSettlementRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		settlementIDStr := chi.URLParam(r, "settlement_id")
		var settlementID pgtype.UUID
		if err := settlementID.Scan(settlementIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid settlement id")
			return
		}

		settlement, err := settlementService.UpdateSettlement(r.Context(), service.UpdateSettlementInput{
			SettlementID:         settlementID,
			Amount:               req.Amount,
			CurrencyCode:         req.CurrencyCode,
			Status:               req.Status,
			PaymentMethod:        req.PaymentMethod,
			TransactionReference: req.TransactionReference,
			Notes:                req.Notes,
			UpdatedBy:            requesterID,
		})
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrSettlementNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrInvalidAmount, service.ErrInvalidStatus:
				statusCode = http.StatusBadRequest
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := settlementToResponse(settlement)
		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func UpdateSettlementStatusHandler(settlementService service.SettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		req, ok := middleware.GetBody[UpdateSettlementStatusRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		settlementIDStr := chi.URLParam(r, "settlement_id")
		var settlementID pgtype.UUID
		if err := settlementID.Scan(settlementIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid settlement id")
			return
		}

		settlement, err := settlementService.UpdateSettlementStatus(r.Context(), service.UpdateSettlementStatusInput{
			SettlementID: settlementID,
			Status:       req.Status,
			UpdatedBy:    requesterID,
		})
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrSettlementNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrInvalidStatus:
				statusCode = http.StatusBadRequest
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := settlementToResponse(settlement)
		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func DeleteSettlementHandler(settlementService service.SettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		settlementIDStr := chi.URLParam(r, "settlement_id")
		var settlementID pgtype.UUID
		if err := settlementID.Scan(settlementIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid settlement id")
			return
		}

		err := settlementService.DeleteSettlement(r.Context(), settlementID, requesterID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrSettlementNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type CreateFriendSettlementRequest struct {
	PayerID              string `json:"payer_id" validate:"required,uuid"`
	PayeeID              string `json:"payee_id" validate:"required,uuid"`
	Amount               string `json:"amount" validate:"required"`
	CurrencyCode         string `json:"currency_code" validate:"omitempty,len=3"`
	Status               string `json:"status" validate:"omitempty,oneof=pending completed cancelled"`
	PaymentMethod        string `json:"payment_method" validate:"max=50"`
	TransactionReference string `json:"transaction_reference" validate:"max=100"`
	Notes                string `json:"notes" validate:"max=500"`
}

func CreateFriendSettlementHandler(friendSettlementService service.FriendSettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		req, ok := middleware.GetBody[CreateFriendSettlementRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		friendIDStr := chi.URLParam(r, "friend_id")
		friendID, err := parseUUID(friendIDStr)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid friend_id")
			return
		}

		var payerID, payeeID pgtype.UUID
		if err := payerID.Scan(req.PayerID); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid payer id")
			return
		}
		if err := payeeID.Scan(req.PayeeID); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid payee id")
			return
		}

		status := req.Status
		if status == "" {
			status = "pending"
		}

		settlement, err := friendSettlementService.CreateFriendSettlement(r.Context(), requesterID, friendID, service.CreateFriendSettlementInput{
			PayerID:              payerID,
			PayeeID:              payeeID,
			Amount:               req.Amount,
			CurrencyCode:         req.CurrencyCode,
			Status:               status,
			PaymentMethod:        req.PaymentMethod,
			TransactionReference: req.TransactionReference,
			Notes:                req.Notes,
		})
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrFriendNotFound:
				statusCode = http.StatusForbidden
			case service.ErrInvalidAmount, service.ErrInvalidStatus:
				statusCode = http.StatusBadRequest
			default:
				statusCode = http.StatusBadRequest
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := settlementToResponse(settlement)
		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func ListFriendSettlementsHandler(friendSettlementService service.FriendSettlementService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		friendIDStr := chi.URLParam(r, "friend_id")
		friendID, err := parseUUID(friendIDStr)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid friend_id")
			return
		}

		rows, err := friendSettlementService.ListFriendSettlements(r.Context(), requesterID, friendID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrFriendNotFound:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusBadRequest
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := make([]SettlementWithUsersResponse, len(rows))
		for i, row := range rows {
			amount := "0"
			if row.Amount.Valid {
				if val, err := row.Amount.Value(); err == nil {
					if str, ok := val.(string); ok {
						amount = str
					}
				}
			}

			payer := UserInfo{
				Email:     row.PayerEmail,
				Name:      row.PayerName.String,
				AvatarURL: row.PayerAvatarUrl.String,
			}
			payee := UserInfo{
				Email:     row.PayeeEmail,
				Name:      row.PayeeName.String,
				AvatarURL: row.PayeeAvatarUrl.String,
			}

			resp[i] = SettlementWithUsersResponse{
				ID:                   row.ID,
				GroupID:              row.GroupID,
				PayerID:              row.PayerID,
				PayeeID:              row.PayeeID,
				Amount:               amount,
				CurrencyCode:         row.CurrencyCode,
				Status:               row.Status,
				PaymentMethod:        row.PaymentMethod.String,
				TransactionReference: row.TransactionReference.String,
				PaidAt:               "", // friend settlements don't expose paid_at separately for now
				Notes:                row.Notes.String,
				CreatedAt:            row.CreatedAt.Time.Format(time.RFC3339),
				CreatedBy:            row.CreatedBy,
				UpdatedAt:            row.UpdatedAt.Time.Format(time.RFC3339),
				UpdatedBy:            row.UpdatedBy,
				Payer:                payer,
				Payee:                payee,
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

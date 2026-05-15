package grpc

import (
	"context"
	"time"

	"finance-tracker/pb"
	"finance-tracker/pkg/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// createTxRequest converts a gRPC CreateTransactionRequest to the internal model.
func createTxRequest(req *pb.CreateTransactionRequest) (models.CreateTransactionRequest, error) {
	if req.GetAccountId() <= 0 {
		return models.CreateTransactionRequest{}, status.Error(codes.InvalidArgument, "account_id is required")
	}
	if req.GetAmount() == "" {
		return models.CreateTransactionRequest{}, status.Error(codes.InvalidArgument, "amount is required")
	}
	if req.GetCurrency() == "" {
		return models.CreateTransactionRequest{}, status.Error(codes.InvalidArgument, "currency is required")
	}
	if req.GetType() == "" {
		return models.CreateTransactionRequest{}, status.Error(codes.InvalidArgument, "type is required")
	}
	if req.GetDescription() == "" {
		return models.CreateTransactionRequest{}, status.Error(codes.InvalidArgument, "description is required")
	}
	if req.GetTransactedAt() == "" {
		return models.CreateTransactionRequest{}, status.Error(codes.InvalidArgument, "transacted_at is required")
	}

	transactedAt, err := parseDate(req.GetTransactedAt())
	if err != nil {
		return models.CreateTransactionRequest{}, status.Errorf(codes.InvalidArgument, "invalid transacted_at format: %v", err)
	}

	categoryID := req.GetCategoryId()
	notes := req.GetNotes()

	modelReq := models.CreateTransactionRequest{
		AccountID:    req.GetAccountId(),
		CategoryID:   ptr(categoryID),
		Amount:       req.GetAmount(),
		Currency:     req.GetCurrency(),
		Type:         req.GetType(),
		Description:  req.GetDescription(),
		Notes:        ptrStr(notes),
		TransactedAt: transactedAt.Format("2006-01-02"),
	}

	return modelReq, nil
}

// toProtoTransaction converts an internal Transaction to a gRPC Transaction.
func toProtoTransaction(tx models.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:           tx.ID,
		AccountId:    tx.AccountID,
		CategoryId:   tx.CategoryID,
		Amount:       tx.Amount,
		Currency:     tx.Currency,
		Type:         tx.Type,
		Description:  tx.Description,
		Notes:        tx.Notes,
		TransactedAt: tx.TransactedAt,
		CreatedAt:    tx.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    tx.UpdatedAt.Format(time.RFC3339),
	}
}

// toProtoTransactions converts a slice of internal Transactions to gRPC Transactions.
func toProtoTransactions(txs []models.Transaction) []*pb.Transaction {
	if txs == nil {
		return nil
	}
	result := make([]*pb.Transaction, 0, len(txs))
	for _, tx := range txs {
		result = append(result, toProtoTransaction(tx))
	}
	return result
}

// createTxQuery converts a gRPC ListTransactionsRequest to the internal query model.
func createTxQuery(req *pb.ListTransactionsRequest) models.ListTransactionsQuery {
	q := models.ListTransactionsQuery{
		Page:  1,
		Limit: 20,
	}

	if req.GetPage() > 0 {
		q.Page = int(req.GetPage())
	}
	if req.GetLimit() > 0 && req.GetLimit() <= 100 {
		q.Limit = int(req.GetLimit())
	}
	if req.GetAccountId() > 0 {
		q.AccountID = ptr(req.GetAccountId())
	}
	if req.GetCategoryId() > 0 {
		q.CategoryID = ptr(req.GetCategoryId())
	}
	if req.GetType() != "" {
		q.Type = &req.Type
	}
	if req.GetFrom() != "" {
		q.From = &req.From
	}
	if req.GetTo() != "" {
		q.To = &req.To
	}

	return q
}

// Create implements the Create RPC.
func (s *Server) Create(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.Transaction, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	modelReq, err := createTxRequest(req)
	if err != nil {
		return nil, err
	}

	result, appErr := s.transactionSvc.Create(ctx, userID, modelReq)
	if appErr != nil {
		return nil, mapErr(appErr)
	}

	return toProtoTransaction(*result), nil
}

// List implements the List RPC.
func (s *Server) List(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	query := createTxQuery(req)

	result, appErr := s.transactionSvc.List(ctx, userID, query)
	if appErr != nil {
		return nil, mapErr(appErr)
	}

	return &pb.ListTransactionsResponse{
		Transactions: toProtoTransactions(result),
	}, nil
}

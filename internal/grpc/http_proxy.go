package grpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"finance-tracker/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// HTTPProxy is a simple HTTP-to-gRPC proxy for browser clients.
// Accepts POST /grpc with JSON body and calls gRPC via HTTP proxy.
type HTTPProxy struct {
	grpcAddr string
}

func NewHTTPProxy(grpcAddr string) *HTTPProxy {
	return &HTTPProxy{grpcAddr: grpcAddr}
}

func (p *HTTPProxy) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract auth token from Authorization header
	authHeader := r.Header.Get("Authorization")
	token := ""
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token = strings.TrimSpace(parts[1])
		}
	}

	// Parse request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Method string          `json:"method"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Method == "" {
		http.Error(w, "method is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, p.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		http.Error(w, "gRPC dial error", http.StatusBadGateway)
		return
	}
	defer conn.Close()

	if token != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
	}

	client := pb.NewTransactionServiceClient(conn)

	switch req.Method {
	case "/finance.v1.TransactionService/Create":
		var in pb.CreateTransactionRequest
		if err := json.Unmarshal(req.Data, &in); err != nil {
			http.Error(w, "invalid create request body", http.StatusBadRequest)
			return
		}
		out, err := client.Create(ctx, &in)
		if err != nil {
			writeProxyError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	case "/finance.v1.TransactionService/List":
		var in pb.ListTransactionsRequest
		if err := json.Unmarshal(req.Data, &in); err != nil {
			http.Error(w, "invalid list request body", http.StatusBadRequest)
			return
		}
		out, err := client.List(ctx, &in)
		if err != nil {
			writeProxyError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out.GetTransactions()})
	case "/finance.v1.TransactionService/Get":
		var in pb.GetTransactionRequest
		if err := json.Unmarshal(req.Data, &in); err != nil {
			http.Error(w, "invalid get request body", http.StatusBadRequest)
			return
		}
		out, err := client.Get(ctx, &in)
		if err != nil {
			writeProxyError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	case "/finance.v1.TransactionService/Update":
		var in pb.UpdateTransactionRequest
		if err := json.Unmarshal(req.Data, &in); err != nil {
			http.Error(w, "invalid update request body", http.StatusBadRequest)
			return
		}
		out, err := client.Update(ctx, &in)
		if err != nil {
			writeProxyError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	case "/finance.v1.TransactionService/Delete":
		var in pb.DeleteTransactionRequest
		if err := json.Unmarshal(req.Data, &in); err != nil {
			http.Error(w, "invalid delete request body", http.StatusBadRequest)
			return
		}
		out, err := client.Delete(ctx, &in)
		if err != nil {
			writeProxyError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	default:
		http.Error(w, "unsupported method", http.StatusBadRequest)
	}
}

// ParseAccessToken extracts the access token from the request.
func ParseAccessToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrAbortHandler
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", http.ErrAbortHandler
	}
	return strings.TrimSpace(parts[1]), nil
}

func writeProxyError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
}

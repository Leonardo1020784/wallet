package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
	"time"
    "log" 
    "wallet/transaction-service/internal/models"
    "github.com/nats-io/nats.go"

)

// TransferMoney handles transferring money from one user to another
func TransferMoney(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var transferRequest models.TransferRequest
        err := json.NewDecoder(r.Body).Decode(&transferRequest)
        if err != nil {
            http.Error(w, "Invalid request payload", http.StatusBadRequest)
            return
        }

        tx, err := db.Begin()
        if err != nil {
            log.Printf("Error beginning transaction: %v", err)
            http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
            return
        }

        // Step 1: Debit sender's account
        var senderBalance float64
        err = tx.QueryRow(`SELECT balance FROM user_balance WHERE user_id = $1`, transferRequest.FromUserID).Scan(&senderBalance)
        if err == sql.ErrNoRows {
            tx.Rollback()
            log.Printf("Sender with ID %s not found", transferRequest.FromUserID)
            http.Error(w, "Sender not found", http.StatusNotFound)
            return
        } else if err != nil {
            tx.Rollback()
            log.Printf("Error fetching sender's balance: %v", err)
            http.Error(w, "Failed to fetch sender's balance", http.StatusInternalServerError)
            return
        }

        if senderBalance < transferRequest.Amount {
            tx.Rollback()
            log.Printf("Insufficient balance for user %s", transferRequest.FromUserID)
            http.Error(w, "Insufficient balance", http.StatusBadRequest)
            return
        }

        _, err = tx.Exec(`UPDATE user_balance SET balance = balance - $1 WHERE user_id = $2`, transferRequest.Amount, transferRequest.FromUserID)
        if err != nil {
            tx.Rollback()
            log.Printf("Error debiting sender's account: %v", err)
            http.Error(w, "Failed to debit sender's account", http.StatusInternalServerError)
            return
        }

        // Log the debit transaction
        _, err = tx.Exec(`INSERT INTO transactions (user_id, amount, transaction_type, created_at) VALUES ($1, $2, 'debit', $3)`,
            transferRequest.FromUserID, transferRequest.Amount, time.Now())
        if err != nil {
            tx.Rollback()
            log.Printf("Error logging debit transaction: %v", err)
            http.Error(w, "Failed to log debit transaction", http.StatusInternalServerError)
            return
        }

        // Step 2: Credit recipient's account
        var recipientBalance float64
        err = tx.QueryRow(`SELECT balance FROM user_balance WHERE user_id = $1`, transferRequest.ToUserID).Scan(&recipientBalance)
        if err == sql.ErrNoRows {
            tx.Rollback()
            log.Printf("Recipient with ID %s not found", transferRequest.ToUserID)
            http.Error(w, "Recipient not found", http.StatusNotFound)
            return
        } else if err != nil {
            tx.Rollback()
            log.Printf("Error fetching recipient's balance: %v", err)
            http.Error(w, "Failed to fetch recipient's balance", http.StatusInternalServerError)
            return
        }

        _, err = tx.Exec(`UPDATE user_balance SET balance = balance + $1 WHERE user_id = $2`, transferRequest.Amount, transferRequest.ToUserID)
        if err != nil {
            tx.Rollback()
            log.Printf("Error crediting recipient's account: %v", err)
            http.Error(w, "Failed to credit recipient's account", http.StatusInternalServerError)
            return
        }

        // Log the credit transaction
        _, err = tx.Exec(`INSERT INTO transactions (user_id, amount, transaction_type, created_at) VALUES ($1, $2, 'credit', $3)`,
            transferRequest.ToUserID, transferRequest.Amount, time.Now())
        if err != nil {
            tx.Rollback()
            log.Printf("Error logging credit transaction: %v", err)
            http.Error(w, "Failed to log credit transaction", http.StatusInternalServerError)
            return
        }

        // Commit the transaction
        err = tx.Commit()
        if err != nil {
            log.Printf("Error committing transaction: %v", err)
            http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
            return
        }

        // Respond with success
        response := map[string]string{"message": "Transfer successful"}
        json.NewEncoder(w).Encode(response)
    }
}


func GetBalance(db *sql.DB) nats.MsgHandler {
    return func(msg *nats.Msg) {
        email := string(msg.Data)
        log.Printf("Received NATS request to fetch balance for email: %s", email)

        // Fetch user balance
        var balance float64
        query := `SELECT balance FROM user_balance WHERE user_id = (SELECT user_id FROM users WHERE email = $1)`
        err := db.QueryRow(query, email).Scan(&balance)
        if err != nil {
            log.Printf("Error fetching balance for email %s: %v", email, err)
            msg.Respond([]byte(`{"error": "User not found"}`))
            return
        }

        log.Printf("Balance for email %s: %f", email, balance)

        // Create response
        response := models.BalanceResponse{
            Email:   email,
            Balance: balance,
        }

        responseBytes, err := json.Marshal(response)
        if err != nil {
            log.Printf("Error marshalling balance response: %v", err)
            msg.Respond([]byte(`{"error": "Internal server error"}`))
            return
        }

        // Respond to NATS request
        msg.Respond(responseBytes)
    }
}


func AddMoney(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var transaction models.TransactionRequest
        err := json.NewDecoder(r.Body).Decode(&transaction)
        if err != nil {
            http.Error(w, "Invalid request payload", http.StatusBadRequest)
            return
        }

        tx, err := db.Begin()
        if err != nil {
            log.Printf("Error beginning transaction: %v", err)
            http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
            return
        }

        // Ensure the user exists in the user_balance table
        var existingBalance float64
        err = tx.QueryRow(`SELECT balance FROM user_balance WHERE user_id = $1`, transaction.UserID).Scan(&existingBalance)
        if err == sql.ErrNoRows {
            log.Printf("User with ID %s not found", transaction.UserID)
            tx.Rollback()
            http.Error(w, "User not found", http.StatusNotFound)
            return
        } else if err != nil {
            log.Printf("Error fetching user balance: %v", err)
            tx.Rollback()
            http.Error(w, "Failed to fetch user balance", http.StatusInternalServerError)
            return
        }

        // Update user balance
        query := `UPDATE user_balance SET balance = balance + $1 WHERE user_id = $2 RETURNING balance`
        var newBalance float64
        err = tx.QueryRow(query, transaction.Amount, transaction.UserID).Scan(&newBalance)
        if err != nil {
            log.Printf("Error updating balance: %v", err)
            tx.Rollback()
            http.Error(w, "Failed to update balance", http.StatusInternalServerError)
            return
        }

        // Insert into transactions table
        query = `INSERT INTO transactions (user_id, amount, transaction_type, created_at) VALUES ($1, $2, 'credit', $3)`
        _, err = tx.Exec(query, transaction.UserID, transaction.Amount, time.Now())
        if err != nil {
            log.Printf("Error logging transaction: %v", err)
            tx.Rollback()
            http.Error(w, "Failed to log transaction", http.StatusInternalServerError)
            return
        }

        // Commit transaction
        err = tx.Commit()
        if err != nil {
            log.Printf("Error committing transaction: %v", err)
            http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
            return
        }

        // Respond with new balance
        response := map[string]interface{}{
            "updated_balance": newBalance,
        }
        json.NewEncoder(w).Encode(response)
    }
}


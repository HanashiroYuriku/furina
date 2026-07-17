package dto

import "encoding/json"

type AuditLogResponse struct {
	ID        string          `json:"id"`
	UserID    string          `json:"userId"`
	Action    string          `json:"action"`
	TableName string          `json:"tableName"`
	RecordID  string          `json:"recordId"`
	OldData   json.RawMessage `json:"oldData"`
	NewData   json.RawMessage `json:"newData"`
	CreatedAt string          `json:"createdAt"`
}

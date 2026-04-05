package api

import "go.mongodb.org/mongo-driver/bson/primitive"

const userIDHeader = "X-User-ID"
const uuidSubjectKey = "uuidSubject"
const apiKeyHeader = "X-Api-Key"

func isValidPolicyID(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}

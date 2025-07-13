// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package policy

type PolicyAttribute string

const (
	BulkEnrollTokenLifetimeDays PolicyAttribute = "BulkEnrollTokenLifetimeDays"
)

type PolicyConditionType string

// policy condition
type PolicyCondition struct {
	Type      PolicyConditionType
	Attribute PolicyAttribute
	Value     interface{}
}

// policy statement
type PolicyStatement struct {
	Allow     bool              `json:"allow"`
	Attribute PolicyAttribute   `json:"attribute,omitempty"`
	Condition []PolicyCondition `json:"condition,omitempty"`
}

// policy data
type Policy struct {
	Version    int                        `json:"version"`
	Attributes map[PolicyAttribute]string `json:"attributes"`
	// policy statement support is pending at this time
	// this is due to repeated keys in incoming map needing
	// special parse for intake validation or processing.
	// this is a cosmetic and ease of use concern to allow
	// Preferred: "condition": { "lte": { "Key": "Value" }
	// Internal: "condition": [ { "n": "lte", "k": "Key", "v": "Value"} ]
}

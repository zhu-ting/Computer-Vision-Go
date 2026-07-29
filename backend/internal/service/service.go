// Package service contains the business logic layer.
// Services orchestrate repository calls, enforce business rules,
// and return DTOs (data transfer objects) that are safe to expose
// to the API layer. They never deal with HTTP concerns directly.
//
// Anti-cheat rule enforced here:
//
//	Responses to the frontend during an active exam MUST NOT include
//	Option.IsCorrect or Question.Analysis. These are only returned
//	after the exam is submitted (handled in a later commit).
package service

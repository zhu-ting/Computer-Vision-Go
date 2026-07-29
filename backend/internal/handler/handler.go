// Package handler contains HTTP request handlers.
// Handlers are deliberately thin:
//  1. Bind & validate the request body / URL params
//  2. Call the appropriate service method
//  3. Write the HTTP response (status code + JSON body)
//
// Handlers never contain business logic or database queries.
package handler

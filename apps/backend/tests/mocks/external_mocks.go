package mocks

// This file contains manually created mock interfaces for external dependencies
// that are difficult to auto-generate due to complex package dependencies.

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/iterator"
)

// MockFirestoreClient is a mock for firestore.Client
type MockFirestoreClient struct {
	collections     map[string]*MockCollection
	transactionFunc func(context.Context, func(context.Context, *firestore.Transaction) error) error
}

// NewMockFirestoreClient creates a new mock Firestore client
func NewMockFirestoreClient() *MockFirestoreClient {
	return &MockFirestoreClient{
		collections: make(map[string]*MockCollection),
	}
}

// Collection returns a mock collection
func (m *MockFirestoreClient) Collection(path string) *MockCollection {
	if m.collections[path] == nil {
		m.collections[path] = &MockCollection{path: path}
	}
	return m.collections[path]
}

// RunTransaction runs a mock transaction
func (m *MockFirestoreClient) RunTransaction(ctx context.Context, f func(context.Context, *firestore.Transaction) error) error {
	if m.transactionFunc != nil {
		return m.transactionFunc(ctx, f)
	}
	// Default: run function with nil transaction
	return f(ctx, nil)
}

// SetTransactionFunc sets the transaction function for testing
func (m *MockFirestoreClient) SetTransactionFunc(f func(context.Context, func(context.Context, *firestore.Transaction) error) error) {
	m.transactionFunc = f
}

// MockCollection is a mock for firestore.CollectionRef
type MockCollection struct {
	path      string
	documents map[string]*MockDocumentRef
}

// Doc returns a mock document reference
func (m *MockCollection) Doc(id string) *MockDocumentRef {
	if m.documents == nil {
		m.documents = make(map[string]*MockDocumentRef)
	}
	if m.documents[id] == nil {
		m.documents[id] = &MockDocumentRef{id: id, collection: m}
	}
	return m.documents[id]
}

// Where returns a mock query
func (m *MockCollection) Where(path, op string, value interface{}) *MockQuery {
	return &MockQuery{
		collection: m,
		conditions: []QueryCondition{{path, op, value}},
	}
}

// MockDocumentRef is a mock for firestore.DocumentRef
type MockDocumentRef struct {
	id         string
	collection *MockCollection
	data       interface{}
	exists     bool
}

// Get returns mock document data
func (m *MockDocumentRef) Get(ctx context.Context) (*MockDocumentSnapshot, error) {
	return &MockDocumentSnapshot{
		data:   m.data,
		exists: m.exists,
		ref:    m,
	}, nil
}

// Set sets mock document data
func (m *MockDocumentRef) Set(ctx context.Context, data interface{}) error {
	m.data = data
	m.exists = true
	return nil
}

// SetMockData sets data for testing
func (m *MockDocumentRef) SetMockData(data interface{}, exists bool) {
	m.data = data
	m.exists = exists
}

// MockDocumentSnapshot is a mock for firestore.DocumentSnapshot
type MockDocumentSnapshot struct {
	data   interface{}
	exists bool
	ref    *MockDocumentRef
}

// DataTo unmarshals document data
func (m *MockDocumentSnapshot) DataTo(v interface{}) error {
	if !m.exists || m.data == nil {
		return firestore.ErrNilDocumentSnapshot
	}
	// In a real implementation, this would properly unmarshal the data
	// For testing, we assume the data is already the correct type
	return nil
}

// Exists returns whether the document exists
func (m *MockDocumentSnapshot) Exists() bool {
	return m.exists
}

// QueryCondition represents a query condition
type QueryCondition struct {
	Path  string
	Op    string
	Value interface{}
}

// MockQuery is a mock for firestore.Query
type MockQuery struct {
	collection *MockCollection
	conditions []QueryCondition
	orderBy    []string
	results    []interface{}
	err        error
}

// Where adds a condition to the query
func (m *MockQuery) Where(path, op string, value interface{}) *MockQuery {
	m.conditions = append(m.conditions, QueryCondition{path, op, value})
	return m
}

// OrderBy adds ordering to the query
func (m *MockQuery) OrderBy(path string, dir firestore.Direction) *MockQuery {
	m.orderBy = append(m.orderBy, path)
	return m
}

// Documents returns a mock iterator
func (m *MockQuery) Documents(ctx context.Context) *MockDocumentIterator {
	return &MockDocumentIterator{
		results: m.results,
		err:     m.err,
		index:   0,
	}
}

// SetMockResults sets the results for testing
func (m *MockQuery) SetMockResults(results []interface{}, err error) {
	m.results = results
	m.err = err
}

// MockDocumentIterator is a mock for firestore.DocumentIterator
type MockDocumentIterator struct {
	results []interface{}
	err     error
	index   int
}

// Next returns the next document
func (m *MockDocumentIterator) Next() (*MockDocumentSnapshot, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.index >= len(m.results) {
		return nil, iterator.Done
	}
	
	result := m.results[m.index]
	m.index++
	
	return &MockDocumentSnapshot{
		data:   result,
		exists: true,
	}, nil
}

// Stop stops the iterator
func (m *MockDocumentIterator) Stop() {
	// No-op for mock
}

// MockFirebaseAuthClient is a mock for auth.Client
type MockFirebaseAuthClient struct {
	users      map[string]*auth.UserRecord
	verifyFunc func(context.Context, string) (*auth.Token, error)
}

// NewMockFirebaseAuthClient creates a new mock Firebase Auth client
func NewMockFirebaseAuthClient() *MockFirebaseAuthClient {
	return &MockFirebaseAuthClient{
		users: make(map[string]*auth.UserRecord),
	}
}

// GetUser returns a mock user record
func (m *MockFirebaseAuthClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	if user, exists := m.users[uid]; exists {
		return user, nil
	}
	return nil, auth.ErrUserNotFound
}

// VerifyIDToken verifies a mock ID token
func (m *MockFirebaseAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	if m.verifyFunc != nil {
		return m.verifyFunc(ctx, idToken)
	}
	// Default: return a valid token
	return &auth.Token{
		UID: "test-uid",
		Claims: map[string]interface{}{
			"email": "test@example.com",
		},
		IssuedAt: time.Now().Unix(),
	}, nil
}

// SetMockUser sets a user for testing
func (m *MockFirebaseAuthClient) SetMockUser(uid, email string) {
	m.users[uid] = &auth.UserRecord{
		UID:   uid,
		Email: email,
	}
}

// SetVerifyFunc sets the verify function for testing
func (m *MockFirebaseAuthClient) SetVerifyFunc(f func(context.Context, string) (*auth.Token, error)) {
	m.verifyFunc = f
}
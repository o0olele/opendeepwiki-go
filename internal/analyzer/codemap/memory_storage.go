package codemap

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// InMemoryEmbeddingStorage implements the EmbeddingStorage interface using in-memory storage
type InMemoryEmbeddingStorage struct {
	Embeddings map[string]EmbeddingRecord
	mutex      sync.RWMutex `gob:"-"`
}

// NewInMemoryEmbeddingStorage creates a new in-memory embedding storage
func NewInMemoryEmbeddingStorage() *InMemoryEmbeddingStorage {

	tmp := &InMemoryEmbeddingStorage{
		Embeddings: make(map[string]EmbeddingRecord),
		mutex:      sync.RWMutex{},
	}
	return tmp
}

func (s *InMemoryEmbeddingStorage) LoadFromFile(filePath string) error {
	embeddings := make(map[string]EmbeddingRecord)

	err := readFromFile(filePath, embeddings, "")
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Embeddings = embeddings
	return nil
}

func (s *InMemoryEmbeddingStorage) SaveToFile(filePath string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return persistToFile(filePath, s.Embeddings, false, "")
}

// StoreEmbedding stores an embedding with metadata
func (s *InMemoryEmbeddingStorage) StoreEmbedding(id string, embedding []float32, content string, metadata map[string]string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Embeddings[id] = EmbeddingRecord{
		ID:        id,
		Content:   content,
		Metadata:  metadata,
		Embedding: embedding,
	}

	return nil
}

// SearchEmbeddings searches for embeddings similar to the query embedding
func (s *InMemoryEmbeddingStorage) SearchEmbeddings(queryEmbedding []float32, filter map[string]string, limit int, minRelevance float64) ([]EmbeddingRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var results []EmbeddingRecord

	// Apply filter and calculate similarity
	for _, record := range s.Embeddings {
		// Apply filter
		if !matchesFilter(record, filter) {
			continue
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryEmbedding, record.Embedding)

		// Apply minimum relevance threshold
		if float64(similarity) < minRelevance {
			continue
		}

		// Create a copy of the record with the score
		recordWithScore := record
		recordWithScore.Score = float64(similarity)

		results = append(results, recordWithScore)
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// matchesFilter checks if a record matches the filter
func matchesFilter(record EmbeddingRecord, filter map[string]string) bool {
	for key, value := range filter {
		if recordValue, ok := record.Metadata[key]; !ok || recordValue != value {
			return false
		}
	}
	return true
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// GetEmbedding gets an embedding by ID
func (s *InMemoryEmbeddingStorage) GetEmbedding(id string) (EmbeddingRecord, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	record, ok := s.Embeddings[id]
	if !ok {
		return EmbeddingRecord{}, fmt.Errorf("embedding not found: %s", id)
	}

	return record, nil
}

// DeleteEmbedding deletes an embedding by ID
func (s *InMemoryEmbeddingStorage) DeleteEmbedding(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.Embeddings[id]; !ok {
		return fmt.Errorf("embedding not found: %s", id)
	}

	delete(s.Embeddings, id)
	return nil
}

// ListEmbeddings lists all embeddings
func (s *InMemoryEmbeddingStorage) ListEmbeddings() []EmbeddingRecord {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var records []EmbeddingRecord
	for _, record := range s.Embeddings {
		records = append(records, record)
	}

	return records
}

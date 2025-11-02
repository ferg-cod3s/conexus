// Package sqlite provides HNSW (Hierarchical Navigable Small World) vector indexing
package sqlite

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/ferg-cod3s/conexus/internal/embedding"
)

// HNSWConfig configures the HNSW index
type HNSWConfig struct {
	M              int     // Maximum number of connections per layer (default: 16)
	MMax           int     // Maximum number of connections for the bottom layer (default: 32)
	MMax0          int     // Maximum number of connections for layer 0 (default: 64)
	EfConstruction int     // Size of the dynamic candidate list during construction (default: 200)
	EfSearch       int     // Size of the dynamic candidate list during search (default: 32)
	ML             float64 // Normalization factor for level generation (default: 1/ln(M))
	MaxLevel       int     // Maximum allowed level (computed automatically)
}

// DefaultHNSWConfig returns sensible defaults for HNSW
func DefaultHNSWConfig() HNSWConfig {
	return HNSWConfig{
		M:              16,
		MMax:           32,
		MMax0:          64,
		EfConstruction: 200,
		EfSearch:       32,
		ML:             1.0 / math.Log(16),
	}
}

// HNSWNode represents a node in the HNSW graph
type HNSWNode struct {
	ID        string           // Document ID
	Vector    embedding.Vector // The vector (normalized)
	Level     int              // Level in the hierarchy
	Neighbors [][]string       // Neighbors per level [level][neighbor_ids]
	Deleted   bool             // Soft delete flag
}

// HNSWIndex implements Hierarchical Navigable Small World for ANN search
type HNSWIndex struct {
	config     HNSWConfig
	nodes      map[string]*HNSWNode // id -> node
	entryPoint string               // Entry point for searches
	maxLevel   int                  // Current maximum level
	mu         sync.RWMutex         // Protects concurrent access
	vectorDim  int                  // Vector dimensionality
}

// NewHNSWIndex creates a new HNSW index
func NewHNSWIndex(config HNSWConfig) *HNSWIndex {
	if config.M == 0 {
		config = DefaultHNSWConfig()
	}

	return &HNSWIndex{
		config:    config,
		nodes:     make(map[string]*HNSWNode),
		maxLevel:  0,
		vectorDim: 0,
	}
}

// Insert adds a vector to the HNSW index
func (hnsw *HNSWIndex) Insert(id string, vector embedding.Vector) error {
	hnsw.mu.Lock()
	defer hnsw.mu.Unlock()

	// Validate vector
	if len(vector) == 0 {
		return fmt.Errorf("cannot insert empty vector")
	}
	if hnsw.vectorDim == 0 {
		hnsw.vectorDim = len(vector)
	} else if len(vector) != hnsw.vectorDim {
		return fmt.Errorf("vector dimension mismatch: expected %d, got %d", hnsw.vectorDim, len(vector))
	}

	// Normalize vector for cosine similarity
	normalizedVector := hnswNormalizeVector(vector)

	// Generate level for this node
	level := hnsw.generateLevel()

	// Create node
	node := &HNSWNode{
		ID:        id,
		Vector:    normalizedVector,
		Level:     level,
		Neighbors: make([][]string, level+1),
		Deleted:   false,
	}

	// Initialize neighbor slices
	for i := 0; i <= level; i++ {
		node.Neighbors[i] = make([]string, 0, hnsw.config.MMax)
	}

	// Insert into graph
	hnsw.insertNode(node)

	// Update max level
	if level > hnsw.maxLevel {
		hnsw.maxLevel = level
	}

	return nil
}

// Search finds the k nearest neighbors using HNSW
func (hnsw *HNSWIndex) Search(queryVector embedding.Vector, k int, ef int) ([]SearchCandidate, error) {
	hnsw.mu.RLock()
	defer hnsw.mu.RUnlock()

	if len(hnsw.nodes) == 0 {
		return []SearchCandidate{}, nil
	}

	// Normalize query vector
	queryNorm := hnswNormalizeVector(queryVector)

	// Start search from entry point
	entryPoint := hnsw.entryPoint
	if entryPoint == "" {
		// Find any non-deleted node
		for id, node := range hnsw.nodes {
			if !node.Deleted {
				entryPoint = id
				break
			}
		}
		if entryPoint == "" {
			return []SearchCandidate{}, nil
		}
	}

	// Perform greedy search from top level down to level 0
	currentNodeID := entryPoint
	for level := hnsw.maxLevel; level > 0; level-- {
		currentNodeID = hnsw.searchLayer(queryNorm, currentNodeID, 1, level)
	}

	// Search level 0 with full candidate list
	candidates := hnsw.searchLayerKNN(queryNorm, currentNodeID, ef, 0)

	// Sort by distance and return top k
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	if len(candidates) > k {
		candidates = candidates[:k]
	}

	return candidates, nil
}

// Remove marks a node as deleted (soft delete)
func (hnsw *HNSWIndex) Remove(id string) error {
	hnsw.mu.Lock()
	defer hnsw.mu.Unlock()

	node, exists := hnsw.nodes[id]
	if !exists {
		return fmt.Errorf("node %s not found", id)
	}

	node.Deleted = true
	return nil
}

// Size returns the number of nodes in the index
func (hnsw *HNSWIndex) Size() int {
	hnsw.mu.RLock()
	defer hnsw.mu.RUnlock()

	count := 0
	for _, node := range hnsw.nodes {
		if !node.Deleted {
			count++
		}
	}
	return count
}

// GetNode retrieves a node by ID (for debugging/testing)
func (hnsw *HNSWIndex) GetNode(id string) (*HNSWNode, bool) {
	hnsw.mu.RLock()
	defer hnsw.mu.RUnlock()

	node, exists := hnsw.nodes[id]
	return node, exists && !node.Deleted
}

// SearchCandidate represents a search result candidate
type SearchCandidate struct {
	ID       string
	Distance float32
}

// insertNode inserts a node into the HNSW graph
func (hnsw *HNSWIndex) insertNode(node *HNSWNode) {
	// If this is the first node, make it the entry point
	if hnsw.entryPoint == "" {
		hnsw.nodes[node.ID] = node
		hnsw.entryPoint = node.ID
		return
	}

	// Start from entry point
	currentNodeID := hnsw.entryPoint

	// Search for insertion points from top level down
	for level := hnsw.maxLevel; level > node.Level; level-- {
		currentNodeID = hnsw.searchLayer(node.Vector, currentNodeID, 1, level)
	}

	// Search at each level from max(node.Level, 0) down to 0
	for level := max(0, node.Level); level >= 0; level-- {
		// Find efConstruction nearest neighbors at this level
		candidates := hnsw.searchLayerKNN(node.Vector, currentNodeID, hnsw.config.EfConstruction, level)

		// Select neighbors for this node
		neighbors := hnsw.selectNeighbors(candidates, hnsw.config.M, level)

		// Connect bidirectional
		for _, neighborID := range neighbors {
			if neighborNode, exists := hnsw.nodes[neighborID]; exists && !neighborNode.Deleted {
				// Add connection from new node to neighbor
				node.Neighbors[level] = append(node.Neighbors[level], neighborID)

				// Add connection from neighbor to new node (with pruning)
				neighborNode.Neighbors[level] = hnsw.pruneConnections(
					append(neighborNode.Neighbors[level], node.ID),
					node.Vector,
					hnsw.getMaxConnections(level),
					level,
				)
			}
		}

		// Update current node for next level
		if len(neighbors) > 0 {
			currentNodeID = neighbors[0]
		}
	}

	// Add node to index
	hnsw.nodes[node.ID] = node
}

// searchLayer performs a greedy search at a specific level
func (hnsw *HNSWIndex) searchLayer(queryVector embedding.Vector, entryPointID string, ef int, level int) string {
	visited := make(map[string]bool)
	candidates := NewCandidateSet()

	// Start with entry point
	entryPoint := hnsw.nodes[entryPointID]
	if entryPoint == nil || entryPoint.Deleted || level > entryPoint.Level {
		return entryPointID
	}

	candidates.Insert(entryPointID, cosineDistance(queryVector, entryPoint.Vector))
	visited[entryPointID] = true

	for !candidates.Empty() {
		// Get closest candidate
		closestID, _ := candidates.Pop()

		// Check termination condition
		if candidates.Len() >= ef {
			_, furthestDist := candidates.PeekFurthest()
			closestNode := hnsw.nodes[closestID]
			if closestNode == nil || cosineDistance(queryVector, closestNode.Vector) > furthestDist {
				return closestID
			}
		}

		// Explore neighbors
		closestNode := hnsw.nodes[closestID]
		if closestNode == nil || level >= len(closestNode.Neighbors) {
			continue
		}

		for _, neighborID := range closestNode.Neighbors[level] {
			if visited[neighborID] {
				continue
			}
			visited[neighborID] = true

			neighbor := hnsw.nodes[neighborID]
			if neighbor == nil || neighbor.Deleted {
				continue
			}

			distance := cosineDistance(queryVector, neighbor.Vector)
			candidates.Insert(neighborID, distance)
		}
	}

	// Return closest found
	if closestID, _ := candidates.PeekClosest(); closestID != "" {
		return closestID
	}
	return entryPointID
}

// searchLayerKNN performs KNN search at a specific level
func (hnsw *HNSWIndex) searchLayerKNN(queryVector embedding.Vector, entryPointID string, ef int, level int) []SearchCandidate {
	visited := make(map[string]bool)
	candidates := NewCandidateSet()
	results := NewCandidateSet()

	// Start with entry point
	entryPoint := hnsw.nodes[entryPointID]
	if entryPoint == nil || entryPoint.Deleted {
		return []SearchCandidate{}
	}

	distance := cosineDistance(queryVector, entryPoint.Vector)
	candidates.Insert(entryPointID, distance)
	results.Insert(entryPointID, distance)
	visited[entryPointID] = true

	for !candidates.Empty() {
		// Get closest candidate
		closestID, closestDist := candidates.Pop()

		// Get furthest result
		_, furthestResultDist := results.PeekFurthest()

		// Termination condition
		if closestDist > furthestResultDist && results.Len() >= ef {
			break
		}

		// Explore neighbors
		closestNode := hnsw.nodes[closestID]
		if closestNode == nil || level >= len(closestNode.Neighbors) {
			continue
		}

		for _, neighborID := range closestNode.Neighbors[level] {
			if visited[neighborID] {
				continue
			}
			visited[neighborID] = true

			neighbor := hnsw.nodes[neighborID]
			if neighbor == nil || neighbor.Deleted {
				continue
			}

			distance := cosineDistance(queryVector, neighbor.Vector)

			// Add to candidates
			candidates.Insert(neighborID, distance)

			// Add to results if it's a candidate
			if results.Len() < ef || distance < furthestResultDist {
				results.Insert(neighborID, distance)
				// Keep only ef best
				if results.Len() > ef {
					results.PopFurthest()
				}
			}
		}
	}

	return results.ToSlice()
}

// selectNeighbors selects the best neighbors using heuristic
func (hnsw *HNSWIndex) selectNeighbors(candidates []SearchCandidate, M int, level int) []string {
	if len(candidates) <= M {
		result := make([]string, len(candidates))
		for i, c := range candidates {
			result[i] = c.ID
		}
		return result
	}

	// Sort by distance
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	// Heuristic selection: keep nodes that are closer to query than to each other
	selected := make([]string, 0, M)
	selected = append(selected, candidates[0].ID)

	for i := 1; i < len(candidates) && len(selected) < M; i++ {
		candidateID := candidates[i].ID
		candidateNode := hnsw.nodes[candidateID]
		if candidateNode == nil {
			continue
		}

		shouldSelect := true
		for _, selectedID := range selected {
			selectedNode := hnsw.nodes[selectedID]
			if selectedNode == nil {
				continue
			}

			// If candidate is closer to existing selected than to query, skip
			distToSelected := cosineDistance(candidateNode.Vector, selectedNode.Vector)
			distToQuery := candidates[i].Distance

			if distToSelected < distToQuery {
				shouldSelect = false
				break
			}
		}

		if shouldSelect {
			selected = append(selected, candidateID)
		}
	}

	return selected
}

// pruneConnections prunes connections to maintain maximum allowed
func (hnsw *HNSWIndex) pruneConnections(connections []string, centerVector embedding.Vector, maxConnections int, level int) []string {
	if len(connections) <= maxConnections {
		return connections
	}

	// Calculate distances from center
	type connection struct {
		id   string
		dist float32
	}

	conns := make([]connection, len(connections))
	for i, connID := range connections {
		connNode := hnsw.nodes[connID]
		if connNode == nil {
			conns[i] = connection{id: connID, dist: math.MaxFloat32}
			continue
		}
		conns[i] = connection{
			id:   connID,
			dist: cosineDistance(centerVector, connNode.Vector),
		}
	}

	// Sort by distance
	sort.Slice(conns, func(i, j int) bool {
		return conns[i].dist < conns[j].dist
	})

	// Keep closest maxConnections
	result := make([]string, maxConnections)
	for i := 0; i < maxConnections; i++ {
		result[i] = conns[i].id
	}

	return result
}

// generateLevel generates a random level using exponential distribution
func (hnsw *HNSWIndex) generateLevel() int {
	level := 0
	for rand.Float64() < hnsw.config.ML && level < hnsw.config.MaxLevel {
		level++
	}
	return level
}

// getMaxConnections returns maximum connections for a level
func (hnsw *HNSWIndex) getMaxConnections(level int) int {
	if level == 0 {
		return hnsw.config.MMax0
	}
	return hnsw.config.MMax
}

// CandidateSet maintains a set of candidates with distances
type CandidateSet struct {
	items map[string]float32
}

func NewCandidateSet() *CandidateSet {
	return &CandidateSet{items: make(map[string]float32)}
}

func (cs *CandidateSet) Insert(id string, distance float32) {
	cs.items[id] = distance
}

func (cs *CandidateSet) Empty() bool {
	return len(cs.items) == 0
}

func (cs *CandidateSet) Len() int {
	return len(cs.items)
}

func (cs *CandidateSet) Pop() (string, float32) {
	var closestID string
	var closestDist float32 = math.MaxFloat32

	for id, dist := range cs.items {
		if dist < closestDist {
			closestID = id
			closestDist = dist
		}
	}

	if closestID != "" {
		delete(cs.items, closestID)
	}

	return closestID, closestDist
}

func (cs *CandidateSet) PeekClosest() (string, float32) {
	var closestID string
	var closestDist float32 = math.MaxFloat32

	for id, dist := range cs.items {
		if dist < closestDist {
			closestID = id
			closestDist = dist
		}
	}

	return closestID, closestDist
}

func (cs *CandidateSet) PeekFurthest() (string, float32) {
	var furthestID string
	var furthestDist float32 = -1

	for id, dist := range cs.items {
		if dist > furthestDist {
			furthestID = id
			furthestDist = dist
		}
	}

	return furthestID, furthestDist
}

func (cs *CandidateSet) PopFurthest() (string, float32) {
	furthestID, furthestDist := cs.PeekFurthest()
	if furthestID != "" {
		delete(cs.items, furthestID)
	}
	return furthestID, furthestDist
}

func (cs *CandidateSet) ToSlice() []SearchCandidate {
	result := make([]SearchCandidate, 0, len(cs.items))
	for id, dist := range cs.items {
		result = append(result, SearchCandidate{ID: id, Distance: dist})
	}
	return result
}

// Utility functions

func hnswNormalizeVector(v embedding.Vector) embedding.Vector {
	norm := vectorMagnitude(v)
	if norm == 0 {
		return v // Avoid division by zero
	}

	normalized := make(embedding.Vector, len(v))
	for i, val := range v {
		normalized[i] = val / norm
	}
	return normalized
}

func cosineDistance(a, b embedding.Vector) float32 {
	// Since vectors are normalized, cosine distance = 1 - cosine similarity
	dotProduct := float32(0)
	for i := range a {
		dotProduct += a[i] * b[i]
	}

	// Clamp to [0, 2] to handle floating point errors
	if dotProduct > 1.0 {
		dotProduct = 1.0
	} else if dotProduct < -1.0 {
		dotProduct = -1.0
	}

	// Cosine distance = 1 - similarity
	return 1.0 - dotProduct
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

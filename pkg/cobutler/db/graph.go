package db

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Graph represents the SQLite database used to store the brain data
type Graph struct {
	Conn  *sql.DB
	order int
}

// Order returns the order of the graph
func (g *Graph) Order() int {
	return g.order
}

// NewGraph creates a new Graph with the specified SQLite database
func NewGraph(dbPath string) (*Graph, error) {
	// Append SQLite performance flags to the DSN
	dsn := dbPath + "?_journal=WAL&_synchronous=OFF&_locking_mode=NORMAL&_cache_size=10000&_busy_timeout=5000"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Retrieve the brain order from the database info table
	var order int
	err = db.QueryRow("SELECT text FROM info WHERE attribute = 'order'").Scan(&order)
	if err != nil {
		return nil, fmt.Errorf("failed to get brain order: %w", err)
	}

	// Set more pragmas for better performance
	pragmas := []string{
		"PRAGMA cache_size=10000",
		"PRAGMA synchronous=OFF",
		"PRAGMA journal_mode=WAL",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA mmap_size=30000000000",
		"PRAGMA page_size=4096",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			// Log but continue if pragma fails
			fmt.Printf("Warning: failed to set %s: %v\n", pragma, err)
		}
	}

	// Set connection pool limits
	db.SetMaxOpenConns(1) // SQLite can only handle one writer at a time
	db.SetMaxIdleConns(1)

	rand.Seed(time.Now().UnixNano())

	graph := &Graph{
		Conn:  db,
		order: order,
	}

	return graph, nil
}

// Close closes the database connection
func (g *Graph) Close() error {
	return g.Conn.Close()
}

// Commit commits the current transaction or starts one if none exists
func (g *Graph) Commit() error {
	// Explicitly try to BEGIN a transaction first
	// If it fails because one is active, that's ok
	g.Conn.Exec("BEGIN")

	// Try to commit regardless of BEGIN result
	_, err := g.Conn.Exec("COMMIT")
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Start a new transaction for the next operations
	_, err = g.Conn.Exec("BEGIN")
	return err
}

// BeginTransaction begins a new transaction if one isn't already active
func (g *Graph) BeginTransaction() error {
	// Check if a transaction is already active by attempting a no-op update
	var inTransaction bool
	err := g.Conn.QueryRow("SELECT 1 FROM sqlite_master LIMIT 0").Scan()
	if err != nil && err.Error() == "sql: Scan called without calling Next" {
		// No error about transaction, so we are not in a transaction
		inTransaction = false
	} else {
		// Either no error or different error - assume we're in a transaction
		inTransaction = true
	}

	if !inTransaction {
		_, err = g.Conn.Exec("BEGIN")
		return err
	}
	return nil
}

// GetTokenByText gets a token ID by its text, optionally creating it if it doesn't exist
func (g *Graph) GetTokenByText(text string, create bool) (int, error) {
	var id int
	err := g.Conn.QueryRow("SELECT id FROM tokens WHERE text = ?", text).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get token: %w", err)
	}
	if !create {
		return 0, nil
	}

	// Check if the token contains a word character
	isWord := 0
	for _, c := range text {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			isWord = 1
			break
		}
	}

	result, err := g.Conn.Exec("INSERT INTO tokens (text, is_word) VALUES (?, ?)", text, isWord)
	if err != nil {
		return 0, fmt.Errorf("failed to insert token: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(lastID), nil
}

// GetNodeByTokens gets a node ID for the specified token IDs
func (g *Graph) GetNodeByTokens(tokens []int) (int, error) {
	if len(tokens) != g.order {
		return 0, fmt.Errorf("expected %d tokens, got %d", g.order, len(tokens))
	}

	// Build the query dynamically based on the order
	var conditions []string
	args := make([]interface{}, 0, g.order)
	for i := 0; i < g.order; i++ {
		conditions = append(conditions, fmt.Sprintf("token%d_id = ?", i))
		args = append(args, tokens[i])
	}

	query := fmt.Sprintf("SELECT id FROM nodes WHERE %s", strings.Join(conditions, " AND "))
	var id int
	err := g.Conn.QueryRow(query, args...).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get node: %w", err)
	}

	// Node not found, create it
	columns := make([]string, 0, g.order)
	placeholders := make([]string, 0, g.order)
	for i := 0; i < g.order; i++ {
		columns = append(columns, fmt.Sprintf("token%d_id", i))
		placeholders = append(placeholders, "?")
	}

	query = fmt.Sprintf("INSERT INTO nodes (count, %s) VALUES (0, %s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := g.Conn.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert node: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(lastID), nil
}

// AddEdge adds an edge between two nodes or increments its count if it already exists
func (g *Graph) AddEdge(prevNode, nextNode int, hasSpace bool) error {
	hasSpaceInt := 0
	if hasSpace {
		hasSpaceInt = 1
	}

	// Try to update an existing edge
	result, err := g.Conn.Exec(
		"UPDATE edges SET count = count + 1 WHERE prev_node = ? AND next_node = ? AND has_space = ?",
		prevNode, nextNode, hasSpaceInt)
	if err != nil {
		return fmt.Errorf("failed to update edge: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// If no rows were affected, we need to insert a new edge
	if rowsAffected == 0 {
		_, err = g.Conn.Exec(
			"INSERT INTO edges (prev_node, next_node, has_space, count) VALUES (?, ?, ?, 1)",
			prevNode, nextNode, hasSpaceInt)
		if err != nil {
			return fmt.Errorf("failed to insert edge: %w", err)
		}
	}

	return nil
}

// GetRandomNodeWithToken returns a random node containing the specified token
func (g *Graph) GetRandomNodeWithToken(tokenID int) (int, error) {
	var count int
	err := g.Conn.QueryRow("SELECT COUNT(*) FROM nodes WHERE token0_id = ?", tokenID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count nodes: %w", err)
	}

	if count == 0 {
		return 0, nil
	}

	// Select a random node
	offset := rand.Intn(count)
	var nodeID int
	err = g.Conn.QueryRow("SELECT id FROM nodes WHERE token0_id = ? LIMIT 1 OFFSET ?", tokenID, offset).Scan(&nodeID)
	if err != nil {
		return 0, fmt.Errorf("failed to get random node: %w", err)
	}

	return nodeID, nil
}

// GetRandomToken returns a random token ID
func (g *Graph) GetRandomToken() (int, error) {
	var count int
	err := g.Conn.QueryRow("SELECT COUNT(*) FROM tokens WHERE text != ''").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tokens: %w", err)
	}

	if count == 0 {
		return 0, fmt.Errorf("no tokens in database")
	}

	// Select a random token
	offset := rand.Intn(count)
	var tokenID int
	err = g.Conn.QueryRow("SELECT id FROM tokens WHERE text != '' LIMIT 1 OFFSET ?", offset).Scan(&tokenID)
	if err != nil {
		return 0, fmt.Errorf("failed to get random token: %w", err)
	}

	return tokenID, nil
}

// GetTextByEdge returns the text and space info for a given edge
func (g *Graph) GetTextByEdge(edgeID int) (string, bool, error) {
	// Get the next node ID for this edge
	var nextNodeID int
	var hasSpace int

	err := g.Conn.QueryRow(`
		SELECT next_node, has_space 
		FROM edges
		WHERE id = ?
	`, edgeID).Scan(&nextNodeID, &hasSpace)

	if err != nil {
		return "", false, err
	}

	// Get the token ID for this node
	var tokenID int
	err = g.Conn.QueryRow(`
		SELECT token0_id FROM nodes
		WHERE id = ?
	`, nextNodeID).Scan(&tokenID)

	if err != nil {
		return "", false, err
	}

	// Get the token text
	var text string
	err = g.Conn.QueryRow(`
		SELECT text FROM tokens
		WHERE id = ?
	`, tokenID).Scan(&text)

	if err != nil {
		return "", false, err
	}

	return text, hasSpace == 1, nil
}

// GetWordTokens returns the token IDs in the node that are actual words
func (g *Graph) GetWordTokens(tokenIDs []int) ([]int, error) {
	if len(tokenIDs) == 0 {
		return nil, nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(tokenIDs))
	args := make([]interface{}, len(tokenIDs))
	for i, id := range tokenIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	// Build and execute the query
	query := fmt.Sprintf("SELECT id FROM tokens WHERE id IN (%s) AND is_word = 1", strings.Join(placeholders, ", "))
	rows, err := g.Conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query word tokens: %w", err)
	}
	defer rows.Close()

	var result []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan token ID: %w", err)
		}
		result = append(result, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating token rows: %w", err)
	}

	return result, nil
}

// SearchRandomWalk performs a random walk from startID to endID in the specified direction
func (g *Graph) SearchRandomWalk(startID, endID int, direction bool) ([]int, error) {
	var edgeIDs []int
	currentID := startID
	maxLength := 15 // Limit depth for better performance (down from 100)

	for i := 0; i < maxLength; i++ {
		// Build a more efficient query with LIMIT to reduce processing time
		query := ""
		if direction {
			// Forward direction (prev_node -> next_node)
			query = "SELECT id, next_node FROM edges WHERE prev_node = ? ORDER BY RANDOM() LIMIT 5"
		} else {
			// Backward direction (next_node -> prev_node)
			query = "SELECT id, prev_node FROM edges WHERE next_node = ? ORDER BY RANDOM() LIMIT 5"
		}

		// Execute the query
		rows, err := g.Conn.Query(query, currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to query edges: %w", err)
		}

		// Collect edges
		var edges []struct {
			ID       int
			TargetID int
		}

		for rows.Next() {
			var id, targetID int
			if err := rows.Scan(&id, &targetID); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan edge: %w", err)
			}

			// Skip self-loops
			if targetID == currentID {
				continue
			}

			edges = append(edges, struct {
				ID       int
				TargetID int
			}{
				ID:       id,
				TargetID: targetID,
			})
		}
		rows.Close()

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating edge rows: %w", err)
		}

		if len(edges) == 0 {
			break // Dead end
		}

		// Choose first edge for speed
		chosenEdge := edges[0]
		if len(edges) > 1 {
			// If we have multiple choices, pick a random one
			chosenEdge = edges[rand.Intn(len(edges))]
		}

		// Add the chosen edge to the path
		edgeIDs = append(edgeIDs, chosenEdge.ID)
		currentID = chosenEdge.TargetID

		// Check if we've reached the end
		if currentID == endID {
			break
		}
	}

	return edgeIDs, nil
}

// FindEdgesForContext finds edges that match a given context of token IDs
func (g *Graph) FindEdgesForContext(tokenIDs []int) ([]int, error) {
	if len(tokenIDs) < 2 {
		return nil, fmt.Errorf("context too short")
	}

	// Get node IDs that contain the token context
	nodeID, err := g.findNodeContainingContext(tokenIDs)
	if err != nil || nodeID == 0 {
		return nil, fmt.Errorf("no matching context found")
	}

	// Get edges that follow this node
	edges, err := g.findEdgesFromNode(nodeID)
	if err != nil {
		return nil, err
	}

	return edges, nil
}

// findNodeContainingContext looks for a node containing a sequence of tokens
func (g *Graph) findNodeContainingContext(tokenIDs []int) (int, error) {
	// For exact match of our order, use GetNodeByTokens
	if len(tokenIDs) == g.order {
		return g.GetNodeByTokens(tokenIDs)
	}

	// For partial matches, we need a custom query
	// This looks for nodes where the last N tokens match the first N tokens of our context
	// where N is min(len(tokenIDs), g.order)

	// Build the match query
	matchLength := min(len(tokenIDs), g.order)
	matchIDs := tokenIDs[0:matchLength]

	// Convert token IDs to string for SQL
	tokenStr := make([]string, len(matchIDs))
	for i, id := range matchIDs {
		tokenStr[i] = fmt.Sprintf("%d", id)
	}

	// Build query to find nodes with matching token sequences
	// This is a simplified version that adapts to the schema
	query := fmt.Sprintf(`
		SELECT id FROM nodes WHERE id IN (
			SELECT nodes.id FROM nodes
			WHERE token0_id IN (%s)
			LIMIT 20
		)
		LIMIT 1
	`, strings.Join(tokenStr, ","))

	var nodeID int
	err := g.Conn.QueryRow(query).Scan(&nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return nodeID, nil
}

// findEdgesFromNode gets edges that follow from the given node
func (g *Graph) findEdgesFromNode(nodeID int) ([]int, error) {
	// Query for edges that start from this node based on schema
	rows, err := g.Conn.Query(`
		SELECT id FROM edges 
		WHERE prev_node = ?
		ORDER BY count DESC
		LIMIT 20
	`, nodeID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []int
	for rows.Next() {
		var edgeID int
		if err := rows.Scan(&edgeID); err != nil {
			return nil, err
		}
		edges = append(edges, edgeID)
	}

	return edges, nil
}

// Helper function for min value
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

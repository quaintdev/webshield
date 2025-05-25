package repository

import (
	"strings"
)

type DomainDataStore struct {
	root *TrieNode
}

func NewDomainDataSTore() *DomainDataStore {
	return &DomainDataStore{
		root: NewTrieNode(),
	}
}

func (d *DomainDataStore) GetDomainCategory(domain string) string {
	return d.Search(domain)
}

func (d *DomainDataStore) AddDomain(domain string, category string) {
	d.Insert(domain, category)
}

// TrieNode represents a node in our domain trie
type TrieNode struct {
	children map[string]*TrieNode
	isEnd    bool
	category string
}

// NewTrieNode creates a new trie node
func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[string]*TrieNode),
		isEnd:    false,
		category: "",
	}
}

// Insert adds a domain and its category to the trie
func (d *DomainDataStore) Insert(domain string, category string) {
	// Reverse the domain parts for the trie (e.g., "example.com" -> "com.example")
	parts := strings.Split(domain, ".")
	reverseArray(parts)
	reverseDomain := strings.Join(parts, ".")

	current := d.root
	for _, part := range strings.Split(reverseDomain, ".") {
		if _, exists := current.children[part]; !exists {
			current.children[part] = NewTrieNode()
		}
		current = current.children[part]
	}

	current.isEnd = true
	current.category = category
}

// Search looks up a domain and returns its category
func (d *DomainDataStore) Search(domain string) string {
	// Reverse the domain parts for searching
	parts := strings.Split(domain, ".")
	reverseArray(parts)
	reverseDomain := strings.Join(parts, ".")

	current := d.root
	bestMatch := ""

	// Track parts to handle wildcard matches
	domainParts := strings.Split(reverseDomain, ".")

	// Keep track of our current path through the trie
	var currentPath []string

	for i, part := range domainParts {
		if node, exists := current.children[part]; exists {
			current = node
			currentPath = append(currentPath, part)

			// If this node marks the end of a domain, remember its category
			if current.isEnd {
				bestMatch = current.category
			}
		} else {
			// No match for this part, so break
			break
		}

		// Check for wildcard match at this level (e.g., *.example.com)
		if wildcard, exists := current.children["*"]; exists && i < len(domainParts)-1 {
			// Wildcard exists, so we have a potential match
			if wildcard.isEnd {
				bestMatch = wildcard.category
			}
		}
	}

	return bestMatch
}

// Helper function to reverse an array in place
func reverseArray(arr []string) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

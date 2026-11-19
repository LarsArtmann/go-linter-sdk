package linter

// ByIDLen returns the size of the internal ID index. Test-only shim (see
// AGENTS.md: tests are black-box; unexported state is reached through this
// file, not by reverting to white-box tests).
func (r *Registry) ByIDLen() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.byID)
}

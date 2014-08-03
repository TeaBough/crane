package crane

import (
	"fmt"
	"strings"
)

// DependencyGraph maps container names
// to their dependencies
type DependencyGraph map[string]*Dependencies

// determineOrder works on the dependency graph and returns the order
// of the given the target (a subset of the graph).
// If force is true, the map will be ordered even if dependencies
// are missing.
// If force is false and the graph cannot be resolved properly,
// an error is returned.
func (graph DependencyGraph) order(target Target, force bool) (order []string, err error) {
	success := true
	for success && len(order) < len(target) {
		success = false

		// Try to resolve container with satisfied dependencies
		for _, name := range target {
			dependencies := graph[name]
			if dependencies.satisfied() {
				// Resolve "name" and continue with next iteration
				success = true
				order = append([]string{name}, order...)
				graph.resolve(name)
				break
			}
		}

		// If we could not resolve a container so far in this iteration,
		// we check if one of the non-targeted dependencies already runs or exists.
		if !success {
			for _, name := range target {
				dependencies := graph[name]
				for _, name := range dependencies.all {
					if !target.includes(name) {
						container := &Container{RawName: name}
						satisfied := false
						if dependencies.mustRun(name) {
							satisfied = container.running()
						} else {
							satisfied = container.exists()
						}
						if satisfied {
							success = true
							graph.resolve(name)
							break
						}
					}
				}
			}
		}

		// If we still have not resolved a dependency, we can't
		// order the given target properly. However, if the order
		// is forced, we just resolve the first non-targeted
		// dependency.
		if !success && force {
			for _, name := range target {
				dependencies := graph[name]
				for _, name := range dependencies.all {
					if !target.includes(name) {
						success = true
						graph.resolve(name)
						break
					}
				}
			}
		}
	}

	// If we the order is not complete yet, the target
	// cannot be resolved (cyclic or missing dependency found).
	if len(order) < len(target) {
		unresolved := []string{}
		for _, name := range target {
			if _, ok := graph[name]; ok {
				unresolved = append(unresolved, name)
			}
		}
		err = fmt.Errorf("Dependencies for container(s) %s could not be resolved. Check for cyclic or missing dependencies.", strings.Join(unresolved, ", "))
	}

	return
}

// resolve deletes the given name from the
// dependency graph and removes it from all
// dependencies.
func (d DependencyGraph) resolve(resolved string) {
	delete(d, resolved)
	for _, dependencies := range d {
		dependencies.remove(resolved)
	}
}

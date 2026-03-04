package dfs

func DetectCycle(graph map[string][]string, startID string) (bool, []string) {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var stack []string
	var cycle []string

	var walk func(string) bool
	walk = func(id string) bool {
		if inStack[id] {
			for i, s := range stack {
				if s == id {
					cycle = make([]string, len(stack[i:]))
					copy(cycle, stack[i:])
					cycle = append(cycle, id)
					break
				}
			}
			return true
		}
		if visited[id] {
			return false
		}
		visited[id] = true
		inStack[id] = true
		stack = append(stack, id)

		for _, next := range graph[id] {
			if walk(next) {
				return true
			}
		}

		stack = stack[:len(stack)-1]
		inStack[id] = false
		return false
	}

	hasCycle := walk(startID)
	return hasCycle, cycle
}

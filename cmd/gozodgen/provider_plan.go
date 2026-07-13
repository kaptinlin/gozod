package main

import "strings"

type providerEdge struct {
	from string
	to   string
}

type generatedProviderPlan struct {
	types       map[string]struct{}
	cyclicEdges map[providerEdge]struct{}
}

func newGeneratedProviderPlan(infos []*GenerationInfo) *generatedProviderPlan {
	plan := &generatedProviderPlan{
		types:       make(map[string]struct{}, len(infos)),
		cyclicEdges: make(map[providerEdge]struct{}),
	}
	for _, info := range infos {
		if info != nil && info.Name != "" {
			plan.types[info.Name] = struct{}{}
		}
	}

	adjacency := make(map[string]map[string]struct{}, len(plan.types))
	for _, info := range infos {
		if info == nil {
			continue
		}
		for i := range info.Fields {
			target, ok := plan.referencedType(info.Fields[i].EffectiveTypeName())
			if !ok {
				continue
			}
			if adjacency[info.Name] == nil {
				adjacency[info.Name] = make(map[string]struct{})
			}
			adjacency[info.Name][target] = struct{}{}
		}
	}

	for from, targets := range adjacency {
		for to := range targets {
			if reachesGeneratedType(to, from, adjacency, make(map[string]struct{})) {
				plan.cyclicEdges[providerEdge{from: from, to: to}] = struct{}{}
			}
		}
	}
	return plan
}

func (p *generatedProviderPlan) referencedType(typeName string) (string, bool) {
	for {
		switch {
		case strings.HasPrefix(typeName, "*"):
			typeName = strings.TrimPrefix(typeName, "*")
		case strings.HasPrefix(typeName, "[]"):
			typeName = strings.TrimPrefix(typeName, "[]")
		case strings.HasPrefix(typeName, "["):
			end := strings.IndexByte(typeName, ']')
			if end < 0 || end == len(typeName)-1 {
				return "", false
			}
			typeName = typeName[end+1:]
		case strings.HasPrefix(typeName, "map["):
			end := strings.LastIndex(typeName, "]")
			if end < 0 || end == len(typeName)-1 {
				return "", false
			}
			typeName = typeName[end+1:]
		default:
			_, ok := p.types[typeName]
			return typeName, ok
		}
	}
}

func (p *generatedProviderPlan) has(typeName string) bool {
	if p == nil {
		return false
	}
	_, ok := p.types[typeName]
	return ok
}

func (p *generatedProviderPlan) isCyclic(from, to string) bool {
	if p == nil {
		return false
	}
	_, ok := p.cyclicEdges[providerEdge{from: from, to: to}]
	return ok
}

func reachesGeneratedType(current, target string, adjacency map[string]map[string]struct{}, seen map[string]struct{}) bool {
	if current == target {
		return true
	}
	if _, ok := seen[current]; ok {
		return false
	}
	seen[current] = struct{}{}
	for next := range adjacency[current] {
		if reachesGeneratedType(next, target, adjacency, seen) {
			return true
		}
	}
	return false
}

package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestGeneratedProviderPlanClassifiesEdgesIndependentlyOfInputOrder(t *testing.T) {
	t.Parallel()

	node := generationInfoWithEdges("Node", "*Node", "[]*Node", "map[string]*Node")
	department := generationInfoWithEdges("Department", "*Employee")
	employee := generationInfoWithEdges("Employee", "*Department")
	company := generationInfoWithEdges("Company", "*Employee")

	orders := [][]*GenerationInfo{
		{node, department, employee, company},
		{company, employee, node, department},
	}
	for _, infos := range orders {
		plan := newGeneratedProviderPlan(infos)

		assert.True(t, plan.isCyclic("Node", "Node"))
		assert.True(t, plan.isCyclic("Department", "Employee"))
		assert.True(t, plan.isCyclic("Employee", "Department"))
		assert.False(t, plan.isCyclic("Company", "Employee"))
		assert.False(t, plan.has("external.Profile"))
	}
}

func generationInfoWithEdges(name string, typeNames ...string) *GenerationInfo {
	fields := make([]tagparser.FieldInfo, len(typeNames))
	for i, typeName := range typeNames {
		fields[i] = tagparser.FieldInfo{
			Name:     "Field",
			Type:     reflect.TypeFor[struct{}](),
			TypeName: typeName,
		}
	}
	return &GenerationInfo{Name: name, Fields: fields}
}

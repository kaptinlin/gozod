package providerfixture

import (
	"reflect"
	"testing"

	"github.com/kaptinlin/gozod"
)

func TestGeneratedProviderGraphConstructsAndParses(t *testing.T) {
	tests := []struct {
		name  string
		parse func() error
	}{
		{
			name: "self cycle pointer edge",
			parse: func() error {
				_, err := (Node{}).Schema().Parse(Node{
					Value: 1,
					Next:  &Node{Value: 2},
				})
				return err
			},
		},
		{
			name: "self cycle slice edge",
			parse: func() error {
				_, err := (Node{}).Schema().Parse(Node{
					Value:    1,
					Children: []*Node{{Value: 2}},
				})
				return err
			},
		},
		{
			name: "self cycle map edge",
			parse: func() error {
				_, err := (Node{}).Schema().Parse(Node{
					Value:          1,
					ChildrenByName: map[string]*Node{"child": {Value: 2}},
				})
				return err
			},
		},
		{
			name: "mutual cycle",
			parse: func() error {
				_, err := (Employee{}).Schema().Parse(Employee{
					Name:       "Ada",
					Department: &Department{Name: "Research"},
				})
				return err
			},
		},
		{
			name: "acyclic edge into cyclic component",
			parse: func() error {
				_, err := (Company{}).Schema().Parse(Company{
					Name:  "Acme",
					Owner: &Employee{Name: "Ada"},
				})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.parse(); err != nil {
				t.Fatalf("generated schema Parse() returned error: %v", err)
			}
		})
	}
}

func TestGeneratedProviderGraphMatchesReflection(t *testing.T) {
	input := Node{
		Value:          1,
		Next:           &Node{Value: 2},
		Children:       []*Node{{Value: 3}},
		ChildrenByName: map[string]*Node{"child": {Value: 4}},
	}

	generated, err := (Node{}).Schema().Parse(input)
	if err != nil {
		t.Fatalf("generated Parse() returned error: %v", err)
	}
	reflected, err := gozod.MustFromStruct[Node]().Parse(input)
	if err != nil {
		t.Fatalf("reflection Parse() returned error: %v", err)
	}
	if !reflect.DeepEqual(generated, reflected) {
		t.Fatalf("generated output = %#v, reflection output = %#v", generated, reflected)
	}

	invalid := Node{Value: 1, Next: &Node{Value: 0}}
	_, generatedErr := (Node{}).Schema().Parse(invalid)
	_, reflectedErr := gozod.MustFromStruct[Node]().Parse(invalid)
	if generatedErr == nil || reflectedErr == nil {
		t.Fatalf("generated error = %v, reflection error = %v", generatedErr, reflectedErr)
	}
	if generatedErr.Error() != reflectedErr.Error() {
		t.Fatalf("generated error = %q, reflection error = %q", generatedErr, reflectedErr)
	}
}

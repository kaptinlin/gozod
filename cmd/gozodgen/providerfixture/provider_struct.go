// Package providerfixture verifies generated schema dependency graphs.
package providerfixture

// Node exercises self-referential generated schemas.
type Node struct {
	Value          int              `json:"value" gozod:"required,gte=1"`
	Next           *Node            `json:"next" gozod:"optional"`
	Children       []*Node          `json:"children" gozod:""`
	ChildrenByName map[string]*Node `json:"children_by_name" gozod:""`
}

// Department exercises one side of a mutual generated-schema cycle.
type Department struct {
	Name      string      `json:"name" gozod:"required"`
	Manager   *Employee   `json:"manager" gozod:""`
	Employees []*Employee `json:"employees" gozod:""`
}

// Employee exercises one side of a mutual generated-schema cycle.
type Employee struct {
	Name       string      `json:"name" gozod:"required"`
	Department *Department `json:"department" gozod:""`
	Reports    []*Employee `json:"reports" gozod:""`
}

// Company exercises an acyclic edge into a cyclic generated-schema component.
type Company struct {
	Name  string    `json:"name" gozod:"required"`
	Owner *Employee `json:"owner" gozod:""`
}

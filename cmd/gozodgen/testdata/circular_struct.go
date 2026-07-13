package testdata

// Node represents a circular reference structure
type Node struct {
	Value          int              `json:"value" gozod:"required"`
	Next           *Node            `json:"next" gozod:""`
	Children       []*Node          `json:"children" gozod:""`
	ChildrenByName map[string]*Node `json:"children_by_name" gozod:""`
}

// Department and Employee demonstrate mutual circular references
type Department struct {
	Name      string      `json:"name" gozod:"required"`
	Manager   *Employee   `json:"manager" gozod:""`
	Employees []*Employee `json:"employees" gozod:""`
}

type Employee struct {
	Name       string      `json:"name" gozod:"required"`
	Department *Department `json:"department" gozod:""`
	Reports    []*Employee `json:"reports" gozod:""`
}

type Company struct {
	Name  string    `json:"name" gozod:"required"`
	Owner *Employee `json:"owner" gozod:""`
}

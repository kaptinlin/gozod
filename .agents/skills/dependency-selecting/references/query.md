# Query & Filtering

## `github.com/agentable/queryparse` — URL Query String Parsing

- Complete URL Query Language specification
- Filtering, sorting, pagination, field selection, resource loading

**When to use:** RESTful API query parameter parsing, complex filter/sort/page from URL.

## `github.com/agentable/queryschema` — Query Schema Building

- 100% Query Protocol v1.0 compliant
- Smart JSON serialization, fluent API
- Complete method chaining
- Cross-language compatible for RESTful APIs

**When to use:** Building query schemas for API responses, defining filterable/sortable fields, API specification.

## `github.com/agentable/condeval` — Dynamic Condition Evaluation

- Evaluates dynamic filter expressions based on Query Protocol
- Type-safe operations, extensible architecture
- High performance

**When to use:** Runtime filter evaluation, dynamic query conditions, rule engine expressions.

## `github.com/agentable/filterschema` — Filter Schema Validation

- Type-safe filter validators for QuerySchema expressions

**When to use:** Validating incoming filter parameters against a schema, ensuring query safety.

## How They Fit Together

```
Client request → queryparse (parse URL query)
                     ↓
               condeval (evaluate conditions against data)
                     ↓
API definition → queryschema (define available filters/sorts)
                     ↓
               filterschema (validate filters against schema)
```

Typical usage:
1. **queryschema** defines what fields are filterable/sortable
2. **queryparse** parses incoming URL query strings
3. **filterschema** validates parsed filters against the schema
4. **condeval** evaluates conditions at runtime

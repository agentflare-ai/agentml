# Cypher List Comprehension Implementation Plan

## Overview

This document outlines the implementation plan for adding Cypher list comprehensions to the SQLite Graph Extension. List comprehensions allow creating new lists by iterating over existing lists with optional filtering and transformation.

## Syntax Specification

### Grammar

```
ListComprehension := '[' Variable 'IN' Expression ['WHERE' Expression] '|' Expression ']'
```

### Components

1. **Variable**: Iterator variable (e.g., `x`, `item`)
2. **Input Expression**: Source list or range (e.g., `[1,2,3]`, `range(0,10)`)
3. **WHERE Clause**: Optional filter predicate
4. **Transform Expression**: Output transformation for each element

### Examples

```cypher
-- Basic mapping
[x IN range(0,5) | x * 2]  -- [0, 2, 4, 6, 8, 10]

-- Filtering only
[x IN range(0,10) WHERE x % 2 = 0]  -- [0, 2, 4, 6, 8, 10]

-- Filtering with transformation
[n IN [1,2,3,4,5] WHERE n > 2 | n * 10]  -- [30, 40, 50]

-- Property access
[p IN nodes WHERE p.age > 25 | p.name]

-- Complex expressions
[x IN list WHERE x IS NOT NULL | toString(x)]
```

## Implementation Architecture

### 1. Lexer Changes (`src/cypher/cypher-lexer.c`)

#### Current State
- Already has `CYPHER_TOK_IN` token
- Already has `CYPHER_TOK_PIPE` token (|)
- Already has `CYPHER_TOK_WHERE` token

#### Required Changes
✅ **No lexer changes needed** - All required tokens already exist.

### 2. AST Changes (`include/cypher.h`, `src/cypher/cypher-ast.c`)

#### New AST Node Type
Add to `CypherAstNodeType` enum:
```c
CYPHER_AST_LIST_COMPREHENSION,  // [x IN list WHERE pred | expr]
```

#### AST Node Structure
The list comprehension AST node will have:
- **Child 0**: Iterator variable (CYPHER_AST_IDENTIFIER)
- **Child 1**: Input expression (any expression)
- **Child 2**: WHERE predicate (optional, can be NULL)
- **Child 3**: Transform expression (any expression)

#### New AST Functions
```c
// Create list comprehension AST node
CypherAst *cypherAstCreateListComprehension(
    const char *zVariable,      // Iterator variable name
    CypherAst *pInputExpr,      // Input list expression
    CypherAst *pWherePred,      // WHERE predicate (optional)
    CypherAst *pTransformExpr,  // Transform expression
    int iLine, int iColumn
);
```

### 3. Parser Changes (`src/cypher/cypher-parser.c`)

#### Integration Point
Modify `parsePrimaryExpression()` to handle list comprehensions when encountering `[`:

```c
// Current logic in parsePrimaryExpression():
if (parser->current.type == CYPHER_TOK_LBRACKET) {
    // Check if this is a list comprehension or list literal
    if (isListComprehension(parser)) {
        return parseListComprehension(parser);
    } else {
        return parseListLiteral(parser);
    }
}
```

#### New Parser Functions

```c
// Check if current '[' starts a list comprehension
static int isListComprehension(CypherParser *parser) {
    // Look ahead to see if pattern is: [ identifier IN ...
    CypherParser saved = *parser;
    
    cypherParserNextToken(parser);  // Skip '['
    if (parser->current.type != CYPHER_TOK_IDENTIFIER) {
        *parser = saved;
        return 0;
    }
    
    cypherParserNextToken(parser);  // Skip identifier
    int isComprehension = (parser->current.type == CYPHER_TOK_IN);
    
    *parser = saved;  // Restore parser state
    return isComprehension;
}

// Parse list comprehension: [var IN expr [WHERE pred] | transform]
static CypherAst *parseListComprehension(CypherParser *parser) {
    int line = parser->current.line;
    int col = parser->current.column;
    
    // Expect '['
    if (!cypherParserExpectToken(parser, CYPHER_TOK_LBRACKET)) {
        return NULL;
    }
    
    // Parse iterator variable
    if (parser->current.type != CYPHER_TOK_IDENTIFIER) {
        cypherParserError(parser, "Expected variable name in list comprehension");
        return NULL;
    }
    char *zVariable = sqlite3_mprintf("%.*s", parser->current.len, parser->current.text);
    cypherParserNextToken(parser);
    
    // Expect 'IN'
    if (!cypherParserExpectToken(parser, CYPHER_TOK_IN)) {
        sqlite3_free(zVariable);
        return NULL;
    }
    
    // Parse input expression
    CypherAst *pInputExpr = parseExpression(parser);
    if (!pInputExpr) {
        sqlite3_free(zVariable);
        return NULL;
    }
    
    // Optional WHERE clause
    CypherAst *pWherePred = NULL;
    if (parser->current.type == CYPHER_TOK_WHERE) {
        cypherParserNextToken(parser);  // Skip WHERE
        pWherePred = parseExpression(parser);
        if (!pWherePred) {
            sqlite3_free(zVariable);
            cypherAstDestroy(pInputExpr);
            return NULL;
        }
    }
    
    // Expect '|'
    if (!cypherParserExpectToken(parser, CYPHER_TOK_PIPE)) {
        sqlite3_free(zVariable);
        cypherAstDestroy(pInputExpr);
        cypherAstDestroy(pWherePred);
        return NULL;
    }
    
    // Parse transform expression
    CypherAst *pTransformExpr = parseExpression(parser);
    if (!pTransformExpr) {
        sqlite3_free(zVariable);
        cypherAstDestroy(pInputExpr);
        cypherAstDestroy(pWherePred);
        return NULL;
    }
    
    // Expect ']'
    if (!cypherParserExpectToken(parser, CYPHER_TOK_RBRACKET)) {
        sqlite3_free(zVariable);
        cypherAstDestroy(pInputExpr);
        cypherAstDestroy(pWherePred);
        cypherAstDestroy(pTransformExpr);
        return NULL;
    }
    
    // Create list comprehension AST node
    CypherAst *pListComp = cypherAstCreateListComprehension(
        zVariable, pInputExpr, pWherePred, pTransformExpr, line, col
    );
    
    sqlite3_free(zVariable);
    return pListComp;
}
```

### 4. Logical Planning (`src/cypher/cypher-logical-plan.c`)

#### Planning Strategy
List comprehensions will be planned as:
1. **Evaluation of input expression** → produces a list
2. **For each element in the list**:
   - Bind iterator variable to current element
   - Evaluate WHERE predicate (if present)
   - If predicate is true, evaluate transform expression
   - Add result to output list

#### Logical Plan Operators
```c
// Logical plan for list comprehension
typedef struct LogicalListComprehension {
    LogicalOperator base;
    char *zVariable;               // Iterator variable name
    LogicalOperator *pInputPlan;   // Plan for input expression
    LogicalOperator *pWherePlan;   // Plan for WHERE predicate (optional)
    LogicalOperator *pTransformPlan; // Plan for transform expression
} LogicalListComprehension;
```

#### Planning Function
```c
LogicalOperator *planListComprehension(
    CypherAst *pAst,
    QueryPlanningContext *pContext
) {
    LogicalListComprehension *pOp = sqlite3_malloc(sizeof(LogicalListComprehension));
    
    // Plan input expression
    pOp->pInputPlan = planExpression(pAst->apChildren[1], pContext);
    
    // Create new scope for iterator variable
    QueryPlanningContext iterContext = *pContext;
    addVariableToScope(&iterContext, pAst->apChildren[0]->zValue);
    
    // Plan WHERE predicate if present
    if (pAst->apChildren[2]) {
        pOp->pWherePlan = planExpression(pAst->apChildren[2], &iterContext);
    }
    
    // Plan transform expression
    pOp->pTransformPlan = planExpression(pAst->apChildren[3], &iterContext);
    
    return (LogicalOperator*)pOp;
}
```

### 5. Physical Planning (`src/cypher/cypher-physical-plan.c`)

#### Physical Operator
```c
typedef struct PhysicalListComprehension {
    PhysicalOperator base;
    char *zVariable;                    // Iterator variable name
    PhysicalOperator *pInputOp;         // Input expression operator
    PhysicalOperator *pWhereOp;         // WHERE predicate operator (optional)
    PhysicalOperator *pTransformOp;     // Transform expression operator
} PhysicalListComprehension;
```

#### Execution Model
```c
int execListComprehension(
    PhysicalListComprehension *pOp,
    ExecutionContext *pContext,
    CypherValue *pResult
) {
    // 1. Execute input expression to get source list
    CypherValue inputList;
    if (pOp->pInputOp->exec(pOp->pInputOp, pContext, &inputList) != CYPHER_OK) {
        return CYPHER_ERROR;
    }
    
    if (inputList.type != CYPHER_VALUE_LIST) {
        cypherValueDestroy(&inputList);
        return CYPHER_ERROR;
    }
    
    // 2. Create output list
    CypherValue outputList;
    cypherValueCreateList(&outputList, 0);
    
    // 3. Iterate over input list
    for (int i = 0; i < inputList.list.nElements; i++) {
        // Bind iterator variable to current element
        ExecutionContext iterContext = *pContext;
        setVariableValue(&iterContext, pOp->zVariable, &inputList.list.apElements[i]);
        
        // 4. Evaluate WHERE predicate if present
        if (pOp->pWhereOp) {
            CypherValue whereResult;
            if (pOp->pWhereOp->exec(pOp->pWhereOp, &iterContext, &whereResult) != CYPHER_OK) {
                cypherValueDestroy(&inputList);
                cypherValueDestroy(&outputList);
                return CYPHER_ERROR;
            }
            
            if (!cypherValueIsTruthy(&whereResult)) {
                cypherValueDestroy(&whereResult);
                continue;  // Skip this element
            }
            cypherValueDestroy(&whereResult);
        }
        
        // 5. Evaluate transform expression
        CypherValue transformResult;
        if (pOp->pTransformOp->exec(pOp->pTransformOp, &iterContext, &transformResult) != CYPHER_OK) {
            cypherValueDestroy(&inputList);
            cypherValueDestroy(&outputList);
            return CYPHER_ERROR;
        }
        
        // 6. Add result to output list
        cypherValueListAppend(&outputList, &transformResult);
    }
    
    cypherValueDestroy(&inputList);
    *pResult = outputList;
    return CYPHER_OK;
}
```

### 6. Expression Evaluation (`src/cypher/cypher-expressions.c`)

#### Value System Integration
List comprehensions will integrate with the existing `CypherValue` system:

```c
// Add to CypherExpressionType enum
CYPHER_EXPR_LIST_COMPREHENSION,

// Evaluation function
int evaluateListComprehension(
    CypherExpression *pExpr,
    ExecutionContext *pContext,
    CypherValue *pResult
) {
    // Delegate to physical operator execution
    return execListComprehension(
        (PhysicalListComprehension*)pExpr->pPhysicalOp,
        pContext,
        pResult
    );
}
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)
- [ ] Add `CYPHER_AST_LIST_COMPREHENSION` AST node type
- [ ] Implement `cypherAstCreateListComprehension()` function
- [ ] Add list comprehension support to `cypherAstNodeTypeName()`
- [ ] Basic AST creation and destruction

### Phase 2: Parser Implementation (Week 2-3)
- [ ] Implement `isListComprehension()` lookahead function
- [ ] Implement `parseListComprehension()` parser function
- [ ] Integrate with `parsePrimaryExpression()`
- [ ] Add comprehensive error handling
- [ ] Write parser unit tests

### Phase 3: Logical Planning (Week 3-4)
- [ ] Implement `LogicalListComprehension` operator
- [ ] Add variable scoping for iterator variables
- [ ] Implement planning functions
- [ ] Add cost estimation for list comprehensions

### Phase 4: Physical Planning & Execution (Week 4-5)
- [ ] Implement `PhysicalListComprehension` operator
- [ ] Add execution context management for iterator variables
- [ ] Implement `execListComprehension()` function
- [ ] Add memory management and error handling

### Phase 5: Expression Integration (Week 5-6)
- [ ] Integrate with expression evaluation system
- [ ] Add type checking and validation
- [ ] Implement proper error messages
- [ ] Add performance optimizations

### Phase 6: Testing & Documentation (Week 6-7)
- [ ] Write comprehensive test suite
- [ ] Add performance benchmarks
- [ ] Write documentation and examples
- [ ] Integration testing with existing Cypher features

## Test Cases

### Basic Functionality
```cypher
-- Simple mapping
RETURN [x IN range(0,3) | x * 2] AS result
-- Expected: [0, 2, 4, 6]

-- Simple filtering
RETURN [x IN range(0,10) WHERE x % 2 = 0] AS result
-- Expected: [0, 2, 4, 6, 8, 10]

-- Filtering with transformation
RETURN [n IN [1,2,3,4,5] WHERE n > 2 | n * 10] AS result
-- Expected: [30, 40, 50]
```

### Complex Expressions
```cypher
-- Property access
MATCH (p:Person)
RETURN [person IN collect(p) WHERE person.age > 25 | person.name] AS adults

-- Nested list comprehensions
RETURN [outer IN [[1,2], [3,4]] | [inner IN outer | inner * 2]] AS nested
-- Expected: [[2, 4], [6, 8]]

-- String operations
RETURN [s IN ["hello", "world"] | upper(s)] AS uppercased
-- Expected: ["HELLO", "WORLD"]
```

### Edge Cases
```cypher
-- Empty list
RETURN [x IN [] | x] AS empty
-- Expected: []

-- NULL handling
RETURN [x IN [1, null, 3] WHERE x IS NOT NULL | x] AS no_nulls
-- Expected: [1, 3]

-- Type coercion
RETURN [x IN [1, 2, 3] | toString(x)] AS strings
-- Expected: ["1", "2", "3"]
```

## Performance Considerations

### Optimizations
1. **Lazy Evaluation**: Only compute transform expressions for elements that pass the WHERE clause
2. **Memory Efficiency**: Use streaming evaluation for large lists when possible
3. **Parallel Processing**: Consider parallel evaluation for independent list elements
4. **Caching**: Cache compiled expressions for iterator variables

### Complexity Analysis
- **Time Complexity**: O(n × (w + t)) where n = input list size, w = WHERE evaluation cost, t = transform cost
- **Space Complexity**: O(m) where m = output list size
- **Memory Usage**: One iterator variable binding per nested comprehension level

## Integration Points

### Existing Features
- **Function Calls**: List comprehensions can call existing functions like `range()`, `collect()`, `size()`
- **Pattern Matching**: Can be used in `MATCH` clauses for complex filtering
- **Property Access**: Full support for property access within comprehensions
- **Variable Scoping**: Proper variable scoping with existing query variables

### Future Enhancements
- **Pattern Comprehensions**: `[p IN (a)-[r]-(b) | p]` syntax
- **Parallel Evaluation**: Multi-threaded evaluation for large lists
- **Query Optimization**: Push-down optimizations for common patterns
- **LIMIT/OFFSET**: Support for partial list evaluation

## Risk Mitigation

### Potential Issues
1. **Variable Name Conflicts**: Iterator variables shadowing existing variables
2. **Memory Usage**: Large lists causing memory pressure
3. **Performance**: Nested comprehensions causing exponential complexity
4. **Parser Ambiguity**: Distinguishing list comprehensions from list literals

### Mitigation Strategies
1. **Scoping Rules**: Clear variable scoping hierarchy
2. **Memory Limits**: Configurable limits on list sizes
3. **Optimization**: Early termination and streaming evaluation
4. **Grammar Design**: Unambiguous grammar with proper lookahead

## Success Metrics

### Functional Requirements
- [ ] All openCypher list comprehension syntax supported
- [ ] Proper variable scoping and shadowing
- [ ] Correct null handling and type coercion
- [ ] Integration with existing functions and operators

### Performance Requirements
- [ ] Sub-millisecond evaluation for small lists (< 100 elements)
- [ ] Linear scaling with list size
- [ ] Memory usage proportional to output size
- [ ] No performance regression on existing features

### Quality Requirements
- [ ] 100% test coverage of new code
- [ ] No memory leaks or buffer overflows
- [ ] Proper error handling and reporting
- [ ] Comprehensive documentation

## Conclusion

This implementation plan provides a comprehensive roadmap for adding Cypher list comprehensions to the SQLite Graph Extension. The phased approach ensures systematic development while maintaining code quality and performance. The design integrates seamlessly with the existing architecture and provides a solid foundation for future enhancements.

The implementation follows established patterns in the codebase and maintains consistency with the openCypher specification. With proper testing and documentation, this feature will significantly enhance the expressiveness and power of Cypher queries in the extension.

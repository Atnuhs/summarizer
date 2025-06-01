# Test Data

This directory contains test cases for the Go competitive programming bundle tool.

## Test Cases

### simple/
- **Purpose**: Tests bundling of a main.go file that only depends on standard library packages
- **Contains**: A simple program that sorts an array of integers using `fmt` and `sort` packages
- **Expected**: The bundler should preserve standard library imports and generate working code

### with-deps/
- **Purpose**: Tests bundling with local package dependencies
- **Contains**: UnionFind and Graph data structures as local packages
- **Expected**: Local packages should be bundled with proper type/function prefixing

### with-remote/
- **Purpose**: Tests bundling with remote package dependencies (samber/lo, samber/mo)
- **Contains**: Functional programming utilities and Option types
- **Status**: Currently has limitations with complex remote packages and generics

### with-unused/
- **Purpose**: Tests dead code elimination capabilities
- **Contains**: Math and utility packages with both used and unused functions/types
- **Status**: Currently does NOT eliminate unused code (documented limitation)
- **Expected unused items**: 
  - Functions: Subtract, Divide, UnusedGlobalFunction, FormatNumber, ReverseString
  - Types: UnusedStruct, AnotherUnusedStruct, FileManager
  - Constants/Variables: UnusedConstant, UnusedVariable, DefaultPath, GlobalCounter

## Running Tests

To test a specific case:
```bash
cd testdata/simple
go run ../../main.go -input main.go -output submit.go
go run submit.go  # Should produce the same output as: go run main.go
```

## Known Limitations

1. **Dead Code Elimination**: Currently all exported symbols are included, regardless of usage
2. **Complex Remote Packages**: Large packages with generics may not bundle correctly
3. **Type Reference Rewriting**: Some type references in composite literals may not be rewritten properly
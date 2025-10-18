## Summary

<!-- Provide a concise description of what this PR does. Focus on the "why" and "what", not the "how". -->


## Type of Change

<!-- Check all that apply -->

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update (changes to docs, examples, or comments)
- [ ] Performance improvement (non-breaking change that improves performance)
- [ ] Code refactoring (no functional changes, code structure improvements)
- [ ] Build/CI changes (changes to build process, CI/CD, dependencies)
- [ ] Test improvements (adding or updating tests)
- [ ] SCXML conformance (changes to improve W3C SCXML compliance)
- [ ] Namespace extension (new or updated namespace implementation)
- [ ] WebAssembly component (WASM-related changes)

## Related Issues

<!-- Link related issues using GitHub keywords: "Fixes #123", "Closes #456", "Relates to #789" -->

- Fixes #
- Relates to #

## Related AEP

<!-- If this PR implements or relates to an AgentML Enhancement Proposal, link it here -->

- [ ] N/A - Not related to an AEP
- [ ] Implements AEP-#### (link to AEP PR or file)
- [ ] Relates to AEP-#### (provide context)

<!-- Example: Implements AEP-0015 (Runtime Plugin API) -->

## What Changed

<!-- Provide a bullet-point list of the key changes made in this PR -->

-
-
-

## Testing

<!-- Describe the testing you performed to verify your changes -->

### Test Commands Run

```bash
# List the commands you ran to test this PR
go test ./...
```

### Test Results

<!-- Summarize test results. If all tests pass, say so. If some fail, explain why. -->

- [ ] All existing tests pass
- [ ] New tests added for new functionality
- [ ] Integration tests pass (if applicable)
- [ ] SCXML conformance tests pass (`go test ./tests/conformance/...`)
- [ ] Manual testing completed

### Test Coverage

<!-- If you added new code, did you add corresponding tests? What's the coverage impact? -->

- Coverage impact: [increase/decrease/unchanged]
- Critical paths covered: [Yes/No]

## SCXML Compliance

<!-- Complete this section if your changes affect SCXML runtime behavior -->

- [ ] N/A - This PR does not affect SCXML compliance
- [ ] Changes improve W3C SCXML conformance (specify which tests now pass)
- [ ] Changes maintain existing conformance level
- [ ] Changes may impact conformance (explain why and plan to address)

**Conformance test results:**
<!-- If applicable, report current pass rate, e.g., "185/193 tests passing (95.9%)" -->

```
Current: X/193 tests passing
Previous: Y/193 tests passing
```

## Breaking Changes

<!-- Complete this section if you checked "Breaking change" above -->

- [ ] N/A - No breaking changes

**If breaking changes exist, describe:**

1. **What breaks:**
2. **Migration path:**
3. **Deprecation timeline:**
4. **Docs updated:** [ ] Yes / [ ] No

## Documentation

<!-- Ensure documentation reflects your changes -->

- [ ] Code comments added/updated (especially for complex logic)
- [ ] README.md updated (if user-facing changes)
- [ ] Documentation in `/docs` updated (if applicable)
- [ ] Examples added/updated in `/examples` (if applicable)
- [ ] API documentation generated/updated (for namespaces)
- [ ] N/A - No documentation changes needed

**Note:** CHANGELOG.md is auto-generated from conventional commits via GoReleaser during releases.

## Visual Evidence

<!-- For UI/UX changes, behavior changes, or new features, add visual proof -->
<!-- Attach screenshots, animated GIFs, or terminal recordings -->
<!-- This guarantees you tested the change and helps reviewers understand the impact -->

<!-- Example:
![Before](url-to-before-screenshot)
![After](url-to-after-screenshot)
-->

## Checklist

<!-- Complete this checklist before submitting your PR -->

### Code Quality

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My code passes `go fmt` and `go vet` checks
- [ ] I have run `golangci-lint` (if applicable) and addressed warnings
- [ ] No new compiler warnings or errors introduced

### Testing & Validation

- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have tested edge cases and error conditions
- [ ] Integration tests pass (if applicable)
- [ ] Manual testing completed and documented above

### Documentation & Communication

- [ ] I have updated relevant documentation
- [ ] My commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) format
- [ ] Breaking changes are clearly documented

### Project-Specific

- [ ] If this PR implements an AEP, I have linked it in the "Related AEP" section above
- [ ] If this PR modifies namespaces, I have updated the namespace documentation
- [ ] If this PR affects WASM components, I have tested with the WIT definitions
- [ ] If this PR changes the SCXML runtime, I have run conformance tests
- [ ] Dependencies are up-to-date and security vulnerabilities addressed
- [ ] I have verified no sensitive data (credentials, keys, secrets) is committed

## Performance Impact

<!-- If applicable, describe any performance implications -->

- [ ] N/A - No performance impact
- [ ] Performance improved (provide benchmarks if available)
- [ ] Performance may be impacted (explain trade-offs)

**Benchmarks (if applicable):**

```
Before: [benchmark results]
After:  [benchmark results]
```

## Security Considerations

<!-- Address any security implications of your changes -->

- [ ] N/A - No security implications
- [ ] This PR addresses a security vulnerability (describe in private first!)
- [ ] I have reviewed the code for potential security issues
- [ ] No new dependencies with known vulnerabilities added
- [ ] Secrets and sensitive data handling reviewed

## Additional Context

<!-- Add any other context, screenshots, logs, or information that would be helpful for reviewers -->
<!-- Include links to related PRs, discussions, RFCs, or AEPs -->



---

## For Reviewers

<!-- This section is for reviewers - authors can skip -->

**Review focus areas:**
- [ ] Code correctness and logic
- [ ] Test coverage and quality
- [ ] Documentation completeness
- [ ] Performance implications
- [ ] Security considerations
- [ ] SCXML compliance (if applicable)
- [ ] Breaking changes properly handled
- [ ] API design (for namespace changes)

**Shepherd:** <!-- To be assigned by maintainers -->

---

**Note:** This PR template helps ensure high-quality contributions. Please fill out all applicable sections. If a section doesn't apply, mark it as "N/A" or remove it. Thank you for contributing to AgentML! 🙌

# Pull Request Templates

This directory contains specialized pull request templates for different types of contributions to AgentML.

## Available Templates

### Default Template

**Location:** `../.github/PULL_REQUEST_TEMPLATE.md`

The default template is used automatically for all standard pull requests. It covers:
- Bug fixes
- New features
- Documentation updates
- Specification updates (.xsd, .wit files)
- Example implementations
- Interface definitions
- Build/CI changes

**When to use:** For all regular code contributions, this template is applied automatically when you create a new PR.

### AEP Template (AgentML Enhancement Proposals)

**Location:** `.github/PULL_REQUEST_TEMPLATE/aep.md`

The AEP template is specifically designed for submitting AgentML Enhancement Proposals (AEPs) - design documents for significant changes to the AgentML specification, runtime, or ecosystem.

**When to use:** When proposing major architectural changes, new features requiring design review, schema changes, or runtime behavior modifications.

**How to use:** When creating a PR for an AEP, append `?template=aep.md` to the PR creation URL, or select "AEP" from the template dropdown.

Example:
```
https://github.com/agentflare-ai/agentml/compare/main...your-branch?template=aep.md
```

## How GitHub PR Templates Work

### Automatic Application

When you create a new pull request without specifying a template, GitHub automatically uses the default template located at `.github/PULL_REQUEST_TEMPLATE.md`.

### Selecting a Specific Template

You can select a specific template in two ways:

#### Method 1: Query Parameter
Add `?template=TEMPLATE_NAME.md` to the PR creation URL:

```
?template=aep.md           # Use the AEP template
```

#### Method 2: GitHub UI
If multiple templates exist, GitHub shows a template picker when you create a new PR. Select the appropriate template from the dropdown menu.

### Multiple Templates

When both `PULL_REQUEST_TEMPLATE.md` (default) and `PULL_REQUEST_TEMPLATE/` directory exist:
- The `.md` file serves as the **default template** for all PRs
- Files inside the directory are **specialized templates** selectable via dropdown or URL parameter

## Template Selection Guide

Use this guide to choose the right template:

| Type of Contribution | Template to Use | How to Access |
|---------------------|-----------------|---------------|
| Bug fix | Default | Automatic |
| New feature (standard) | Default | Automatic |
| Documentation update | Default | Automatic |
| Specification update (.xsd, .wit) | Default | Automatic |
| Example implementation | Default | Automatic |
| Interface definition | Default | Automatic |
| **AEP submission** | **AEP** | **`?template=aep.md`** |
| **Major design proposal** | **AEP** | **`?template=aep.md`** |
| **Spec changes (significant)** | **AEP** | **`?template=aep.md`** |

## Contributing

### For Contributors

1. **Read the template carefully** - Each section serves a purpose for reviewers and maintainers
2. **Fill out all applicable sections** - If a section doesn't apply, mark it as "N/A" or remove it
3. **Be thorough** - Provide sufficient context, testing results, and documentation
4. **Use conventional commits** - Follow the [Conventional Commits](https://www.conventionalcommits.org/) format
5. **Test your changes** - Ensure all tests pass before submitting

### For Maintainers

1. **Review template completeness** - Ensure contributors filled out required sections
2. **Request updates if needed** - Ask for clarification or additional information
3. **Use the checklist** - The "For Reviewers" section helps ensure thorough review
4. **Assign shepherds for AEPs** - AEP PRs require a designated shepherd

## Template Maintenance

### When to Update Templates

- **New contribution patterns emerge** - Add sections for new types of changes
- **Process changes** - Update checklists and requirements as the project evolves
- **Feedback from contributors** - Improve clarity based on common questions
- **Tool changes** - Update references to CI/CD, testing tools, or workflows

### How to Update Templates

1. Propose template changes via a regular PR (use the default template)
2. Discuss with maintainers in the PR comments
3. Update all affected templates consistently
4. Document changes in the PR description
5. Get approval from at least two maintainers

## Additional Resources

- [Creating a pull request template](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/creating-a-pull-request-template-for-your-repository) - GitHub Docs
- [AgentML Contributing Guidelines](../../docs/contributing/guidelines.mdx) - Contribution guidelines
- [AgentML AEP Process](../../aeps/README.md) - Enhancement proposal process
- [Conventional Commits](https://www.conventionalcommits.org/) - Commit message format

---

**Questions?** Open a discussion or ask in your pull request comments.

# Contributing to AgentML

Thank you for your interest in contributing to AgentML! We're building the universal language for AI agents—think HTML for the web, but for agentic systems. Your contributions help shape the future of how agents are built and deployed.

## 🌟 Why Contribute?

AgentML is built in public with the community. We're creating an open standard that:
- Eliminates vendor lock-in for AI agents
- Provides deterministic, auditable behavior via state machines
- Enables interoperability across frameworks and runtimes
- Is based on the battle-tested W3C SCXML specification

Your contributions directly impact how developers worldwide build reliable, portable agents.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Submission Guidelines](#submission-guidelines)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Enhancement Proposals (AEPs)](#enhancement-proposals-aeps)
- [Community](#community)
- [Recognition](#recognition)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please be:

- **Respectful**: Value diverse perspectives and experiences
- **Constructive**: Provide helpful feedback and be open to receiving it
- **Collaborative**: Work together toward shared goals
- **Professional**: Keep discussions focused and productive

While we don't yet have a formal Code of Conduct document, these principles guide our community. We're building this in the open—help us establish a culture of excellence and inclusivity.

## How Can I Contribute?

### 🐛 Report Bugs

Found a bug? Help us improve stability:

1. Check [existing issues](https://github.com/agentflare-ai/agentml/issues) to avoid duplicates
2. Use the bug report template when creating an issue
3. Include:
   - AgentML/agentmlx version (`agentmlx --version`)
   - Operating system and architecture
   - Minimal reproduction case (agent file or code)
   - Expected vs actual behavior
   - Error messages and stack traces

**Example:**
```markdown
## Bug: State transition fails in parallel states

**Version**: agentmlx 0.1.0  
**OS**: Ubuntu 22.04 (amd64)

**Minimal reproduction**: [attached .aml file]

**Expected**: Both parallel states complete  
**Actual**: Only first state completes  

**Error**: [paste error output]
```

### 💡 Suggest Features

Have an idea? We want to hear it:

1. Check [GitHub Discussions](https://github.com/agentflare-ai/agentml/discussions) for similar ideas
2. For minor features, open an issue with:
   - Use case description
   - Proposed solution
   - Example usage
   - Why existing approaches don't work
3. For major features, consider proposing an [AEP](#enhancement-proposals-aeps)

### 📝 Improve Documentation

Documentation is crucial for adoption:

- Fix typos and clarify confusing sections
- Add examples and tutorials
- Improve API documentation
- Create video walkthroughs
- Translate documentation (future)

Documentation lives in:
- `docs/` - Main documentation (MDX format)
- `examples/` - Sample agent files
- `README.md` - Project overview
- Namespace READMEs - Extension-specific docs

### 🔧 Submit Code Changes

Ready to code? Follow these steps:

1. **Check for existing work**: Search issues and PRs to avoid duplicates
2. **Discuss first**: For non-trivial changes, open an issue or discussion first
3. **Follow the process**: See [Submission Guidelines](#submission-guidelines) below

### 🧪 Test and Validate

Help ensure quality:

- Run SCXML conformance tests and report results
- Test on different platforms (Linux, macOS, Windows, ARM)
- Validate example agents
- Test new features with real-world use cases
- Report test coverage gaps

### 🔌 Create Extensions

Build custom namespaces:

- Implement new LLM integrations
- Add vector databases or memory systems
- Create domain-specific actions
- Build IOProcessors for new protocols

See [docs/extensions/custom.mdx](docs/extensions/custom.mdx) for guidance.

### 🗣️ Share and Advocate

- Share your AgentML agents and use cases
- Write blog posts or tutorials
- Present at conferences or meetups
- Participate in community discussions
- Help answer questions from other users

## Development Setup

### Prerequisites

**For Specification & Docs:**
- Node.js 18+ or 20+
- npm or pnpm
- Git

**For Runtime Development:**
- Go 1.21+
- Git
- Make (optional)

### Quick Start

**1. Fork and Clone**

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/agentml.git
cd agentml

# Add upstream remote
git remote add upstream https://github.com/agentflare-ai/agentml.git
```

**2. Install Dependencies**

```bash
# For documentation development
npm install

# For Go modules (if working on runtime packages)
go mod download
```

**3. Create a Branch**

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

**4. Make Changes**

Follow our [Development Guide](docs/contributing/development.mdx) for detailed setup instructions, including:
- Development workflows
- Testing strategies
- Debugging tools
- Hot reload setup

**5. Test Your Changes**

```bash
# Run tests (Go packages)
go test ./...

# Validate agents
agentmlx validate examples/your-agent.aml

# Build and test documentation
npm run docs:dev
```

## Submission Guidelines

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) for clear, semantic commit history:

```bash
# Format: <type>(<scope>): <subject>

# Examples:
git commit -m "feat(gemini): Add streaming support for Gemini extension"
git commit -m "fix(scxml): Resolve state transition bug in parallel states"
git commit -m "docs(quick-start): Update examples with event schemas"
git commit -m "test(conformance): Add datamodel tests for SCXML compliance"
git commit -m "refactor(parser): Simplify XML namespace handling"
git commit -m "perf(interpreter): Optimize state transition lookup"
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring (no functional changes)
- `perf`: Performance improvements
- `chore`: Build process, dependencies, tooling
- `style`: Formatting, missing semicolons, etc.
- `ci`: CI/CD configuration changes

**Scopes** (optional): `scxml`, `gemini`, `ollama`, `memory`, `parser`, `conformance`, etc.

### Pull Request Process

**1. Update Documentation**

- Add/update relevant docs in `docs/`
- Update README if user-facing changes exist
- Add examples to demonstrate new features
- Update namespace documentation for extensions

**2. Add Tests**

All code changes require tests:
- Unit tests for new functions/methods
- Integration tests for features
- SCXML conformance tests if affecting runtime behavior
- Example agents that demonstrate usage

**3. Run Quality Checks**

```bash
# Format code
go fmt ./...

# Run linters
golangci-lint run

# Vet code
go vet ./...

# Run all tests
go test ./... -cover

# Run conformance tests
go test ./tests/conformance/...
```

**4. Push and Create PR**

```bash
git push origin feature/your-feature-name
```

Open a pull request on GitHub using our [PR template](.github/PULL_REQUEST_TEMPLATE.md). Ensure you:
- Provide a clear description of changes
- Link related issues (`Fixes #123`, `Relates to #456`)
- Check all applicable boxes in the template
- Add visual evidence (screenshots, terminal output) for behavior changes
- Indicate if this implements an AEP

**5. Address Review Feedback**

- Respond to review comments promptly
- Make requested changes
- Ask questions if feedback is unclear
- Be open to iteration and collaboration

### Review Timeline

- **Initial review**: Within 1 week for most PRs
- **Feedback**: Constructive suggestions from maintainers
- **Iteration**: Work together to refine the contribution
- **Approval**: When ready, PR is approved by maintainers
- **Merge**: Maintainer merges to main branch

Be patient and responsive. We're all volunteers building this together.

## Coding Standards

### Go Code Style

Follow Go best practices and conventions:

```go
// ✅ Good: Clear, idiomatic Go
func ProcessEvent(ctx context.Context, event *Event) error {
    if event == nil {
        return fmt.Errorf("event cannot be nil")
    }
    
    // Process the event
    return nil
}

// ❌ Bad: Poor naming, no error handling
func pe(e interface{}) {
    // Process
}
```

**Guidelines:**
- Use `gofmt` for formatting (enforced)
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Write clear, concise comments explaining "why", not "what"
- Keep functions focused and testable (single responsibility)
- Use meaningful variable names (avoid single letters except in loops/short scopes)
- Handle all errors explicitly
- Use structured logging with context

**Required Tools:**
```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

### AgentML/SCXML Files

Follow W3C SCXML and XML best practices:

```xml
<!-- ✅ Good: Clear structure, proper indentation, meaningful IDs -->
<agent xmlns="github.com/agentflare-ai/agentml/agent"
       datamodel="ecmascript"
       import:gemini="github.com/agentflare-ai/agentml/gemini">

  <datamodel>
    <data id="user_message" expr="''" />
    <data id="response" expr="''" />
  </datamodel>

  <state id="await_input">
    <onentry>
      <log expr="'Waiting for user input'" />
    </onentry>
    
    <transition event="user.message" target="process_message" />
  </state>

  <state id="process_message">
    <!-- Process the user message with LLM -->
    <onentry>
      <gemini:generate
        model="gemini-2.0-flash-exp"
        location="_event"
        promptexpr="'Respond to: ' + user_message" />
    </onentry>
    
    <transition event="response.ready" target="send_response" />
  </state>
</agent>
```

**Guidelines:**
- Use 2-space indentation
- Close all tags properly
- Use meaningful state IDs (`process_payment` not `state_5`)
- Add comments for complex logic
- Follow W3C SCXML specification
- Use `event:schema` with detailed descriptions to guide LLMs
- Keep agent files focused (decompose complex agents with `<invoke>`)
- Prefer external scripts over inline for better tooling support

### Documentation Style

Documentation is written in MDX (Markdown with JSX):

```mdx
---
title: "Feature Name"
description: "Brief, compelling description for SEO and previews"
---

# Feature Name

Opening paragraph that clearly explains what this is and why it matters.

## Usage

Practical example showing the feature in action:

```xml
<agent xmlns="github.com/agentflare-ai/agentml/agent">
  <!-- Clear, minimal example -->
</agent>
```

## Best Practices

1. **Use descriptive state IDs**: Makes debugging easier
2. **Validate with schemas**: Ensures reliability
3. **Test edge cases**: Cover error scenarios

## Common Pitfalls

- **Issue**: Description of problem
  **Solution**: How to fix it
```

## Testing Requirements

All contributions must include appropriate tests. We aim for high test coverage and reliability.

### Go Package Tests

**Unit Tests:**
```bash
# Run all tests
go test ./...

# Run specific package
go test ./pkg/scxml/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Structure:**
```go
func TestEventMatching(t *testing.T) {
    tests := []struct {
        name     string
        event    string
        pattern  string
        expected bool
    }{
        {"exact match", "user.login", "user.login", true},
        {"wildcard match", "user.login", "user.*", true},
        {"no match", "user.login", "admin.*", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MatchEvent(tt.event, tt.pattern)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### SCXML Conformance Tests

AgentML aims for W3C SCXML conformance. Current status: **185/193 tests passing (95.9%)**

```bash
# Run conformance tests
go test ./tests/conformance/...

# Run specific conformance test
go test -run TestDatamodel ./tests/conformance/...
```

When making changes that affect runtime behavior:
1. Run conformance tests before and after
2. Document any conformance impact in your PR
3. If fixing conformance, note which tests now pass
4. Reference [W3C SCXML spec](https://www.w3.org/TR/scxml/) for clarification

### Integration Tests

```bash
# Run integration tests
go test ./tests/integration/...

# Run specific test
go test -run TestChatbotAgent ./tests/integration/...
```

### Agent Validation Tests

```bash
# Validate agent syntax
agentmlx validate examples/chatbot.aml

# Validate all examples
for agent in examples/*.aml; do
    agentmlx validate "$agent"
done
```

## Documentation

Good documentation is as important as good code.

### What to Document

**For New Features:**
- Add usage guide in `docs/`
- Include practical examples
- Document configuration options
- Add troubleshooting section

**For Extensions/Namespaces:**
- Create namespace README
- Document all actions and attributes
- Provide example agents
- Explain use cases

**For Breaking Changes:**
- Update migration guide
- Document deprecated features
- Provide timeline for removal
- Include before/after examples

**For Configuration:**
- Document environment variables
- Explain default values
- Show example configurations

### Documentation Structure

```
docs/
├── getting-started/
│   ├── overview.mdx
│   ├── installation.mdx
│   └── quick-start.mdx
├── concepts/
│   ├── state-machines.mdx
│   ├── events-schemas.mdx
│   └── namespaces.mdx
├── architecture/
│   ├── interpreter.mdx
│   ├── io-processors.mdx
│   └── namespace-system.mdx
├── extensions/
│   ├── gemini.mdx
│   ├── memory.mdx
│   └── custom.mdx
├── best-practices/
│   ├── error-handling.mdx
│   ├── performance.mdx
│   └── testing.mdx
└── contributing/
    ├── development.mdx
    └── guidelines.mdx
```

### Live Documentation Preview

```bash
npm run docs:dev
# Visit http://localhost:3000
```

## Enhancement Proposals (AEPs)

For significant changes to AgentML—such as new language features, runtime behavior modifications, or architectural decisions—please submit an **AgentML Enhancement Proposal (AEP)**.

### When to Write an AEP

**Requires an AEP:**
- New schema elements or attributes
- Changes to runtime semantics
- New standard namespaces
- Breaking changes to APIs
- Major architectural changes

**Does NOT require an AEP:**
- Bug fixes
- Documentation improvements
- Performance optimizations (unless changing semantics)
- Minor extensions or helpers
- Test improvements

### AEP Process

1. **Socialize the idea**: Discuss in [GitHub Discussions](https://github.com/agentflare-ai/agentml/discussions) first
2. **Use the template**: Copy `aeps/AEP-0000-template.md`
3. **Write the AEP**: Complete all required sections
4. **Submit as PR**: Open a PR to `aeps/` (number will be assigned)
5. **Iterate**: Address feedback during review
6. **FCP**: Final Comment Period announced by shepherd
7. **Decision**: Approvers accept or reject
8. **Implementation**: If accepted, create tracking issues and implement

See [aeps/README.md](aeps/README.md) for complete details on the AEP process, lifecycle, and requirements.

## Community

### Getting Help

- **Documentation**: Read the [full AgentML docs](docs/getting-started/overview.mdx)
- **GitHub Issues**: Search [existing issues](https://github.com/agentflare-ai/agentml/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/agentflare-ai/agentml/discussions)
- **Discord**: Join our community (coming soon)
- **Twitter/X**: Follow [@agentflareai](https://twitter.com/agentflareai)

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions, ideas, and general discussion
- **Pull Requests**: Code review and collaboration
- **Discord** (coming soon): Real-time chat and community support

### Be a Good Citizen

- **Search first**: Before asking, search existing issues and discussions
- **Be specific**: Provide context, examples, and details
- **Be patient**: Maintainers and community members are volunteers
- **Be helpful**: Answer questions when you can
- **Be respectful**: Treat others as you'd like to be treated

## Recognition

We value every contribution and recognize contributors:

- **CONTRIBUTORS.md**: All contributors listed (coming soon)
- **Release notes**: Notable contributions highlighted
- **Documentation credits**: Authors acknowledged in docs
- **Community spotlight**: Featured in discussions and social media

Your contributions, big or small, help build the universal language for AI agents. Thank you! 🙌

## License

By contributing to AgentML, you agree that your contributions will be licensed under the [MIT License](LICENSE).

## Additional Resources

- [Development Guide](docs/contributing/development.mdx) - Detailed development setup
- [Contributing Guidelines](docs/contributing/guidelines.mdx) - Additional guidelines
- [AEP Process](aeps/README.md) - Enhancement proposal workflow
- [W3C SCXML Specification](https://www.w3.org/TR/scxml/) - Reference standard
- [Effective Go](https://golang.org/doc/effective_go.html) - Go style guide

---

**Questions?** Open a [discussion](https://github.com/agentflare-ai/agentml/discussions) or reach out to the maintainer team. We're here to help you succeed in contributing to AgentML!

**Let's build the future of AI agents together.** ✨


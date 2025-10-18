# AgentML Enhancement Proposals (AEPs)

This directory contains AgentML Enhancement Proposals (AEPs), the design documents used to propose, discuss, and document significant changes to the AgentML specification, runtime implementations, tooling, and ecosystem.

## What are AEPs?

AEPs are structured documents that describe proposed enhancements to AgentML. They serve as the primary mechanism for proposing major features, collecting community input on design decisions, and documenting the rationale behind changes to the AgentML ecosystem.

An AEP should provide a clear technical specification, motivating use cases, and consideration of alternatives. The process ensures that changes to AgentML are well-vetted, documented, and aligned with the project's goals.

## Why AEPs Exist

AEPs serve several critical purposes:

- **Transparency**: All stakeholders can see what changes are being proposed and why
- **Discussion**: Provide a focal point for technical debate and consensus-building
- **Documentation**: Create a historical record of design decisions and their rationale
- **Quality**: Ensure thorough review before committing to significant changes
- **Coordination**: Help maintainers and contributors align on roadmap and priorities

AEPs are used for changes that impact the schema definition language, runtime behavior, standard library, tooling interfaces, or cross-cutting architectural concerns. Routine bug fixes, documentation improvements, and minor enhancements typically do not require an AEP.

## Directory Layout

```
aeps/
   README.md              # This file
   AEP-0000-template.md   # Template for new proposals
   AEP-####-title.md      # Individual proposals
```

Each AEP is a standalone Markdown file. Supporting materials (diagrams, code samples, etc.) may be included in subdirectories named `AEP-####-assets/`.

## Proposal Lifecycle

AEPs progress through a defined lifecycle with clear state transitions:

### 1. Idea
Initial concept discussed informally (e.g., GitHub Discussions, community forums). Not yet formalized as an AEP.

### 2. Draft
Author creates an AEP document using the template and opens a pull request. The proposal is assigned a number and enters community review. Feedback is incorporated iteratively.

### 3. Review
The AEP has been refined and is undergoing active technical review. Shepherds and reviewers provide detailed feedback. The author addresses concerns and updates the proposal.

### 4. Final Comment Period (FCP)
Major concerns have been addressed. The Shepherd announces an FCP (typically 7-14 days) for final objections. The community is invited to provide last comments before a decision.

### 5. Accepted
The AEP has been approved by designated Approvers. The proposal is now part of the official roadmap and ready for implementation.

### 6. Implemented
The changes described in the AEP have been implemented in the relevant repositories (schema, runtime, tooling, etc.).

### 7. Released
The implementation has been included in a versioned release of AgentML. The AEP is considered complete.

**Alternative outcomes:**
- **Rejected**: The proposal was declined during review or FCP
- **Withdrawn**: The author chose to withdraw the proposal
- **Superseded**: Replaced by a newer AEP

## Roles

### Author
The individual(s) who write and champion the AEP. Responsible for incorporating feedback, updating the proposal, and driving it through the lifecycle.

### Shepherd
An experienced maintainer assigned to guide the AEP through the process. Shepherds help refine the proposal, facilitate discussion, and determine when to move to FCP.

### Reviewer
Community members and maintainers who provide technical feedback on the proposal. Anyone may participate as a reviewer.

### Approver
Designated maintainers with authority to accept or reject AEPs. Typically core team members or domain experts with decision-making responsibility.

## File Naming Convention

AEPs follow a strict naming convention:

```
AEP-####-short-title.md
```

- **####**: Zero-padded four-digit number (e.g., `0001`, `0042`, `0123`)
- **short-title**: Lowercase, hyphen-separated descriptive title
- **Extension**: Always `.md` (Markdown)

**Examples:**
```
AEP-0000-template.md
AEP-0015-runtime-plugin-api.md
AEP-0042-schema-version-negotiation.md
AEP-0108-tool-use-tracing.md
```

Numbers are assigned sequentially by maintainers when the AEP PR is opened.

## How to Submit a Proposal

1. **Socialize the idea**: Discuss your concept in GitHub Discussions or community channels to gauge interest and refine the approach

2. **Copy the template**: Use `AEP-0000-template.md` as a starting point for your proposal

3. **Write the AEP**: Complete all required sections with sufficient technical detail

4. **Open a pull request**: Submit your AEP as a PR to this repository. Leave the number as `####` initially; a maintainer will assign the official number

5. **Iterate**: Address feedback from reviewers and the assigned Shepherd. Update the PR as needed

6. **FCP and decision**: Once major concerns are resolved, the Shepherd will announce FCP. After FCP concludes, Approvers will accept or reject the proposal

7. **Implementation tracking**: If accepted, create implementation tracking issues and link them in the AEP

## Acceptance Rules

For an AEP to be accepted, it must:

- Address a clear need with demonstrated use cases
- Provide a complete technical specification
- Consider alternatives and explain why they were not chosen
- Document backward compatibility impact and migration path
- Achieve consensus among reviewers during FCP
- Receive explicit approval from at least two Approvers

Approvers may reject an AEP if it:

- Conflicts with project goals or philosophy
- Introduces unacceptable complexity or maintenance burden
- Has unresolved technical or design concerns
- Lacks sufficient community support

## GitHub Labels

AEPs use the following labels for tracking:

- `aep`: Marks a pull request or issue as related to an AEP
- `aep:draft`: Proposal is in draft state
- `aep:review`: Proposal is under active review
- `aep:fcp`: Proposal is in Final Comment Period
- `aep:accepted`: Proposal has been accepted
- `aep:rejected`: Proposal was rejected
- `aep:withdrawn`: Author withdrew the proposal
- `aep:implemented`: Changes have been implemented
- `aep:released`: Implementation is included in a release

## Example AEP

**Filename**: `AEP-0015-runtime-plugin-api.md`

**Summary**: Proposes a standardized plugin API for AgentML runtimes, enabling third-party extensions to hook into execution lifecycle events, modify tool resolution, and inject custom observability instrumentation.

**Status**: Accepted (FCP completed 2024-02-15, implementation in progress)

## Additional Resources

- [CONTRIBUTING.md](../CONTRIBUTING.md) - General contribution guidelines
- [GOVERNANCE.md](../GOVERNANCE.md) - Project governance model and decision-making process
- [AgentML Specification](../agentml.xsd) - Core schema and runtime specification
- [GitHub Discussions](https://github.com/agentflare/agentml/discussions) - Discuss ideas before drafting an AEP

---

**Questions?** Open a discussion or reach out to the maintainer team.

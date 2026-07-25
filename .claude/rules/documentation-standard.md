# Documentation standard

When generating or updating documentation, produce documentation that is optimized for both human developers and AI assistants.

## Principles

- Write for humans first; structure for AI retrieval.
- Be concise, explicit, and avoid unnecessary prose.
- Explain both **what** the component does and **why** it exists.
- State assumptions, constraints, and invariants explicitly.
- Use consistent terminology throughout the project.
- Prefer short, self-contained sections over long narratives.
- Use Markdown headings, bullet points, and tables where appropriate.
- Make every document understandable without requiring unrelated context.

## Standard structure

Each document should include, when applicable:

1. Overview — Purpose, Responsibilities, Scope
2. Architecture — Position within the system, Upstream dependencies, Downstream consumers
3. Data Flow — Inputs, Processing, Outputs
4. Business Rules — Validation rules, Constraints, Assumptions, Invariants
5. Dependencies — Internal modules, External services, Database tables, Configuration
6. API / Interface — Public methods/endpoints, Parameters, Return values, Error cases
7. Security — Authentication, Authorization, Sensitive data handling
8. Performance — Caching, Concurrency, Scalability considerations
9. Failure Modes — Expected failures, Recovery strategy, Logging and monitoring
10. Related Components
11. Future Considerations — Known limitations, Technical debt, Planned improvements

## Rules

- Keep documents modular and focused on a single topic.
- Avoid ambiguous or inferred behavior—be explicit.
- Document architectural decisions and the rationale behind them.
- Cross-reference related documents instead of duplicating content.
- Prefer examples over lengthy explanations.
- Include Mermaid diagrams when they improve understanding.
- Treat documentation as part of the codebase and keep it synchronized with implementation.

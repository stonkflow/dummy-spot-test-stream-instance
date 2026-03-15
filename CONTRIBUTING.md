Contributing Guide
===

# Workflow

## 0. Fork and Clone the Repository

Before making any changes, [fork](https://github.com/stonkflow/dummy-spot-test-stream-instance/fork) the repository to your own GitHub account and clone it to your local machine.

## 1. Open an Issue

Before starting any work, please [open](https://github.com/stonkflow/dummy-spot-test-stream-instance/issues/new) an issue to describe the problem or proposed change.

This helps maintainers and contributors:

- understand the motivation for the change
- discuss possible solutions
- avoid duplicated work
- agree on the scope before implementation

When creating an issue, try to include:

- a clear and concise description of the problem or idea
- the expected behavior or outcome
- relevant context (use cases, examples, logs, etc.)
- if applicable, a brief outline of the proposed approach

Once the issue is discussed and accepted, you can proceed with the implementation.

## 2. Create a New Branch

After the issue has been discussed and accepted, create a new branch for your work.

Branches should be created from the latest `main` branch and should have a descriptive name that reflects the purpose of the change.

## 3. Make Changes and Commit

Implement your changes in the created branch. Keep commits focused, small, and logically separated when possible.

Contributors are encouraged to follow the [Conventional Commits](https://www.conventionalcommits.org/) specification when writing commit messages. This helps maintain a clean project history and improves readability of the change log.

When preparing commits for a pull request, try to organize them by purpose:

- **refactor** — code changes that restructure existing code without adding functionality, preparing the codebase for new functionality.
- **feat** — implementation of new functionality.
- **test** — adding or updating unit tests.
- **docs** — updates or additions to project documentation (`.md` files).
- **chore** — maintenance tasks and other changes that do not fall into the categories above.

Example commit messages:

```
refactor: simplify kafka consumer initialization
feat: add market data stream iterator
test: add unit tests for websocket reconnect logic
docs: update README with usage example
chore: update CI configuration
```

Keeping commits well-structured makes the review process easier and helps maintain a clear and understandable project history.

## 4. Open a Pull Request

Once your changes are ready, [open](https://github.com/stonkflow/dummy-spot-test-stream-instance/compare) a pull request against the `main` branch.

In the pull request description, provide:

- a clear summary of the changes
- any relevant notes that may help reviewers understand the implementation

The pull request should also reference the related issue. To link the pull request to the issue and automatically close it after merging, include:

```
Close #<issue-number>
```

Example:

```
Close #42 This ensures the pull request is properly connected to the corresponding task and keeps the project workflow organized.
```

# Breaking change

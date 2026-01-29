# Feature Ideas & Roadmap

This document outlines planned features, improvements, and maintenance tasks for the Daily Dash application.

## 🛠 User Experience & Customization

### Location & Configuration
- **Road Condition Selection**: Implement a UI for selecting the region or specific road segments for condition reports.
- **Weather Location Selection**: Allow users to search for and set their preferred city/location for weather updates.
- **Persistent Settings**: Ensure that user selections (bus stops, weather location, road regions) are saved to a config file and restored upon restarting the application.

### Search & Filtering Improvements
- **Cleaner List View**: Remove duplicate entries and hide internal IDs from the main list display to reduce clutter.

## 📚 Documentation
- **Component Guides**: Create detailed documentation for each dashboard widget (Transit, Weather, Road), explaining how they work, their API sources, and configuration options.
- **Skills & Workflows**: Document standardized workflows ("skills") for common developer tasks like adding a new API source.

## ⚙️ Development & Quality Assurance

### Automation
- **Linting & Formatting**: Set up automated tooling (Makefiles, pre-commit hooks) to enforce code formatting and linting standards automatically.
- **CI/CD**: Explore options for continuous integration checks.

### Code Quality & Security
- **Code Review**: Conduct a thorough review of the codebase to identify refactoring opportunities and architectural improvements.
- **Security Review**: Audit dependencies and API handling logic for potential security vulnerabilities.
- **Test Coverage**: Increase the coverage of automated tests, particularly for the UI update loops and API error handling scenarios.

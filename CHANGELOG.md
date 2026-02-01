# Changelog

All notable changes to the Hopfield Assignment Problem Solver will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-01

### Added
- Comprehensive test suite with 29 unit tests and 13 integration tests
- Docker containerization with health checks and service dependencies
- Nginx reverse proxy with load balancing configuration
- Makefile for build automation and development workflow
- API documentation and deployment guides
- Contributing guidelines and development setup instructions

### Fixed
- Go module import paths corrected for proper package structure
- Hopfield algorithm activation function now handles both scalars and numpy arrays
- API error handling improved for missing cost_matrix fields and empty requests
- Matrix validation now properly rejects non-square matrices
- Health check configuration fixed to use curl instead of missing requests library
- Service startup dependencies restored for proper container orchestration

### Changed
- Updated Python version requirement to 3.11+
- Improved error messages and validation logic
- Enhanced test expectations to match algorithm behavior

### Technical Details
- **Go API Gateway**: Built with Gin framework, handles REST endpoints
- **Python Hopfield Solver**: Flask-based service with numpy implementation
- **Containerization**: Multi-service Docker setup with health monitoring
- **Testing**: 100% pass rate across all test suites
- **Documentation**: Complete API reference and deployment guides</content>
<parameter name="filePath">/home/roberto/repos/HopfieldAssigmentProblemSolver/CHANGELOG.md
# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do Not** disclose publicly

Please do not open a public GitHub issue for security vulnerabilities.

### 2. Report privately

Send an email to [security@yourproject.com](mailto:security@yourproject.com) with:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### 3. Response timeline

- **Initial response**: Within 48 hours
- **Status update**: Within 1 week
- **Fix timeline**: Depends on severity
  - Critical: 1-7 days
  - High: 1-2 weeks
  - Medium: 2-4 weeks
  - Low: Next release cycle

### 4. Disclosure process

1. We will investigate and validate the vulnerability
2. We will develop and test a fix
3. We will release a security patch
4. After the patch is released, we will publicly disclose the vulnerability

## Security Best Practices

When using this API:

1. **API Keys**
   - Never commit API keys to version control
   - Rotate API keys regularly
   - Use environment variables for API keys
   - Limit API key permissions to minimum required

2. **Network Security**
   - Use HTTPS in production
   - Implement rate limiting
   - Use firewall rules to restrict access
   - Monitor for unusual activity

3. **Input Validation**
   - Always validate cost matrix size
   - Check for malformed requests
   - Implement request size limits

4. **Authentication**
   - Use strong API keys (minimum 32 characters)
   - Implement API key expiration
   - Log all authentication failures

5. **Dependencies**
   - Keep dependencies updated
   - Use Dependabot for automatic updates
   - Review security advisories regularly

## Known Security Considerations

### Rate Limiting

The API includes basic rate limiting, but you should implement additional controls at the infrastructure level for production deployments.

### Input Validation

Matrix size is limited to 50x50 to prevent resource exhaustion attacks. Cost values are checked for NaN/Inf to prevent algorithm failures.

### Authentication

API key authentication is optional by default (for development). **You must enable it in production** by setting the `API_KEY` environment variable.

## Security Updates

We will announce security updates through:
- GitHub Security Advisories
- Release notes
- Email notifications (for registered users)

## Bug Bounty Program

We do not currently have a bug bounty program, but we appreciate responsible disclosure and will acknowledge contributors in our security advisories.

## Questions?

For security-related questions, contact [security@yourproject.com](mailto:security@yourproject.com).

For general questions, open a GitHub issue.

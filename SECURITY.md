# Security Policy

## Protecting identity credentials

- Never commit connection strings, storage account keys, tokens, or identity credentials.
- Keep them in Azure configuration (`SQLConnectionString`, `CloudConnectionString`), environment variables, or GitHub Actions secrets.
- Enable **Secret scanning** and **Push protection** under Settings > Code security.
- If a credential is ever committed, revoke and rotate it immediately; deleting the commit is not enough.

## Reporting a vulnerability

Please use GitHub's private vulnerability reporting (Security tab > Report a vulnerability) instead of opening a public issue.

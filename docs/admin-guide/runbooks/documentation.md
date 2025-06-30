# Documentation (this website)

Documents can be viewed at <https://ubiquitycluster.github.io/ubiquity/>.

## GitHub Pages Setup

The documentation is automatically built and deployed using GitHub Actions to GitHub Pages with a custom domain.

### Configuration Files

- **MkDocs Configuration**: `mkdocs.yml` - Site configuration and navigation
- **GitHub Actions**: `.github/workflows/docs.yml` - Automated build and deployment
- **Dependencies**: `requirements.txt` - Python package dependencies
- **Custom Domain**: `docs/CNAME` - Custom domain configuration

## Local Development

To edit and view locally, run:

```bash
# Install dependencies
pip install -r requirements.txt

# Serve locally with hot reload
mkdocs serve
```

Then visit [localhost:8000](http://localhost:8000)

Alternatively, if available:
```bash
make docs
```

## Deployment

Documentation is automatically deployed when:
- Changes are pushed to the `main` branch
- Files in the `docs/` directory are modified
- The `mkdocs.yml` configuration is updated

The GitHub Action will:
1. Build the static site using MkDocs Material
2. Deploy to GitHub Pages
3. Serve at the standard GitHub Pages URL `ubiquitycluster.github.io/ubiquity/`

## Features

- **Material Design**: Modern, responsive design
- **Search**: Full-text search functionality  
- **Navigation**: Expandable navigation with indexes
- **Code Copying**: Copy buttons for code blocks
- **Mermaid Diagrams**: Support for Mermaid diagrams
- **Emoji Support**: GitHub-style emoji rendering

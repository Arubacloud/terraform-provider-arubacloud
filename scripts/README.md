# Documentation Scripts

This directory contains scripts to enhance the generated Terraform provider documentation.

## format-docs.sh / format-docs.ps1

These scripts post-process the generated documentation to separate **Arguments** (input fields) from **Attributes** (output fields), making it clearer for users which fields they can configure and which are read-only outputs.

### Usage

The script is automatically run as part of `make docs`, but you can also run it manually:

**On Linux/macOS/WSL:**
```bash
bash scripts/format-docs.sh docs
```

**On Windows (PowerShell):**
```powershell
.\scripts\format-docs.ps1 -DocsDir docs
```

### What it does

The script transforms the schema sections in generated documentation from:

```markdown
## Schema

### Required
- `field1` (String) Description

### Optional
- `field2` (String) Description

### Read-Only
- `id` (String) Resource ID
```

To:

```markdown
## Schema

### Arguments

The following arguments are supported:

#### Required
- `field1` (String) Description

#### Optional
- `field2` (String) Description

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only
- `id` (String) Resource ID
```

This makes it clearer that "Arguments" are inputs and "Attributes" are outputs.

---
page_title: "OpenTofu Registry Submission - ArubaCloud Provider"
subcategory: "Development"
description: |-
  Step-by-step guide to publishing the ArubaCloud provider to the OpenTofu registry.
---

# OpenTofu Registry Submission

This guide documents the one-time process of submitting the ArubaCloud provider to the [OpenTofu Registry](https://registry.opentofu.org). Once done, users can install the provider via `tofu init` using the source `registry.opentofu.org/Arubacloud/arubacloud`.

The OpenTofu Registry is separate from the Terraform Registry. Publishing to `registry.terraform.io` does not automatically make the provider available on `registry.opentofu.org`.

## Prerequisites

- Maintainer access to the `Arubacloud/terraform-provider-arubacloud` GitHub repository.
- The GPG private key used to sign releases (stored as the `GPG_PRIVATE_KEY` repository secret).
- At least one tagged GitHub release with signed artifacts already published.

## Steps

### 1. Export the GPG public key

```bash
# List keys to find the fingerprint
gpg --list-secret-keys --keyid-format LONG

# Export the public key in ASCII-armored format
gpg --armor --export <KEY_FINGERPRINT> > arubacloud-gpg-public.asc
```

Note the 16-character key ID (last 16 chars of the fingerprint).

### 2. Fork the opentofu/registry repository

Fork https://github.com/opentofu/registry and clone it locally.

### 3. Create the provider directory

```bash
mkdir -p providers/a/arubacloud
```

### 4. Create providers/a/arubacloud/arubacloud.json

```json
{
  "versions": []
}
```

The registry bot populates the `versions` array automatically by scanning GitHub releases. Leave it empty.

### 5. Create providers/a/arubacloud/arubacloud-keys.json

```json
{
  "keys": [
    {
      "key_id": "<16-CHAR-KEY-ID>",
      "ascii_armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\n<base64-encoded-key>\n-----END PGP PUBLIC KEY BLOCK-----\n"
    }
  ]
}
```

Replace `<16-CHAR-KEY-ID>` with the key ID from step 1 and paste the full ASCII-armored public key as a single JSON string with `\n` for line breaks.

A helper to format the key as a JSON string:

```bash
python3 -c "
import json, sys
key = open('arubacloud-gpg-public.asc').read()
print(json.dumps(key))
"
```

### 6. Open a PR to opentofu/registry

Push your branch and open a pull request. The registry bot (`@opentofu-registry-bot`) will:

1. Verify the GPG signature on the latest GitHub release checksum file.
2. Auto-merge the PR if verification passes.

Monitor the PR for bot feedback. Common issues:

| Error | Fix |
|---|---|
| `signature verification failed` | The key in `arubacloud-keys.json` does not match the key used to sign the release. Re-export and re-check. |
| `no releases found` | Ensure the GitHub repo has at least one tagged release with `*_SHA256SUMS` and `*_SHA256SUMS.sig` artifacts. |
| `key_id mismatch` | The `key_id` field must be exactly 16 hexadecimal characters (uppercase). |

### 7. Verify

Once the PR is merged, the provider should appear at:

```
https://registry.opentofu.org/providers/Arubacloud/arubacloud
```

Test with a minimal configuration:

```hcl
terraform {
  required_providers {
    arubacloud = {
      source  = "registry.opentofu.org/Arubacloud/arubacloud"
      version = ">= 0.3.0"
    }
  }
}
```

Run `tofu init` — the provider should download successfully.

## Future releases

Once registered, new GitHub releases are picked up automatically by the OpenTofu registry bot — no further manual steps are needed per release.

The release workflow (`.github/workflows/release.yml`) already signs the `SHA256SUMS` file with GPG, which is the artifact the OpenTofu registry bot verifies.

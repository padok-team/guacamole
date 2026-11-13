# Guacamole test fixtures

Hand-crafted Terraform modules and Terragrunt layers used to exercise every
Guacamole static check. The folder is split into **`pass/`** (fixtures that must
produce **no** error) and **`fail/`** (fixtures that must trigger **exactly one**
check each), so the expected result of a run is unambiguous.

Each `fail/` fixture is named after the check it targets (`<check_id>_<reason>`)
and is crafted to trip **only** that check — it stays clean for every other one.
This way, running the whole `fail/` folder lets you match one ❌ to one fixture.

## Layout

```
tests/
├── modules/                        # Terraform modules -> `guacamole static module`
│   ├── pass/
│   │   └── well_formed/            # everything correct: all module checks OK
│   └── fail/
│       ├── tf_mod_001_remote_module_not_pinned/
│       ├── tf_mod_002_provider_in_module/
│       ├── tf_mod_003_required_provider_operator/
│       ├── tf_nam_001_resource_not_this/
│       ├── tf_nam_002_not_snake_case/
│       ├── tf_nam_003_stuttering/
│       ├── tf_nam_004_var_number_mismatch/
│       ├── tf_nam_005_multiple_named_this/
│       ├── tf_var_001_missing_description/
│       ├── tf_var_002_type_any/
│       ├── tf_out_001_missing_description/
│       └── tf_dat_001_datasource_computed/
└── layers/                         # Terragrunt layers -> `guacamole static layer`
    ├── root.hcl                    # shared parent config (find_in_parent_folders)
    ├── modules/bucket/             # local module called by the TG_INP_001 fixtures
    ├── pass/
    │   ├── app/                    # inputs spread across files, no duplicate
    │   ├── tg_inp_001_declared_inputs/        # every input is declared in the module
    │   ├── tg_inp_001_registry_source/        # remote source: layer skipped
    │   └── tg_inp_001_missing_local_module/   # unreadable module: layer skipped, run goes on
    └── fail/
        ├── app/                    # `region` duplicated across included files
        ├── tg_inp_001_undeclared_input/           # `unknown_input` not declared in the module
        └── tg_inp_001_undeclared_input_in_merge/  # same, given through `merge()`
```

## How to run

Layer checks can be run from any directory, including the repo root.

```bash
# Option A — run inside the Docker test image (fixtures baked in at /tests,
# see scripts/docker-test-version.sh):
./scripts/docker-test-version.sh
# then, inside the container:
guacamole static module -p /tests/modules/pass    # expect: all ✅ (100%)
guacamole static module -p /tests/modules/fail -v # expect: every check ❌
guacamole static layer  -p /tests/layers/pass     # expect: all ✅ (100%)
guacamole static layer  -p /tests/layers/fail     # expect: TG_DRY_001 ❌ + TG_INP_001 ❌

# Option B — run the local binary from anywhere:
guacamole static module -p /path/to/guacamole/tests/modules/pass   # expect: all ✅ (100%)
guacamole static module -p /path/to/guacamole/tests/modules/fail   # expect: every check ❌
guacamole static layer  -p /path/to/guacamole/tests/layers/pass    # expect: all ✅ (100%)
guacamole static layer  -p /path/to/guacamole/tests/layers/fail    # expect: TG_DRY_001 ❌ + TG_INP_001 ❌

# Add -v to see, for each failing check, the exact fixture path that triggered it.
```

## Coverage

| Check        | Description                                                        | `fail/` fixture                          |
| ------------ | ------------------------------------------------------------------ | ---------------------------------------- |
| `TF_MOD_001` | Remote module call should be pinned to a specific version          | `tf_mod_001_remote_module_not_pinned`    |
| `TF_MOD_002` | Provider should be defined by the consumer of the module           | `tf_mod_002_provider_in_module`          |
| `TF_MOD_003` | Required provider versions should use the `~>` operator            | `tf_mod_003_required_provider_operator`  |
| `TF_NAM_001` | Unique-type resource should be named `this` / `these`              | `tf_nam_001_resource_not_this`           |
| `TF_NAM_002` | `snake_case` should be used for all resource names                 | `tf_nam_002_not_snake_case`              |
| `TF_NAM_003` | No stuttering between a resource name and its type                 | `tf_nam_003_stuttering`                  |
| `TF_NAM_004` | A collection-typed variable should have a plural name or collection-equivalent word | `tf_nam_004_var_number_mismatch` |
| `TF_NAM_005` | Several resources of a type must not be named `this` / `these`     | `tf_nam_005_multiple_named_this`         |
| `TF_VAR_001` | A variable should contain a description                            | `tf_var_001_missing_description`         |
| `TF_VAR_002` | A variable should declare a specific type (not `any`)              | `tf_var_002_type_any`                    |
| `TF_OUT_001` | An output should contain a description                             | `tf_out_001_missing_description`         |
| `TF_DAT_001` | A data source should not depend on a value computed during apply   | `tf_dat_001_datasource_computed`         |
| `TG_DRY_001` | No duplicate inputs within a layer                                 | `layers/fail/app`                        |
| `TG_INP_001` | Terragrunt inputs should be declared as variables in the module    | `layers/fail/tg_inp_001_undeclared_input` |

### TG_INP_001 notes

- `terraform.source` is evaluated by Terragrunt: the `tg_inp_001_undeclared_input` /
  `tg_inp_001_declared_inputs` pair builds it with `get_terragrunt_dir()` in an
  included file, with inputs in another one (context pattern).
- Only local sources are checked. Remote ones (`registry_source`) and unreadable
  ones (`missing_local_module`) are skipped, with a warning for the latter.
- Known limitation: inputs that are not an object literal (e.g.
  `yamldecode(file(...))`, `local.x`, or a `merge()` containing one) are not read,
  so nothing is reported.

### Not covered here

`TF_MOD_004` (use `for_each` instead of `count`) is a **state** check
(`guacamole state`): it reads a real Terraform state/plan and therefore needs an
initialized layer with provider access, which cannot be reproduced with static
fixtures.

### On pull requests

The [`iac-score`](../.github/workflows/iac-score.yaml) workflow builds
Guacamole from the pull request and runs `guacamole ci` on every fixture of this
folder. The result is posted as a PR comment: each `fail/` fixture should show
its single ❌ and each `pass/` fixture 100%. When adding a check, add its
fixtures here to see it covered in the comment.

## Isolation tricks

Some checks overlap by nature, so a few fixtures use two resources of the same
type on purpose:

- `tf_nam_002` / `tf_nam_003` declare **two** resources of the same type so that
  `TF_NAM_001` (which only fires when a type is unique in the module) stays green
  and the fixture demonstrates a single check.

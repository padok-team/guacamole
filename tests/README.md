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
    ├── pass/app/                   # inputs spread across files, no duplicate
    └── fail/app/                   # `region` duplicated across included files
```

## How to run

> ⚠️ **Run Guacamole from a directory that is _not_ a parent of these fixtures**
> (e.g. from `/tmp`, using an absolute `-p`, or inside the Docker test image).
> Terragrunt's layer discovery treats any path located under the current working
> directory as its download dir and silently skips it, so
> `guacamole static layer` finds **0 layers** when you run it from the repo root
> against `./tests`. This only affects local runs from the repo; in real usage
> the scanned codebase is a separate directory. See `test-version.sh`, which
> bakes these fixtures into the image at `/tests`.

```bash
# From the repo root, point at an absolute path from a neutral CWD:
cd /tmp
guacamole static module -p /path/to/guacamole/tests/modules/pass   # expect: all ✅ (100%)
guacamole static module -p /path/to/guacamole/tests/modules/fail   # expect: every check ❌
guacamole static layer  -p /path/to/guacamole/tests/layers/pass    # expect: TG_DRY_001 ✅
guacamole static layer  -p /path/to/guacamole/tests/layers/fail    # expect: TG_DRY_001 ❌

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
| `TF_NAM_004` | A collection-typed variable should have a plural name              | `tf_nam_004_var_number_mismatch`         |
| `TF_NAM_005` | Several resources of a type must not be named `this` / `these`     | `tf_nam_005_multiple_named_this`         |
| `TF_VAR_001` | A variable should contain a description                            | `tf_var_001_missing_description`         |
| `TF_VAR_002` | A variable should declare a specific type (not `any`)              | `tf_var_002_type_any`                    |
| `TF_OUT_001` | An output should contain a description                             | `tf_out_001_missing_description`         |
| `TF_DAT_001` | A data source should not depend on a value computed during apply   | `tf_dat_001_datasource_computed`         |
| `TG_DRY_001` | No duplicate inputs within a layer                                 | `layers/fail`                            |

### Not covered here

`TF_MOD_004` (use `for_each` instead of `count`) is a **state** check
(`guacamole state`): it reads a real Terraform state/plan and therefore needs an
initialized layer with provider access, which cannot be reproduced with static
fixtures.

## Isolation tricks

Some checks overlap by nature, so a few fixtures use two resources of the same
type on purpose:

- `tf_nam_002` / `tf_nam_003` declare **two** resources of the same type so that
  `TF_NAM_001` (which only fires when a type is unique in the module) stays green
  and the fixture demonstrates a single check.

# Craft 0.1.41 language example

This project demonstrates multiline expressions, trailing commas, a project module shared by two files, and an explicit `import module`.

From this directory, run `craft check`, `craft fmt`, `craft fmt --check`, then `craft run`. The expected printed value is `12`; the second format pass should make no changes. The private `add` function is shared inside `demo.numbers` but cannot be called from `main.craft` through the import. This example is ready for user testing; it has not been run during implementation.

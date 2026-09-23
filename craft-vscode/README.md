# Craft Language Support 0.2.1

Local extension ID: `craft-local.craft-language-support`. Requires VS Code 1.85+.
Base language helpers target Craft 0.1.2; Task/Timer and the `task3`/`timer3`
snippets require CLI 0.1.3. Imports/exports, Fn/Bytes/Context/HTTP types and
`import4`/`fn4`/`http4` snippets require CLI 0.1.4. Highlighting itself needs neither Craft nor Go.

## Install and choose colors

1. In **VS Code**, open Extensions → `…` → **Install from VSIX…** and select
   `dist/craft-language-support-0.2.1.vsix`. This updates the same extension ID;
   only the latest language VSIX is needed. Do not use Visual Studio Installer.
2. Open a `.craft` file and confirm the language mode is Craft.
3. Run **Preferences: Color Theme** and choose **Craft Dark** or **Craft Light**.
   Dark+, Light+ and your existing theme remain available; installation does not switch themes.
4. For icons, install the separate Craft Forge File Icons VSIX and run
   **Preferences: File Icon Theme**. Color Theme and File Icon Theme are separate settings.

Open `fixtures/highlighting.craft` to compare comments, Thai, strings, escapes,
numbers, nested types, comparisons, selections and deliberately incomplete code.
Use **Developer: Inspect Editor Tokens and Scopes** to inspect expected scopes.
Try zoom 100%, 125%, 150%, plus the built-in High Contrast themes.
The fixture is an editor gallery, not a runnable Craft program.

## Editing

Line comments, braces, brackets, parentheses, quote pairing, indentation-based
folding and snippets are included. Angle brackets are not auto-closed.
Snippet prefixes: `main`, `func`, `struct`, `let`, `var`, `if`, `iflet`, `while`,
`for`, `foreach`, `try`, `defer`, `test`, `task3`, `timer3`, `import4`, `fn4`, `http4`.
Choose another language mode or uninstall this extension in VS Code to revert;
the icon extension and `.toml`/`.json` associations are independent.

The grammar is lexical, not a type checker. Ambiguous identifiers keep a general
variable scope. It does not validate programs or implement semantic tokens,
hover, completion, definition, Problems integration or Format Document.
These remain separate level-B work after the CLI protocol design in Rev.3.
There is no workspace execution, activation code, telemetry or settings mutation.

## Build and user tests

From the repository root: `powershell -ExecutionPolicy Bypass -File scripts/build.ps1 -Language`.
From this directory: `npm install --ignore-scripts`, then `npm test`.
Tests use the TextMate/Oniguruma tokenizer and include a 4.5:1 contrast assertion
for token colors against background, current line and selection. Prepared tests
are not evidence that they passed; visual acceptance is performed by the user.

## Provenance

Grammar and both palettes were created for this project. No third-party theme
assets are embedded. MIT license follows the existing local icon extension's
license convention; `craft-local` does not claim a Marketplace publisher account.
Implementation references: [VS Code Syntax Highlight Guide](https://code.visualstudio.com/api/language-extensions/syntax-highlight-guide)
and [Language Configuration Guide](https://code.visualstudio.com/api/language-extensions/language-configuration-guide).


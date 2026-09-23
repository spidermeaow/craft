// User-run tests. Uses VS Code's TextMate/Oniguruma tokenizer, not JS regex emulation.
const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const tm = require('vscode-textmate');
const onig = require('vscode-oniguruma');
const root = path.resolve(__dirname, '..');
const ready = (async () => {
  const wasm = fs.readFileSync(require.resolve('vscode-oniguruma/release/onig.wasm'));
  await onig.loadWASM(wasm.buffer.slice(wasm.byteOffset, wasm.byteOffset + wasm.byteLength));
  const registry = new tm.Registry({
    onigLib: Promise.resolve({createOnigScanner: p => new onig.OnigScanner(p), createOnigString: s => new onig.OnigString(s)}),
    loadGrammar: async scope => scope === 'source.craft' ? tm.parseRawGrammar(fs.readFileSync(path.join(root, 'syntaxes/craft.tmLanguage.json'), 'utf8'), 'craft.json') : null
  });
  return registry.loadGrammar('source.craft');
})();

test('expected token scopes and boundaries', async () => {
  const grammar = await ready;
  for (const [line, needle, scope] of [
    ['let gift: String = "if // var"', 'let', 'storage.type'],
    ['let gift: String = "if // var"', 'gift', 'variable'],
    ['let gift: String = "if // var"', 'if // var', 'string.quoted.double'],
    ['let ifไทย: Int = 1', 'ifไทย', 'variable'],
    ['let ค่า: Float = 1e-3', '1e-3', 'constant.numeric'],
    ['for i in 1..3 { }', '..', 'keyword.operator'],
    ['// if true "String"', 'if', 'comment.line'],
    ['func ฝาก(บัญชี: Account): Int {', 'ฝาก', 'entity.name.function'],
    ['struct Account {', 'Account', 'entity.name.type'],
    ['let account: Account = value', 'Account', 'entity.name.type'],
    ['let a: Map<String, Map<String, Int?>> = {}', 'Map', 'support.type'],
    ['let a: Int[]? = null', 'null', 'constant.language'],
    ['std.timer.after(delay, callback)', 'std.timer', 'support.namespace'],
    ['thing.trim().upper()', 'upper', 'entity.name.function'],
    ['call(name: "value")', 'name', 'variable.other.property'],
    ['enum import lambda async await', 'enum', 'variable.other'],
    ['let Tasking: Int = 1', 'Tasking', 'variable'],
    ['let t: Task = std.task.spawn(worker)', 'Task', 'support.type']
    ,['import api "api"', 'import', 'keyword.control']
    ,['export func handler(): HttpResponse {', 'export', 'storage.type']
    ,['let f: Fn<HttpResponse, HttpRequest> = handler', 'Fn', 'support.type']
    ,['let b: Bytes = value', 'Bytes', 'support.type']
    ,['let import: Int = 1', 'import', 'variable']
  ]) {
    const at = line.indexOf(needle);
    const token = grammar.tokenizeLine(line, tm.INITIAL).tokens.find(t => t.startIndex <= at && t.endIndex > at);
    assert.ok(token.scopes.some(s => s.startsWith(scope)), `${line}: ${needle} => ${token.scopes}`);
  }
});

test('unfinished string and comment do not leak into next line', async () => {
  const grammar = await ready;
  for (const first of ['let s: String = "unfinished', '// "unfinished']) {
    const state = grammar.tokenizeLine(first, tm.INITIAL).ruleStack;
    const tokens = grammar.tokenizeLine('let next: Int = 7', state).tokens;
    assert.ok(tokens[0].scopes.some(s => s.startsWith('storage.type')));
    assert.ok(!tokens.some(t => t.scopes.some(s => s.startsWith('string.') || s.startsWith('comment.'))));
  }
});

test('fixtures tokenize and theme contrast stays readable', async () => {
  const grammar = await ready;
  let state = tm.INITIAL;
  for (const line of fs.readFileSync(path.join(root, 'fixtures/highlighting.craft'), 'utf8').split(/\r?\n/)) {
    state = grammar.tokenizeLine(line, state).ruleStack;
  }
  const luminance = hex => {
    const rgb = hex.slice(1).match(/../g).map(s => parseInt(s, 16) / 255).map(v => v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4);
    return rgb[0] * 0.2126 + rgb[1] * 0.7152 + rgb[2] * 0.0722;
  };
  for (const name of ['dark', 'light']) {
    const theme = JSON.parse(fs.readFileSync(path.join(root, `themes/craft-${name}-color-theme.json`), 'utf8'));
    for (const background of ['editor.background', 'editor.selectionBackground', 'editor.lineHighlightBackground']) {
      const b = luminance(theme.colors[background]);
      for (const rule of theme.tokenColors) {
        const f = luminance(rule.settings.foreground);
        assert.ok((Math.max(f, b) + 0.05) / (Math.min(f, b) + 0.05) >= 4.5, `${name}: ${rule.scope} on ${background}`);
      }
    }
  }
});

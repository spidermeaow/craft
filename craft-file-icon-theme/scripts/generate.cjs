// The neutral manifest owns associations; VS Code only consumes the generated adapter.
const fs = require('node:fs');
const path = require('node:path');
const root = path.resolve(__dirname, '..');
const manifest = JSON.parse(fs.readFileSync(path.join(root, 'theme.json'), 'utf8'));
const pkg = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'));
if (manifest.version !== pkg.version || manifest.displayName !== pkg.displayName || manifest.license !== pkg.license) {
  throw Error('Keep extension version/displayName/license consistent with theme.json');
}
const iconRoot = path.resolve(root, manifest.iconDirectory);
if (!iconRoot.startsWith(root + path.sep)) throw Error('Icon directory must stay inside the extension');
const theme = {iconDefinitions: {}, file: '_file', folder: '_folder', folderExpanded: '_folder',
  rootFolder: '_folder', rootFolderExpanded: '_folder', fileExtensions: {}, fileNames: {}, folderNames: {}, folderNamesExpanded: {}};
const skipped = [];
const icons = new Set(Object.values(manifest.fileAssociations));
for (const id of [...icons].sort()) {
  if (!/^[a-z][a-z0-9-]*$/.test(id)) throw Error(`Invalid icon ID: ${id}`);
  const file = path.join(iconRoot, id + '.svg');
  const svg = fs.readFileSync(file, 'utf8');
  if (!svg.includes('<svg') || !svg.includes('viewBox=') || /<(script|foreignObject)\b|(?:href\s*=\s*["'](?:https?:|\/\/))|\bon\w+\s*=/i.test(svg)) throw Error(`Invalid or externally dependent SVG: ${id}`);
  theme.iconDefinitions[id] = {iconPath: '../' + path.relative(root, file).split(path.sep).join('/')};
}
// Small neutral fallbacks supplement (and never replace) the six prepared assets.
theme.iconDefinitions._file = {iconPath: '../adapter-icons/file.svg'};
theme.iconDefinitions._folder = {iconPath: '../adapter-icons/folder.svg'};
for (const file of ['file.svg', 'folder.svg']) {
  const svg = fs.readFileSync(path.join(root, 'adapter-icons', file), 'utf8');
  if (!svg.includes('<svg') || !svg.includes('viewBox=')) throw Error(`Invalid fallback SVG: ${file}`);
}
for (const [rule, icon] of Object.entries(manifest.fileAssociations)) {
  let match;
  if ((match = /^\*\.([\w.-]+)$/.exec(rule))) theme.fileExtensions[match[1]] = icon;
  else if ((match = /^([\w.-]+)\/\*\.([\w.-]+)$/.exec(rule))) theme.fileExtensions[match[1] + '/' + match[2]] = icon;
  else if ((match = /^([\w.-]+)\/$/.exec(rule))) {
    theme.folderNames[match[1]] = icon; theme.folderNamesExpanded[match[1]] = icon;
  } else if (/^[\w.-]+$/.test(rule)) theme.fileNames[rule] = icon;
  else skipped.push({rule, icon, reason: 'VS Code native file icon associations do not support this glob; use the matching parent-folder rule or the generic extension icon.'});
}
const outputs = {
  'craft-forge-icon-theme.json': theme,
  'association-report.json': {source: '../theme.json', adapter: 'vscode', skipped,
    notes: ['Parent rules match only the immediate parent name, at any depth.', 'Craft Forge is a standalone theme; mappings from the previously selected theme are not merged.']}
};
fs.mkdirSync(path.join(root, 'generated'), {recursive: true});
for (const [name, value] of Object.entries(outputs)) {
  const file = path.join(root, 'generated', name), text = JSON.stringify(value, null, 2) + '\n';
  if (process.argv.includes('--check')) {
    if (!fs.existsSync(file) || fs.readFileSync(file, 'utf8') !== text) throw Error(`Stale generated adapter: ${name}; run npm run generate`);
  } else fs.writeFileSync(file, text);
}
for (const item of skipped) console.log(`Documented fallback: ${item.rule}`);
console.log(process.argv.includes('--check') ? 'Generated adapter is up to date.' : 'Generated VS Code icon theme.');
